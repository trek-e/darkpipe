// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package mailmigrate

import (
	"context"
	"fmt"
	"log"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav/carddav"
)

// AddressBookInfo holds information about an address book.
type AddressBookInfo struct {
	Name         string
	Path         string
	ContactCount int
}

// CardSyncResult holds the result of syncing a single address book.
type CardSyncResult struct {
	Book    string
	Created int
	Merged  int
	Skipped int
	Errors  int
}

// CardSyncAllResult holds the result of syncing all address books.
type CardSyncAllResult struct {
	Books    []CardSyncResult
	Total    int
	Created  int
	Merged   int
	Skipped  int
	Errors   int
}

// CardDryRunResult holds the result of a dry-run.
type CardDryRunResult struct {
	Books []AddressBookInfo
	Total int
}

// CardDAVSync handles CardDAV contact synchronization.
type CardDAVSync struct {
	Source    *carddav.Client
	Dest      *carddav.Client
	State     *MigrationState
	StatePath string
	MergeMode string // "append", "overwrite", "skip"

	// Callbacks for progress reporting
	OnProgress func(book string, current, total int)
}

// NewCardDAVSync creates a new CardDAV sync instance.
func NewCardDAVSync(source, dest *carddav.Client, state *MigrationState, statePath string) *CardDAVSync {
	return &CardDAVSync{
		Source:    source,
		Dest:      dest,
		State:     state,
		StatePath: statePath,
		MergeMode: "append", // default merge mode
	}
}

// DiscoverAddressBooks finds all address books on the source server.
func (c *CardDAVSync) DiscoverAddressBooks(ctx context.Context, principal string) ([]AddressBookInfo, error) {
	homeSet, err := c.Source.FindAddressBookHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("find address book home set: %w", err)
	}

	books, err := c.Source.FindAddressBooks(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("find address books: %w", err)
	}

	var result []AddressBookInfo
	for _, book := range books {
		// Query to count contacts
		query := carddav.AddressBookQuery{
			DataRequest: carddav.AddressDataRequest{
				Props: []string{"*"}, // AllProp
			},
		}

		objects, err := c.Source.QueryAddressBook(ctx, book.Path, &query)
		if err != nil {
			log.Printf("Warning: failed to query address book %s: %v", book.Path, err)
			continue
		}

		result = append(result, AddressBookInfo{
			Name:         book.Name,
			Path:         book.Path,
			ContactCount: len(objects),
		})
	}

	return result, nil
}

// SyncAddressBook syncs a single address book from source to destination.
func (c *CardDAVSync) SyncAddressBook(ctx context.Context, sourcePath, destPath string) (*CardSyncResult, error) {
	result := &CardSyncResult{
		Book: sourcePath,
	}

	// Query all contacts from source
	sourceQuery := carddav.AddressBookQuery{
		DataRequest: carddav.AddressDataRequest{
			Props: []string{"*"}, // AllProp
		},
	}

	sourceObjects, err := c.Source.QueryAddressBook(ctx, sourcePath, &sourceQuery)
	if err != nil {
		return nil, fmt.Errorf("query source address book: %w", err)
	}

	// Query all contacts from destination (for merge detection)
	destQuery := carddav.AddressBookQuery{
		DataRequest: carddav.AddressDataRequest{
			Props: []string{"*"},
		},
	}

	destObjects, err := c.Dest.QueryAddressBook(ctx, destPath, &destQuery)
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

	total := len(sourceObjects)
	for i, sourceObj := range sourceObjects {
		// Extract primary email
		email := extractPrimaryEmail(sourceObj.Card)
		if email == "" {
			log.Printf("Warning: contact has no email, skipping")
			result.Errors++
			continue
		}

		// Check if already migrated
		if c.State.IsContactMigrated(email) {
			result.Skipped++
			if c.OnProgress != nil {
				c.OnProgress(sourcePath, i+1, total)
			}
			continue
		}

		// Check if contact exists in destination
		destObj, exists := destByEmail[email]

		var cardToWrite vcard.Card
		var merged bool

		if exists {
			// Contact exists - apply merge logic
			switch c.MergeMode {
			case "append":
				// Fill empty fields from source, don't overwrite
				mergedCard, _ := mergeContact(destObj.Card, sourceObj.Card)
				cardToWrite = mergedCard
				merged = true

			case "overwrite":
				// Replace with source
				cardToWrite = sourceObj.Card

			case "skip":
				// Skip entirely
				result.Skipped++
				if c.OnProgress != nil {
					c.OnProgress(sourcePath, i+1, total)
				}
				continue
			}
		} else {
			// New contact
			cardToWrite = sourceObj.Card
		}

		// PUT contact to destination
		contactPath := fmt.Sprintf("%s/%s.vcf", destPath, email)
		_, err = c.Dest.PutAddressObject(ctx, contactPath, cardToWrite)
		if err != nil {
			log.Printf("Warning: failed to put contact %s: %v", email, err)
			result.Errors++
			continue
		}

		// Mark as migrated
		if err := c.State.MarkContactMigrated(email); err != nil {
			log.Printf("Warning: failed to mark contact %s as migrated: %v", email, err)
		}

		if merged {
			result.Merged++
		} else {
			result.Created++
		}

		if c.OnProgress != nil {
			c.OnProgress(sourcePath, i+1, total)
		}
	}

	// Save state after address book completion
	if err := c.State.Save(); err != nil {
		log.Printf("Warning: failed to save state: %v", err)
	}

	return result, nil
}

