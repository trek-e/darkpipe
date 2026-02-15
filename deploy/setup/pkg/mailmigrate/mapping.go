package mailmigrate

import (
	"regexp"
	"strings"
)

// FolderMapper handles provider-specific folder name mappings and skip rules
type FolderMapper struct {
	Mappings        map[string]string // source folder -> destination folder
	SkipFolders     map[string]bool   // source folder -> should skip
	LabelsAsFolders bool              // Gmail: create subfolders instead of keywords
}

// NewFolderMapper creates a mapper with provider-specific defaults and user overrides
func NewFolderMapper(provider string, overrides map[string]string) *FolderMapper {
	mapper := &FolderMapper{
		Mappings:        make(map[string]string),
		SkipFolders:     make(map[string]bool),
		LabelsAsFolders: false,
	}

	// Apply provider-specific defaults
	switch strings.ToLower(provider) {
	case "gmail":
		// Gmail special folders (per locked decisions)
		mapper.Mappings["[Gmail]/Sent Mail"] = "Sent"
		mapper.Mappings["[Gmail]/Drafts"] = "Drafts"
		mapper.Mappings["[Gmail]/Trash"] = "Trash"
		mapper.Mappings["[Gmail]/Spam"] = "Junk"

		// Skip Gmail virtual folders
		mapper.SkipFolders["[Gmail]/All Mail"] = true
		mapper.SkipFolders["[Gmail]/Important"] = true
		mapper.SkipFolders["[Gmail]/Starred"] = true

	case "outlook":
		// Outlook special folders
		mapper.Mappings["Deleted Items"] = "Trash"
		mapper.Mappings["Sent Items"] = "Sent"
		mapper.Mappings["Junk Email"] = "Junk"

		// Skip Outlook clutter
		mapper.SkipFolders["Clutter"] = true

	case "generic", "icloud", "mailcow", "mailu", "docker-mailserver":
		// No default mappings for generic/other providers
		// Pass-through folder names as-is
	}

	// Apply user overrides
	for source, dest := range overrides {
		if dest == "" {
			// Empty string means skip this folder
			mapper.SkipFolders[source] = true
			delete(mapper.Mappings, source)
		} else {
			mapper.Mappings[source] = dest
			delete(mapper.SkipFolders, source)
		}
	}

	return mapper
}

// Map returns the destination folder name and whether to skip
// If no mapping exists, returns source folder name (pass-through)
func (m *FolderMapper) Map(sourceFolder string) (destFolder string, skip bool) {
	// Check if folder should be skipped
	if m.SkipFolders[sourceFolder] {
		return "", true
	}

	// Check for explicit mapping
	if dest, exists := m.Mappings[sourceFolder]; exists {
		return dest, false
	}

	// No mapping, pass through source name
	return sourceFolder, false
}

// AllMappings returns all configured mappings for dry-run preview
func (m *FolderMapper) AllMappings() map[string]string {
	result := make(map[string]string)

	// Copy all mappings
	for source, dest := range m.Mappings {
		result[source] = dest
	}

	// Add skip markers
	for source := range m.SkipFolders {
		result[source] = "(skip)"
	}

	return result
}

// LabelToKeyword converts Gmail label to valid IMAP keyword atom
// IMAP keywords are atoms: no spaces, special chars replaced
// Per research Pattern 5 and locked decisions
func (m *FolderMapper) LabelToKeyword(label string) string {
	// Replace spaces with underscores
	keyword := strings.ReplaceAll(label, " ", "_")

	// Remove special characters (keep alphanumeric, underscore, hyphen)
	regex := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	keyword = regex.ReplaceAllString(keyword, "")

	// IMAP keywords can't start with a digit (make it start with label_)
	if len(keyword) > 0 && keyword[0] >= '0' && keyword[0] <= '9' {
		keyword = "label_" + keyword
	}

	// Empty after sanitization, use default
	if keyword == "" {
		keyword = "custom_label"
	}

	return keyword
}
