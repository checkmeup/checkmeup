package email

import (
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/resend/resend-go/v2"
)

const founderAddress = "andrew@checkmeup.net"

// FounderNotifier emails the founder directly for internal, non-tenant-facing
// notifications — currently just in-app feature suggestions (EP-23, no public
// board: DocsView.vue's "Need help?" section). See Sender for the customer-
// facing send methods.
type FounderNotifier struct {
	client *resend.Client
}

func NewFounderNotifier(apiKey string) *FounderNotifier {
	return &FounderNotifier{client: newResendClient(apiKey)}
}

func (f *FounderNotifier) SendFeatureSuggestion(fromEmail, text string) error {
	if f.client == nil {
		slog.Warn("email sending skipped: RESEND_API_KEY not set", "from", fromEmail)
		return nil
	}

	escapedText := strings.ReplaceAll(html.EscapeString(text), "\n", "<br>")
	body := fmt.Sprintf(`
<p>New feature suggestion from %s:</p>
<blockquote>%s</blockquote>
`, html.EscapeString(fromEmail), escapedText)

	_, err := f.client.Emails.Send(&resend.SendEmailRequest{
		From:    fromAddress,
		To:      []string{founderAddress},
		Subject: "Checkmeup: new feature suggestion",
		Html:    body,
	})
	return err
}