// SyncAll discovers and syncs all address books.
func (c *CardDAVSync) SyncAll(ctx context.Context, sourcePrincipal, destPrincipal string) (*CardSyncAllResult, error) {
	result := &CardSyncAllResult{}

	// Discover address books on source
	books, err := c.DiscoverAddressBooks(ctx, sourcePrincipal)
	if err != nil {
		return nil, fmt.Errorf("discover address books: %w", err)
	}

	result.Total = len(books)

	// Discover destination home set
	destHomeSet, err := c.Dest.FindAddressBookHomeSet(ctx, destPrincipal)
	if err != nil {
		return nil, fmt.Errorf("find destination address book home set: %w", err)
	}

	// Sync each address book
	for _, book := range books {
		// Use same book name on destination
		destPath := fmt.Sprintf("%s/%s", destHomeSet, book.Name)

		syncResult, err := c.SyncAddressBook(ctx, book.Path, destPath)
		if err != nil {
			log.Printf("Warning: failed to sync address book %s: %v", book.Name, err)
			result.Books = append(result.Books, CardSyncResult{
				Book:   book.Name,
				Errors: book.ContactCount,
			})
			result.Errors += book.ContactCount
			continue
		}

		result.Books = append(result.Books, *syncResult)
		result.Created += syncResult.Created
		result.Merged += syncResult.Merged
		result.Skipped += syncResult.Skipped
		result.Errors += syncResult.Errors
	}

	return result, nil
}

// DryRun lists address books and contact counts without syncing.
func (c *CardDAVSync) DryRun(ctx context.Context, sourcePrincipal string) (*CardDryRunResult, error) {
	books, err := c.DiscoverAddressBooks(ctx, sourcePrincipal)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, book := range books {
		total += book.ContactCount
	}

	return &CardDryRunResult{
		Books: books,
		Total: total,
	}, nil
}

// mergeContact merges two vCards by filling empty fields from source.
// Returns the merged card and a list of merge decisions for logging.
func mergeContact(dest, source vcard.Card) (vcard.Card, []string) {
	merged := dest
	var decisions []string

	// Fields to consider for merge
	fieldsToMerge := []string{
		vcard.FieldTelephone,
		vcard.FieldAddress,
		vcard.FieldBirthday,
		vcard.FieldNote,
		vcard.FieldOrganization,
		vcard.FieldTitle,
		vcard.FieldURL,
		vcard.FieldPhoto,
		vcard.FieldNickname,
	}

	for _, field := range fieldsToMerge {
		destField := dest.Get(field)
		sourceField := source.Get(field)

		// If destination field is empty and source has a value, copy from source
		if destField == nil && sourceField != nil {
			merged.Set(field, sourceField)
			decisions = append(decisions, fmt.Sprintf("filled %s from source", field))
		}
		// If destination has a value, keep it (don't overwrite)
	}

	return merged, decisions
}

// extractPrimaryEmail extracts the first email from a vCard.
func extractPrimaryEmail(card vcard.Card) string {
	if email := card.PreferredValue(vcard.FieldEmail); email != "" {
		return email
	}
	emails := card.Values(vcard.FieldEmail)
	if len(emails) > 0 {
		return emails[0]
	}
	return ""
}
