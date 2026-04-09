package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationNotifier は予約確定・キャンセル時の通知インターフェース。
// 実装はすべてのエラーをログに記録し、呼び出し元には返さない（fire-and-forget）。
type ReservationNotifier interface {
	// NotifyCreated は予約確定後に LINE + メールを送信する。
	NotifyCreated(ctx context.Context, appt *model.ReservationAppointment, customer *model.ReservationCustomer)
	// NotifyCancelled は予約キャンセル後に LINE + メールを送信する。
	NotifyCancelled(ctx context.Context, appt *model.ReservationAppointment, customer *model.ReservationCustomer)
}

// ReservationNotificationConfig はSMTP共有設定を保持する。
// LINE アクセストークン・通知先メールアドレスはクリニックごとに DB から取得する。
type ReservationNotificationConfig struct {
	// SMTP設定（SMTPHost が空文字=無効）
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

type reservationNotificationService struct {
	cfg         ReservationNotificationConfig
	settingRepo repository.ReservationSettingRepository
}

// NewReservationNotificationService は通知サービスを初期化して返す。
// SMTP設定が空の場合はメール送信をスキップする。
// LINE アクセストークン・通知先メールは予約のクリニック設定（DB）から都度取得する。
func NewReservationNotificationService(cfg ReservationNotificationConfig, settingRepo repository.ReservationSettingRepository) ReservationNotifier {
	return &reservationNotificationService{
		cfg:         cfg,
		settingRepo: settingRepo,
	}
}

// ---- NotifyCreated ----

func (s *reservationNotificationService) NotifyCreated(
	ctx context.Context,
	appt *model.ReservationAppointment,
	customer *model.ReservationCustomer,
) {
	lineText := s.buildCreatedLineMessage(appt)
	emailSubject, emailBody := s.buildCreatedEmail(appt, customer)
	clinicID := appt.ClinicID

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// クリニックごとの設定を DB から取得（LINE アクセストークン・通知先メール）
		setting, err := s.settingRepo.FindByClinicID(bgCtx, clinicID)
		if err != nil {
			slog.ErrorContext(bgCtx, "notify created: failed to load clinic setting",
				"clinic_id", clinicID,
				"reservation_id", appt.ID,
				"error", err.Error(),
			)
			return
		}

		if customer != nil && customer.LineUserID != "" && setting.LineAccessToken != "" {
			lineSvc := NewLineMessagingService(setting.LineAccessToken)
			if err := lineSvc.PushText(bgCtx, customer.LineUserID, lineText); err != nil {
				slog.ErrorContext(bgCtx, "LINE notify created failed",
					"reservation_id", appt.ID,
					"error", err.Error(),
				)
			}
		}

		if setting.NotificationEmail != "" {
			if err := s.sendEmail(bgCtx, setting.NotificationEmail, emailSubject, emailBody); err != nil {
				slog.ErrorContext(bgCtx, "email notify created failed",
					"reservation_id", appt.ID,
					"error", err.Error(),
				)
			}
		}
	}()
}

// ---- NotifyCancelled ----

func (s *reservationNotificationService) NotifyCancelled(
	ctx context.Context,
	appt *model.ReservationAppointment,
	customer *model.ReservationCustomer,
) {
	lineMsg := s.buildCancelledLineMessage(appt)
	emailSubject, emailBody := s.buildCancelledEmail(appt, customer)
	clinicID := appt.ClinicID

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// クリニックごとの設定を DB から取得（LINE アクセストークン・通知先メール）
		setting, err := s.settingRepo.FindByClinicID(bgCtx, clinicID)
		if err != nil {
			slog.ErrorContext(bgCtx, "notify cancelled: failed to load clinic setting",
				"clinic_id", clinicID,
				"reservation_id", appt.ID,
				"error", err.Error(),
			)
			return
		}

		if customer != nil && customer.LineUserID != "" && setting.LineAccessToken != "" {
			lineSvc := NewLineMessagingService(setting.LineAccessToken)
			if err := lineSvc.PushText(bgCtx, customer.LineUserID, lineMsg); err != nil {
				slog.ErrorContext(bgCtx, "LINE notify cancelled failed",
					"reservation_id", appt.ID,
					"error", err.Error(),
				)
			}
		}

		if setting.NotificationEmail != "" {
			if err := s.sendEmail(bgCtx, setting.NotificationEmail, emailSubject, emailBody); err != nil {
				slog.ErrorContext(bgCtx, "email notify cancelled failed",
					"reservation_id", appt.ID,
					"error", err.Error(),
				)
			}
		}
	}()
}

// ---- LINE メッセージテンプレート ----

func (s *reservationNotificationService) buildCreatedLineMessage(appt *model.ReservationAppointment) string {
	var sb strings.Builder
	sb.WriteString("ご予約を承りました。\n\n")
	sb.WriteString(fmt.Sprintf("■ 予約番号: R-%06d\n", appt.ID))
	sb.WriteString(fmt.Sprintf("■ 日時: %s\n", formatDateTimeJP(appt.StartTime, appt.EndTime)))
	if appt.ServiceType != nil {
		sb.WriteString(fmt.Sprintf("■ メニュー: %s\n", appt.ServiceType.Name))
	}
	if appt.Doctor != nil {
		sb.WriteString(fmt.Sprintf("■ 担当: %s\n", appt.Doctor.Name))
	}
	if appt.Pet != nil {
		petInfo := appt.Pet.Name
		sb.WriteString(fmt.Sprintf("■ ペット: %s\n", petInfo))
	}
	sb.WriteString("\nキャンセルはLINEメニューの\n「予約確認・キャンセル」から行えます。")
	return sb.String()
}

