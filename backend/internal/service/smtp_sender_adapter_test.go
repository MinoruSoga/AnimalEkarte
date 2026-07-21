package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// TestToSMTPConfig_FieldMapping は reservation.SMTPConfig→smtpConfig の写像で
// フィールドの取り違え（例: User と Pass の入れ替わり）が起きないことを固定する
// （R④ レビュー HIGH 指摘: closure 経由の統合はどのテストも実写像を通らないため）。
func TestToSMTPConfig_FieldMapping(t *testing.T) {
	got := toSMTPConfig(reservation.SMTPConfig{Host: "h", Port: "587", User: "u", Pass: "secret"})
	assert.Equal(t, smtpConfig{Host: "h", Port: "587", User: "u", Pass: "secret"}, got)
}
