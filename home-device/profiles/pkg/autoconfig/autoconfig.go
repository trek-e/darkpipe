// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package autoconfig

import (
	"bytes"
	"fmt"
	"text/template"
)

const autoconfigTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<clientConfig version="1.1">
  <emailProvider id="{{.Domain}}">
    <domain>{{.Domain}}</domain>
    <displayName>{{.Domain}}</displayName>
    <displayShortName>{{.Domain}}</displayShortName>
    <incomingServer type="imap">
      <hostname>{{.MailHostname}}</hostname>
      <port>993</port>
      <socketType>SSL</socketType>
      <authentication>password-cleartext</authentication>
      <username>%EMAILADDRESS%</username>
    </incomingServer>
    <outgoingServer type="smtp">
      <hostname>{{.MailHostname}}</hostname>
      <port>587</port>
      <socketType>STARTTLS</socketType>
      <authentication>password-cleartext</authentication>
      <username>%EMAILADDRESS%</username>
    </outgoingServer>
  </emailProvider>
</clientConfig>`

// AutoconfigData contains the data for generating Mozilla autoconfig XML.
type AutoconfigData struct {
	Domain       string
	MailHostname string
}

// GenerateAutoconfig generates Mozilla/Thunderbird autoconfig XML.
// This should be served at:
// - /.well-known/autoconfig/mail/config-v1.1.xml
// - autoconfig.<domain>/mail/config-v1.1.xml
func GenerateAutoconfig(domain, mailHostname string) ([]byte, error) {
	tmpl, err := template.New("autoconfig").Parse(autoconfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	data := AutoconfigData{
		Domain:       domain,
		MailHostname: mailHostname,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
