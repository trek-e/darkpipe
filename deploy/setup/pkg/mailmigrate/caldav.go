// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"context"
	"fmt"
	"log"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

// CalendarInfo holds information about a calendar.
type CalendarInfo struct {
	Name       string
	Path       string
	EventCount int
}

// CalSyncResult holds the result of syncing a single calendar.
type CalSyncResult struct {
	Calendar string
	Migrated int
	Skipped  int
	Errors   int
}

// CalSyncAllResult holds the result of syncing all calendars.
type CalSyncAllResult struct {
	Calendars []CalSyncResult
	Total     int
	Migrated  int
	Skipped   int
	Errors    int
}

// CalDryRunResult holds the result of a dry-run.
type CalDryRunResult struct{
	Calendars []CalendarInfo
	Total     int
}

// CalDAVSync handles CalDAV calendar synchronization.
type CalDAVSync struct {
	Source   *caldav.Client
	Dest     *caldav.Client
	State    *MigrationState
	StatePath string

	// Callbacks for progress reporting
	OnProgress      func(calendar string, current, total int)
	OnCalendarStart func(calendar string, total int)
	OnCalendarDone  func(calendar string)
}

// NewCalDAVSync creates a new CalDAV sync instance.
func NewCalDAVSync(source, dest *caldav.Client, state *MigrationState, statePath string) *CalDAVSync {
	return &CalDAVSync{
		Source:    source,
		Dest:      dest,
		State:     state,
		StatePath: statePath,
	}
}

// DiscoverCalendars finds all calendars on the source server.
func (c *CalDAVSync) DiscoverCalendars(ctx context.Context, principal string) ([]CalendarInfo, error) {
	homeSet, err := c.Source.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("find calendar home set: %w", err)
	}

	calendars, err := c.Source.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("find calendars: %w", err)
	}

	var result []CalendarInfo
	for _, cal := range calendars {
		// Query to count events
		query := caldav.CalendarQuery{
			CompRequest: caldav.CalendarCompRequest{
				Name: "VCALENDAR",
				Comps: []caldav.CalendarCompRequest{
					{Name: "VEVENT"},
				},
			},
		}

		objects, err := c.Source.QueryCalendar(ctx, cal.Path, &query)
		if err != nil {
			log.Printf("Warning: failed to query calendar %s: %v", cal.Path, err)
			continue
		}

		result = append(result, CalendarInfo{
			Name:       cal.Name,
			Path:       cal.Path,
			EventCount: len(objects),
		})
	}

	return result, nil
}

// SyncCalendar syncs a single calendar from source to destination.
func (c *CalDAVSync) SyncCalendar(ctx context.Context, sourcePath, destPath string) (*CalSyncResult, error) {
	result := &CalSyncResult{
		Calendar: sourcePath,
	}

	// Query all events from source
	query := caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:  "VCALENDAR",
			Props: []string{"*"}, // AllProp
			Comps: []caldav.CalendarCompRequest{
				{Name: "VEVENT", Props: []string{"*"}},
			},
		},
	}

	objects, err := c.Source.QueryCalendar(ctx, sourcePath, &query)
	if err != nil {
		return nil, fmt.Errorf("query source calendar: %w", err)
	}

	total := len(objects)
	if c.OnCalendarStart != nil {
		c.OnCalendarStart(sourcePath, total)
	}

	for i, obj := range objects {
		// Extract UID from iCal data
		uid, err := extractEventUID(obj.Data)
		if err != nil {
			log.Printf("Warning: failed to extract UID from event: %v", err)
			result.Errors++
			continue
		}

		// Check if already migrated
		if c.State.IsCalEventMigrated(uid) {
			result.Skipped++
			if c.OnProgress != nil {
				c.OnProgress(sourcePath, i+1, total)
			}
			continue
		}

		// PUT event to destination
		eventPath := fmt.Sprintf("%s/%s.ics", destPath, uid)
		_, err = c.Dest.PutCalendarObject(ctx, eventPath, obj.Data)
		if err != nil {
			log.Printf("Warning: failed to put event %s: %v", uid, err)
			result.Errors++
			continue
		}

		// Mark as migrated
		if err := c.State.MarkCalEventMigrated(uid); err != nil {
			log.Printf("Warning: failed to mark event %s as migrated: %v", uid, err)
		}

		result.Migrated++
		if c.OnProgress != nil {
			c.OnProgress(sourcePath, i+1, total)
		}
	}

	// Save state after calendar completion
	if err := c.State.Save(); err != nil {
		log.Printf("Warning: failed to save state: %v", err)
	}

	if c.OnCalendarDone != nil {
		c.OnCalendarDone(sourcePath)
	}

	return result, nil
}

// SyncAll discovers and syncs all calendars.
func (c *CalDAVSync) SyncAll(ctx context.Context, sourcePrincipal, destPrincipal string) (*CalSyncAllResult, error) {
	result := &CalSyncAllResult{}

	// Discover calendars on source
	calendars, err := c.DiscoverCalendars(ctx, sourcePrincipal)
	if err != nil {
		return nil, fmt.Errorf("discover calendars: %w", err)
	}

	result.Total = len(calendars)

	// Discover destination home set
	destHomeSet, err := c.Dest.FindCalendarHomeSet(ctx, destPrincipal)
	if err != nil {
		return nil, fmt.Errorf("find destination calendar home set: %w", err)
	}

	// Sync each calendar
	for _, cal := range calendars {
		// Use same calendar name on destination
		destPath := fmt.Sprintf("%s/%s", destHomeSet, cal.Name)

		syncResult, err := c.SyncCalendar(ctx, cal.Path, destPath)
		if err != nil {
			log.Printf("Warning: failed to sync calendar %s: %v", cal.Name, err)
			result.Calendars = append(result.Calendars, CalSyncResult{
				Calendar: cal.Name,
				Errors:   cal.EventCount,
			})
			result.Errors += cal.EventCount
			continue
		}

		result.Calendars = append(result.Calendars, *syncResult)
		result.Migrated += syncResult.Migrated
		result.Skipped += syncResult.Skipped
		result.Errors += syncResult.Errors
	}

	return result, nil
}

// DryRun lists calendars and event counts without syncing.
func (c *CalDAVSync) DryRun(ctx context.Context, sourcePrincipal string) (*CalDryRunResult, error) {
	calendars, err := c.DiscoverCalendars(ctx, sourcePrincipal)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, cal := range calendars {
		total += cal.EventCount
	}

	return &CalDryRunResult{
		Calendars: calendars,
		Total:     total,
	}, nil
}

// extractEventUID extracts the UID from iCal data.
func extractEventUID(data *ical.Calendar) (string, error) {
	for _, comp := range data.Children {
		if comp.Name == "VEVENT" {
			uid := comp.Props.Get("UID")
			if uid != nil {
				return uid.Value, nil
			}
		}
	}
	return "", fmt.Errorf("no UID found in VEVENT")
}
