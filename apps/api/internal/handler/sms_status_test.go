package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
)

func TestHandleSMSStatus(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	authH, _, _ := testNotificationChannelHandler(t)
	h := NewSMSStatusHandler(pool, email.NewSender(""), "right-token")

	// sentSMS records an alert sms the way the worker does after Twilio accepts it.
	sentSMS := func(t *testing.T) (db.NotificationChannel, string) {
		t.Helper()
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		c, err := queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
			OrgID: orgID, Type: db.NotificationChannelTypeSms, Name: "SMS", Config: []byte(`{"phone_number":"+15005550006"}`),
		})
		if err != nil {
			t.Fatalf("create channel: %v", err)
		}
		sid := "SM" + uuid.NewString()
		if err := queries.InsertSMSMessage(context.Background(), db.InsertSMSMessageParams{Sid: sid, OrgID: orgID, ChannelID: c.ID}); err != nil {
			t.Fatalf("insert sms message: %v", err)
		}
		return c, sid
	}
	post := func(token, sid string) *httptest.ResponseRecorder {
		form := url.Values{"MessageSid": {sid}, "MessageStatus": {"undelivered"}, "ErrorCode": {"30003"}}
		req := httptest.NewRequest(http.MethodPost, "/webhook/twilio/sms-status/"+token, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = withURLParam(req, "token", token)
		w := httptest.NewRecorder()
		h.HandleStatus(w, req)
		return w
	}
	lastDelivery := func(t *testing.T, c db.NotificationChannel) string {
		t.Helper()
		got, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: c.ID, OrgID: c.OrgID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		return got.LastDeliveryDetail.String
	}

	t.Run("a wrong token is a 404 and changes nothing", func(t *testing.T) {
		c, sid := sentSMS(t)
		if w := post("wrong-token", sid); w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
		if got := lastDelivery(t, c); got != "" {
			t.Fatalf("last delivery = %q, want none recorded", got)
		}
	})

	t.Run("the right token records the final status on the channel", func(t *testing.T) {
		c, sid := sentSMS(t)
		if w := post("right-token", sid); w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if got := lastDelivery(t, c); got != "undelivered (error 30003)" {
			t.Fatalf("last delivery = %q, want %q", got, "undelivered (error 30003)")
		}
	})
}
