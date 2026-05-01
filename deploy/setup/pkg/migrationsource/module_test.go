package migrationsource

import (
	"context"
	"testing"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
)

func TestCapabilities(t *testing.T) {
	m := New()
	p := &providers.GenericProvider{CalDAVURL: "https://cal.example.com", CardDAVURL: "https://card.example.com"}
	caps := m.Capabilities(p)
	if !caps[CapabilityCalDAV] || !caps[CapabilityCardDAV] {
		t.Fatalf("expected caldav/carddav capabilities true, got %#v", caps)
	}
}

func TestDiscoverMetadata(t *testing.T) {
	m := New()
	p := &providers.GenericProvider{IMAPHost: "imap.example.com", IMAPPort: 993, UseTLS: true}
	meta, err := m.DiscoverMetadata(context.Background(), p)
	if err != nil {
		t.Fatalf("DiscoverMetadata error: %v", err)
	}
	if meta.ProviderSlug != "generic" {
		t.Fatalf("expected generic slug, got %s", meta.ProviderSlug)
	}
	if meta.Endpoints["imap"] == "" {
		t.Fatalf("expected imap endpoint")
	}
}

func TestClassifyError(t *testing.T) {
	err := classifyError(assertErr("login failed"))
	if _, ok := err.(ErrAuthFailed); !ok {
		t.Fatalf("expected ErrAuthFailed, got %T", err)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
