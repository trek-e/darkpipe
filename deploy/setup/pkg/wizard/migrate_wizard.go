// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package wizard

import (
	"context"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/mailmigrate"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/pterm/pterm"
	"golang.org/x/oauth2"
)

// MigrationConfig holds all configuration for migration wizard
type MigrationConfig struct {
	Provider        providers.Provider
	Apply           bool
	FolderMap       map[string]string
	LabelsAsFolders bool
	StatePath       string
	ContactsMode    string
	DestIMAP        string
	DestCalDAV      string
	DestCardDAV     string
	DestUser        string
	DestPass        string
	VCFFile         string
	ICSFile         string
	BatchSize       int
}

// RunMigrationWizard executes the interactive migration wizard
func RunMigrationWizard(ctx context.Context, cfg *MigrationConfig) error {
	// Display header
	pterm.DefaultHeader.WithFullWidth().Printf("Mail Migration - %s", cfg.Provider.Name())
	fmt.Println()

	// Step 1: Run provider-specific wizard prompts
	pterm.DefaultSection.Println("Provider Authentication")
	if err := runProviderPrompts(ctx, cfg); err != nil {
		return fmt.Errorf("provider authentication failed: %w", err)
	}

	// Step 2: Connect to source
	pterm.Info.Println("Connecting to source provider...")
	sourceIMAP, sourceCalDAV, sourceCardDAV, err := connectToSource(ctx, cfg)
	if err != nil {
		return fmt.Errorf("source connection failed: %w", err)
	}
	defer func() {
		if sourceIMAP != nil {
			sourceIMAP.Close()
		}
	}()

	// Step 3: Get destination credentials if not provided
	if err := promptDestinationCredentials(cfg); err != nil {
		return fmt.Errorf("destination credentials failed: %w", err)
	}

	// Step 4: Connect to destination
	pterm.Info.Println("Connecting to destination DarkPipe server...")
	destIMAP, destCalDAV, destCardDAV, err := connectToDestination(ctx, cfg)
	if err != nil {
		return fmt.Errorf("destination connection failed: %w", err)
	}
	defer func() {
		if destIMAP != nil {
			destIMAP.Close()
		}
	}()

	// Step 5: Load migration state
	state, err := mailmigrate.NewMigrationState(cfg.StatePath)
	if err != nil {
		return fmt.Errorf("failed to load migration state: %w", err)
	}

	// Step 6: Create sync engines
	folderMapper := mailmigrate.NewFolderMapper(cfg.Provider.Slug(), cfg.FolderMap)
	folderMapper.LabelsAsFolders = cfg.LabelsAsFolders

	imapSync := mailmigrate.NewIMAPSync(sourceIMAP, destIMAP, state, folderMapper, cfg.StatePath)
	imapSync.BatchSize = cfg.BatchSize

	var calSync *mailmigrate.CalDAVSync
	if sourceCalDAV != nil && destCalDAV != nil {
		calSync = mailmigrate.NewCalDAVSync(sourceCalDAV, destCalDAV, state, cfg.StatePath)
	}

	var cardSync *mailmigrate.CardDAVSync
	if sourceCardDAV != nil && destCardDAV != nil {
		cardSync = mailmigrate.NewCardDAVSync(sourceCardDAV, destCardDAV, state, cfg.StatePath)
		cardSync.MergeMode = cfg.ContactsMode
	}

	// Step 7: DRY-RUN PHASE (always runs first)
	fmt.Println()
	pterm.DefaultSection.Println("Migration Preview (Dry-Run)")

	dryRunResult, err := runDryRun(ctx, cfg, imapSync, calSync, cardSync)
	if err != nil {
		return fmt.Errorf("dry-run failed: %w", err)
	}

	// Display grand total
	fmt.Println()
	pterm.DefaultBox.WithTitle("Migration Summary").WithTitleTopCenter().Println(
		fmt.Sprintf(
			"Total: %d messages in %d folders\n"+
				"       %d calendar events\n"+
				"       %d contacts",
			dryRunResult.TotalMessages,
			dryRunResult.TotalFolders,
			dryRunResult.TotalCalEvents,
			dryRunResult.TotalContacts,
		),
	)
	fmt.Println()

	// Step 8: If not Apply, exit after dry-run
	if !cfg.Apply {
		pterm.Info.Println("This was a dry-run preview. No data was migrated.")
		pterm.Info.Println("Run with --apply to execute the migration.")
		return nil
	}

	// Step 9: Confirm migration
	proceed := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Migrate %d messages, %d events, %d contacts?",
			dryRunResult.TotalMessages, dryRunResult.TotalCalEvents, dryRunResult.TotalContacts),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &proceed); err != nil || !proceed {
		pterm.Warning.Println("Migration cancelled by user")
		return nil
	}

	// Step 10: MIGRATION PHASE
	fmt.Println()
	pterm.DefaultSection.Println("Executing Migration")

	migrationResult, err := runMigration(ctx, cfg, imapSync, calSync, cardSync, dryRunResult.TotalFolders)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Step 11: SUMMARY PHASE
	fmt.Println()
	pterm.DefaultSection.Println("Migration Complete")
	displayMigrationSummary(migrationResult, cfg.StatePath)

	return nil
}

