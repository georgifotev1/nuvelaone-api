package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/resend/resend-go/v3"
	"html/template"
)

//go:embed templates/*.html
var templates embed.FS

type resendClient struct {
	client    *resend.Client
	fromEmail string
	devEmail  string
}

func NewResendClient(apiKey, fromEmail, devEmail string) Mailer {
	return &resendClient{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		devEmail:  devEmail,
	}
}

func (r *resendClient) Send(email EmailData) error {
	to := email.To
	if r.devEmail != "" {
		to = []string{r.devEmail}
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

	from := email.From
	if from == "" {
		from = r.fromEmail
	}

	_, err = r.client.Emails.Send(&resend.SendEmailRequest{
		From:    from,
		To:      to,
		Subject: email.Subject,
		Html:    htmlBody.String(),
	})
	if err != nil {
		return fmt.Errorf("mailer.Send: resend: %w", err)
	}

	return nil
}
