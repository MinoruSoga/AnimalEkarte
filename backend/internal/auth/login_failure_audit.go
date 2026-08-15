package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const loginFailureAuditTimeout = 5 * time.Second

func (h *HTTPHandler) enqueueLoginFailureAudit(
	accountID uint64,
	knownAccount bool,
	clientIP, userAgent string,
) {
	if !knownAccount {
		return
	}
	h.loginAuditOnce.Do(func() {
		if h.loginAuditSlots == nil {
			h.loginAuditSlots = make(chan struct{}, loginFailureAuditWorkers)
		}
	})
	select {
	case h.loginAuditSlots <- struct{}{}:
	default:
		logLoginFailureAuditFallback(accountID, "worker_saturated")
		return
	}

	if !h.loginAuditGate.Register() {
		<-h.loginAuditSlots
		logLoginFailureAuditFallback(accountID, "registration_closed")
		return
	}
	sharedkernel.GoSafe("login failure audit", func() {
		defer h.loginAuditGate.Done()
		defer func() { <-h.loginAuditSlots }()
		ctx, cancel := context.WithTimeout(
			context.Background(),
			loginFailureAuditTimeout,
		)
		defer cancel()
		h.AuditKnownAccountLoginFailure(
			ctx,
			accountID,
			clientIP,
			userAgent,
		)
	})
}

func logLoginFailureAuditFallback(accountID uint64, reason string) {
	slog.Error(
		"login failure audit persistence fallback",
		slog.String("audit_action", model.AuditActionAuthLoginFailure),
		slog.Uint64("account_id", accountID),
		slog.String("reason", reason),
	)
}

// Wait drains bounded authentication background work during server shutdown.
func (h *HTTPHandler) Wait() {
	if h == nil {
		return
	}
	h.loginAuditGate.CloseAndWait()
}
