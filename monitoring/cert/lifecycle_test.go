package cert

import (
	"errors"
	"testing"
)

type fakeWatcher struct {
	info *CertInfo
	err  error
}

func (f fakeWatcher) CheckCert(string) (*CertInfo, error) { return f.info, f.err }

type fakeRenewer struct{ err error }

func (f fakeRenewer) Renew(RotatorConfig) error { return f.err }

type fakeReloader struct{ err error }

func (f fakeReloader) ReloadServices() error { return f.err }

func TestLifecycle_NoRenew(t *testing.T) {
	m := NewLifecycleManager(
		fakeWatcher{info: &CertInfo{ShouldRenew: false}},
		fakeRenewer{},
		fakeReloader{},
	)
	res, err := m.Run("/tmp/cert.pem", RotatorConfig{})
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if !res.Checked || res.Renewed || res.Reloaded { t.Fatalf("unexpected result: %+v", res) }
}

func TestLifecycle_RenewAndReload(t *testing.T) {
	m := NewLifecycleManager(
		fakeWatcher{info: &CertInfo{ShouldRenew: true}},
		fakeRenewer{},
		fakeReloader{},
	)
	res, err := m.Run("/tmp/cert.pem", RotatorConfig{})
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if !res.Checked || !res.Renewed || !res.Reloaded { t.Fatalf("unexpected result: %+v", res) }
}

func TestLifecycle_RenewError(t *testing.T) {
	m := NewLifecycleManager(
		fakeWatcher{info: &CertInfo{ShouldRenew: true}},
		fakeRenewer{err: errors.New("boom")},
		fakeReloader{},
	)
	res, err := m.Run("/tmp/cert.pem", RotatorConfig{})
	if err == nil { t.Fatal("expected err") }
	if !res.Checked || res.Renewed || res.Reloaded { t.Fatalf("unexpected result: %+v", res) }
}

func TestLifecycle_ReloadError(t *testing.T) {
	m := NewLifecycleManager(
		fakeWatcher{info: &CertInfo{ShouldRenew: true}},
		fakeRenewer{},
		fakeReloader{err: errors.New("reload fail")},
	)
	res, err := m.Run("/tmp/cert.pem", RotatorConfig{})
	if err == nil { t.Fatal("expected err") }
	if !res.Checked || !res.Renewed || res.Reloaded { t.Fatalf("unexpected result: %+v", res) }
}
