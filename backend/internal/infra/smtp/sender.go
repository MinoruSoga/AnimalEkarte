// Package smtp provides the context-aware SMTP transport shared by domain
// notification services.
package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	netsmtp "net/smtp"
	"strconv"
	"strings"
	"time"
)

// Config contains SMTP connection settings.
type Config struct {
	Host string
	Port string
	User string
	Pass string
	// AllowInsecureLoopback permits plaintext SMTP only for an explicit
	// loopback development server. Ports 465 and 587 are never weakened.
	AllowInsecureLoopback bool
}

type transportMode uint8

const (
	opportunisticSTARTTLS transportMode = iota
	implicitTLS
	requiredSTARTTLS
)

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
	if (cfg.User == "") != (cfg.Pass == "") {
		return fmt.Errorf("smtp send: user and password must be configured together")
	}
	if err := ValidateEnvelopeAddress(from); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := ValidateEnvelopeAddress(to); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	mode, err := modeForConfig(cfg)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	port := strings.TrimSpace(cfg.Port)
	address := net.JoinHostPort(cfg.Host, port)
	var auth netsmtp.Auth
	if cfg.User != "" {
		auth = netsmtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return sendError(ctx, err)
	}
	defer func() { _ = connection.Close() }()

	if deadline, ok := ctx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			return sendError(ctx, err)
		}
	}
	stopCancellation := context.AfterFunc(ctx, func() {
		_ = connection.SetDeadline(time.Now())
	})
	defer stopCancellation()

	client, err := newClient(ctx, connection, cfg.Host, mode)
	if err != nil {
		return sendError(ctx, err)
	}
	defer func() { _ = client.Close() }()

	if auth != nil {
		if supported, _ := client.Extension("AUTH"); !supported {
			return fmt.Errorf("smtp send: AUTH is configured but not advertised by server")
		}
		if err := client.Auth(auth); err != nil {
			return sendError(ctx, err)
		}
	}

	if err := client.Mail(from); err != nil {
		return sendError(ctx, err)
	}
	if err := client.Rcpt(to); err != nil {
		return sendError(ctx, err)
	}

	writer, err := client.Data()
	if err != nil {
		return sendError(ctx, err)
	}
	if _, err := writer.Write(message); err != nil {
		return sendError(ctx, err)
	}
	if err := writer.Close(); err != nil {
		return sendError(ctx, err)
	}
	if err := client.Quit(); err != nil {
		return sendError(ctx, err)
	}
	return nil
}

func modeForPort(port string) transportMode {
	mode, err := modeForPortWithLookup(port, net.LookupPort)
	if err != nil {
		return requiredSTARTTLS
	}
	return mode
}

type lookupPortFunc func(network, service string) (int, error)

func modeForPortWithLookup(
	port string,
	lookupPort lookupPortFunc,
) (transportMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(port))
	numeric := strings.TrimPrefix(normalized, "+")
	if numeric != "" && strings.IndexFunc(numeric, func(character rune) bool {
		return character < '0' || character > '9'
	}) == -1 {
		number, err := strconv.ParseUint(numeric, 10, 16)
		if err != nil {
			return requiredSTARTTLS, fmt.Errorf("invalid SMTP port %q: %w", port, err)
		}
		return modeForPortNumber(int(number)), nil
	}

	number, err := lookupPort("tcp", normalized)
	if err != nil {
		return requiredSTARTTLS, fmt.Errorf(
			"resolve SMTP service %q: %w",
			normalized,
			err,
		)
	}
	return modeForPortNumber(number), nil
}

func modeForPortNumber(port int) transportMode {
	switch port {
	case 465:
		return implicitTLS
	case 587:
		return requiredSTARTTLS
	default:
		return opportunisticSTARTTLS
	}
}

func modeForConfig(cfg Config) (transportMode, error) {
	return modeForConfigWithLookup(cfg, net.LookupPort)
}

func modeForConfigWithLookup(
	cfg Config,
	lookupPort lookupPortFunc,
) (transportMode, error) {
	mode, err := modeForPortWithLookup(cfg.Port, lookupPort)
	if err != nil {
		return requiredSTARTTLS, err
	}
	if mode != opportunisticSTARTTLS {
		return mode, nil
	}
	if !cfg.AllowInsecureLoopback {
		return requiredSTARTTLS, nil
	}
	if !isLoopbackHost(cfg.Host) {
		return requiredSTARTTLS, fmt.Errorf(
			"insecure SMTP is allowed only for an explicit loopback host",
		)
	}
	return opportunisticSTARTTLS, nil
}

func isLoopbackHost(host string) bool {
	normalized := strings.TrimSpace(host)
	if strings.EqualFold(normalized, "localhost") {
		return true
	}
	ip := net.ParseIP(normalized)
	return ip != nil && ip.IsLoopback()
}

func newClient(
	ctx context.Context,
	connection net.Conn,
	host string,
	mode transportMode,
) (*netsmtp.Client, error) {
	return newClientWithTLSConfig(ctx, connection, host, mode, strictTLSConfig(host))
}

func strictTLSConfig(host string) *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: host,
	}
}

func newClientWithTLSConfig(
	ctx context.Context,
	connection net.Conn,
	host string,
	mode transportMode,
	tlsConfig *tls.Config,
) (*netsmtp.Client, error) {
	if mode == implicitTLS {
		tlsConnection := tls.Client(connection, tlsConfig)
		if err := tlsConnection.HandshakeContext(ctx); err != nil {
			return nil, fmt.Errorf("implicit TLS handshake: %w", err)
		}
		client, err := netsmtp.NewClient(tlsConnection, host)
		if err != nil {
			return nil, err
		}
		if err := client.Hello("localhost"); err != nil {
			_ = client.Close()
			return nil, err
		}
		return client, nil
	}

	client, err := netsmtp.NewClient(connection, host)
	if err != nil {
		return nil, err
	}
	if err := client.Hello("localhost"); err != nil {
		_ = client.Close()
		return nil, err
	}

	supported, _ := client.Extension("STARTTLS")
	if mode == requiredSTARTTLS && !supported {
		_ = client.Close()
		return nil, fmt.Errorf("STARTTLS is required for this SMTP connection")
	}
	if supported {
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, err
		}
	}
	return client, nil
}

func sendError(ctx context.Context, err error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return fmt.Errorf("smtp send: %w", contextErr)
	}
	var networkErr net.Error
	if _, ok := ctx.Deadline(); ok &&
		errors.As(err, &networkErr) &&
		networkErr.Timeout() {
		return fmt.Errorf("smtp send: %w", context.DeadlineExceeded)
	}
	return fmt.Errorf("smtp send: %w", err)
}
