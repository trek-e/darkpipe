package autodiscover

import (
	"bytes"
	"fmt"
	"text/template"
)

const autodiscoverTemplate = `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006">
  <Response xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a">
    <Account>
      <AccountType>email</AccountType>
      <Action>settings</Action>
      <Protocol>
        <Type>IMAP</Type>
        <Server>{{.MailHostname}}</Server>
        <Port>993</Port>
        <DomainRequired>off</DomainRequired>
        <LoginName>{{.Email}}</LoginName>
        <SPA>off</SPA>
        <SSL>on</SSL>
        <AuthRequired>on</AuthRequired>
      </Protocol>
      <Protocol>
        <Type>SMTP</Type>
        <Server>{{.MailHostname}}</Server>
        <Port>587</Port>
        <DomainRequired>off</DomainRequired>
        <LoginName>{{.Email}}</LoginName>
        <SPA>off</SPA>
        <Encryption>TLS</Encryption>
        <AuthRequired>on</AuthRequired>
        <UsePOPAuth>on</UsePOPAuth>
        <SMTPLast>off</SMTPLast>
      </Protocol>
    </Account>
  </Response>
</Autodiscover>`

// AutodiscoverData contains the data for generating Outlook autodiscover XML.
type AutodiscoverData struct {
	Email        string
	MailHostname string
}

// GenerateAutodiscover generates Microsoft Outlook autodiscover XML.
// This should be served at:
// - autodiscover.<domain>/autodiscover/autodiscover.xml
func GenerateAutodiscover(email, mailHostname string) ([]byte, error) {
	tmpl, err := template.New("autodiscover").Parse(autodiscoverTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	data := AutodiscoverData{
		Email:        email,
		MailHostname: mailHostname,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
