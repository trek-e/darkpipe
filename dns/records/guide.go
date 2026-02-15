// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package records

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// AllRecords holds all DNS records for a domain.
type AllRecords struct {
	Domain             string
	SPF                SPFRecord
	DKIM               DKIMRecord
	DMARC              DMARCRecord
	MX                 MXRecord
	SRV                []SRVRecord
	AutodiscoverCNAMEs []DNSRecord
}

// GenerateGuide creates a markdown guide for manual DNS setup.
// Includes record values, explanations, provider documentation links, and verification instructions.
func GenerateGuide(records AllRecords) string {
	var sb strings.Builder

	sb.WriteString("# DNS Records for ")
	sb.WriteString(records.Domain)
	sb.WriteString("\n\n")

	sb.WriteString("This guide contains all DNS records needed for email authentication and delivery.\n")
	sb.WriteString("Copy these values into your DNS provider's control panel.\n\n")

	sb.WriteString("---\n\n")

	// SPF Record
	sb.WriteString("## SPF Record (Sender Policy Framework)\n\n")
	sb.WriteString("**What it does:** Declares which IP addresses are authorized to send email from your domain. ")
	sb.WriteString("Prevents spammers from spoofing your domain.\n\n")
	sb.WriteString("**Record Type:** `TXT`\n\n")
	sb.WriteString("**Name/Host:** `@` (or your domain)\n\n")
	sb.WriteString("**Value:**\n```\n")
	sb.WriteString(records.SPF.Value)
	sb.WriteString("\n```\n\n")

	// DKIM Record
	sb.WriteString("## DKIM Record (DomainKeys Identified Mail)\n\n")
	sb.WriteString("**What it does:** Cryptographically signs your emails so recipients can verify they weren't modified in transit. ")
	sb.WriteString("Essential for deliverability at Gmail, Outlook, Yahoo.\n\n")
	sb.WriteString("**Record Type:** `TXT`\n\n")
	sb.WriteString("**Name/Host:** `")
	sb.WriteString(records.DKIM.Domain)
	sb.WriteString("`\n\n")
	sb.WriteString("**Value:**\n```\n")
	sb.WriteString(records.DKIM.Value)
	sb.WriteString("\n```\n\n")

	// DMARC Record
	sb.WriteString("## DMARC Record (Domain-based Message Authentication)\n\n")
	sb.WriteString("**What it does:** Tells receiving mail servers what to do if SPF or DKIM checks fail. ")
	sb.WriteString("Protects your domain from phishing and spoofing.\n\n")
	sb.WriteString("**Record Type:** `TXT`\n\n")
	sb.WriteString("**Name/Host:** `")
	sb.WriteString(records.DMARC.Domain)
	sb.WriteString("`\n\n")
	sb.WriteString("**Value:**\n```\n")
	sb.WriteString(records.DMARC.Value)
	sb.WriteString("\n```\n\n")
	sb.WriteString("**Policy Progression:** Start with `p=none` to monitor, then move to `p=quarantine` after verifying reports, ")
	sb.WriteString("and finally `p=reject` for full enforcement.\n\n")

	// MX Record
	sb.WriteString("## MX Record (Mail Exchange)\n\n")
	sb.WriteString("**What it does:** Directs incoming email to your cloud relay server.\n\n")
	sb.WriteString("**Record Type:** `MX`\n\n")
	sb.WriteString("**Name/Host:** `@` (or your domain)\n\n")
	sb.WriteString("**Priority:** `")
	sb.WriteString(fmt.Sprintf("%d", records.MX.Priority))
	sb.WriteString("`\n\n")
	sb.WriteString("**Value/Hostname:** `")
	sb.WriteString(records.MX.Hostname)
	sb.WriteString("`\n\n")

	// SRV Records (RFC 6186)
	if len(records.SRV) > 0 {
		sb.WriteString("## SRV Records (Email Autodiscovery - RFC 6186)\n\n")
		sb.WriteString("**What they do:** Enable automatic email client configuration for IMAP and SMTP. ")
		sb.WriteString("Thunderbird, Apple Mail, and other clients use these to discover server settings.\n\n")
		sb.WriteString("**Record Type:** `SRV`\n\n")
		sb.WriteString("**Add these records:**\n```\n")
		for _, srv := range records.SRV {
			sb.WriteString(srv.String())
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	// Autodiscover CNAMEs
	if len(records.AutodiscoverCNAMEs) > 0 {
		sb.WriteString("## Autodiscover CNAME Records\n\n")
		sb.WriteString("**What they do:** Enable automatic configuration for Thunderbird (autoconfig) and Outlook (autodiscover).\n\n")
		sb.WriteString("**Record Type:** `CNAME`\n\n")
		sb.WriteString("**Add these records:**\n```\n")
		for _, cname := range records.AutodiscoverCNAMEs {
			sb.WriteString(cname.String())
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("---\n\n")

	// Provider Documentation Links
	sb.WriteString("## How to Add DNS Records (By Provider)\n\n")
	sb.WriteString("- **Cloudflare:** https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/\n")
	sb.WriteString("- **Route53 (AWS):** https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-creating.html\n")
	sb.WriteString("- **GoDaddy:** https://www.godaddy.com/help/add-a-txt-record-19232\n")
	sb.WriteString("- **Namecheap:** https://www.namecheap.com/support/knowledgebase/article.aspx/317/2237/how-do-i-add-txtspfdkimdmarc-records-for-my-domain/\n")
	sb.WriteString("- **Google Domains:** https://support.google.com/domains/answer/3290350\n\n")

	// Verification Section
	sb.WriteString("---\n\n")
	sb.WriteString("## Verification\n\n")
	sb.WriteString("After adding these records, verify they're working:\n\n")
	sb.WriteString("### Built-in Validator\n")
	sb.WriteString("```bash\n")
	sb.WriteString("darkpipe dns-setup --validate-only\n")
	sb.WriteString("```\n\n")
	sb.WriteString("### External Tools\n")
	sb.WriteString("- **MXToolbox:** https://mxtoolbox.com/SuperTool.aspx (check SPF, DKIM, DMARC, MX)\n")
	sb.WriteString("- **mail-tester.com:** https://www.mail-tester.com/ (send test email, get deliverability score)\n")
	sb.WriteString("- **Google Admin Toolbox:** https://toolbox.googleapps.com/apps/checkmx/ (MX record checker)\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("**Note:** DNS propagation can take 5-15 minutes. ")
	sb.WriteString("If verification fails immediately after adding records, wait a few minutes and try again.\n")

	return sb.String()
}

// PrintRecords prints DNS records to the terminal with color formatting.
func PrintRecords(w io.Writer, records AllRecords, useColor bool) {
	if !useColor {
		// Disable color output
		color.NoColor = true
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Fprintf(w, "\n%s\n\n", bold("DNS Records Setup Checklist"))

	// SPF
	fmt.Fprintf(w, "%s SPF Record\n", green("✓"))
	fmt.Fprintf(w, "  Type:  %s\n", cyan("TXT"))
	fmt.Fprintf(w, "  Name:  %s\n", cyan("@"))
	fmt.Fprintf(w, "  Value: %s\n\n", yellow(records.SPF.Value))

	// DKIM
	fmt.Fprintf(w, "%s DKIM Record\n", green("✓"))
	fmt.Fprintf(w, "  Type:  %s\n", cyan("TXT"))
	fmt.Fprintf(w, "  Name:  %s\n", cyan(records.DKIM.Domain))
	fmt.Fprintf(w, "  Value: %s\n\n", yellow(truncate(records.DKIM.Value, 80)))

	// DMARC
	fmt.Fprintf(w, "%s DMARC Record\n", green("✓"))
	fmt.Fprintf(w, "  Type:  %s\n", cyan("TXT"))
	fmt.Fprintf(w, "  Name:  %s\n", cyan(records.DMARC.Domain))
	fmt.Fprintf(w, "  Value: %s\n\n", yellow(records.DMARC.Value))

	// MX
	fmt.Fprintf(w, "%s MX Record\n", green("✓"))
	fmt.Fprintf(w, "  Type:     %s\n", cyan("MX"))
	fmt.Fprintf(w, "  Name:     %s\n", cyan("@"))
	fmt.Fprintf(w, "  Priority: %s\n", cyan(fmt.Sprintf("%d", records.MX.Priority)))
	fmt.Fprintf(w, "  Value:    %s\n\n", yellow(records.MX.Hostname))

	// SRV Records (RFC 6186)
	if len(records.SRV) > 0 {
		fmt.Fprintf(w, "%s SRV Records (Email Autodiscovery - RFC 6186)\n", green("✓"))
		for _, srv := range records.SRV {
			fmt.Fprintf(w, "  %s\n", yellow(srv.String()))
		}
		fmt.Fprintf(w, "\n")
	}

	// Autodiscover CNAMEs
	if len(records.AutodiscoverCNAMEs) > 0 {
		fmt.Fprintf(w, "%s Autodiscover CNAME Records\n", green("✓"))
		for _, cname := range records.AutodiscoverCNAMEs {
			fmt.Fprintf(w, "  %s\n", yellow(cname.String()))
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "%s\n", bold("Next Steps:"))
	fmt.Fprintf(w, "1. Add these records to your DNS provider\n")
	fmt.Fprintf(w, "2. Wait 5-15 minutes for DNS propagation\n")
	fmt.Fprintf(w, "3. Run: %s\n\n", cyan("darkpipe dns-setup --validate-only"))
}

// PrintJSON outputs records in JSON format for machine parsing.
func PrintJSON(w io.Writer, records AllRecords) error {
	recordsMap := map[string]interface{}{
		"spf": map[string]interface{}{
			"type":  "TXT",
			"name":  "@",
			"value": records.SPF.Value,
		},
		"dkim": map[string]interface{}{
			"type":  "TXT",
			"name":  records.DKIM.Domain,
			"value": records.DKIM.Value,
		},
		"dmarc": map[string]interface{}{
			"type":  "TXT",
			"name":  records.DMARC.Domain,
			"value": records.DMARC.Value,
		},
		"mx": map[string]interface{}{
			"type":     "MX",
			"name":     "@",
			"priority": records.MX.Priority,
			"value":    records.MX.Hostname,
		},
	}

	// Add SRV records if present
	if len(records.SRV) > 0 {
		srvRecords := make([]map[string]interface{}, len(records.SRV))
		for i, srv := range records.SRV {
			srvRecords[i] = map[string]interface{}{
				"type":     "SRV",
				"service":  srv.Service,
				"proto":    srv.Proto,
				"domain":   srv.Domain,
				"priority": srv.Priority,
				"weight":   srv.Weight,
				"port":     srv.Port,
				"target":   srv.Target,
			}
		}
		recordsMap["srv"] = srvRecords
	}

	// Add autodiscover CNAMEs if present
	if len(records.AutodiscoverCNAMEs) > 0 {
		cnameRecords := make([]map[string]interface{}, len(records.AutodiscoverCNAMEs))
		for i, cname := range records.AutodiscoverCNAMEs {
			cnameRecords[i] = map[string]interface{}{
				"type":   cname.Type,
				"name":   cname.Name,
				"target": cname.Target,
			}
		}
		recordsMap["autodiscover_cnames"] = cnameRecords
	}

	output := map[string]interface{}{
		"domain":  records.Domain,
		"records": recordsMap,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// SaveGuide writes the DNS guide to a file.
func SaveGuide(path string, records AllRecords) error {
	content := GenerateGuide(records)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write guide to %s: %w", path, err)
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
