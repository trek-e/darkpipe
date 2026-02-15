package mailmigrate

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

// VCFImportResult holds the result of importing a VCF file.
type VCFImportResult struct {
	Total   int
	Created int
	Merged  int
	Skipped int
	Errors  int
}

// VCFDryRunResult holds the result of a VCF dry-run.
type VCFDryRunResult struct {
	Total        int
	WithEmail    int
	WithoutEmail int
}

// ICSImportResult holds the result of importing an ICS file.
type ICSImportResult struct {
	Total    int
	Imported int
	Skipped  int
	Errors   int
}

// ICSDryRunResult holds the result of an ICS dry-run.
type ICSDryRunResult struct {
	Total int
}

// ImportVCF imports contacts from a VCF file to a CardDAV server.
func ImportVCF(ctx context.Context, filePath string, dest *carddav.Client, destPath string, state *MigrationState, statePath string, mergeMode string, onProgress func(current, total int)) (*VCFImportResult, error) {
	result := &VCFImportResult{}

	// Parse VCF file
	cards, err := parseVCFFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VCF file: %w", err)
	}

	result.Total = len(cards)

	// Query all contacts from destination (for merge detection)
	destQuery := carddav.AddressBookQuery{
		DataRequest: carddav.AddressDataRequest{
			Props: []string{"*"},
		},
	}

	destObjects, err := dest.QueryAddressBook(ctx, destPath, &destQuery)
	if err != nil {
		// If destination doesn't exist, assume empty
		destObjects = nil
	}

	// Build destination email index
	destByEmail := make(map[string]*carddav.AddressObject)
	for i := range destObjects {
		email := extractPrimaryEmail(destObjects[i].Card)
		if email != "" {
			destByEmail[email] = &destObjects[i]
		}
	}

	// Import each card
	for i, card := range cards {
		email := extractPrimaryEmail(card)
		if email == "" {
			log.Printf("Warning: contact has no email, skipping")
			result.Errors++
			continue
		}

		// Check if already migrated
		if state.IsContactMigrated(email) {
			result.Skipped++
			if onProgress != nil {
				onProgress(i+1, result.Total)
			}
			continue
		}

		// Check if contact exists in destination
		destObj, exists := destByEmail[email]

		var cardToWrite vcard.Card
		var merged bool

		if exists {
			// Contact exists - apply merge logic
			switch mergeMode {
			case "append":
				// Fill empty fields from source, don't overwrite
				mergedCard, _ := mergeContact(destObj.Card, card)
				cardToWrite = mergedCard
				merged = true

			case "overwrite":
				// Replace with source
				cardToWrite = card

			case "skip":
				// Skip entirely
				result.Skipped++
				if onProgress != nil {
					onProgress(i+1, result.Total)
				}
				continue
			}
		} else {
			// New contact
			cardToWrite = card
		}

		// PUT contact to destination
		contactPath := fmt.Sprintf("%s/%s.vcf", destPath, email)
		_, err = dest.PutAddressObject(ctx, contactPath, cardToWrite)
		if err != nil {
			log.Printf("Warning: failed to put contact %s: %v", email, err)
			result.Errors++
			continue
		}

		// Mark as migrated
		if err := state.MarkContactMigrated(email); err != nil {
			log.Printf("Warning: failed to mark contact %s as migrated: %v", email, err)
		}

		if merged {
			result.Merged++
		} else {
			result.Created++
		}

		if onProgress != nil {
			onProgress(i+1, result.Total)
		}

		// Periodic save
		if (i+1)%100 == 0 {
			state.Save()
		}
	}

	// Final save
	state.Save()

	return result, nil
}

// DryRunVCF analyzes a VCF file without importing.
func DryRunVCF(filePath string) (*VCFDryRunResult, error) {
	cards, err := parseVCFFile(filePath)
	if err != nil {
		return nil, err
	}

	result := &VCFDryRunResult{
		Total: len(cards),
	}

	for _, card := range cards {
		email := extractPrimaryEmail(card)
		if email != "" {
			result.WithEmail++
		} else {
			result.WithoutEmail++
		}
	}

	return result, nil
}

// ImportICS imports calendar events from an ICS file to a CalDAV server.
func ImportICS(ctx context.Context, filePath string, dest *caldav.Client, destPath string, state *MigrationState, statePath string, onProgress func(current, total int)) (*ICSImportResult, error) {
	result := &ICSImportResult{}

	// Parse ICS file
	calendars, err := parseICSFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICS file: %w", err)
	}

	// Count events
	var events []*ical.Component
	for _, cal := range calendars {
		for _, comp := range cal.Children {
			if comp.Name == "VEVENT" {
				events = append(events, comp)
			}
		}
	}

	result.Total = len(events)

	// Import each event
	for i, event := range events {
		// Extract UID
		uidProp := event.Props.Get("UID")
		if uidProp == nil {
			log.Printf("Warning: event has no UID, skipping")
			result.Errors++
			continue
		}
		uid := uidProp.Value

		// Check if already migrated
		if state.IsCalEventMigrated(uid) {
			result.Skipped++
			if onProgress != nil {
				onProgress(i+1, result.Total)
			}
			continue
		}

		// Wrap event in VCALENDAR envelope
		cal := ical.NewCalendar()
		cal.Props.SetText("VERSION", "2.0")
		cal.Props.SetText("PRODID", "-//DarkPipe//Mail Migration//EN")
		cal.Children = append(cal.Children, event)

		// PUT to destination
		eventPath := fmt.Sprintf("%s/%s.ics", destPath, uid)
		_, err = dest.PutCalendarObject(ctx, eventPath, cal)
		if err != nil {
			log.Printf("Warning: failed to put event %s: %v", uid, err)
			result.Errors++
			continue
		}

		// Mark as migrated
		if err := state.MarkCalEventMigrated(uid); err != nil {
			log.Printf("Warning: failed to mark event %s as migrated: %v", uid, err)
		}

		result.Imported++

		if onProgress != nil {
			onProgress(i+1, result.Total)
		}

		// Periodic save
		if (i+1)%100 == 0 {
			state.Save()
		}
	}

	// Final save
	state.Save()

	return result, nil
}

// DryRunICS analyzes an ICS file without importing.
func DryRunICS(filePath string) (*ICSDryRunResult, error) {
	calendars, err := parseICSFile(filePath)
	if err != nil {
		return nil, err
	}

	// Count events
	eventCount := 0
	for _, cal := range calendars {
		for _, comp := range cal.Children {
			if comp.Name == "VEVENT" {
				eventCount++
			}
		}
	}

	return &ICSDryRunResult{
		Total: eventCount,
	}, nil
}

// parseVCFFile parses a VCF file that may contain multiple vCards.
func parseVCFFile(path string) ([]vcard.Card, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cards []vcard.Card
	decoder := vcard.NewDecoder(strings.NewReader(string(data)))

	for {
		card, err := decoder.Decode()
		if err != nil {
			// End of file or error
			break
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// parseICSFile parses an ICS file.
func parseICSFile(path string) ([]*ical.Calendar, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var calendars []*ical.Calendar
	decoder := ical.NewDecoder(strings.NewReader(string(data)))

	for {
		cal, err := decoder.Decode()
		if err != nil {
			// End of file or error
			break
		}
		calendars = append(calendars, cal)
	}

	return calendars, nil
}
