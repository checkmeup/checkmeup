package handler

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/worker"
)

// SMSStatusHandler receives Twilio's status callbacks for alert SMS (ADR
// alerting), so a message the carrier fails after Twilio accepted it still
// counts as failed.
type SMSStatusHandler struct {
	token string
	n     worker.Notifiers
}

// NewSMSStatusHandler takes the token the callback URL must carry
// (twilio.StatusCallbackToken).
func NewSMSStatusHandler(pool *pgxpool.Pool, mailer *email.Sender, token string) *SMSStatusHandler {
	return &SMSStatusHandler{
		token: token,
		n:     worker.Notifiers{Queries: db.New(pool), Mailer: mailer, Logger: slog.Default()},
	}
}

// HandleStatus POST /webhook/twilio/sms-status/{token}
func (h *SMSStatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if subtle.ConstantTimeCompare([]byte(chi.URLParam(r, "token")), []byte(h.token)) != 1 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	worker.HandleSMSStatus(r.Context(), h.n, r.PostForm.Get("MessageSid"), r.PostForm.Get("MessageStatus"), r.PostForm.Get("ErrorCode"))
	w.WriteHeader(http.StatusNoContent)
}
