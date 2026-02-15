// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/wizard"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate email, contacts, and calendars from an existing provider",
	Long: `Migrate email, contacts, and calendars from an existing provider.

Supported providers:
  - gmail: Google Gmail/Workspace (IMAP, CalDAV, CardDAV via OAuth2)
  - outlook: Microsoft Outlook/365 (IMAP, CalDAV, CardDAV via OAuth2)
  - icloud: Apple iCloud (IMAP, CalDAV, CardDAV with app-specific password)
  - mailcow: MailCow mail server (IMAP + API for user export)
  - mailu: Mailu mail server (IMAP + API for user export)
  - docker-mailserver: Docker Mailserver (IMAP only)
  - generic: Any IMAP/CalDAV/CardDAV server

By default, the command performs a dry-run to preview what would be migrated.
Use --apply to execute the actual migration.

Examples:
  # Preview Gmail migration
  darkpipe-setup migrate --from gmail

  # Execute Gmail migration with custom folder mappings
  darkpipe-setup migrate --from gmail --apply --folder-map "[Gmail]/All Mail:Archive"

  # Migrate from generic IMAP server
  darkpipe-setup migrate --from generic --apply

  # Import contacts from VCF file
  darkpipe-setup migrate --from generic --vcf-file contacts.vcf --apply
`,
	RunE: runMigrate,
}

var (
	migrateFrom         string
	migrateApply        bool
	migrateFolderMap    string
	migrateLabelsAsFolders bool
	migrateStateFile    string
	migrateContactsMode string
	migrateIMAPDest     string
	migrateCalDAVDest   string
	migrateCardDAVDest  string
	migrateVCFFile      string
	migrateICSFile      string
	migrateBatchSize    int
	migrateDestUser     string
	migrateDestPass     string
)

func init() {
	migrateCmd.Flags().StringVar(&migrateFrom, "from", "", "Provider slug (gmail, outlook, icloud, mailcow, mailu, docker-mailserver, generic)")
	migrateCmd.Flags().BoolVar(&migrateApply, "apply", false, "Execute migration (without this, dry-run only)")
	migrateCmd.Flags().StringVar(&migrateFolderMap, "folder-map", "", "Custom folder mappings: \"Source:Dest,Source2:Dest2\"")
	migrateCmd.Flags().BoolVar(&migrateLabelsAsFolders, "labels-as-folders", false, "Map Gmail labels to folders instead of keywords (Dovecot/Maddy fallback)")
	migrateCmd.Flags().StringVar(&migrateStateFile, "state-file", "/data/migration-state.json", "Path to state file")
	migrateCmd.Flags().StringVar(&migrateContactsMode, "contacts-mode", "append", "Contact merge mode (append/overwrite/skip)")
	migrateCmd.Flags().StringVar(&migrateIMAPDest, "imap-dest", "", "Destination IMAP server (default: from DarkPipe config)")
	migrateCmd.Flags().StringVar(&migrateCalDAVDest, "caldav-dest", "", "Destination CalDAV server (default: from DarkPipe config)")
	migrateCmd.Flags().StringVar(&migrateCardDAVDest, "carddav-dest", "", "Destination CardDAV server (default: from DarkPipe config)")
	migrateCmd.Flags().StringVar(&migrateVCFFile, "vcf-file", "", "Path to .vcf file for contact import")
	migrateCmd.Flags().StringVar(&migrateICSFile, "ics-file", "", "Path to .ics file for calendar import")
	migrateCmd.Flags().IntVar(&migrateBatchSize, "batch-size", 50, "IMAP fetch batch size")
	migrateCmd.Flags().StringVar(&migrateDestUser, "dest-user", "", "Destination mailbox username")
	migrateCmd.Flags().StringVar(&migrateDestPass, "dest-pass", "", "Destination mailbox password")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// If --from not provided, list supported providers
	if migrateFrom == "" {
		pterm.DefaultHeader.Println("Supported Providers")
		fmt.Println()

		supportedProviders := [][]string{
			{"Provider", "Description", "Features"},
			{"gmail", "Google Gmail/Workspace", "IMAP, CalDAV, CardDAV (OAuth2)"},
			{"outlook", "Microsoft Outlook/365", "IMAP, CalDAV, CardDAV (OAuth2)"},
			{"icloud", "Apple iCloud", "IMAP, CalDAV, CardDAV (app password)"},
			{"mailcow", "MailCow mail server", "IMAP + API (user export)"},
			{"mailu", "Mailu mail server", "IMAP + API (user export)"},
			{"docker-mailserver", "Docker Mailserver", "IMAP only"},
			{"generic", "Any IMAP server", "IMAP, CalDAV, CardDAV (flexible)"},
		}

		pterm.DefaultTable.WithHasHeader().WithData(supportedProviders).Render()
		fmt.Println()
		pterm.Info.Println("Use: darkpipe-setup migrate --from <provider>")
		return nil
	}

	// Get provider from registry
	provider, err := providers.DefaultRegistry.Get(migrateFrom)
	if err != nil {
		return fmt.Errorf("provider %q not found. Run 'darkpipe-setup migrate' to see supported providers", migrateFrom)
	}

	// Parse folder mappings
	folderMappings := make(map[string]string)
	if migrateFolderMap != "" {
		for _, mapping := range strings.Split(migrateFolderMap, ",") {
			parts := strings.SplitN(mapping, ":", 2)
			if len(parts) == 2 {
				folderMappings[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Create migration config
	cfg := &wizard.MigrationConfig{
		Provider:         provider,
		Apply:            migrateApply,
		FolderMap:        folderMappings,
		LabelsAsFolders:  migrateLabelsAsFolders,
		StatePath:        migrateStateFile,
		ContactsMode:     migrateContactsMode,
		DestIMAP:         migrateIMAPDest,
		DestCalDAV:       migrateCalDAVDest,
		DestCardDAV:      migrateCardDAVDest,
		DestUser:         migrateDestUser,
		DestPass:         migrateDestPass,
		VCFFile:          migrateVCFFile,
		ICSFile:          migrateICSFile,
		BatchSize:        migrateBatchSize,
	}

	// Run migration wizard
	ctx := context.Background()
	if err := wizard.RunMigrationWizard(ctx, cfg); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
