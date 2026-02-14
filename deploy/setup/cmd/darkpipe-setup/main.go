package main

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/compose"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/migrate"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/secrets"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/validate"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "darkpipe-setup",
		Short: "DarkPipe interactive setup with live validation",
		Long:  "Interactive setup tool for DarkPipe mail server deployment",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("darkpipe-setup version %s\n", version)
		},
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Run interactive setup wizard",
		Long:  "Interactive wizard that collects configuration and generates docker-compose.yml",
		Run:   runSetup,
	}

	rootCmd.AddCommand(versionCmd, setupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runSetup(cmd *cobra.Command, args []string) {
	pterm.DefaultHeader.WithFullWidth().Println("DarkPipe Setup Wizard")
	pterm.Info.Println("This wizard will help you configure DarkPipe mail server")
	fmt.Println()

	cfg := config.DefaultConfig()

	// Check for existing configuration
	if _, err := os.Stat(config.ConfigFile); err == nil {
		existingCfg, err := config.LoadConfig(config.ConfigFile)
		if err != nil {
			pterm.Warning.Printf("Failed to load existing config: %v\n", err)
		} else {
			pterm.Info.Printf("Found existing configuration (version %s)\n", existingCfg.Version)

			upgrade := false
			prompt := &survey.Confirm{
				Message: "Upgrade existing configuration?",
				Default: true,
			}
			if err := survey.AskOne(prompt, &upgrade); err != nil {
				pterm.Error.Println("Setup cancelled")
				return
			}

			if upgrade {
				if err := migrate.Migrate(existingCfg); err != nil {
					pterm.Error.Printf("Migration failed: %v\n", err)
					return
				}
				cfg = existingCfg
			} else {
				pterm.Warning.Println("Using existing configuration without changes")
				return
			}
		}
	}

	// Ask setup mode (Quick vs Advanced)
	var setupMode string
	modePrompt := &survey.Select{
		Message: "Setup mode:",
		Options: []string{"Quick (recommended defaults)", "Advanced (customize everything)"},
		Default: "Quick (recommended defaults)",
	}
	if err := survey.AskOne(modePrompt, &setupMode); err != nil {
		pterm.Error.Println("Setup cancelled")
		return
	}

	isQuick := (setupMode == "Quick (recommended defaults)")

	fmt.Println()
	pterm.DefaultSection.Println("Configuration")

	// Mail domain (always ask)
	if err := askMailDomain(cfg); err != nil {
		pterm.Error.Printf("Setup failed: %v\n", err)
		return
	}

	// Relay hostname (always ask)
	if err := askRelayHostname(cfg); err != nil {
		pterm.Error.Printf("Setup failed: %v\n", err)
		return
	}

	// Admin email (always ask)
	if err := askAdminEmail(cfg); err != nil {
		pterm.Error.Printf("Setup failed: %v\n", err)
		return
	}

	// Advanced mode: ask additional questions
	if !isQuick {
		if err := askMailServer(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}

		if err := askWebmail(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}

		if err := askCalendar(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}

		if err := askTransport(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}

		if err := askQueueEnabled(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}

		if err := askStrictMode(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}
	}

	// Validate configuration
	if err := migrate.ValidateConfig(cfg); err != nil {
		pterm.Error.Printf("Configuration validation failed: %v\n", err)
		return
	}

	// Display summary
	fmt.Println()
	pterm.DefaultSection.Println("Configuration Summary")
	displaySummary(cfg)

	// Confirm
	proceed := false
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with this configuration?",
		Default: true,
	}
	if err := survey.AskOne(confirmPrompt, &proceed); err != nil || !proceed {
		pterm.Warning.Println("Setup cancelled")
		return
	}

	// Generate outputs
	fmt.Println()
	pterm.DefaultSection.Println("Generating Configuration Files")

	progressbar, _ := pterm.DefaultProgressbar.WithTotal(4).WithTitle("Setup Progress").Start()

	// Generate secrets
	progressbar.UpdateTitle("Generating Docker secrets...")
	if err := secrets.GenerateSecrets(config.SecretsDir); err != nil {
		pterm.Error.Printf("Failed to generate secrets: %v\n", err)
		return
	}
	progressbar.Increment()

	// Generate docker-compose.yml
	progressbar.UpdateTitle("Creating docker-compose.yml...")
	if err := compose.Generate(cfg, "docker-compose.yml"); err != nil {
		pterm.Error.Printf("Failed to generate compose file: %v\n", err)
		return
	}
	progressbar.Increment()

	// Save configuration
	progressbar.UpdateTitle("Saving configuration...")
	if err := config.SaveConfig(cfg, config.ConfigFile); err != nil {
		pterm.Error.Printf("Failed to save config: %v\n", err)
		return
	}
	progressbar.Increment()

	// Create marker file
	progressbar.UpdateTitle("Creating setup marker...")
	if err := os.WriteFile(".darkpipe-configured", []byte("configured"), 0644); err != nil {
		pterm.Error.Printf("Failed to create marker file: %v\n", err)
		return
	}
	progressbar.Increment()

	progressbar.Stop()

	// Success output
	fmt.Println()
	pterm.DefaultBox.WithTitle("Setup Complete!").WithTitleTopCenter().Println(
		"DarkPipe has been configured successfully.\n\n" +
			"Next steps:\n" +
			"  1. Review docker-compose.yml\n" +
			"  2. Set up DNS records: darkpipe-dns-setup --domain " + cfg.MailDomain + "\n" +
			"  3. Start services: docker compose up -d\n" +
			"  4. Check logs: docker compose logs -f",
	)
}

