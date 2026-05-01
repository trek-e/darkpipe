// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package wizard

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/imapsync"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/mailmigrate"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/migrationdest"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/migrationsource"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/pterm/pterm"
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
	return NewFlow().Run(ctx, cfg)
}

// runProviderPrompts executes provider-specific authentication prompts
func runProviderPrompts(ctx context.Context, cfg *MigrationConfig) error {
	sourceModule := migrationsource.New()
	return sourceModule.Authenticate(ctx, cfg.Provider)
}

func showSourceMetadata(ctx context.Context, cfg *MigrationConfig) error {
	sourceModule := migrationsource.New()
	meta, err := sourceModule.DiscoverMetadata(ctx, cfg.Provider)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Source Metadata")
	fmt.Printf("Provider: %s (%s)\n", meta.ProviderName, meta.ProviderSlug)

	endpointRows := [][]string{{"Endpoint", "Value"}}
	for _, k := range []string{"imap", "caldav", "carddav"} {
		v := meta.Endpoints[k]
		if v == "" {
			v = "-"
		}
		endpointRows = append(endpointRows, []string{k, v})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(endpointRows).Render()

	capRows := [][]string{{"Capability", "Supported"}}
	for _, c := range []migrationsource.Capability{
		migrationsource.CapabilityLabels,
		migrationsource.CapabilityCalDAV,
		migrationsource.CapabilityCardDAV,
		migrationsource.CapabilityAPI,
	} {
		capRows = append(capRows, []string{string(c), fmt.Sprintf("%t", meta.Capabilities[c])})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(capRows).Render()

	if len(meta.Counts) > 0 {
		countRows := [][]string{{"Metric", "Count"}}
		for k, v := range meta.Counts {
			countRows = append(countRows, []string{k, fmt.Sprintf("%d", v)})
		}
		_ = pterm.DefaultTable.WithHasHeader().WithData(countRows).Render()
	}
	fmt.Println()
	return nil
}

// connectToSource connects to source IMAP/CalDAV/CardDAV
func connectToSource(ctx context.Context, cfg *MigrationConfig) (*imapclient.Client, *caldav.Client, *carddav.Client, error) {
	sourceModule := migrationsource.New()
	adapters, err := sourceModule.OpenAdapters(ctx, cfg.Provider)
	if err != nil {
		return nil, nil, nil, err
	}
	return adapters.IMAP, adapters.CalDAV, adapters.CardDAV, nil
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
	dm := migrationdest.New()
	adapters, err := dm.Connect(ctx, migrationdest.Config{
		DestIMAP: cfg.DestIMAP,
		DestCalDAV: cfg.DestCalDAV,
		DestCardDAV: cfg.DestCardDAV,
		DestUser: cfg.DestUser,
		DestPass: cfg.DestPass,
		TLSPolicy: migrationdest.RequireTLS,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("destination connection failed: %w", err)
	}
	for _, w := range adapters.Warnings {
		pterm.Warning.Printf("Destination adapter warning: %v\n", w)
	}
	return adapters.IMAP, adapters.CalDAV, adapters.CardDAV, nil
}

// DryRunResult holds aggregated dry-run results
type DryRunResult struct {
	TotalFolders   int
	TotalMessages  uint32
	TotalCalEvents int
	TotalContacts  int
}

// runDryRun executes dry-run for all data types
func runDryRun(ctx context.Context, cfg *MigrationConfig, imapSync imapsync.Module, calSync *mailmigrate.CalDAVSync, cardSync *mailmigrate.CardDAVSync) (*DryRunResult, error) {
	result := &DryRunResult{}

	// IMAP dry-run
	pterm.Info.Println("Scanning IMAP folders...")
	imapDryRun, err := imapSync.Preview(ctx)
	if err != nil {
		return nil, fmt.Errorf("IMAP dry-run failed: %w", err)
	}

	// Display folder table
	folderData := [][]string{{"Folder", "Messages", "Mapped To", "Status"}}
	for _, folder := range imapDryRun.Folders {
		status := "migrate"
		if folder.Skipped {
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
func runMigration(ctx context.Context, cfg *MigrationConfig, imapSync imapsync.Module, calSync *mailmigrate.CalDAVSync, cardSync *mailmigrate.CardDAVSync, totalFolders int) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Create progress tracker
	progress := NewMigrationProgress()
	progress.Start()
	progress.SetOverall(totalFolders)

	// Wire up progress callbacks
	imapSync.SetProgressCallbacks(progress.UpdateFolder, progress.StartFolder, progress.CompleteFolder)

	// Run IMAP sync
	imapResult, err := imapSync.Execute(ctx)
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

