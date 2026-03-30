package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// writeAuditLog は監査ログを best-effort で書き込む共通ヘルパー。
// 書き込み失敗はメイン処理に影響しない（ログ出力のみ）。
func (h *Handler) writeAuditLog(c *gin.Context, action, resource string, resourceID *uint64, oldValue, newValue []byte) {
	ctx := c.Request.Context()

	var clinicID *uint64
	if val, exists := c.Get("clinic_id"); exists {
		if s, ok := val.(string); ok {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				clinicID = &id
			}
		}
	}

	var actorID *uint64
	if id, ok := extractUserID(c); ok {
		actorID = &id
	}

	actorType := "anonymous"
	if val, exists := c.Get("user_type"); exists {
		if s, ok := val.(string); ok && s != "" {
			actorType = s
		}
	}

	log := &model.AuditLog{
		ClinicID:   clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}
	if err := h.auditRepo.Create(ctx, log); err != nil {
		slog.WarnContext(ctx, "audit log write failed",
			slog.String("action", action),
			slog.String("resource", resource),
			slog.String("error", err.Error()))
	}
}
