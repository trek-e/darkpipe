package validator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// mockDNSServer creates a test DNS server that responds to queries.
type mockDNSServer struct {
	server   *dns.Server
	addr     string
	handlers map[string]map[uint16]func(*dns.Msg) *dns.Msg // domain -> qtype -> handler
}

func newMockDNSServer() *mockDNSServer {
	m := &mockDNSServer{
		handlers: make(map[string]map[uint16]func(*dns.Msg) *dns.Msg),
	}

	// Create a custom handler function for this specific server instance
	handler := func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)

		// Find handler for this question
		if len(r.Question) > 0 {
			qname := r.Question[0].Name
			qtype := r.Question[0].Qtype

			if typeHandlers, ok := m.handlers[qname]; ok {
				if h, ok := typeHandlers[qtype]; ok {
					response := h(r)
					w.WriteMsg(response)
					return
				}
			}
		}

		// Default: NXDOMAIN
		msg.Rcode = dns.RcodeNameError
		w.WriteMsg(msg)
	}

	mux := dns.NewServeMux()
	mux.HandleFunc(".", handler)

	m.server = &dns.Server{
		Addr:    "127.0.0.1:0",
		Net:     "udp",
		Handler: mux,
	}
	return m
}

func (m *mockDNSServer) start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := m.server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	// Wait for server to start (with timeout)
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-errCh:
			return err
		case <-timeout:
			return fmt.Errorf("timeout waiting for DNS server to start")
		case <-ticker.C:
			if m.server.PacketConn != nil {
				m.addr = m.server.PacketConn.LocalAddr().String()
				return nil
			}
		}
	}
}

func (m *mockDNSServer) stop() {
	if m.server != nil {
		m.server.Shutdown()
	}
}

func (m *mockDNSServer) addTXTHandler(domain string, records []string) {
	fqdn := dns.Fqdn(domain)
	if m.handlers[fqdn] == nil {
		m.handlers[fqdn] = make(map[uint16]func(*dns.Msg) *dns.Msg)
	}

	m.handlers[fqdn][dns.TypeTXT] = func(r *dns.Msg) *dns.Msg {
		msg := new(dns.Msg)
		msg.SetReply(r)

		for _, record := range records {
			txt := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   fqdn,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Txt: []string{record},
			}
			msg.Answer = append(msg.Answer, txt)
		}
		return msg
	}
}