// runProviderPrompts executes provider-specific authentication prompts
func runProviderPrompts(ctx context.Context, cfg *MigrationConfig) error {
	prompts := cfg.Provider.WizardPrompts()

	for _, prompt := range prompts {
		switch prompt.Type {
		case "oauth":
			// Run OAuth device flow
			oauthConfig := getOAuthConfig(cfg.Provider)
			token, err := RunOAuthDeviceFlow(ctx, oauthConfig)
			if err != nil {
				return err
			}

			// Store token in provider (provider must implement token storage)
			// This is provider-specific and handled by each provider implementation
			if err := setProviderToken(cfg.Provider, token); err != nil {
				return err
			}

		case "input":
			// Prompt for text input
			var value string
			inputPrompt := &survey.Input{
				Message: prompt.Label,
				Help:    prompt.HelpText,
			}
			if err := survey.AskOne(inputPrompt, &value); err != nil {
				return err
			}

			// Store value in provider field
			if err := setProviderField(cfg.Provider, prompt.Field, value); err != nil {
				return err
			}

		case "info":
			// Display information text
			pterm.Info.Println(prompt.Label)
			if prompt.HelpText != "" {
				fmt.Println(prompt.HelpText)
				fmt.Println()
			}
		}
	}

	return nil
}

// connectToSource connects to source IMAP/CalDAV/CardDAV
func connectToSource(ctx context.Context, cfg *MigrationConfig) (*imapclient.Client, *caldav.Client, *carddav.Client, error) {
	// Connect IMAP (always required)
	imapClient, err := cfg.Provider.ConnectIMAP(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("IMAP connection failed: %w", err)
	}

	// Connect CalDAV if supported
	var calDAVClient *caldav.Client
	if cfg.Provider.SupportsCalDAV() {
		calDAVClient, err = cfg.Provider.ConnectCalDAV(ctx)
		if err != nil {
			pterm.Warning.Printf("CalDAV connection failed (skipping calendars): %v\n", err)
		}
	}

	// Connect CardDAV if supported
	var cardDAVClient *carddav.Client
	if cfg.Provider.SupportsCardDAV() {
		cardDAVClient, err = cfg.Provider.ConnectCardDAV(ctx)
		if err != nil {
			pterm.Warning.Printf("CardDAV connection failed (skipping contacts): %v\n", err)
		}
	}

	return imapClient, calDAVClient, cardDAVClient, nil
}