func askMailDomain(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Primary mail domain:",
		Default: cfg.MailDomain,
	}
	validator := survey.WithValidator(func(val interface{}) error {
		domain := val.(string)
		spinner, _ := pterm.DefaultSpinner.Start("Validating DNS for " + domain + "...")

		err := validate.ValidateDomain(domain)
		if err != nil {
			spinner.Fail("DNS validation warning: " + err.Error())
			pterm.Warning.Println("You can continue setup, but make sure to configure DNS before starting services")
			return nil // Allow continuation
		}

		spinner.Success("DNS validation passed")
		return nil
	})

	return survey.AskOne(prompt, &cfg.MailDomain, validator)
}

func askRelayHostname(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Cloud relay hostname (must have port 25 open):",
		Default: cfg.RelayHostname,
	}
	validator := survey.WithValidator(func(val interface{}) error {
		hostname := val.(string)
		spinner, _ := pterm.DefaultSpinner.Start("Testing SMTP port 25 on " + hostname + "...")

		err := validate.ValidateSMTPPort(hostname, 25, 5*time.Second)
		if err != nil {
			spinner.Fail("Port 25 test warning: " + err.Error())
			pterm.Warning.Println("Many VPS providers block port 25. Ensure your provider allows SMTP.")
			return nil // Allow continuation
		}

		spinner.Success("Port 25 is accessible")
		return nil
	})

	return survey.AskOne(prompt, &cfg.RelayHostname, validator)
}

func askAdminEmail(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Admin email address:",
		Default: cfg.AdminEmail,
	}
	return survey.AskOne(prompt, &cfg.AdminEmail)
}

func askMailServer(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Mail server component:",
		Options: []string{"stalwart", "maddy", "postfix-dovecot"},
		Default: cfg.MailServer,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"stalwart":        "Modern, all-in-one (Rust, IMAP4rev2/JMAP/CalDAV/CardDAV)",
				"maddy":           "Minimal, Go-based single binary",
				"postfix-dovecot": "Traditional, battle-tested MTA+IMAP",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.MailServer)
}

func askWebmail(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Webmail component:",
		Options: []string{"none", "roundcube", "snappymail"},
		Default: cfg.Webmail,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"none":       "No webmail (IMAP only)",
				"roundcube":  "Traditional, feature-rich, PHP-based",
				"snappymail": "Modern, fast, lightweight (recommended)",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Webmail)
}

func askCalendar(cfg *config.Config) error {
	options := []string{"none", "radicale", "builtin"}

	// Filter out builtin if not using Stalwart
	if cfg.MailServer != "stalwart" {
		options = []string{"none", "radicale"}
		if cfg.Calendar == "builtin" {
			cfg.Calendar = "radicale" // Auto-default to radicale
		}
	}

	prompt := &survey.Select{
		Message: "Calendar/Contacts component:",
		Options: options,
		Default: cfg.Calendar,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"none":     "No calendar/contacts",
				"radicale": "Standalone CalDAV/CardDAV server",
				"builtin":  "Stalwart built-in CalDAV/CardDAV",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Calendar)
}

func askTransport(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Transport layer:",
		Options: []string{"wireguard", "mtls"},
		Default: cfg.Transport,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"wireguard": "WireGuard VPN (recommended, simpler NAT traversal)",
				"mtls":      "Mutual TLS (for restrictive networks)",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Transport)
}

func askQueueEnabled(cfg *config.Config) error {
	prompt := &survey.Confirm{
		Message: "Enable message queuing (for offline relay)?",
		Default: cfg.QueueEnabled,
	}
	return survey.AskOne(prompt, &cfg.QueueEnabled)
}

func askStrictMode(cfg *config.Config) error {
	prompt := &survey.Confirm{
		Message: "Enable TLS strict mode?",
		Default: cfg.StrictMode,
	}
	return survey.AskOne(prompt, &cfg.StrictMode)
}

func displaySummary(cfg *config.Config) {
	data := [][]string{
		{"Mail Domain", cfg.MailDomain},
		{"Relay Hostname", cfg.RelayHostname},
		{"Admin Email", cfg.AdminEmail},
		{"Mail Server", cfg.MailServer},
		{"Webmail", cfg.Webmail},
		{"Calendar/Contacts", cfg.Calendar},
		{"Transport", cfg.Transport},
		{"Queue Enabled", fmt.Sprintf("%v", cfg.QueueEnabled)},
		{"TLS Strict Mode", fmt.Sprintf("%v", cfg.StrictMode)},
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Setting", "Value"},
	}).WithData(data).Render()
}