func (s *reservationNotificationService) buildCancelledLineMessage(appt *model.ReservationAppointment) string {
	var sb strings.Builder
	sb.WriteString("以下のご予約をキャンセルしました。\n\n")
	sb.WriteString(fmt.Sprintf("■ 予約番号: R-%06d\n", appt.ID))
	sb.WriteString(fmt.Sprintf("■ 日時: %s\n", formatDateTimeJP(appt.StartTime, appt.EndTime)))
	if appt.ServiceType != nil {
		sb.WriteString(fmt.Sprintf("■ メニュー: %s\n", appt.ServiceType.Name))
	}
	sb.WriteString("\n再度のご予約はLINEメニューの\n「予約する」から行えます。")
	return sb.String()
}

// ---- メールテンプレート ----

func (s *reservationNotificationService) buildCreatedEmail(
	appt *model.ReservationAppointment,
	customer *model.ReservationCustomer,
) (subject, body string) {
	courseName := ""
	if appt.ServiceType != nil {
		courseName = appt.ServiceType.Name
	}
	dateStr := appt.StartTime.Format("2006年01月02日 15:04")

	subject = fmt.Sprintf("【予約通知】%s 様 - %s (%s)", customerDisplayName(customer), courseName, dateStr)

	var sb strings.Builder
	sb.WriteString("新規予約が入りました。\n\n")
	sb.WriteString(fmt.Sprintf("■ 予約番号: R-%06d\n", appt.ID))
	if customer != nil {
		sb.WriteString(fmt.Sprintf("■ お名前: %s\n", customer.DisplayName))
		sb.WriteString(fmt.Sprintf("■ 本名: %s\n", customer.RealName))
		if len(customer.AdditionalFields) > 0 {
			var fields map[string]any
			if err := json.Unmarshal(customer.AdditionalFields, &fields); err == nil {
				if phone, ok := fields["phone"].(string); ok && phone != "" {
					sb.WriteString(fmt.Sprintf("■ 電話番号: %s\n", phone))
				}
				if note, ok := fields["note"].(string); ok && note != "" {
					sb.WriteString(fmt.Sprintf("■ 診察内容: %s\n", note))
				}
			}
		}
	}
	if appt.Owner != nil {
		sb.WriteString(fmt.Sprintf("■ 飼い主名: %s\n", appt.Owner.OwnerName))
	}
	if appt.Pet != nil {
		sb.WriteString(fmt.Sprintf("■ ペット: %s\n", appt.Pet.Name))
	}
	sb.WriteString(fmt.Sprintf("■ コース: %s\n", courseName))
	if appt.Doctor != nil {
		sb.WriteString(fmt.Sprintf("■ 担当: %s\n", appt.Doctor.Name))
	}
	sb.WriteString(fmt.Sprintf("■ 日時: %s〜%s\n",
		appt.StartTime.Format("2006年01月02日(月) 15:04"),
		appt.EndTime.Format("15:04"),
	))
	if appt.Notes != "" {
		sb.WriteString(fmt.Sprintf("■ 要望: %s\n", appt.Notes))
	} else {
		sb.WriteString("■ 要望: （なし）\n")
	}
	sb.WriteString("■ 予約元: LINE\n")
	body = sb.String()
	return
}

func (s *reservationNotificationService) buildCancelledEmail(
	appt *model.ReservationAppointment,
	customer *model.ReservationCustomer,
) (subject, body string) {
	courseName := ""
	if appt.ServiceType != nil {
		courseName = appt.ServiceType.Name
	}
	dateStr := appt.StartTime.Format("2006年01月02日 15:04")

	subject = fmt.Sprintf("【予約キャンセル】%s 様 - %s (%s)", customerDisplayName(customer), courseName, dateStr)

	var sb strings.Builder
	sb.WriteString("以下の予約がキャンセルされました。\n\n")
	sb.WriteString(fmt.Sprintf("■ 予約番号: R-%06d\n", appt.ID))
	if customer != nil {
		sb.WriteString(fmt.Sprintf("■ お名前: %s\n", customer.DisplayName))
	}
	sb.WriteString(fmt.Sprintf("■ コース: %s\n", courseName))
	sb.WriteString(fmt.Sprintf("■ 日時: %s〜%s\n",
		appt.StartTime.Format("2006年01月02日(月) 15:04"),
		appt.EndTime.Format("15:04"),
	))
	body = sb.String()
	return
}

// ---- SMTP送信 ----

func (s *reservationNotificationService) sendEmail(_ context.Context, to, subject, body string) error {
	if s.cfg.SMTPHost == "" {
		return nil
	}

	from := s.cfg.SMTPFrom
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := s.cfg.SMTPHost + ":" + s.cfg.SMTPPort
	var auth smtp.Auth
	if s.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

// ---- ユーティリティ ----

func customerDisplayName(c *model.ReservationCustomer) string {
	if c == nil {
		return "不明"
	}
	if c.RealName != "" {
		return c.RealName
	}
	return c.DisplayName
}

func formatDateTimeJP(start, end time.Time) string {
	return fmt.Sprintf("%s〜%s",
		start.Format("2006年01月02日(月) 15:04"),
		end.Format("15:04"),
	)
}
