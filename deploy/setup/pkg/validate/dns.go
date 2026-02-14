package validate

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// DNS servers to use for validation
var dnsServers = []string{
	"8.8.8.8:53",   // Google DNS
	"1.1.1.1:53",   // Cloudflare DNS
	"208.67.222.222:53", // OpenDNS
}

// ValidateDomain checks if a domain has valid MX and A/AAAA records
func ValidateDomain(domain string) error {
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	// Check MX records
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeMX)

	var mxFound bool
	for _, server := range dnsServers {
		r, _, err := c.Exchange(m, server)
		if err == nil && len(r.Answer) > 0 {
			mxFound = true
			break
		}
	}

	if !mxFound {
		return fmt.Errorf("no MX records found for %s - set up DNS before running setup", domain)
	}

	// Check A/AAAA records
	m = new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	var aFound bool
	for _, server := range dnsServers {
		r, _, err := c.Exchange(m, server)
		if err == nil && len(r.Answer) > 0 {
			aFound = true
			break
		}
	}

	if !aFound {
		// Try AAAA
		m = new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)

		for _, server := range dnsServers {
			r, _, err := c.Exchange(m, server)
			if err == nil && len(r.Answer) > 0 {
				aFound = true
				break
			}
		}
	}

	if !aFound {
		return fmt.Errorf("no A or AAAA records found for %s - set up DNS before running setup", domain)
	}

	return nil
}

// ValidateSPF checks if a domain has an SPF record
func ValidateSPF(domain string) (bool, error) {
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)

	for _, server := range dnsServers {
		r, _, err := c.Exchange(m, server)
		if err != nil {
			continue
		}

		for _, ans := range r.Answer {
			if txt, ok := ans.(*dns.TXT); ok {
				for _, str := range txt.Txt {
					if len(str) > 6 && str[:7] == "v=spf1 " {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// ValidateDKIM checks if a DKIM key record exists
func ValidateDKIM(domain, selector string) (bool, error) {
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	dkimDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(dkimDomain), dns.TypeTXT)

	for _, server := range dnsServers {
		r, _, err := c.Exchange(m, server)
		if err != nil {
			continue
		}

		if len(r.Answer) > 0 {
			return true, nil
		}
	}

	return false, nil
}

// ValidateDMARC checks if a DMARC record exists
func ValidateDMARC(domain string) (bool, error) {
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	dmarcDomain := fmt.Sprintf("_dmarc.%s", domain)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(dmarcDomain), dns.TypeTXT)

	for _, server := range dnsServers {
		r, _, err := c.Exchange(m, server)
		if err != nil {
			continue
		}

		for _, ans := range r.Answer {
			if txt, ok := ans.(*dns.TXT); ok {
				for _, str := range txt.Txt {
					if len(str) > 8 && str[:8] == "v=DMARC1" {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
