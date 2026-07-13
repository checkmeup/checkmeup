package handler

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// validateChannelConfig checks config has the field(s) required for type.
// Other notification_channel_type values get added (their own migration +
// their own case here) when their epic actually ships (ADR-023).
func validateChannelConfig(channelType string, config map[string]any) error {
	switch channelType {
	case "telegram":
		chatID, _ := config["chatId"].(string)
		if strings.TrimSpace(chatID) == "" {
			return errors.New("chatId is required")
		}
	case "email":
		addr, _ := config["email"].(string)
		if strings.TrimSpace(addr) == "" {
			return errors.New("email is required")
		}
	case "webhook":
		return validateWebhookURL(config)
	case "slack":
		return validateSlackURL(config)
	case "sms":
		return validateSMSConfig(config)
	default:
		return errors.New("unsupported channel type")
	}
	return nil
}

// e164Pattern matches E.164 phone numbers (US-1901): a leading +, no leading
// zero, up to 15 digits total.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// validateSMSConfig enforces US-1901: a valid E.164 phone number, and an
// explicit opt-in checkbox — a TCPA-style regulatory requirement for
// automated texts, not satisfied just by providing a number (ADR-029).
// consent is required true on every request that reaches here; callers that
// want to carry an existing consent forward without re-prompting the user
// (see resolveUpdatedChannelConfig, for an unchanged phone number on update)
// inject consent: true themselves before calling validateChannelConfig.
func validateSMSConfig(config map[string]any) error {
	phone, _ := config["phone_number"].(string)
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("phone_number is required")
	}
	if !e164Pattern.MatchString(phone) {
		return errors.New("phone_number must be in E.164 format (e.g. +14155551234)")
	}
	if !consentGiven(config["consent"]) {
		return errors.New("consent is required — check the opt-in box before saving")
	}
	return nil
}

// consentGiven reports true for either a JSON boolean true or the string
// "true" — config values arrive as any (json.Unmarshal into map[string]any),
// but every other channel's config is plain strings on the wire (chatId,
// email, url...), so the frontend sends "true" here too rather than being
// the one field that needs a non-string type.
func consentGiven(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true"
	default:
		return false
	}
}

// finalizeSMSConsent strips the client-supplied consent flag (already
// validated true by validateSMSConfig) and stamps a server-set consent_at —
// never trusting a client-supplied timestamp — unless one was already
// carried forward from an existing channel (see resolveUpdatedChannelConfig).
func finalizeSMSConsent(config map[string]any) {
	if _, ok := config["consent_at"]; !ok {
		config["consent_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	delete(config, "consent")
}

// validateWebhookURL enforces the US-1401 AC that a webhook URL must be
// https://. Doesn't require config["secret"] — that's generated server-side
// (see ensureWebhookSecret), never supplied by the client.
func validateWebhookURL(config map[string]any) error {
	rawURL, _ := config["url"].(string)
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("url is required")
	}
	if !strings.HasPrefix(rawURL, "https://") {
		return errors.New("url must start with https://")
	}
	return nil
}

// validateSlackURL enforces US-1701: the URL must match the Slack Incoming
// Webhook pattern (https://hooks.slack.com/...). The URL itself is the
// credential — no separate secret like the generic webhook channel (US-1401).
func validateSlackURL(config map[string]any) error {
	rawURL, _ := config["url"].(string)
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("url is required")
	}
	if !strings.HasPrefix(rawURL, "https://hooks.slack.com/") {
		return errors.New("url must be a Slack Incoming Webhook URL (https://hooks.slack.com/...)")
	}
	return nil
}
