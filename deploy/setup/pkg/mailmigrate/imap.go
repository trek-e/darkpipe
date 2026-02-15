package mailmigrate

import (
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// IMAPSync manages IMAP mail migration between source and destination servers
type IMAPSync struct {
	Source           *imapclient.Client
	Dest             *imapclient.Client
	State            *MigrationState
	Mapper           *FolderMapper
	StatePath        string
	LabelsAsKeywords bool
	BatchSize        int
	OnProgress       func(folder string, current, total int)
	OnFolderStart    func(folder string, total int)
	OnFolderDone     func(folder string)
}

// FolderInfo contains metadata about a source folder
type FolderInfo struct {
	Name        string
	Messages    uint32
	MappedTo    string
	Skip        bool
}

// FolderResult contains the result of syncing a single folder
type FolderResult struct {
	Folder   string
	Migrated int
	Skipped  int
	Errors   int
}

// SyncResult contains the overall result of syncing all folders
type SyncResult struct {
	Folders        []FolderResult
	TotalMigrated  int
	TotalSkipped   int
	TotalErrors    int
}

// DryRunFolder contains dry-run preview info for a folder
type DryRunFolder struct {
	Name       string
	Messages   uint32
	MappedTo   string
	WillSkip   bool
}

// DryRunResult contains dry-run preview for all folders
type DryRunResult struct {
	Folders       []DryRunFolder
	TotalMessages uint32
}

// NewIMAPSync creates a new IMAP sync engine
func NewIMAPSync(source, dest *imapclient.Client, state *MigrationState, mapper *FolderMapper, statePath string) *IMAPSync {
	return &IMAPSync{
		Source:           source,
		Dest:             dest,
		State:            state,
		Mapper:           mapper,
		StatePath:        statePath,
		LabelsAsKeywords: false,
		BatchSize:        50,
	}
}

// ListSourceFolders lists all source mailboxes with message counts
func (s *IMAPSync) ListSourceFolders(ctx context.Context) ([]FolderInfo, error) {
	// List all mailboxes
	listCmd := s.Source.List("", "*", nil)
	mailboxes, err := listCmd.Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %w", err)
	}

	var folders []FolderInfo
	for _, mbox := range mailboxes {
		// Get folder mapping
		mappedTo, skip := s.Mapper.Map(mbox.Mailbox)

		info := FolderInfo{
			Name:     mbox.Mailbox,
			Messages: 0, // Will be populated by STATUS command
			MappedTo: mappedTo,
			Skip:     skip,
		}

		// Get message count via STATUS (faster than SELECT)
		if !skip {
			statusCmd := s.Source.Status(mbox.Mailbox, &imap.StatusOptions{
				NumMessages: true,
			})
			statusData, err := statusCmd.Wait()
			if err == nil && statusData != nil && statusData.NumMessages != nil {
				info.Messages = *statusData.NumMessages
			}
		}

		folders = append(folders, info)
	}

	return folders, nil
}

// DryRun shows preview of what would be migrated without transferring
func (s *IMAPSync) DryRun(ctx context.Context) (*DryRunResult, error) {
	folders, err := s.ListSourceFolders(ctx)
	if err != nil {
		return nil, err
	}

	result := &DryRunResult{
		Folders:       make([]DryRunFolder, 0, len(folders)),
		TotalMessages: 0,
	}

	for _, folder := range folders {
		dryFolder := DryRunFolder{
			Name:     folder.Name,
			Messages: folder.Messages,
			MappedTo: folder.MappedTo,
			WillSkip: folder.Skip,
		}
		result.Folders = append(result.Folders, dryFolder)

		if !folder.Skip {
			result.TotalMessages += folder.Messages
		}
	}

	return result, nil
}

