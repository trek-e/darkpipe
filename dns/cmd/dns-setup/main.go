// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/dns/records"
	setupmodule "github.com/darkpipe/darkpipe/dns/setup"
	"github.com/darkpipe/darkpipe/dns/validator"
	"github.com/fatih/color"
)

var (
	// Required flags
	domain        = flag.String("domain", "", "Domain to configure (required)")
	relayHostname = flag.String("relay-hostname", "", "Cloud relay FQDN (required)")
	relayIP       = flag.String("relay-ip", "", "Cloud relay public IP (required)")

	// Mode flags
	apply         = flag.Bool("apply", false, "Actually create DNS records (dry-run by default)")
	validateOnly  = flag.Bool("validate-only", false, "Only validate existing DNS records")
	rotateDKIM    = flag.Bool("rotate-dkim", false, "Rotate DKIM key (generate new selector)")
	jsonOutput    = flag.Bool("json", false, "JSON output mode")
	verboseOutput = flag.Bool("verbose", false, "Verbose output (include per-record apply details)")

	// DKIM flags
	dkimKeyDir        = flag.String("dkim-key-dir", "/etc/darkpipe/dkim", "DKIM key storage location")
	dkimSelectorPrefix = flag.String("dkim-selector-prefix", "darkpipe", "DKIM selector prefix")
	dkimKeyBits       = flag.Int("dkim-key-bits", 2048, "DKIM key size in bits")

	// DMARC flags
	dmarcPolicy = flag.String("dmarc-policy", "none", "DMARC policy (none/quarantine/reject)")
	dmarcRua    = flag.String("dmarc-rua", "", "DMARC aggregate report email")
	dmarcRuf    = flag.String("dmarc-ruf", "", "DMARC forensic report email")

	// Test flags
	sendTest = flag.String("send-test", "", "Send test email to this address after setup")

	// Output flags
	recordsFile = flag.String("records-file", "DNS-RECORDS.md", "Path to save DNS records guide")
)