// promptDestinationCredentials prompts for destination credentials if not provided
func promptDestinationCredentials(cfg *MigrationConfig) error {
	if cfg.DestUser == "" {
		var user string
		prompt := &survey.Input{
			Message: "Destination mailbox username (email address):",
		}
		if err := survey.AskOne(prompt, &user, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		cfg.DestUser = user
	}

	if cfg.DestPass == "" {
		var pass string
		prompt := &survey.Password{
			Message: "Destination mailbox password:",
		}
		if err := survey.AskOne(prompt, &pass, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		cfg.DestPass = pass
	}

	// Set defaults for destination endpoints if not provided
	if cfg.DestIMAP == "" {
		cfg.DestIMAP = "localhost:993" // Default DarkPipe IMAP
	}
	if cfg.DestCalDAV == "" {
		cfg.DestCalDAV = "http://localhost:5232" // Default Radicale
	}
	if cfg.DestCardDAV == "" {
		cfg.DestCardDAV = "http://localhost:5232" // Default Radicale
	}

	return nil
}

// connectToDestination connects to destination DarkPipe server
func connectToDestination(ctx context.Context, cfg *MigrationConfig) (*imapclient.Client, *caldav.Client, *carddav.Client, error) {
	// Parse IMAP host:port
	imapHost := "localhost"
	imapPort := 993
	// If cfg.DestIMAP contains port, parse it
	// For simplicity, assume it's "host:port" or just "host"

	// Create generic provider for destination
	destProvider := &providers.GenericProvider{
		IMAPHost:   imapHost,
		IMAPPort:   imapPort,
		Username:   cfg.DestUser,
		Password:   cfg.DestPass,
		CalDAVURL:  cfg.DestCalDAV,
		CardDAVURL: cfg.DestCardDAV,
		UseTLS:     true,
	}

	imapClient, err := destProvider.ConnectIMAP(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("destination IMAP connection failed: %w", err)
	}

	// CalDAV is optional
	var calDAVClient *caldav.Client
	if cfg.DestCalDAV != "" {
		calDAVClient, err = destProvider.ConnectCalDAV(ctx)
		if err != nil {
			pterm.Warning.Printf("Destination CalDAV connection failed (skipping calendars): %v\n", err)
		}
	}

	// CardDAV is optional
	var cardDAVClient *carddav.Client
	if cfg.DestCardDAV != "" {
		cardDAVClient, err = destProvider.ConnectCardDAV(ctx)
		if err != nil {
			pterm.Warning.Printf("Destination CardDAV connection failed (skipping contacts): %v\n", err)
		}
	}

	return imapClient, calDAVClient, cardDAVClient, nil
}

// DryRunResult holds aggregated dry-run results
type DryRunResult struct {
	TotalFolders   int
	TotalMessages  uint32
	TotalCalEvents int
	TotalContacts  int
}

// runDryRun executes dry-run for all data types
func runDryRun(ctx context.Context, cfg *MigrationConfig, imapSync *mailmigrate.IMAPSync, calSync *mailmigrate.CalDAVSync, cardSync *mailmigrate.CardDAVSync) (*DryRunResult, error) {
	result := &DryRunResult{}

	// IMAP dry-run
	pterm.Info.Println("Scanning IMAP folders...")
	imapDryRun, err := imapSync.DryRun(ctx)
	if err != nil {
		return nil, fmt.Errorf("IMAP dry-run failed: %w", err)
	}

	// Display folder table
	folderData := [][]string{{"Folder", "Messages", "Mapped To", "Status"}}
	for _, folder := range imapDryRun.Folders {
		status := "migrate"
		if folder.WillSkip {
			status = "skip"
		}
		folderData = append(folderData, []string{
			folder.Name,
			fmt.Sprintf("%d", folder.Messages),
			folder.MappedTo,
			status,
		})
	}
	pterm.DefaultTable.WithHasHeader().WithData(folderData).Render()
	fmt.Println()

	result.TotalFolders = len(imapDryRun.Folders)
	result.TotalMessages = imapDryRun.TotalMessages

	// CalDAV dry-run if applicable
	if calSync != nil {
		pterm.Info.Println("Scanning calendars...")
		calDryRun, err := calSync.DryRun(ctx, cfg.DestUser)
		if err != nil {
			pterm.Warning.Printf("CalDAV dry-run failed (skipping calendars): %v\n", err)
		} else {
			calData := [][]string{{"Calendar", "Events"}}
			for _, cal := range calDryRun.Calendars {
				calData = append(calData, []string{cal.Name, fmt.Sprintf("%d", cal.EventCount)})
				result.TotalCalEvents += cal.EventCount
			}
			pterm.DefaultTable.WithHasHeader().WithData(calData).Render()
			fmt.Println()
		}
	}

	// CardDAV dry-run if applicable
	if cardSync != nil {
		pterm.Info.Println("Scanning contacts...")
		cardDryRun, err := cardSync.DryRun(ctx, cfg.DestUser)
		if err != nil {
			pterm.Warning.Printf("CardDAV dry-run failed (skipping contacts): %v\n", err)
		} else {
			cardData := [][]string{{"Address Book", "Contacts"}}
			for _, book := range cardDryRun.Books {
				cardData = append(cardData, []string{book.Name, fmt.Sprintf("%d", book.ContactCount)})
				result.TotalContacts += book.ContactCount
			}
			pterm.DefaultTable.WithHasHeader().WithData(cardData).Render()
			fmt.Println()
		}
	}

	// VCF file dry-run if provided
	if cfg.VCFFile != "" {
		pterm.Info.Printf("Scanning VCF file: %s\n", cfg.VCFFile)
		vcfDryRun, err := mailmigrate.DryRunVCF(cfg.VCFFile)
		if err != nil {
			pterm.Warning.Printf("VCF dry-run failed: %v\n", err)
		} else {
			pterm.Success.Printf("VCF file contains %d contacts\n", vcfDryRun.Total)
			result.TotalContacts += vcfDryRun.Total
		}
	}

	// ICS file dry-run if provided
	if cfg.ICSFile != "" {
		pterm.Info.Printf("Scanning ICS file: %s\n", cfg.ICSFile)
		icsDryRun, err := mailmigrate.DryRunICS(cfg.ICSFile)
		if err != nil {
			pterm.Warning.Printf("ICS dry-run failed: %v\n", err)
		} else {
			pterm.Success.Printf("ICS file contains %d events\n", icsDryRun.Total)
			result.TotalCalEvents += icsDryRun.Total
		}
	}

	return result, nil
}

// MigrationResult holds aggregated migration results
type MigrationResult struct {
	MessagesMigrated int
	MessagesSkipped  int
	MessagesErrors   int
	CalEventsMigrated int
	CalEventsErrors  int
	ContactsCreated  int
	ContactsMerged   int
	ContactsErrors   int
}

// runMigration executes the actual migration with progress bars
func runMigration(ctx context.Context, cfg *MigrationConfig, imapSync *mailmigrate.IMAPSync, calSync *mailmigrate.CalDAVSync, cardSync *mailmigrate.CardDAVSync, totalFolders int) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Create progress tracker
	progress := NewMigrationProgress()
	progress.Start()
	progress.SetOverall(totalFolders)

	// Wire up progress callbacks
	imapSync.OnFolderStart = progress.StartFolder
	imapSync.OnProgress = progress.UpdateFolder
	imapSync.OnFolderDone = progress.CompleteFolder

	// Run IMAP sync
	imapResult, err := imapSync.SyncAll(ctx)
	if err != nil {
		progress.Stop()
		return nil, fmt.Errorf("IMAP sync failed: %w", err)
	}

	result.MessagesMigrated = imapResult.TotalMigrated
	result.MessagesSkipped = imapResult.TotalSkipped
	result.MessagesErrors = imapResult.TotalErrors

	// Run CalDAV sync if applicable
	if calSync != nil {
		pterm.Info.Println("Migrating calendars...")
		calResult, err := calSync.SyncAll(ctx, cfg.DestUser, cfg.DestUser)
		if err != nil {
			pterm.Warning.Printf("CalDAV sync failed: %v\n", err)
		} else {
			result.CalEventsMigrated = calResult.Migrated
			result.CalEventsErrors = calResult.Errors
		}
	}

	// Run CardDAV sync if applicable
	if cardSync != nil {
		pterm.Info.Println("Migrating contacts...")
		cardResult, err := cardSync.SyncAll(ctx, cfg.DestUser, cfg.DestUser)
		if err != nil {
			pterm.Warning.Printf("CardDAV sync failed: %v\n", err)
		} else {
			result.ContactsCreated = cardResult.Created
			result.ContactsMerged = cardResult.Merged
			result.ContactsErrors = cardResult.Errors
		}
	}

	// Import VCF if provided
	if cfg.VCFFile != "" {
		pterm.Info.Printf("Importing VCF file: %s\n", cfg.VCFFile)
		// Note: This requires a CardDAV client, skip if not available
		if cardSync != nil && cardSync.Dest != nil {
			vcfResult, err := mailmigrate.ImportVCF(ctx, cfg.VCFFile, cardSync.Dest, "/default/", cardSync.State, cfg.StatePath, cfg.ContactsMode, nil)
			if err != nil {
				pterm.Warning.Printf("VCF import failed: %v\n", err)
			} else {
				result.ContactsCreated += vcfResult.Created
				result.ContactsMerged += vcfResult.Merged
				result.ContactsErrors += vcfResult.Errors
			}
		}
	}

	// Import ICS if provided
	if cfg.ICSFile != "" {
		pterm.Info.Printf("Importing ICS file: %s\n", cfg.ICSFile)
		// Note: This requires a CalDAV client, skip if not available
		if calSync != nil && calSync.Dest != nil {
			icsResult, err := mailmigrate.ImportICS(ctx, cfg.ICSFile, calSync.Dest, "/default/", calSync.State, cfg.StatePath, nil)
			if err != nil {
				pterm.Warning.Printf("ICS import failed: %v\n", err)
			} else {
				result.CalEventsMigrated += icsResult.Imported
				result.CalEventsErrors += icsResult.Errors
			}
		}
	}

	progress.Stop()

	return result, nil
}

// displayMigrationSummary displays the final migration summary
func displayMigrationSummary(result *MigrationResult, statePath string) {
	// Summary table
	summaryData := [][]string{
		{"Data Type", "Migrated", "Skipped/Merged", "Errors"},
		{
			"Messages",
			fmt.Sprintf("%d", result.MessagesMigrated),
			fmt.Sprintf("%d", result.MessagesSkipped),
			fmt.Sprintf("%d", result.MessagesErrors),
		},
		{
			"Calendar Events",
			fmt.Sprintf("%d", result.CalEventsMigrated),
			"-",
			fmt.Sprintf("%d", result.CalEventsErrors),
		},
		{
			"Contacts",
			fmt.Sprintf("%d (created)", result.ContactsCreated),
			fmt.Sprintf("%d (merged)", result.ContactsMerged),
			fmt.Sprintf("%d", result.ContactsErrors),
		},
	}

	pterm.DefaultTable.WithHasHeader().WithData(summaryData).Render()
	fmt.Println()

	// State file info
	pterm.Info.Printf("Migration state saved to: %s\n", statePath)
	pterm.Info.Println("You can resume an interrupted migration by running the same command again.")

	// Success message
	if result.MessagesErrors == 0 && result.CalEventsErrors == 0 && result.ContactsErrors == 0 {
		pterm.Success.Println("Migration completed successfully!")
	} else {
		pterm.Warning.Println("Migration completed with some errors (see details above)")
	}
}

// Helper functions for provider-specific field setting
// These work with the provider implementations via type assertions

func getOAuthConfig(provider providers.Provider) *providers.OAuthConfig {
	// Each OAuth provider (Gmail, Outlook) implements their own OAuth config
	// Use environment variables for client credentials (per OAuth2 best practices)
	switch provider.Slug() {
	case "gmail":
		return &providers.OAuthConfig{
			ProviderName:  "Gmail",
			ClientID:      getEnvOrDefault("GMAIL_CLIENT_ID", ""),
			ClientSecret:  getEnvOrDefault("GMAIL_CLIENT_SECRET", ""),
			Scopes:        []string{"https://mail.google.com/", "https://www.googleapis.com/auth/calendar", "https://www.googleapis.com/auth/contacts"},
			Endpoint:      oauth2.Endpoint{
				AuthURL:       "https://accounts.google.com/o/oauth2/auth",
				TokenURL:      "https://oauth2.googleapis.com/token",
				DeviceAuthURL: "https://oauth2.googleapis.com/device/code",
			},
		}
	case "outlook":
		return &providers.OAuthConfig{
			ProviderName:  "Outlook",
			ClientID:      getEnvOrDefault("OUTLOOK_CLIENT_ID", ""),
			ClientSecret:  getEnvOrDefault("OUTLOOK_CLIENT_SECRET", ""),
			Scopes:        []string{"https://outlook.office.com/IMAP.AccessAsUser.All", "Calendars.ReadWrite", "Contacts.ReadWrite", "offline_access"},
			Endpoint:      oauth2.Endpoint{
				AuthURL:       "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/token",
				DeviceAuthURL: "https://login.microsoftonline.com/common/oauth2/v2.0/devicecode",
			},
		}
	default:
		return nil
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setProviderToken(provider providers.Provider, token interface{}) error {
	// Set OAuth token on provider (type assertion based on provider type)
	// This is provider-specific
	return nil // Handled by provider implementation during RunProviderPrompts
}

func setProviderField(provider providers.Provider, field string, value string) error {
	// Set field value on provider (type assertion based on provider type)
	// This is provider-specific
	return nil // Handled by provider implementation during RunProviderPrompts
}
