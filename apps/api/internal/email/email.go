package email

import (
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

const fromAddress = "checkmeup <noreply@checkmeup.net>"

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
<p>You requested a password reset for your checkmeup account.</p>
<p><a href="%s">Reset your password</a></p>
<p>This link expires in 1 hour. If you did not request this, ignore this email.</p>
`, resetURL)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{to},
		Subject: "Reset your checkmeup password",
		Html:    html,
	})
	return err
}