func main() {
	flag.Parse()

	// Override from environment variables if not set via flags
	loadEnvDefaults()

	ctx := context.Background()

	// Validate-only mode
	if *validateOnly {
		if err := runValidateOnly(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Rotate DKIM mode
	if *rotateDKIM {
		if err := runRotateDKIM(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "DKIM rotation failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default mode: full DNS setup
	if err := runFullSetup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "DNS setup failed: %v\n", err)
		os.Exit(1)
	}
}

func loadEnvDefaults() {
	if *domain == "" {
		*domain = os.Getenv("DARKPIPE_DOMAIN")
	}
	if *relayHostname == "" {
		*relayHostname = os.Getenv("RELAY_HOSTNAME")
	}
	if *relayIP == "" {
		*relayIP = os.Getenv("RELAY_IP")
	}
	if *dkimKeyDir == "/etc/darkpipe/dkim" {
		if dir := os.Getenv("DKIM_KEY_DIR"); dir != "" {
			*dkimKeyDir = dir
		}
	}
	if *dmarcRua == "" {
		*dmarcRua = os.Getenv("DMARC_RUA")
	}
	if *dmarcRuf == "" {
		*dmarcRuf = os.Getenv("DMARC_RUF")
	}
}

func validateRequiredFlags() error {
	if *domain == "" {
		return fmt.Errorf("--domain is required (or set DARKPIPE_DOMAIN)")
	}
	if *relayHostname == "" {
		return fmt.Errorf("--relay-hostname is required (or set RELAY_HOSTNAME)")
	}
	if *relayIP == "" {
		return fmt.Errorf("--relay-ip is required (or set RELAY_IP)")
	}
	return nil
}

func runValidateOnly(ctx context.Context) error {
	if err := validateRequiredFlags(); err != nil {
		return err
	}

	m := setupmodule.New()
	res, err := m.Validate(ctx, setupmodule.ValidateInput{
		Domain:             *domain,
		RelayHostname:      *relayHostname,
		RelayIP:            *relayIP,
		DKIMSelectorPrefix: *dkimSelectorPrefix,
	})
	if err != nil {
		return err
	}

	if !*jsonOutput {
		fmt.Printf("Validating DNS records for %s...\n\n", *domain)
	}

	report := res.Report
	ptrResult := res.PTR
	srvResult := res.SRV
	cnameResult := res.Autodiscover
	allPassed := res.AllPassed

	// Output results
	if *jsonOutput {
		output := map[string]interface{}{
			"domain":     *domain,
			"timestamp":  report.Timestamp.Format(time.RFC3339),
			"all_passed": allPassed,
			"checks": map[string]interface{}{
				"spf":               resultToJSON(findResult(report.Results, "SPF")),
				"dkim":              resultToJSON(findResult(report.Results, "DKIM")),
				"dmarc":             resultToJSON(findResult(report.Results, "DMARC")),
				"mx":                resultToJSON(findResult(report.Results, "MX")),
				"srv":               resultToJSON(srvResult),
				"autodiscover_cname": resultToJSON(cnameResult),
				"ptr":   ptrResultToJSON(ptrResult),
			},
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Human-friendly output
	printValidationResults(report.Results, ptrResult, srvResult, cnameResult)

	if allPassed {
		color.Green("\n✓ All DNS records are configured correctly!\n")
		return nil
	}

	color.Red("\n✗ Some DNS records are missing or incorrect.\n")
	fmt.Println("\nRun without --validate-only to see setup instructions.")
	return fmt.Errorf("validation failed")
}

func runRotateDKIM(ctx context.Context) error {
	if err := validateRequiredFlags(); err != nil {
		return err
	}

	m := setupmodule.New()
	rot, err := m.RotateDKIM(ctx, setupmodule.RotateInput{
		Domain:             *domain,
		DKIMKeyDir:         *dkimKeyDir,
		DKIMSelectorPrefix: *dkimSelectorPrefix,
		DKIMKeyBits:        *dkimKeyBits,
	})
	if err != nil {
		return fmt.Errorf("failed to rotate DKIM key: %w", err)
	}

	fmt.Printf("Rotating DKIM key for %s\n", *domain)
	fmt.Printf("Current selector: %s\n", rot.OldSelector)
	fmt.Printf("New selector: %s\n\n", rot.NewSelector)
	fmt.Printf("✓ Generated new DKIM key pair\n")
	fmt.Printf("  Private key: %s/%s.private.pem\n", *dkimKeyDir, rot.NewSelector)
	fmt.Printf("  Public key: %s/%s.public.pem\n\n", *dkimKeyDir, rot.NewSelector)
	dkimRecord := rot.Record

	fmt.Printf("Add this DKIM TXT record to your DNS:\n\n")
	fmt.Printf("Type: TXT\n")
	fmt.Printf("Name: %s\n", dkimRecord.Domain)
	fmt.Printf("Value: %s\n\n", dkimRecord.Value)

	fmt.Printf("Next steps:\n")
	fmt.Printf("1. Add the new DKIM record to your DNS provider\n")
	fmt.Printf("2. Wait for DNS propagation (5-15 minutes)\n")
	fmt.Printf("3. Update your mail server to sign with selector: %s\n", rot.NewSelector)
	fmt.Printf("4. Wait 7 days for old signatures to expire\n")
	fmt.Printf("5. Remove the old DKIM record: %s._domainkey.%s\n", rot.OldSelector, *domain)

	return nil
}

func runFullSetup(ctx context.Context) error {
	if err := validateRequiredFlags(); err != nil {
		return err
	}

	if !*jsonOutput {
		color.Cyan("DarkPipe DNS Setup\n")
		color.Cyan("==================\n\n")
		fmt.Printf("Domain: %s\n", *domain)
		fmt.Printf("Relay: %s (%s)\n\n", *relayHostname, *relayIP)
	}

	m := setupmodule.New()
	plan, privateKey, err := m.Plan(ctx, setupmodule.PlanInput{
		Domain:             *domain,
		RelayHostname:      *relayHostname,
		RelayIP:            *relayIP,
		DKIMKeyDir:         *dkimKeyDir,
		DKIMSelectorPrefix: *dkimSelectorPrefix,
		DKIMKeyBits:        *dkimKeyBits,
		DMARCPolicy:        *dmarcPolicy,
		DMARCRUA:           *dmarcRua,
		DMARCRUF:           *dmarcRuf,
	})
	if err != nil {
		return fmt.Errorf("failed to build DNS plan: %w", err)
	}
	selector := plan.Selector
	privateKeyPath := filepath.Join(*dkimKeyDir, selector+".private.pem")
	if !*jsonOutput {
		color.Green("✓ DKIM material ready\n")
		fmt.Printf("  Selector: %s\n", selector)
		fmt.Printf("  Private key: %s\n", privateKeyPath)
	}
	allRecords := plan.Records

	// Step 3: Output records
	if !*jsonOutput {
		fmt.Println()
		records.PrintRecords(os.Stdout, allRecords, true)
	}

	// Step 4: Save guide to file
	if !*jsonOutput {
		if err := records.SaveGuide(*recordsFile, allRecords); err != nil {
			color.Yellow("Warning: Failed to save DNS records file: %v\n", err)
		} else {
			fmt.Printf("\n✓ DNS records saved to: %s\n", *recordsFile)
		}
	}

	// Step 5: Apply if requested
	var applyRes *setupmodule.ApplyResult
	if *apply {
		if !*jsonOutput {
			fmt.Println("\nApplying DNS records via detected provider...")
		}
		applyRes, err = m.Apply(ctx, plan)
		if err != nil {
			if _, ok := err.(setupmodule.ErrManualGuideRequired); ok {
				if !*jsonOutput {
					color.Yellow("Automatic apply unavailable for detected provider.")
					fmt.Println("Use generated DNS-RECORDS guide for manual setup.")
				}
			} else {
				return fmt.Errorf("failed to apply DNS records: %w", err)
			}
		} else if !*jsonOutput {
			color.Green("✓ Apply complete")
			fmt.Printf("  Applied: %d\n", applyRes.Applied)
			fmt.Printf("  Skipped: %d\n", applyRes.Skipped)
			fmt.Printf("  Failed:  %d\n", applyRes.Failed)
			if *verboseOutput {
				for _, r := range applyRes.Records {
					fmt.Printf("  - %s %s %s (%s, retryable=%t)\n", r.Action, r.RecordType, r.Name, r.ReasonCode, r.Retryable)
				}
			}
		}
	}

	if *jsonOutput {
		payload := map[string]interface{}{
			"domain":  allRecords.Domain,
			"records": allRecords,
		}
		if applyRes != nil {
			payload["apply"] = applyRes
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	}

	// Step 6: Send test email if requested
	if *sendTest != "" {
		fmt.Printf("\nSending test email to %s...\n", *sendTest)
		if err := m.SendAuthTest(ctx, setupmodule.AuthTestInput{
			Domain:        *domain,
			RelayHostname: *relayHostname,
			To:            *sendTest,
			PrivateKey:    privateKey,
			Selector:      selector,
		}); err != nil {
			color.Red("Failed to send test email: %v\n", err)
		}
	}

	return nil
}

func findResult(results []validator.CheckResult, recordType string) validator.CheckResult {
	for _, r := range results {
		if r.RecordType == recordType {
			return r
		}
	}
	return validator.CheckResult{RecordType: recordType, Pass: false, Error: "not checked"}
}

func resultToJSON(result validator.CheckResult) map[string]interface{} {
	return map[string]interface{}{
		"pass":     result.Pass,
		"expected": result.Expected,
		"actual":   result.Actual,
		"error":    result.Error,
	}
}

func ptrResultToJSON(result validator.PTRResult) map[string]interface{} {
	return map[string]interface{}{
		"pass":          result.Pass,
		"ip":            result.IP,
		"expected":      result.ExpectedHostname,
		"ptr_names":     result.PTRNames,
		"forward_match": result.ForwardMatch,
		"error":         result.Error,
	}
}

func printValidationResults(results []validator.CheckResult, ptrResult validator.PTRResult, srvResult validator.CheckResult, cnameResult validator.CheckResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("Validation Results:")
	fmt.Println()

	for _, result := range results {
		status := red("✗ FAIL")
		if result.Pass {
			status = green("✓ PASS")
		}

		fmt.Printf("%s %s\n", status, result.RecordType)
		fmt.Printf("  Name: %s\n", yellow(result.Name))

		if result.Pass {
			if result.Actual != "" {
				fmt.Printf("  Value: %s\n", truncate(result.Actual, 80))
			}
		} else {
			if result.Error != "" {
				fmt.Printf("  Error: %s\n", red(result.Error))
			}
		}
		fmt.Println()
	}

	// SRV result
	printCheckResult(srvResult, green, red, yellow)

	// Autodiscover CNAME result
	printCheckResult(cnameResult, green, red, yellow)

	// PTR result
	status := red("✗ FAIL")
	if ptrResult.Pass {
		status = green("✓ PASS")
	}

	fmt.Printf("%s PTR (Reverse DNS)\n", status)
	fmt.Printf("  IP: %s\n", yellow(ptrResult.IP))
	fmt.Printf("  Expected: %s\n", yellow(ptrResult.ExpectedHostname))

	if ptrResult.Pass {
		fmt.Printf("  PTR Records: %s\n", strings.Join(ptrResult.PTRNames, ", "))
	} else {
		if ptrResult.Error != "" {
			// Split error message at newlines for better formatting
			errorLines := strings.Split(ptrResult.Error, "\n")
			fmt.Printf("  Error: %s\n", red(errorLines[0]))
			for _, line := range errorLines[1:] {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("         %s\n", line)
				}
			}
		}
	}
}

// printCheckResult is a helper to print a single CheckResult.
func printCheckResult(result validator.CheckResult, green, red, yellow func(...interface{}) string) {
	status := red("✗ FAIL")
	if result.Pass {
		status = green("✓ PASS")
	}

	fmt.Printf("%s %s\n", status, result.RecordType)
	fmt.Printf("  Name: %s\n", yellow(result.Name))

	if result.Pass {
		if result.Actual != "" {
			fmt.Printf("  Value: %s\n", truncate(result.Actual, 80))
		}
	} else {
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", red(result.Error))
		}
	}
	fmt.Println()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
