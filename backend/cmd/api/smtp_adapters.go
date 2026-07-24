package main

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/infra/smtp"
	"github.com/animal-ekarte/backend/internal/reservation"
)

func smtpConfigFromAuth(config *auth.PasswordResetConfig) (smtp.Config, error) {
	if config == nil {
		return smtp.Config{}, fmt.Errorf("auth SMTP config is required")
	}
	return smtp.Config{
		Host: config.SMTPHost,
		Port: config.SMTPPort,
		User: config.SMTPUser,
		Pass: config.SMTPPass,
	}, nil
}

func sendAuthPasswordResetMail(
	ctx context.Context,
	config *auth.PasswordResetConfig,
	from, to string,
	message []byte,
) error {
	smtpConfig, err := smtpConfigFromAuth(config)
	if err != nil {
		return err
	}
	return smtp.Send(ctx, smtpConfig, from, to, message)
}

func smtpConfigFromReservation(config reservation.SMTPConfig) smtp.Config {
	return smtp.Config{
		Host: config.Host,
		Port: config.Port,
		User: config.User,
		Pass: config.Pass,
	}
}

func sendReservationMail(
	ctx context.Context,
	config reservation.SMTPConfig,
	from, to string,
	message []byte,
) error {
	return smtp.Send(
		ctx,
		smtpConfigFromReservation(config),
		from,
		to,
		message,
	)
}
