// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package setup

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/darkpipe/darkpipe/dns/authtest"
	"github.com/darkpipe/darkpipe/dns/dkim"
	"github.com/darkpipe/darkpipe/dns/provider"
	"github.com/darkpipe/darkpipe/dns/records"
	"github.com/darkpipe/darkpipe/dns/validator"
)

type PlanInput struct {
	Domain             string
	RelayHostname      string
	RelayIP            string
	DKIMKeyDir         string
	DKIMSelectorPrefix string
	DKIMKeyBits        int
	DMARCPolicy        string
	DMARCRUA           string
	DMARCRUF           string
}

type ValidateInput struct {
	Domain             string
	RelayHostname      string
	RelayIP            string
	DKIMSelectorPrefix string
}

type RotateInput struct {
	Domain             string
	DKIMKeyDir         string
	DKIMSelectorPrefix string
	DKIMKeyBits        int
}

type AuthTestInput struct {
	Domain        string
	RelayHostname string
	To            string
	PrivateKey    *rsa.PrivateKey
	Selector      string
}

type Plan struct {
	Domain   string
	Selector string
	Records  records.AllRecords
}

type ApplyResult struct {
	Applied int `json:"applied"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
	Records []RecordApplyResult `json:"records"`
}

type ValidateResult struct {
	AllPassed   bool
	Report      validator.ValidationReport
	PTR         validator.PTRResult
	SRV         validator.CheckResult
	Autodiscover validator.CheckResult
}

type RotateResult struct {
	OldSelector string
	NewSelector string
	Record      records.DKIMRecord
}

type ErrManualGuideRequired struct {
	Provider string
	Domain   string
}

func (e ErrManualGuideRequired) Error() string {
	return fmt.Sprintf("manual DNS guide required: provider=%s domain=%s", e.Provider, e.Domain)
}

type Module interface {
	Plan(ctx context.Context, in PlanInput) (*Plan, *rsa.PrivateKey, error)
	Apply(ctx context.Context, plan *Plan) (*ApplyResult, error)
	Validate(ctx context.Context, in ValidateInput) (*ValidateResult, error)
	RotateDKIM(ctx context.Context, in RotateInput) (*RotateResult, error)
	SendAuthTest(ctx context.Context, in AuthTestInput) error
}

type DefaultModule struct{}

func New() Module { return &DefaultModule{} }

func (m *DefaultModule) Plan(ctx context.Context, in PlanInput) (*Plan, *rsa.PrivateKey, error) {
	_ = ctx
	selector := dkim.GetCurrentSelector(in.DKIMSelectorPrefix)
	privateKeyPath := filepath.Join(in.DKIMKeyDir, selector+".private.pem")

	var privateKey *rsa.PrivateKey
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		key, err := dkim.GenerateKeyPair(in.DKIMKeyBits)
		if err != nil {
			return nil, nil, err
		}
		if err := dkim.SaveKeyPair(key, in.DKIMKeyDir, selector); err != nil {
			return nil, nil, err
		}
		privateKey = key
	} else {
		key, err := dkim.LoadPrivateKey(privateKeyPath)
		if err != nil {
			return nil, nil, err
		}
		privateKey = key
	}

	publicKeyBase64, err := dkim.PublicKeyBase64(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	allRecords := records.AllRecords{
		Domain: in.Domain,
		SPF:    records.GenerateSPF(in.Domain, in.RelayIP, nil),
		DKIM:   records.GenerateDKIMRecord(in.Domain, selector, publicKeyBase64),
		DMARC: records.GenerateDMARC(in.Domain, records.DMARCOptions{
			Policy: in.DMARCPolicy,
			RUA:    in.DMARCRUA,
			RUF:    in.DMARCRUF,
		}),
		MX:                 records.GenerateMX(in.Domain, in.RelayHostname, 10),
		SRV:                records.GenerateSRVRecords(in.Domain, in.RelayHostname),
		AutodiscoverCNAMEs: records.GenerateAutodiscoverCNAME(in.Domain, in.RelayHostname),
	}

	return &Plan{Domain: in.Domain, Selector: selector, Records: allRecords}, privateKey, nil
}

func (m *DefaultModule) Apply(ctx context.Context, plan *Plan) (*ApplyResult, error) {
	adapter, err := NewProviderAdapter(ctx, plan.Domain)
	if err != nil {
		return nil, err
	}

	result := &ApplyResult{Records: []RecordApplyResult{}}
	for _, rec := range flattenRecords(plan) {
		ar := adapter.ApplyRecord(ctx, rec)
		result.Records = append(result.Records, ar)
		switch ar.Action {
		case "create", "update":
			result.Applied++
		case "skip":
			result.Skipped++
		default:
			result.Failed++
		}
	}
	return result, nil
}

func (m *DefaultModule) Validate(ctx context.Context, in ValidateInput) (*ValidateResult, error) {
	selector := dkim.GetCurrentSelector(in.DKIMSelectorPrefix)
	checker := validator.NewChecker(nil)
	report := checker.CheckAll(ctx, in.Domain, in.RelayIP, in.RelayHostname, selector)
	ptrResult := validator.CheckPTR(ctx, in.RelayIP, in.RelayHostname)
	srvResult := checker.CheckSRV(ctx, in.Domain)
	cnameResult := checker.CheckAutodiscoverCNAMEs(ctx, in.Domain, in.RelayHostname)

	return &ValidateResult{
		AllPassed:    report.AllPassed && ptrResult.Pass && srvResult.Pass && cnameResult.Pass,
		Report:       report,
		PTR:          ptrResult,
		SRV:          srvResult,
		Autodiscover: cnameResult,
	}, nil
}

func (m *DefaultModule) RotateDKIM(ctx context.Context, in RotateInput) (*RotateResult, error) {
	_ = ctx
	newSelector := dkim.GetNextSelector(in.DKIMSelectorPrefix)
	privateKey, err := dkim.GenerateKeyPair(in.DKIMKeyBits)
	if err != nil {
		return nil, err
	}
	if err := dkim.SaveKeyPair(privateKey, in.DKIMKeyDir, newSelector); err != nil {
		return nil, err
	}
	publicKeyBase64, err := dkim.PublicKeyBase64(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	rec := records.GenerateDKIMRecord(in.Domain, newSelector, publicKeyBase64)
	return &RotateResult{
		OldSelector: dkim.GetCurrentSelector(in.DKIMSelectorPrefix),
		NewSelector: newSelector,
		Record:      rec,
	}, nil
}

func (m *DefaultModule) SendAuthTest(ctx context.Context, in AuthTestInput) error {
	testCfg := authtest.TestEmailConfig{
		From:         fmt.Sprintf("test@%s", in.Domain),
		To:           in.To,
		RelayHost:    in.RelayHostname,
		RelayPort:    25,
		DKIMKey:      in.PrivateKey,
		DKIMDomain:   in.Domain,
		DKIMSelector: in.Selector,
	}
	return authtest.SendTestEmail(ctx, testCfg)
}

func flattenRecords(plan *Plan) []provider.Record {
	ttl := 300
	recs := []provider.Record{
		{Type: "TXT", Name: plan.Records.SPF.Domain, Content: plan.Records.SPF.Value, TTL: ttl},
		{Type: "TXT", Name: plan.Records.DKIM.Domain, Content: plan.Records.DKIM.Value, TTL: ttl},
		{Type: "TXT", Name: plan.Records.DMARC.Domain, Content: plan.Records.DMARC.Value, TTL: ttl},
	}
	mxP := plan.Records.MX.Priority
	recs = append(recs, provider.Record{Type: "MX", Name: plan.Records.MX.Domain, Content: plan.Records.MX.Hostname, TTL: ttl, Priority: &mxP})
	for _, c := range plan.Records.AutodiscoverCNAMEs {
		recs = append(recs, provider.Record{Type: "CNAME", Name: c.Name, Content: c.Target, TTL: ttl})
	}
	for _, s := range plan.Records.SRV {
		content := fmt.Sprintf("%d %d %d %s", s.Priority, s.Weight, s.Port, s.Target)
		name := fmt.Sprintf("%s.%s.%s", s.Service, s.Proto, s.Domain)
		recs = append(recs, provider.Record{Type: "SRV", Name: name, Content: content, TTL: ttl})
	}
	return recs
}


