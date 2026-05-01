// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package cert

import (
	"fmt"
)

// CertInspector is seam for certificate state inspection.
type CertInspector interface {
	CheckCert(certPath string) (*CertInfo, error)
}

// Renewer is seam for renewal execution.
type Renewer interface {
	Renew(config RotatorConfig) error
}

// Reloader is seam for service reload execution.
type Reloader interface {
	ReloadServices() error
}

// RetryRenewer adapts existing RenewWithRetry fn to Renewer seam.
type RetryRenewer struct{}

func (RetryRenewer) Renew(config RotatorConfig) error { return RenewWithRetry(config) }

// LifecycleResult reports lifecycle step outcomes.
type LifecycleResult struct {
	Checked bool
	Renewed bool
	Reloaded bool
}

// LifecycleManager orchestrates check -> renew -> reload.
type LifecycleManager struct {
	watcher CertInspector
	renewer Renewer
	reloader Reloader
}

func NewLifecycleManager(watcher CertInspector, renewer Renewer, reloader Reloader) *LifecycleManager {
	return &LifecycleManager{watcher: watcher, renewer: renewer, reloader: reloader}
}

func (m *LifecycleManager) Run(certPath string, cfg RotatorConfig) (*LifecycleResult, error) {
	info, err := m.watcher.CheckCert(certPath)
	if err != nil {
		return nil, fmt.Errorf("check cert: %w", err)
	}
	res := &LifecycleResult{Checked: true}
	if !info.ShouldRenew {
		return res, nil
	}
	if err := m.renewer.Renew(cfg); err != nil {
		return res, fmt.Errorf("renew cert: %w", err)
	}
	res.Renewed = true
	if err := m.reloader.ReloadServices(); err != nil {
		return res, fmt.Errorf("reload services: %w", err)
	}
	res.Reloaded = true
	return res, nil
}
