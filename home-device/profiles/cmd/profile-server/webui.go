package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

//go:embed templates/*.html static/*.css
var embedFS embed.FS

// WebUIHandler handles web UI requests for device management
type WebUIHandler struct {
	AppPassStore apppassword.Store
	TokenStore   qrcode.TokenStore
	Config       ServerConfig
	templates    *template.Template
}

// NewWebUIHandler creates a new web UI handler
func NewWebUIHandler(appPassStore apppassword.Store, tokenStore qrcode.TokenStore, config ServerConfig) *WebUIHandler {
	tmpl, err := template.ParseFS(embedFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &WebUIHandler{
		AppPassStore: appPassStore,
		TokenStore:   tokenStore,
		Config:       config,
		templates:    tmpl,
	}
}

// Device represents a device (app password) in the UI
type Device struct {
	ID          string
	Email       string
	DeviceName  string
	CreatedAt   time.Time
	LastUsedAt  time.Time
}

// AddDeviceData holds data for the add device page
type AddDeviceData struct {
	Email   string
	Error   string
	Success bool
}

// AddDeviceResultData holds data for the add device result page
type AddDeviceResultData struct {
	Email        string
	DeviceName   string
	AppPassword  string
	Platform     string
	QRCodeData   string        // base64-encoded PNG
	ProfileURL   string
	Instructions template.HTML // HTML-safe instructions
}

// DeviceListData holds data for the device list page
type DeviceListData struct {
	Email   string
	Devices []Device
	Error   string
	Success string
}

// HandleDeviceList displays all devices for the authenticated user
func (h *WebUIHandler) HandleDeviceList(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// List app passwords for this user
	appPasswords, err := h.AppPassStore.List(email)
	if err != nil {
		log.Printf("Failed to list app passwords for %s: %v", email, err)
		h.renderDeviceList(w, email, nil, "Failed to load devices", "")
		return
	}

	// Convert to Device structs
	devices := make([]Device, len(appPasswords))
	for i, ap := range appPasswords {
		devices[i] = Device{
			ID:         ap.ID,
			Email:      ap.Email,
			DeviceName: ap.DeviceName,
			CreatedAt:  ap.CreatedAt,
			LastUsedAt: ap.LastUsedAt,
		}
	}

	success := r.URL.Query().Get("success")
	h.renderDeviceList(w, email, devices, "", success)
}

// HandleAddDevice displays the add device form
func (h *WebUIHandler) HandleAddDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method == http.MethodGet {
		h.renderAddDevice(w, email, "", false)
		return
	}

	if r.Method == http.MethodPost {
		h.processAddDevice(w, r, email)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// processAddDevice handles the form submission
func (h *WebUIHandler) processAddDevice(w http.ResponseWriter, r *http.Request, email string) {
	if err := r.ParseForm(); err != nil {
		h.renderAddDevice(w, email, "Invalid form data", false)
		return
	}

	deviceName := r.FormValue("device_name")
	platform := r.FormValue("platform")

	if deviceName == "" {
		h.renderAddDevice(w, email, "Device name is required", false)
		return
	}

	if platform == "" {
		h.renderAddDevice(w, email, "Platform is required", false)
		return
	}

	// Generate app password
	plainPassword, err := apppassword.GenerateAppPassword()
	if err != nil {
		log.Printf("Failed to generate app password: %v", err)
		h.renderAddDevice(w, email, "Failed to generate password", false)
		return
	}

	// Store app password
	_, err = h.AppPassStore.Create(email, deviceName, plainPassword)
	if err != nil {
		log.Printf("Failed to create app password: %v", err)
		h.renderAddDevice(w, email, "Failed to save password", false)
		return
	}

	// Generate platform-specific instructions and QR code
	var qrCodeData string
	var profileURL string
	var instructions string

	if platform == "ios" || platform == "macos" {
		// Generate single-use token for .mobileconfig download
		expiry := time.Now().Add(15 * time.Minute)
		token, err := h.TokenStore.Create(email, expiry)
		if err != nil {
			log.Printf("Failed to create token: %v", err)
			h.renderAddDevice(w, email, "Failed to create download token", false)
			return
		}

		profileURL = fmt.Sprintf("https://%s/profile/download?token=%s", h.Config.Hostname, token)

		// Generate QR code
		qrPNG, err := qrcode.GenerateQRCodePNG(profileURL, 256)
		if err != nil {
			log.Printf("Failed to generate QR code: %v", err)
			h.renderAddDevice(w, email, "Failed to generate QR code", false)
			return
		}
		qrCodeData = base64.StdEncoding.EncodeToString(qrPNG)

		instructions = fmt.Sprintf(`
<h3>iOS/macOS Setup</h3>
<ol>
<li>Scan the QR code below with your device camera, OR</li>
<li>Click the "Download Profile" button below</li>
<li>Follow the prompts to install the configuration profile</li>
<li>Your email, calendar, and contacts will sync automatically</li>
</ol>
<p><strong>Token expires:</strong> %s (%s)</p>
<p><a href="%s" class="button">Download Profile (.mobileconfig)</a></p>
`, expiry.Format(time.RFC3339), time.Until(expiry).Round(time.Second), profileURL)

	} else if platform == "android" {
		// For Android, generate QR code with autoconfig URL
		autoconfigURL := fmt.Sprintf("https://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", h.Config.Hostname, email)

		qrPNG, err := qrcode.GenerateQRCodePNG(autoconfigURL, 256)
		if err != nil {
			log.Printf("Failed to generate QR code: %v", err)
			h.renderAddDevice(w, email, "Failed to generate QR code", false)
			return
		}
		qrCodeData = base64.StdEncoding.EncodeToString(qrPNG)
		profileURL = autoconfigURL

		instructions = fmt.Sprintf(`
<h3>Android Setup</h3>
<ol>
<li>Open your email app (Gmail, K-9 Mail, etc.)</li>
<li>Add a new account</li>
<li>Enter your email: <strong>%s</strong></li>
<li>Enter this app password: <strong>%s</strong></li>
<li>Follow the app's configuration wizard</li>
</ol>
<p><strong>Manual Settings:</strong></p>
<ul>
<li>IMAP Server: %s (Port 993, SSL)</li>
<li>SMTP Server: %s (Port 587, STARTTLS)</li>
<li>Username: %s</li>
</ul>
`, email, plainPassword, h.Config.Hostname, h.Config.Hostname, email)

	} else if platform == "thunderbird" || platform == "outlook" {
		instructions = fmt.Sprintf(`
<h3>%s Setup</h3>
<p>Just enter your email address and this app password - %s will auto-discover the settings:</p>
<ol>
<li>Open %s</li>
<li>Add a new email account</li>
<li>Email: <strong>%s</strong></li>
<li>Password: <strong>%s</strong></li>
<li>Click "Continue" - settings will be detected automatically</li>
</ol>
`, platform, platform, platform, email, plainPassword)

	} else {
		// Other/Manual
		instructions = fmt.Sprintf(`
<h3>Manual Setup</h3>
<p>Configure your email client with these settings:</p>
<ul>
<li><strong>Email:</strong> %s</li>
<li><strong>Password:</strong> %s</li>
<li><strong>IMAP Server:</strong> %s</li>
<li><strong>IMAP Port:</strong> 993 (SSL/TLS)</li>
<li><strong>SMTP Server:</strong> %s</li>
<li><strong>SMTP Port:</strong> 587 (STARTTLS)</li>
<li><strong>Username:</strong> %s (full email address)</li>
</ul>
`, email, plainPassword, h.Config.Hostname, h.Config.Hostname, email)
	}

	// Render result page
	data := AddDeviceResultData{
		Email:        email,
		DeviceName:   deviceName,
		AppPassword:  plainPassword,
		Platform:     platform,
		QRCodeData:   qrCodeData,
		ProfileURL:   profileURL,
		Instructions: template.HTML(instructions),
	}

	if err := h.templates.ExecuteTemplate(w, "add_device_result.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleRevokeDevice revokes an app password
func (h *WebUIHandler) HandleRevokeDevice(w http.ResponseWriter, r *http.Request) {
	_, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/devices?error=Invalid+form+data", http.StatusSeeOther)
		return
	}

	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		http.Redirect(w, r, "/devices?error=Missing+device+ID", http.StatusSeeOther)
		return
	}

	// Revoke the app password
	if err := h.AppPassStore.Revoke(deviceID); err != nil {
		log.Printf("Failed to revoke device %s: %v", deviceID, err)
		http.Redirect(w, r, "/devices?error=Failed+to+revoke+device", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/devices?success=Device+revoked+successfully", http.StatusSeeOther)
}

// authenticate checks Basic Auth credentials
func (h *WebUIHandler) authenticate(w http.ResponseWriter, r *http.Request) (string, bool) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Device Management"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", false
	}

	// Verify credentials against app password store
	// The username should be the email address
	if !strings.Contains(username, "@") {
		w.Header().Set("WWW-Authenticate", `Basic realm="Device Management"`)
		http.Error(w, "Unauthorized - use your email as username", http.StatusUnauthorized)
		return "", false
	}

	// Verify password (this will check the main account password, not app passwords)
	// For simplicity in v1, we'll use the admin credentials for the web UI
	// In production, this should verify against the actual mail server
	if username != h.Config.AdminUser || password != h.Config.AdminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="Device Management"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", false
	}

	return username, true
}

// renderDeviceList renders the device list template
func (h *WebUIHandler) renderDeviceList(w http.ResponseWriter, email string, devices []Device, errorMsg, successMsg string) {
	data := DeviceListData{
		Email:   email,
		Devices: devices,
		Error:   errorMsg,
		Success: successMsg,
	}

	if err := h.templates.ExecuteTemplate(w, "device_list.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// renderAddDevice renders the add device template
func (h *WebUIHandler) renderAddDevice(w http.ResponseWriter, email string, errorMsg string, success bool) {
	data := AddDeviceData{
		Email:   email,
		Error:   errorMsg,
		Success: success,
	}

	if err := h.templates.ExecuteTemplate(w, "add_device.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ServeStatic serves static assets
func (h *WebUIHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" || strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}

	content, err := embedFS.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set content type based on extension
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	w.Write(content)
}
