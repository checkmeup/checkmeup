package email

import (
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/resend/resend-go/v2"
)

const fromAddress = "Checkmeup <noreply@checkmeup.net>"
const founderAddress = "andrew@checkmeup.net"

type Sender struct {
	client *resend.Client
	// apiKey is empty in dev when RESEND_API_KEY is not set
	apiKey string
}

func NewSender(apiKey string) *Sender {
	var client *resend.Client
	if apiKey != "" {
		client = resend.NewClient(apiKey)
	}
	return &Sender{client: client, apiKey: apiKey}
}

func (s *Sender) SendPasswordReset(to, resetURL string) error {
	if s.client == nil {
		slog.Warn("email sending skipped: RESEND_API_KEY not set", "to", to)
		return nil
	}

	html := fmt.Sprintf(`
<p>You requested a password reset for your Checkmeup account.</p>
<p><a href="%s">Reset your password</a></p>
<p>This link expires in 1 hour. If you did not request this, ignore this email.</p>
`, resetURL)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{to},
		Subject: "Reset your Checkmeup password",
		Html:    html,
	})
	return err
}

// SendAlertEmail sends a monitor alert (down/recovery) to an org's alert email address.
func (s *Sender) SendAlertEmail(to, subject, html string) error {
	if s.client == nil {
		slog.Warn("email sending skipped: RESEND_API_KEY not set", "to", to)
		return nil
	}

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})
	return err
}

// SendTestAlertEmail verifies deliverability before a user saves an alert email address.
func (s *Sender) SendTestAlertEmail(to string) error {
	return s.SendAlertEmail(to, "Checkmeup: test alert", "<p>✅ Checkmeup is connected! You'll receive alerts here.</p>")
}

func (s *Sender) SendFeatureSuggestion(fromEmail, text string) error {
	if s.client == nil {
		slog.Warn("email sending skipped: RESEND_API_KEY not set", "from", fromEmail)
		return nil
	}

	escapedText := strings.ReplaceAll(html.EscapeString(text), "\n", "<br>")
	body := fmt.Sprintf(`
<p>New feature suggestion from %s:</p>
<blockquote>%s</blockquote>
`, html.EscapeString(fromEmail), escapedText)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{founderAddress},
		Subject: "Checkmeup: new feature suggestion",
		Html:    body,
	})
	return err
}
