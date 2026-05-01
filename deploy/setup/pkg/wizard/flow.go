// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package wizard

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/imapsync"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/mailmigrate"
	"github.com/pterm/pterm"
)

// Flow orchestrates migration wizard phases behind one seam.
type Flow struct{}

func NewFlow() *Flow { return &Flow{} }

func (f *Flow) Run(ctx context.Context, cfg *MigrationConfig) error {
	pterm.DefaultHeader.WithFullWidth().Printf("Mail Migration - %s", cfg.Provider.Name())
	fmt.Println()

	pterm.DefaultSection.Println("Provider Authentication")
	if err := runProviderPrompts(ctx, cfg); err != nil { return fmt.Errorf("provider authentication failed: %w", err) }
	if err := showSourceMetadata(ctx, cfg); err != nil { pterm.Warning.Printf("Could not load source metadata preview: %v\n", err) }

	pterm.Info.Println("Connecting to source provider...")
	sourceIMAP, sourceCalDAV, sourceCardDAV, err := connectToSource(ctx, cfg)
	if err != nil { return fmt.Errorf("source connection failed: %w", err) }
	defer func() { if sourceIMAP != nil { sourceIMAP.Close() } }()

	if err := promptDestinationCredentials(cfg); err != nil { return fmt.Errorf("destination credentials failed: %w", err) }

	pterm.Info.Println("Connecting to destination DarkPipe server...")
	destIMAP, destCalDAV, destCardDAV, err := connectToDestination(ctx, cfg)
	if err != nil { return fmt.Errorf("destination connection failed: %w", err) }
	defer func() { if destIMAP != nil { destIMAP.Close() } }()

	state, err := mailmigrate.NewMigrationState(cfg.StatePath)
	if err != nil { return fmt.Errorf("failed to load migration state: %w", err) }

	folderMapper := mailmigrate.NewFolderMapper(cfg.Provider.Slug(), cfg.FolderMap)
	folderMapper.LabelsAsFolders = cfg.LabelsAsFolders
	imapSync := imapsync.New(sourceIMAP, destIMAP, state, folderMapper, cfg.StatePath)
	imapSync.SetBatchSize(cfg.BatchSize)

	var calSync *mailmigrate.CalDAVSync
	if sourceCalDAV != nil && destCalDAV != nil {
		calSync = mailmigrate.NewCalDAVSync(sourceCalDAV, destCalDAV, state, cfg.StatePath)
	}
	var cardSync *mailmigrate.CardDAVSync
	if sourceCardDAV != nil && destCardDAV != nil {
		cardSync = mailmigrate.NewCardDAVSync(sourceCardDAV, destCardDAV, state, cfg.StatePath)
		cardSync.MergeMode = cfg.ContactsMode
	}

	fmt.Println()
	pterm.DefaultSection.Println("Migration Preview (Dry-Run)")
	dryRunResult, err := runDryRun(ctx, cfg, imapSync, calSync, cardSync)
	if err != nil { return fmt.Errorf("dry-run failed: %w", err) }

	fmt.Println()
	pterm.DefaultBox.WithTitle("Migration Summary").WithTitleTopCenter().Println(
		fmt.Sprintf("Total: %d messages in %d folders\n       %d calendar events\n       %d contacts", dryRunResult.TotalMessages, dryRunResult.TotalFolders, dryRunResult.TotalCalEvents, dryRunResult.TotalContacts),
	)
	fmt.Println()

	if !cfg.Apply {
		pterm.Info.Println("This was a dry-run preview. No data was migrated.")
		pterm.Info.Println("Run with --apply to execute the migration.")
		return nil
	}

	return runApplyPhase(ctx, cfg, imapSync, calSync, cardSync, dryRunResult)
}

func runApplyPhase(ctx context.Context, cfg *MigrationConfig, imapSync imapsync.Module, calSync *mailmigrate.CalDAVSync, cardSync *mailmigrate.CardDAVSync, dryRunResult *DryRunResult) error {
	proceed := false
	confirmPrompt := &survey.Confirm{Message: fmt.Sprintf("Migrate %d messages, %d events, %d contacts?", dryRunResult.TotalMessages, dryRunResult.TotalCalEvents, dryRunResult.TotalContacts), Default: false}
	if err := survey.AskOne(confirmPrompt, &proceed); err != nil || !proceed {
		pterm.Warning.Println("Migration cancelled by user")
		return nil
	}
	fmt.Println()
	pterm.DefaultSection.Println("Executing Migration")
	migrationResult, err := runMigration(ctx, cfg, imapSync, calSync, cardSync, dryRunResult.TotalFolders)
	if err != nil { return fmt.Errorf("migration failed: %w", err) }
	fmt.Println()
	pterm.DefaultSection.Println("Migration Complete")
	displayMigrationSummary(migrationResult, cfg.StatePath)
	return nil
}