func (m *mockDNSServer) addMXHandler(domain string, records []struct{ priority uint16; hostname string }) {
	fqdn := dns.Fqdn(domain)
	if m.handlers[fqdn] == nil {
		m.handlers[fqdn] = make(map[uint16]func(*dns.Msg) *dns.Msg)
	}

	m.handlers[fqdn][dns.TypeMX] = func(r *dns.Msg) *dns.Msg {
		msg := new(dns.Msg)
		msg.SetReply(r)

		for _, record := range records {
			mx := &dns.MX{
				Hdr: dns.RR_Header{
					Name:   fqdn,
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Preference: record.priority,
				Mx:         dns.Fqdn(record.hostname),
			}
			msg.Answer = append(msg.Answer, mx)
		}
		return msg
	}
}

func TestCheckSPF_Pass(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("example.com", []string{"v=spf1 ip4:1.2.3.4 -all"})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckSPF(context.Background(), "example.com", "1.2.3.4")

	if !result.Pass {
		t.Errorf("Expected SPF check to pass, got: %s", result.Error)
	}
	if result.RecordType != "SPF" {
		t.Errorf("Expected RecordType=SPF, got %s", result.RecordType)
	}
	if result.Expected != "ip4:1.2.3.4" {
		t.Errorf("Expected ip4:1.2.3.4, got %s", result.Expected)
	}
}

func TestCheckSPF_NoRecord(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("example.com", []string{}) // No records

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckSPF(context.Background(), "example.com", "1.2.3.4")

	if result.Pass {
		t.Error("Expected SPF check to fail when no record exists")
	}
	if result.Error != "No SPF record found" {
		t.Errorf("Expected 'No SPF record found' error, got: %s", result.Error)
	}
}

func TestCheckSPF_MultipleSPFRecords(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	// Multiple SPF records - RFC 7208 violation
	mock.addTXTHandler("example.com", []string{
		"v=spf1 ip4:1.2.3.4 -all",
		"v=spf1 ip4:5.6.7.8 -all",
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckSPF(context.Background(), "example.com", "1.2.3.4")

	if result.Pass {
		t.Error("Expected SPF check to fail when multiple SPF records exist")
	}
	if !contains(result.Error, "Multiple SPF records") {
		t.Errorf("Expected 'Multiple SPF records' error, got: %s", result.Error)
	}
}

func TestCheckSPF_WrongIP(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("example.com", []string{"v=spf1 ip4:5.6.7.8 -all"})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckSPF(context.Background(), "example.com", "1.2.3.4")

	if result.Pass {
		t.Error("Expected SPF check to fail when IP doesn't match")
	}
	if !contains(result.Error, "does not contain ip4:1.2.3.4") {
		t.Errorf("Expected IP mismatch error, got: %s", result.Error)
	}
}

func TestCheckDKIM_Pass(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("darkpipe-2026q1._domainkey.example.com", []string{
		"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDKIM(context.Background(), "example.com", "darkpipe-2026q1")

	if !result.Pass {
		t.Errorf("Expected DKIM check to pass, got: %s", result.Error)
	}
	if result.RecordType != "DKIM" {
		t.Errorf("Expected RecordType=DKIM, got %s", result.RecordType)
	}
}

func TestCheckDKIM_NotFound(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("darkpipe-2026q1._domainkey.example.com", []string{}) // No record

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDKIM(context.Background(), "example.com", "darkpipe-2026q1")

	if result.Pass {
		t.Error("Expected DKIM check to fail when selector not found")
	}
	if result.Error != "No DKIM record found" {
		t.Errorf("Expected 'No DKIM record found' error, got: %s", result.Error)
	}
}

func TestCheckDKIM_MissingVersion(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("darkpipe-2026q1._domainkey.example.com", []string{
		"k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDKIM(context.Background(), "example.com", "darkpipe-2026q1")

	if result.Pass {
		t.Error("Expected DKIM check to fail when v=DKIM1 is missing")
	}
	if !contains(result.Error, "does not start with v=DKIM1") {
		t.Errorf("Expected version error, got: %s", result.Error)
	}
}

func TestCheckDKIM_EmptyPublicKey(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("darkpipe-2026q1._domainkey.example.com", []string{
		"v=DKIM1; k=rsa; p=",
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDKIM(context.Background(), "example.com", "darkpipe-2026q1")

	if result.Pass {
		t.Error("Expected DKIM check to fail when p= is empty")
	}
	if !contains(result.Error, "p= value is empty") {
		t.Errorf("Expected empty p= error, got: %s", result.Error)
	}
}

func TestCheckDMARC_Pass(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("_dmarc.example.com", []string{
		"v=DMARC1; p=none; sp=quarantine; rua=mailto:dmarc@example.com",
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDMARC(context.Background(), "example.com")

	if !result.Pass {
		t.Errorf("Expected DMARC check to pass, got: %s", result.Error)
	}
	if result.RecordType != "DMARC" {
		t.Errorf("Expected RecordType=DMARC, got %s", result.RecordType)
	}
	if !contains(result.Actual, "policy: none") {
		t.Errorf("Expected policy to be reported in Actual, got: %s", result.Actual)
	}
}

func TestCheckDMARC_NotFound(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addTXTHandler("_dmarc.example.com", []string{}) // No record

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckDMARC(context.Background(), "example.com")

	if result.Pass {
		t.Error("Expected DMARC check to fail when no record exists")
	}
	if result.Error != "No DMARC record found" {
		t.Errorf("Expected 'No DMARC record found' error, got: %s", result.Error)
	}
}

func TestCheckMX_Pass(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addMXHandler("example.com", []struct{ priority uint16; hostname string }{
		{10, "mail.example.com"},
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckMX(context.Background(), "example.com", "mail.example.com")

	if !result.Pass {
		t.Errorf("Expected MX check to pass, got: %s", result.Error)
	}
	if result.RecordType != "MX" {
		t.Errorf("Expected RecordType=MX, got %s", result.RecordType)
	}
}

func TestCheckMX_NotFound(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addMXHandler("example.com", []struct{ priority uint16; hostname string }{}) // No records

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckMX(context.Background(), "example.com", "mail.example.com")

	if result.Pass {
		t.Error("Expected MX check to fail when no records exist")
	}
	if result.Error != "No MX records found" {
		t.Errorf("Expected 'No MX records found' error, got: %s", result.Error)
	}
}

func TestCheckMX_WrongHostname(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	mock.addMXHandler("example.com", []struct{ priority uint16; hostname string }{
		{10, "other.example.com"},
	})

	checker := NewChecker([]string{mock.addr})
	result := checker.CheckMX(context.Background(), "example.com", "mail.example.com")

	if result.Pass {
		t.Error("Expected MX check to fail when hostname doesn't match")
	}
	if !contains(result.Error, "not found in MX records") {
		t.Errorf("Expected hostname mismatch error, got: %s", result.Error)
	}
}

func TestCheckAll(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	// Set up all records
	mock.addTXTHandler("example.com", []string{"v=spf1 ip4:1.2.3.4 -all"})
	mock.addTXTHandler("darkpipe-2026q1._domainkey.example.com", []string{
		"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
	})
	mock.addTXTHandler("_dmarc.example.com", []string{
		"v=DMARC1; p=none; sp=quarantine",
	})
	mock.addMXHandler("example.com", []struct{ priority uint16; hostname string }{
		{10, "mail.example.com"},
	})

	checker := NewChecker([]string{mock.addr})
	report := checker.CheckAll(context.Background(), "example.com", "1.2.3.4", "mail.example.com", "darkpipe-2026q1")

	if len(report.Results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(report.Results))
	}

	if !report.AllPassed {
		t.Error("Expected all checks to pass")
		for _, result := range report.Results {
			if !result.Pass {
				t.Logf("Failed check: %s - %s", result.RecordType, result.Error)
			}
		}
	}

	// Verify all record types are present
	types := make(map[string]bool)
	for _, result := range report.Results {
		types[result.RecordType] = true
	}

	expected := []string{"SPF", "DKIM", "DMARC", "MX"}
	for _, recordType := range expected {
		if !types[recordType] {
			t.Errorf("Missing %s check in results", recordType)
		}
	}
}

func TestCheckAll_SomeFailures(t *testing.T) {
	mock := newMockDNSServer()
	defer mock.stop()
	if err := mock.start(); err != nil {
		t.Fatalf("Failed to start mock DNS server: %v", err)
	}

	// Set up only some records (SPF and MX pass, DKIM and DMARC fail)
	mock.addTXTHandler("example.com", []string{"v=spf1 ip4:1.2.3.4 -all"})
	mock.addMXHandler("example.com", []struct{ priority uint16; hostname string }{
		{10, "mail.example.com"},
	})

	checker := NewChecker([]string{mock.addr})
	report := checker.CheckAll(context.Background(), "example.com", "1.2.3.4", "mail.example.com", "darkpipe-2026q1")

	if report.AllPassed {
		t.Error("Expected some checks to fail")
	}

	passCount := 0
	failCount := 0
	for _, result := range report.Results {
		if result.Pass {
			passCount++
		} else {
			failCount++
		}
	}

	if passCount != 2 {
		t.Errorf("Expected 2 passing checks (SPF, MX), got %d", passCount)
	}
	if failCount != 2 {
		t.Errorf("Expected 2 failing checks (DKIM, DMARC), got %d", failCount)
	}
}

func TestNewChecker_DefaultServers(t *testing.T) {
	checker := NewChecker(nil)
	if len(checker.servers) != 2 {
		t.Errorf("Expected 2 default servers, got %d", len(checker.servers))
	}
	if checker.servers[0] != "8.8.8.8:53" {
		t.Errorf("Expected first server to be 8.8.8.8:53, got %s", checker.servers[0])
	}
	if checker.servers[1] != "1.1.1.1:53" {
		t.Errorf("Expected second server to be 1.1.1.1:53, got %s", checker.servers[1])
	}
}

func TestNewChecker_CustomServers(t *testing.T) {
	customServers := []string{"9.9.9.9:53"}
	checker := NewChecker(customServers)
	if len(checker.servers) != 1 {
		t.Errorf("Expected 1 custom server, got %d", len(checker.servers))
	}
	if checker.servers[0] != "9.9.9.9:53" {
		t.Errorf("Expected server to be 9.9.9.9:53, got %s", checker.servers[0])
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || fmt.Sprintf("%s", s)[0:min(len(s), len(substr))] == substr || stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
