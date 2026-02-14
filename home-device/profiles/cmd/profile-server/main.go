package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

func main() {
	// Check for CLI subcommands
	if len(os.Args) > 1 && os.Args[1] == "qr" {
		RunQRCommand(os.Args[2:])
		return
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

	// Create web UI handler
	webUI := NewWebUIHandler(appPassStore, tokenStore, config)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/profile/download", handler.HandleProfileDownload)
	mux.HandleFunc("/mail/config-v1.1.xml", handler.HandleAutoconfig)
	mux.HandleFunc("/autodiscover/autodiscover.xml", handler.HandleAutodiscover)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/qr/generate", handler.HandleQRGenerate)
	mux.HandleFunc("/qr/image", handler.HandleQRImage)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
