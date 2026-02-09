// Package main provides the entrypoint for the cloud relay daemon.
//
// The relay daemon listens on localhost:10025 for SMTP connections from
// Postfix and forwards mail to the home device via WireGuard or mTLS transport.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/notify"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/smtp"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/tls"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting DarkPipe cloud relay daemon...")

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config: listen=%s transport=%s home=%s strict_mode=%v webhook=%v",
		cfg.ListenAddr, cfg.TransportType, cfg.HomeDeviceAddr, cfg.StrictModeEnabled, cfg.WebhookURL != "")

	// Set up notification system if webhook is configured
	var notifier notify.Notifier
	if cfg.WebhookURL != "" {
		log.Printf("Enabling webhook notifications to %s", cfg.WebhookURL)
		webhookNotifier := notify.NewWebhookNotifier(cfg.WebhookURL)
		notifier = notify.NewMultiNotifier(webhookNotifier)
		defer notifier.Close()
	} else {
		// Use a no-op notifier if webhook is not configured
		notifier = &noopNotifier{}
	}

	// Apply strict mode configuration if enabled
	if cfg.StrictModeEnabled {
		log.Println("Applying strict TLS mode to Postfix...")
		strictMode := tls.NewStrictMode(true)
		if err := strictMode.GeneratePolicyMap(); err != nil {
			log.Printf("WARNING: Failed to generate TLS policy map: %v", err)
		}
		if err := strictMode.ApplyToPostfix(); err != nil {
			log.Printf("WARNING: Failed to apply strict mode to Postfix: %v", err)
		}
	}

	// TLS monitor infrastructure is ready
	// The actual log monitoring will be set up in Task 2 when we modify entrypoint.sh
	// to pipe Postfix logs to the monitor
	if notifier != nil {
		log.Println("TLS monitor ready (will be activated via entrypoint.sh)")
	}

	// Create appropriate forwarder based on transport type
	var forwarder forward.Forwarder
	if cfg.TransportType == "mtls" {
		log.Println("Initializing mTLS forwarder...")
		forwarder, err = forward.NewMTLSForwarder(
			cfg.CACertPath,
			cfg.ClientCertPath,
			cfg.ClientKeyPath,
			cfg.HomeDeviceAddr,
		)
		if err != nil {
			log.Fatalf("Failed to create mTLS forwarder: %v", err)
		}
	} else {
		log.Println("Initializing WireGuard forwarder...")
		forwarder = forward.NewWireGuardForwarder(cfg.HomeDeviceAddr)
	}
	defer forwarder.Close()

	// Create and start SMTP server
	server := smtp.NewServer(forwarder, cfg)
	log.Printf("Relay daemon listening on %s (forwarding to %s via %s)", cfg.ListenAddr, cfg.HomeDeviceAddr, cfg.TransportType)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutdown signal received, stopping server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server (blocks until shutdown)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Relay daemon stopped")
}

// noopNotifier is a no-op notifier used when webhook notifications are disabled.
type noopNotifier struct{}

func (n *noopNotifier) Send(ctx context.Context, event notify.Event) error {
	return nil
}

func (n *noopNotifier) Close() error {
	return nil
}
