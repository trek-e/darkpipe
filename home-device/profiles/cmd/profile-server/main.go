// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/cert"
	"github.com/darkpipe/darkpipe/monitoring/delivery"
	"github.com/darkpipe/darkpipe/monitoring/health"
	"github.com/darkpipe/darkpipe/monitoring/queue"
	"github.com/darkpipe/darkpipe/monitoring/status"
	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

func main() {
	// Check for CLI subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "qr":
			RunQRCommand(os.Args[2:])
			return
		case "status":
			status.RunStatusCommand(os.Args[2:])
			return
		}
	}

	// Load configuration from environment
	config := ServerConfig{
		Domain:      getEnv("MAIL_DOMAIN", "example.com"),
		Hostname:    getEnv("MAIL_HOSTNAME", "mail.example.com"),
		CalDAVURL:   getEnv("CALDAV_URL", ""),
		CardDAVURL:  getEnv("CARDDAV_URL", ""),
		CalDAVPort:  getEnvInt("CALDAV_PORT", 443),
		CardDAVPort: getEnvInt("CARDDAV_PORT", 443),
		AdminUser:   getEnv("ADMIN_USER", "admin"),
		AdminPass:   getEnv("ADMIN_PASSWORD", ""),
	}

	port := getEnv("PROFILE_SERVER_PORT", "8090")
	mailServerType := getEnv("MAIL_SERVER_TYPE", "stalwart")

	if config.AdminPass == "" {
		log.Println("WARNING: ADMIN_PASSWORD not set, QR generation endpoints will be insecure")
	}

	// Initialize app password store based on mail server type
	var appPassStore apppassword.Store

	switch mailServerType {
	case "stalwart":
		appPassStore = apppassword.NewStalwartStore()
		log.Printf("Using Stalwart app password store")

	case "dovecot":
		storePath := getEnv("APP_PASSWORD_STORE_PATH", "/data/app-passwords.json")
		appPassStore = apppassword.NewDovecotStore(storePath)
		log.Printf("Using Dovecot app password store (path: %s)", storePath)

	case "maddy":
		storePath := getEnv("APP_PASSWORD_STORE_PATH", "/data/app-passwords.json")
		appPassStore = apppassword.NewMaddyStore(storePath)
		log.Printf("Using Maddy app password store (path: %s)", storePath)

	case "postfix-dovecot":
		// Same as dovecot
		storePath := getEnv("APP_PASSWORD_STORE_PATH", "/data/app-passwords.json")
		appPassStore = apppassword.NewDovecotStore(storePath)
		log.Printf("Using Dovecot app password store (path: %s)", storePath)

	default:
		log.Fatalf("Unknown MAIL_SERVER_TYPE: %s (supported: stalwart, dovecot, maddy, postfix-dovecot)", mailServerType)
	}

	// Initialize generators
	profileGen := &mobileconfig.ProfileGenerator{}

	// Initialize token store
	tokenStore := qrcode.NewMemoryTokenStore()

	// Create handler
	handler := &ProfileHandler{
		ProfileGen:   profileGen,
		TokenStore:   tokenStore,
		AppPassStore: appPassStore,
		Config:       config,
	}

	// Initialize monitoring subsystem
	healthChecker := health.NewChecker()
	healthChecker.RegisterCheck("postfix", health.CheckPostfix)
	healthChecker.RegisterCheck("imap", health.CheckIMAP)

	deliveryTracker := delivery.NewDeliveryTracker(0) // 0 = use default (1000)

	// Certificate watcher with paths from env (comma-separated)
	var certPaths []string
	if paths := os.Getenv("MONITOR_CERT_PATHS"); paths != "" {
		certPaths = strings.Split(paths, ",")
	}
	certWatcher := cert.NewCertWatcher(certPaths)

	// Status aggregator wires all monitoring sources together
	aggregator := status.NewStatusAggregator(
		healthChecker,
		queue.GetQueueStats,
		deliveryTracker,
		certWatcher,
	)

	// Load status dashboard template
	statusTmpl, err := template.ParseFiles("templates/status.html")
	if err != nil {
		log.Printf("WARNING: Could not load status template: %v (dashboard disabled)", err)
	}

	// Start push-based monitoring if configured
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if healthcheckURL := os.Getenv("MONITOR_HEALTHCHECK_URL"); healthcheckURL != "" {
		pinger := status.NewHealthchecksPinger(healthcheckURL, 0)
		go pinger.Run(ctx, func() (*status.SystemStatus, error) {
			return aggregator.GetStatus(ctx)
		})
		log.Printf("Push monitoring enabled (URL: %s)", healthcheckURL)
	}

	// Start delivery log parser in background
	logPath := getEnv("MONITOR_LOG_PATH", "/var/log/mail.log")
	go startLogParser(ctx, logPath, deliveryTracker)

	// Create web UI handler
	webUI := NewWebUIHandler(appPassStore, tokenStore, config)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/profile/download", handler.HandleProfileDownload)
	mux.HandleFunc("/mail/config-v1.1.xml", handler.HandleAutoconfig)
	mux.HandleFunc("/autodiscover/autodiscover.xml", handler.HandleAutodiscover)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/health/live", health.LivenessHandler(healthChecker))
	mux.HandleFunc("/health/ready", health.ReadinessHandler(healthChecker))
	mux.HandleFunc("/qr/generate", handler.HandleQRGenerate)
	mux.HandleFunc("/qr/image", handler.HandleQRImage)

	// Status dashboard and API routes
	if statusTmpl != nil {
		mux.HandleFunc("/status", status.HandleDashboard(aggregator, statusTmpl))
	}
	mux.HandleFunc("/status/api", status.HandleStatusAPI(aggregator))

	// Web UI routes
	mux.HandleFunc("/devices", webUI.HandleDeviceList)
	mux.HandleFunc("/devices/add", webUI.HandleAddDevice)
	mux.HandleFunc("/devices/revoke", webUI.HandleRevokeDevice)
	mux.HandleFunc("/static/", webUI.ServeStatic)

	// Wrap with logging middleware
	loggedMux := LogRequest(mux)

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Profile server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel monitoring background goroutines
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv gets environment variable with fallback to default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with fallback to default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// startLogParser reads Postfix log lines and feeds them to the delivery tracker.
func startLogParser(ctx context.Context, logPath string, tracker *delivery.DeliveryTracker) {
	parser := &delivery.Parser{}

	// Try to open the log file, retry periodically if not available yet
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		f, err := os.Open(logPath)
		if err != nil {
			log.Printf("Waiting for mail log at %s: %v", logPath, err)
			time.Sleep(30 * time.Second)
			continue
		}

		// Tail the log file
		scanner := newLineScanner(f)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				f.Close()
				return
			default:
			}

			entry, err := parser.ParseLine(scanner.Text())
			if err == nil && entry != nil {
				tracker.Record(entry)
			}
		}

		f.Close()
		// File rotated or closed, wait and retry
		time.Sleep(5 * time.Second)
	}
}

// newLineScanner wraps bufio.Scanner for line-by-line reading.
func newLineScanner(f *os.File) *lineScanner {
	return &lineScanner{file: f}
}

type lineScanner struct {
	file *os.File
	buf  [4096]byte
	line string
	pos  int64
}

func (s *lineScanner) Scan() bool {
	// Seek to current position and read new data
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := s.file.Read(buf)
		if n == 0 || err != nil {
			// No more data, wait for more
			time.Sleep(1 * time.Second)
			continue
		}
		if buf[0] == '\n' {
			s.line = string(line)
			return true
		}
		line = append(line, buf[0])
	}
}

func (s *lineScanner) Text() string {
	return s.line
}
