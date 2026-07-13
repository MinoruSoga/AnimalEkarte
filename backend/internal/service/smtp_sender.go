package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// smtpConfig は SMTP 送信に必要な接続設定。
type smtpConfig struct {
	Host string
	Port string
	User string
	Pass string
}

// validateSMTPLine は net/smtp.SendMail 内部の validateLine 相当のガードである。
// MAIL FROM/RCPT TO 等のエンベロープコマンドに CR/LF を許すと SMTP コマンド
// インジェクションになるため、手動展開の smtp クライアント呼び出し前に拒否する。
func validateSMTPLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return fmt.Errorf("smtp: A line must not contain CR or LF")
	}
	return nil
}

// sendSMTPMail は ctx 対応の SMTP 送信を行う。
// smtp.SendMail は ctx を受け付けないため deadline が SMTP サーバ無応答時に
// 伝播しない（backend/CLAUDE.md X-18）。DialContext + conn.SetDeadline で
// net/smtp.SendMail 相当の手順を ctx 対応に手動展開する。
func sendSMTPMail(ctx context.Context, cfg smtpConfig, from, to string, msg []byte) error {
	// net/smtp.SendMail の validateLine 相当: エンベロープコマンド（MAIL FROM/RCPT TO）に
	// CR/LF を許すと SMTP コマンドインジェクションになるため、dial する前に拒否する。
	if err := validateSMTPLine(from); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := validateSMTPLine(to); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	addr := cfg.Host + ":" + cfg.Port
	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
	}

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
	}

	// net/smtp.SendMail は AUTH extension をサーバが広告している場合のみ Auth を試みる
	// （未広告サーバへの無条件 Auth 呼び出しは、AUTH 非対応サーバに対する不要な認証失敗を
	// 招き得る）。手動展開でも同じガードを踏襲する。
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("smtp send: %w", err)
			}
		}
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	if err := c.Quit(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
