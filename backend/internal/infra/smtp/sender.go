// Package smtp provides the context-aware SMTP transport shared by domain
// notification services.
package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	netsmtp "net/smtp"
	"strings"
	"time"
)

// Config contains SMTP connection settings.
type Config struct {
	Host string
	Port string
	User string
	Pass string
}

// ValidateEnvelopeAddress rejects SMTP command injection in MAIL FROM and
// RCPT TO values.
func ValidateEnvelopeAddress(address string) error {
	if strings.ContainsAny(address, "\n\r") {
		return fmt.Errorf("smtp: A line must not contain CR or LF")
	}
	return nil
}

// Send delivers one message while respecting context cancellation and
// deadlines for every network operation.
func Send(ctx context.Context, cfg Config, from, to string, message []byte) error {
	if err := ValidateEnvelopeAddress(from); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := ValidateEnvelopeAddress(to); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	address := net.JoinHostPort(cfg.Host, cfg.Port)
	var auth netsmtp.Auth
	if cfg.User != "" {
		auth = netsmtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	defer func() { _ = connection.Close() }()

	if deadline, ok := ctx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
	}
	stopCancellation := context.AfterFunc(ctx, func() {
		_ = connection.SetDeadline(time.Now())
	})
	defer stopCancellation()

	client, err := netsmtp.NewClient(connection, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	defer func() { _ = client.Close() }()

	if supported, _ := client.Extension("STARTTLS"); supported {
		if err := client.StartTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.Host,
		}); err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
	}

	if auth != nil {
		if supported, _ := client.Extension("AUTH"); supported {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp send: %w", err)
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
