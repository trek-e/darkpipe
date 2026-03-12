// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/profiles/internal/logutil"
	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/autoconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/autodiscover"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// profileDebug controls whether full email addresses are logged.
// Set PROFILE_DEBUG=true to enable verbose logging.
var profileDebug = strings.EqualFold(os.Getenv("PROFILE_DEBUG"), "true")

// logEmail returns the email address redacted for logging, unless PROFILE_DEBUG is set.
func logEmail(email string) string {
	if profileDebug {
		return email
	}
	return logutil.RedactEmail(email)
}

// ServerConfig holds the configuration for the profile server.
type ServerConfig struct {
	Domain       string
	Hostname     string
	CalDAVURL    string
	CardDAVURL   string
	CalDAVPort   int
	CardDAVPort  int
	AdminUser    string
	AdminPass    string
}

// ProfileHandler handles all profile-related HTTP requests.
type ProfileHandler struct {
	ProfileGen   *mobileconfig.ProfileGenerator
	TokenStore   qrcode.TokenStore
	AppPassStore apppassword.Store
	Config       ServerConfig
}

// HandleProfileDownload handles GET /profile/download?token=<token>
func (h *ProfileHandler) HandleProfileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusBadRequest)
		return
	}

	// Validate token (single-use)
	email, valid, err := h.TokenStore.Validate(token)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid, expired, or already-used token", http.StatusUnauthorized)
		return
	}

	// Generate new app password for this device
	deviceName := fmt.Sprintf("QR-%d", time.Now().Unix())
	plainPassword, err := apppassword.GenerateAppPassword()
	if err != nil {
		log.Printf("App password generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = h.AppPassStore.Create(email, deviceName, plainPassword)
	if err != nil {
		log.Printf("App password storage error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate .mobileconfig profile
	profileCfg := mobileconfig.ProfileConfig{
		Domain:       h.Config.Domain,
		MailHostname: h.Config.Hostname,
		Email:        email,
		AppPassword:  plainPassword,
		CalDAVURL:    h.Config.CalDAVURL,
		CardDAVURL:   h.Config.CardDAVURL,
		CalDAVPort:   h.Config.CalDAVPort,
		CardDAVPort:  h.Config.CardDAVPort,
	}

	profileData, err := h.ProfileGen.GenerateProfile(profileCfg)
	if err != nil {
		log.Printf("Profile generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Serve as .mobileconfig file
	w.Header().Set("Content-Type", "application/x-apple-aspen-config")
	w.Header().Set("Content-Disposition", `attachment; filename="darkpipe-mail.mobileconfig"`)
	w.WriteHeader(http.StatusOK)
	w.Write(profileData)

	log.Printf("Profile downloaded for %s (token: %s, device: %s)", logEmail(email), token[:8], deviceName)
}

// HandleAutoconfig handles GET /mail/config-v1.1.xml?emailaddress=<email>
func (h *ProfileHandler) HandleAutoconfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Email address is optional (Thunderbird sends it, but we don't require it)
	emailAddress := r.URL.Query().Get("emailaddress")
	_ = emailAddress // Not used in generation for now

	xml, err := autoconfig.GenerateAutoconfig(h.Config.Domain, h.Config.Hostname)
	if err != nil {
		log.Printf("Autoconfig generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(xml)
}

// HandleAutodiscover handles GET/POST /autodiscover/autodiscover.xml
func (h *ProfileHandler) HandleAutodiscover(w http.ResponseWriter, r *http.Request) {
	var email string

	if r.Method == http.MethodPost {
		// Outlook sends POST with XML body containing email
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Simple extraction of email from XML (basic approach)
		// Full XML parsing would be more robust, but this is sufficient
		bodyStr := string(body)
		email = extractEmailFromAutodiscoverXML(bodyStr)
	}

	// Generate autodiscover XML
	xml, err := autodiscover.GenerateAutodiscover(email, h.Config.Hostname)
	if err != nil {
		log.Printf("Autodiscover generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(xml)
}

// HandleHealth handles GET /health
func (h *ProfileHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleQRGenerate handles GET /qr/generate?email=<email>
// Authenticated endpoint that generates a QR code and returns PNG image.
func (h *ProfileHandler) HandleQRGenerate(w http.ResponseWriter, r *http.Request) {
	// Basic auth check
	user, pass, ok := r.BasicAuth()
	if !ok || user != h.Config.AdminUser || pass != h.Config.AdminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="QR Generation"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}

	// Generate QR code URL
	url, err := qrcode.GenerateQRCode(h.Config.Hostname, email, h.TokenStore)
	if err != nil {
		log.Printf("QR code generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate PNG image
	png, err := qrcode.GenerateQRCodePNG(url, 256)
	if err != nil {
		log.Printf("QR code PNG generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return as downloadable PNG
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="qr-%s.png"`, email))
	w.WriteHeader(http.StatusOK)
	w.Write(png)

	log.Printf("QR code generated for %s", logEmail(email))
}

// HandleQRImage handles GET /qr/image?email=<email>
// Same as HandleQRGenerate but returns inline image (for webmail embedding).
func (h *ProfileHandler) HandleQRImage(w http.ResponseWriter, r *http.Request) {
	// Basic auth check
	user, pass, ok := r.BasicAuth()
	if !ok || user != h.Config.AdminUser || pass != h.Config.AdminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="QR Generation"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}

	// Generate QR code URL
	url, err := qrcode.GenerateQRCode(h.Config.Hostname, email, h.TokenStore)
	if err != nil {
		log.Printf("QR code generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate PNG image
	png, err := qrcode.GenerateQRCodePNG(url, 256)
	if err != nil {
		log.Printf("QR code PNG generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return as inline image
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

// extractEmailFromAutodiscoverXML extracts email from Outlook autodiscover request XML.
// This is a simple string search approach; full XML parsing would be more robust.
func extractEmailFromAutodiscoverXML(xml string) string {
	// Look for <EMailAddress>user@example.com</EMailAddress>
	start := "<EMailAddress>"
	end := "</EMailAddress>"

	startIdx := len(start)
	endIdx := len(xml)

	if idx := findString(xml, start); idx >= 0 {
		startIdx = idx + len(start)
	} else {
		return ""
	}

	if idx := findString(xml[startIdx:], end); idx >= 0 {
		endIdx = startIdx + idx
	} else {
		return ""
	}

	return xml[startIdx:endIdx]
}

// findString is a helper to find substring index.
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// LogRequest is middleware for request logging in JSON format.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lw, r)

		logEntry := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      logutil.RedactQueryParams(r.URL.RawQuery),
			"status":     lw.statusCode,
			"duration":   time.Since(start).Milliseconds(),
			"remote":     r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}

		logJSON, _ := json.Marshal(logEntry)
		log.Println(string(logJSON))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
