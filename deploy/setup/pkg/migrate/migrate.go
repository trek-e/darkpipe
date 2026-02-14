package migrate

import (
	"fmt"
	"log"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
)

// NeedsMigration checks if a config needs to be migrated
func NeedsMigration(cfg *config.Config) bool {
	return cfg.Version < config.CurrentVersion
}

// Migrate applies all necessary migrations to bring config to current version
func Migrate(cfg *config.Config) error {
	if !NeedsMigration(cfg) {
		return nil
	}

	log.Printf("Migrating configuration from version %s to %s", cfg.Version, config.CurrentVersion)

	// Apply migrations in sequence
	// No migrations yet (we're at v1)
	// Future migrations would be added here:
	// if cfg.Version == "1" {
	//     if err := migrateV1ToV2(cfg); err != nil {
	//         return err
	//     }
	//     cfg.Version = "2"
	// }

	cfg.Version = config.CurrentVersion
	log.Printf("Migration complete: now at version %s", cfg.Version)

	return nil
}

// migrateV1ToV2 is a placeholder for future schema migration
func migrateV1ToV2(cfg *config.Config) error {
	log.Println("Migrating from v1 to v2...")

	// Example migration logic:
	// - Add new field with default value
	// - Rename fields
	// - Convert data formats
	// - Etc.

	// For now, this is a no-op placeholder
	log.Println("  Added new field X with default value Y")

	return nil
}

// ValidateConfig performs sanity checks on migrated config
func ValidateConfig(cfg *config.Config) error {
	if cfg.Version != config.CurrentVersion {
		return fmt.Errorf("config version mismatch: expected %s, got %s", config.CurrentVersion, cfg.Version)
	}

	if cfg.MailDomain == "" {
		return fmt.Errorf("mail_domain is required")
	}

	if cfg.RelayHostname == "" {
		return fmt.Errorf("relay_hostname is required")
	}

	if cfg.AdminEmail == "" {
		return fmt.Errorf("admin_email is required")
	}

	validMailServers := map[string]bool{"stalwart": true, "maddy": true, "postfix-dovecot": true}
	if !validMailServers[cfg.MailServer] {
		return fmt.Errorf("invalid mail_server: %s (must be stalwart, maddy, or postfix-dovecot)", cfg.MailServer)
	}

	validWebmail := map[string]bool{"none": true, "roundcube": true, "snappymail": true}
	if !validWebmail[cfg.Webmail] {
		return fmt.Errorf("invalid webmail: %s (must be none, roundcube, or snappymail)", cfg.Webmail)
	}

	validCalendar := map[string]bool{"none": true, "radicale": true, "builtin": true}
	if !validCalendar[cfg.Calendar] {
		return fmt.Errorf("invalid calendar: %s (must be none, radicale, or builtin)", cfg.Calendar)
	}

	// Calendar validation: builtin only works with Stalwart
	if cfg.Calendar == "builtin" && cfg.MailServer != "stalwart" {
		return fmt.Errorf("builtin calendar only available with Stalwart mail server")
	}

	validTransport := map[string]bool{"wireguard": true, "mtls": true}
	if !validTransport[cfg.Transport] {
		return fmt.Errorf("invalid transport: %s (must be wireguard or mtls)", cfg.Transport)
	}

	return nil
}
