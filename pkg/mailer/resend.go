package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/resend/resend-go/v3"
)

//go:embed templates/*.html
var templates embed.FS

type resendClient struct {
	client    *resend.Client
	fromEmail string
}

func NewResendClient(apiKey, fromEmail string) Mailer {
	return &resendClient{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

func (r *resendClient) Send(email EmailData) error {
	if email.From == "" {
		email.From = r.fromEmail
	}

	templatePath := "templates/" + email.Template + ".html"
	templateContent, err := templates.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("mailer.Send: read template %s: %w", email.Template, err)
	}

	tmpl, err := template.New(email.Template).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("mailer.Send: parse template %s: %w", email.Template, err)
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, email.Data); err != nil {
		return fmt.Errorf("mailer.Send: execute template %s: %w", email.Template, err)
	}

	_, err = r.client.Emails.Send(&resend.SendEmailRequest{
		From:    email.From,
		To:      email.To,
		Subject: email.Subject,
		Html:    htmlBody.String(),
	})
	if err != nil {
		return fmt.Errorf("mailer.Send: resend: %w", err)
	}

	return nil
}
