package service

import (
	"context"

	smtptransport "github.com/animal-ekarte/backend/internal/infra/smtp"
)

type smtpConfig = smtptransport.Config

func validateSMTPLine(line string) error {
	return smtptransport.ValidateEnvelopeAddress(line)
}

func sendSMTPMail(ctx context.Context, cfg smtpConfig, from, to string, message []byte) error {
	return smtptransport.Send(ctx, cfg, from, to, message)
}