// SyncFolder syncs a single folder from source to destination
func (s *IMAPSync) SyncFolder(ctx context.Context, sourceFolder string) (*FolderResult, error) {
	result := &FolderResult{
		Folder: sourceFolder,
	}

	// Map folder name
	destFolder, skip := s.Mapper.Map(sourceFolder)
	if skip {
		return result, nil
	}

	// Select source folder (read-only)
	selectData, err := s.Source.Select(sourceFolder, &imap.SelectOptions{
		ReadOnly: true,
	}).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to select source folder %q: %w", sourceFolder, err)
	}

	totalMessages := selectData.NumMessages
	if totalMessages == 0 {
		return result, nil
	}

	// Notify folder start
	if s.OnFolderStart != nil {
		s.OnFolderStart(sourceFolder, int(totalMessages))
	}

	// Create destination folder if it doesn't exist
	if err := s.ensureDestFolder(destFolder); err != nil {
		return nil, fmt.Errorf("failed to create destination folder %q: %w", destFolder, err)
	}

	// Fetch all messages in batches
	var seqSet imap.SeqSet
	seqSet.AddRange(1, totalMessages)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		BodySection: []*imap.FetchItemBodySection{
			{
				Specifier: imap.PartSpecifierNone, // Fetch entire message
			},
		},
		InternalDate: true,
	}

	fetchCmd := s.Source.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	processed := 0
	for {
		msgData := fetchCmd.Next()
		if msgData == nil {
			break
		}

		// Collect message data into buffer
		msg, err := msgData.Collect()
		if err != nil {
			result.Errors++
			continue
		}

		processed++

		// Extract Message-ID from envelope
		messageID := ""
		if msg.Envelope != nil && msg.Envelope.MessageID != "" {
			messageID = msg.Envelope.MessageID
		} else if msg.Envelope != nil {
			// No Message-ID, use fallback hash (pitfall #5)
			from := ""
			if len(msg.Envelope.From) > 0 {
				from = msg.Envelope.From[0].Addr()
			}
			subject := msg.Envelope.Subject
			date := msg.Envelope.Date
			messageID = "fallback:" + from + ":" + subject + ":" + date.Format(time.RFC3339)
		} else {
			// No envelope at all, skip
			result.Errors++
			continue
		}

		// Check if already migrated
		if s.State.IsMessageMigrated(messageID) {
			result.Skipped++
			if s.OnProgress != nil {
				s.OnProgress(sourceFolder, processed, int(totalMessages))
			}
			continue
		}

		// Get message body from BodySection
		msgBody := msg.FindBodySection(&imap.FetchItemBodySection{
			Specifier: imap.PartSpecifierNone,
		})

		if len(msgBody) == 0 {
			result.Errors++
			continue
		}

		// APPEND to destination with original flags and INTERNALDATE
		appendErr := s.appendMessage(ctx, destFolder, msgBody, msg.Flags, msg.InternalDate)
		if appendErr != nil {
			result.Errors++
			continue
		}

		// Mark as migrated
		if err := s.State.MarkMessageMigrated(messageID); err != nil {
			// Log but continue
		}

		result.Migrated++

		// Progress callback
		if s.OnProgress != nil {
			s.OnProgress(sourceFolder, processed, int(totalMessages))
		}

		// Periodic save (auto-save handles this, but manual save every batch for safety)
		if processed%s.BatchSize == 0 {
			s.State.Save()
		}
	}

	// Final save
	s.State.Save()

	// Notify folder done
	if s.OnFolderDone != nil {
		s.OnFolderDone(sourceFolder)
	}

	return result, nil
}

// SyncAll syncs all non-skipped folders sequentially
func (s *IMAPSync) SyncAll(ctx context.Context) (*SyncResult, error) {
	folders, err := s.ListSourceFolders(ctx)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Folders: make([]FolderResult, 0, len(folders)),
	}

	for _, folder := range folders {
		if folder.Skip {
			continue
		}

		folderResult, err := s.SyncFolder(ctx, folder.Name)
		if err != nil {
			// Log error but continue with other folders
			folderResult = &FolderResult{
				Folder: folder.Name,
				Errors: 1,
			}
		}

		result.Folders = append(result.Folders, *folderResult)
		result.TotalMigrated += folderResult.Migrated
		result.TotalSkipped += folderResult.Skipped
		result.TotalErrors += folderResult.Errors
	}

	return result, nil
}

// ensureDestFolder creates destination folder if it doesn't exist
func (s *IMAPSync) ensureDestFolder(folder string) error {
	// Try to create folder (will fail if exists, which is OK)
	createCmd := s.Dest.Create(folder, nil)
	createCmd.Wait() // Ignore error - folder might already exist

	return nil
}

// appendMessage uploads a message to destination with preserved date and flags
func (s *IMAPSync) appendMessage(ctx context.Context, folder string, msg []byte, flags []imap.Flag, internalDate time.Time) error {
	options := &imap.AppendOptions{
		Flags: flags,
		Time:  internalDate, // Preserve original date (locked decision)
	}

	appendCmd := s.Dest.Append(folder, int64(len(msg)), options)

	if _, err := appendCmd.Write(msg); err != nil {
		return err
	}

	if err := appendCmd.Close(); err != nil {
		return err
	}

	// Wait for completion
	_, err := appendCmd.Wait()
	return err
}
