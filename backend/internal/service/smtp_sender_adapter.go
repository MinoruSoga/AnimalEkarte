package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// toSMTPConfig は reservation.SMTPConfig（通知 closure の公開型）を smtp_sender の
// 内部 config へ写像する。フィールド対応の正しさは smtp_sender_adapter_test.go が固定する。
func toSMTPConfig(cfg reservation.SMTPConfig) smtpConfig {
	return smtpConfig{Host: cfg.Host, Port: cfg.Port, User: cfg.User, Pass: cfg.Pass}
}

// smtpSendAdapter は reservation 通知サービスへ注入する sendMail closure の実体。
func smtpSendAdapter(ctx context.Context, cfg reservation.SMTPConfig, from, to string, msg []byte) error {
	return sendSMTPMail(ctx, toSMTPConfig(cfg), from, to, msg)
}
