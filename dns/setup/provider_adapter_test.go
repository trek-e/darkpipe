package setup

import (
	"context"
	"errors"
	"testing"

	"github.com/darkpipe/darkpipe/dns/provider"
)

type mockDNSProvider struct {
	list   []provider.Record
	errList error
	errCreate error
	errUpdate error
}

func (m *mockDNSProvider) CreateRecord(ctx context.Context, rec provider.Record) error { return m.errCreate }
func (m *mockDNSProvider) UpdateRecord(ctx context.Context, recordID string, rec provider.Record) error { return m.errUpdate }
func (m *mockDNSProvider) ListRecords(ctx context.Context, filter provider.RecordFilter) ([]provider.Record, error) {
	if m.errList != nil { return nil, m.errList }
	return m.list, nil
}
func (m *mockDNSProvider) DeleteRecord(ctx context.Context, recordID string) error { return nil }
func (m *mockDNSProvider) GetZoneID(ctx context.Context, domain string) (string, error) { return "z", nil }
func (m *mockDNSProvider) Name() string { return "mock" }

func newTestAdapter(inner provider.DNSProvider) *ProviderAdapter {
	return &ProviderAdapter{
		inner: inner,
		ctx: ProviderContext{Domain: "example.com", ZoneID: "z"},
		caps: ProviderCapabilities{
			CanCreate: true, CanUpdate: true, CanList: true,
			Types: map[string]bool{"TXT": true, "MX": true, "CNAME": true, "A": true, "SRV": false},
		},
	}
}

func TestApplyRecord_UnsupportedType(t *testing.T) {
	a := newTestAdapter(&mockDNSProvider{})
	res := a.ApplyRecord(context.Background(), provider.Record{Type: "SRV", Name: "_imaps._tcp.example.com"})
	if res.Action != "skip" || res.ReasonCode != "unsupported_record_type" || res.Retryable {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestApplyRecord_ListFailureRetryable(t *testing.T) {
	a := newTestAdapter(&mockDNSProvider{errList: errors.New("dns down")})
	res := a.ApplyRecord(context.Background(), provider.Record{Type: "TXT", Name: "example.com", Content: "v=spf1"})
	if res.Action != "fail" || res.ReasonCode != "list_failed" || !res.Retryable {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestApplyRecord_CreatePath(t *testing.T) {
	a := newTestAdapter(&mockDNSProvider{list: []provider.Record{}})
	res := a.ApplyRecord(context.Background(), provider.Record{Type: "TXT", Name: "example.com", Content: "v=spf1"})
	if res.Action != "create" || res.ReasonCode != "created" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestApplyRecord_UpdateFailureRetryable(t *testing.T) {
	a := newTestAdapter(&mockDNSProvider{list: []provider.Record{{ID: "1", Type: "TXT", Name: "example.com", Content: "old"}}, errUpdate: errors.New("denied")})
	res := a.ApplyRecord(context.Background(), provider.Record{Type: "TXT", Name: "example.com", Content: "new"})
	if res.Action != "fail" || res.ReasonCode != "update_failed" || !res.Retryable {
		t.Fatalf("unexpected result: %+v", res)
	}
}
