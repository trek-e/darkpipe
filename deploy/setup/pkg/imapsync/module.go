// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package imapsync

import (
	"context"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/mailmigrate"
	"github.com/emersion/go-imap/v2/imapclient"
)

type FolderPreview struct {
	Name     string
	Messages uint32
	MappedTo string
	Skipped  bool
}

type Preview struct {
	Folders       []FolderPreview
	TotalMessages uint32
}

type FolderResult struct {
	Folder   string
	Migrated int
	Skipped  int
	Errors   int
}

type ExecuteResult struct {
	Folders        []FolderResult
	TotalMigrated  int
	TotalSkipped   int
	TotalErrors    int
}

type Module interface {
	Preview(ctx context.Context) (*Preview, error)
	Execute(ctx context.Context) (*ExecuteResult, error)
	ListFolders(ctx context.Context) ([]FolderPreview, error)
	SyncFolder(ctx context.Context, folder string) (*FolderResult, error)
	SetBatchSize(size int)
	SetProgressCallbacks(onProgress func(folder string, current, total int), onFolderStart func(folder string, total int), onFolderDone func(folder string))
}

type Adapter struct {
	inner *mailmigrate.IMAPSync
}

func New(source, dest *imapclient.Client, state *mailmigrate.MigrationState, mapper *mailmigrate.FolderMapper, statePath string) Module {
	return &Adapter{inner: mailmigrate.NewIMAPSync(source, dest, state, mapper, statePath)}
}

func (a *Adapter) SetBatchSize(size int) {
	a.inner.BatchSize = size
}

func (a *Adapter) SetProgressCallbacks(onProgress func(folder string, current, total int), onFolderStart func(folder string, total int), onFolderDone func(folder string)) {
	a.inner.OnProgress = onProgress
	a.inner.OnFolderStart = onFolderStart
	a.inner.OnFolderDone = onFolderDone
}

func (a *Adapter) Preview(ctx context.Context) (*Preview, error) {
	dr, err := a.inner.DryRun(ctx)
	if err != nil {
		return nil, err
	}
	out := &Preview{TotalMessages: dr.TotalMessages, Folders: make([]FolderPreview, 0, len(dr.Folders))}
	for _, f := range dr.Folders {
		out.Folders = append(out.Folders, FolderPreview{Name: f.Name, Messages: f.Messages, MappedTo: f.MappedTo, Skipped: f.WillSkip})
	}
	return out, nil
}

func (a *Adapter) ListFolders(ctx context.Context) ([]FolderPreview, error) {
	folders, err := a.inner.ListSourceFolders(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]FolderPreview, 0, len(folders))
	for _, f := range folders {
		out = append(out, FolderPreview{Name: f.Name, Messages: f.Messages, MappedTo: f.MappedTo, Skipped: f.Skip})
	}
	return out, nil
}

func (a *Adapter) SyncFolder(ctx context.Context, folder string) (*FolderResult, error) {
	fr, err := a.inner.SyncFolder(ctx, folder)
	if err != nil {
		return nil, err
	}
	return &FolderResult{Folder: fr.Folder, Migrated: fr.Migrated, Skipped: fr.Skipped, Errors: fr.Errors}, nil
}

func (a *Adapter) Execute(ctx context.Context) (*ExecuteResult, error) {
	r, err := a.inner.SyncAll(ctx)
	if err != nil {
		return nil, err
	}
	out := &ExecuteResult{TotalMigrated: r.TotalMigrated, TotalSkipped: r.TotalSkipped, TotalErrors: r.TotalErrors, Folders: make([]FolderResult, 0, len(r.Folders))}
	for _, f := range r.Folders {
		out.Folders = append(out.Folders, FolderResult{Folder: f.Folder, Migrated: f.Migrated, Skipped: f.Skipped, Errors: f.Errors})
	}
	return out, nil
}
