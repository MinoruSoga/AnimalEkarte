package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationNotifier は予約確定・キャンセル時の通知インターフェース。
// 実装はすべてのエラーをログに記録し、呼び出し元には返さない（fire-and-forget）。
type ReservationNotifier interface {
	// NotifyCreated は予約確定後に LINE + メールを送信する。
	NotifyCreated(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
	// NotifyCancelled は予約キャンセル後に LINE + メールを送信する。
	NotifyCancelled(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
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
	// FrontendURL はパスワードリセットメール等に含めるフロントエンドのベースURL。
	FrontendURL string
}

type reservationNotificationService struct {
	cfg         ReservationNotificationConfig
	settingRepo repository.LineReservationSettingRepository
	// cipher は line_access_token を復号するために使う（H-4）。
	// nil の場合は復号なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
	cipher *crypto.AESGCMCipher
}

// NewReservationNotificationService は通知サービスを初期化して返す。
// SMTP設定が空の場合はメール送信をスキップする。
// LINE アクセストークン・通知先メールは予約のクリニック設定（DB）から都度取得する。
// cipher が nil の場合は復号なしで動作する（lstep 連携と同一の cipher を再利用する）。
func NewReservationNotificationService(cfg *ReservationNotificationConfig, settingRepo repository.LineReservationSettingRepository, cipher *crypto.AESGCMCipher) ReservationNotifier {
	return &reservationNotificationService{
		cfg:         *cfg,
		settingRepo: settingRepo,
		cipher:      cipher,
	}
}

// ---- NotifyCreated ----

func (s *reservationNotificationService) NotifyCreated(
	ctx context.Context,
	appt *model.Reservation,
	customer *model.LineCustomer,
) {
	lineText := s.buildCreatedLineMessage(appt)
	emailSubject, emailBody := s.buildCreatedEmail(appt, customer)
	clinicID := appt.ClinicID

	goSafe("reservation notify created", func() { //nolint:contextcheck,gosec // 意図的: 通知はリクエスト完了後も継続するため独立した background context を使用
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

		// DB 上の line_access_token は暗号文（H-4）。レガシー平文行はそのまま返る。
		accessToken := decryptLineCredential(bgCtx, s.cipher, setting.LineAccessToken)
		if customer != nil && customer.LineUserID != "" && accessToken != "" {
			lineSvc := NewLineMessagingService(accessToken)
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
	})
}

// ---- NotifyCancelled ----

func (s *reservationNotificationService) NotifyCancelled(
	ctx context.Context,
	appt *model.Reservation,
	customer *model.LineCustomer,
) {
	lineMsg := s.buildCancelledLineMessage(appt)
	emailSubject, emailBody := s.buildCancelledEmail(appt, customer)
	clinicID := appt.ClinicID

	goSafe("reservation notify cancelled", func() { //nolint:contextcheck,gosec // 意図的: 通知はリクエスト完了後も継続するため独立した background context を使用
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

		// DB 上の line_access_token は暗号文（H-4）。レガシー平文行はそのまま返る。
		accessToken := decryptLineCredential(bgCtx, s.cipher, setting.LineAccessToken)
		if customer != nil && customer.LineUserID != "" && accessToken != "" {
			lineSvc := NewLineMessagingService(accessToken)
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
	})
}

// ---- ヘルパー: LINE表示名フォールバック ----

func reservationTypeDisplayName(st *model.ReservationType) string {
	if st == nil {
		return ""
	}
	if st.ReservationDisplayName != "" {
		return st.ReservationDisplayName
	}
	if st.ShowShortName && st.ShortName != "" {
		return st.ShortName
	}
	return st.Name
}

func staffDisplayName(s *model.Staff) string {
	if s == nil {
		return ""
	}
	if s.ReservationDisplayName != "" {
		return s.ReservationDisplayName
	}
	return s.Name
}

// ---- LINE メッセージテンプレート ----

func (s *reservationNotificationService) buildCreatedLineMessage(appt *model.Reservation) string {
	var sb strings.Builder
	sb.WriteString("ご予約を承りました。\n\n")
	fmt.Fprintf(&sb, "■ 予約番号: R-%06d\n", appt.ID)
	fmt.Fprintf(&sb, "■ 日時: %s\n", formatDateTimeJP(appt.StartTime, appt.EndTime))
	if name := reservationTypeDisplayName(appt.ReservationType); name != "" {
		fmt.Fprintf(&sb, "■ メニュー: %s\n", name)
	}
	if name := staffDisplayName(appt.Doctor); name != "" {
		fmt.Fprintf(&sb, "■ 担当: %s\n", name)
	}
	if petNames := extractPetNamesFromCustomerFields(appt); petNames != "" {
		fmt.Fprintf(&sb, "■ ペット: %s\n", petNames)
	}
	sb.WriteString("\nキャンセルはLINEメニューの\n「予約確認・キャンセル」から行えます。")
	return sb.String()
}

func (s *reservationNotificationService) buildCancelledLineMessage(appt *model.Reservation) string {
	var sb strings.Builder
	sb.WriteString("以下のご予約をキャンセルしました。\n\n")
	fmt.Fprintf(&sb, "■ 予約番号: R-%06d\n", appt.ID)
	fmt.Fprintf(&sb, "■ 日時: %s\n", formatDateTimeJP(appt.StartTime, appt.EndTime))
	if name := reservationTypeDisplayName(appt.ReservationType); name != "" {
		fmt.Fprintf(&sb, "■ メニュー: %s\n", name)
	}
	sb.WriteString("\n再度のご予約はLINEメニューの\n「予約する」から行えます。")
	return sb.String()
}

// ---- メールテンプレート ----

func (s *reservationNotificationService) buildCreatedEmail(
	appt *model.Reservation,
	customer *model.LineCustomer,
) (subject, body string) {
	courseName := ""
	if appt.ReservationType != nil {
		courseName = appt.ReservationType.Name
	}
	dateStr := appt.StartTime.Format("2006年01月02日 15:04")

	subject = fmt.Sprintf("【予約通知】%s 様 - %s (%s)", customerDisplayName(customer), courseName, dateStr)

	var sb strings.Builder
	sb.WriteString("新規予約が入りました。\n\n")
	fmt.Fprintf(&sb, "■ 予約番号: R-%06d\n", appt.ID)
	if customer != nil {
		fmt.Fprintf(&sb, "■ お名前: %s\n", customer.DisplayName)
		fmt.Fprintf(&sb, "■ 本名: %s\n", customer.RealName)
		if len(customer.AdditionalFields) > 0 {
			var fields map[string]any
			if err := json.Unmarshal(customer.AdditionalFields, &fields); err == nil {
				if phone, ok := fields["phone"].(string); ok && phone != "" {
					fmt.Fprintf(&sb, "■ 電話番号: %s\n", phone)
				}
				if note, ok := fields["note"].(string); ok && note != "" {
					fmt.Fprintf(&sb, "■ 診察内容: %s\n", note)
				}
			}
		}
	}
	if appt.Owner != nil {
		fmt.Fprintf(&sb, "■ 飼い主名: %s\n", appt.Owner.Name)
	}
	if petNames := extractPetNamesFromCustomerFields(appt); petNames != "" {
		fmt.Fprintf(&sb, "■ ペット: %s\n", petNames)
	}
	fmt.Fprintf(&sb, "■ コース: %s\n", courseName)
	if appt.Doctor != nil {
		fmt.Fprintf(&sb, "■ 担当: %s\n", appt.Doctor.Name)
	}
	fmt.Fprintf(&sb, "■ 日時: %s〜%s\n",
		formatDateJPWithTime(appt.StartTime),
		appt.EndTime.Format("15:04"),
	)
	if appt.Notes != "" {
		fmt.Fprintf(&sb, "■ 要望: %s\n", appt.Notes)
	} else {
		sb.WriteString("■ 要望: （なし）\n")
	}
	sb.WriteString("■ 予約元: LINE\n")
	body = sb.String()
	return
}

func (s *reservationNotificationService) buildCancelledEmail(
	appt *model.Reservation,
	customer *model.LineCustomer,
) (subject, body string) {
	courseName := ""
	if appt.ReservationType != nil {
		courseName = appt.ReservationType.Name
	}
	dateStr := appt.StartTime.Format("2006年01月02日 15:04")

	subject = fmt.Sprintf("【予約キャンセル】%s 様 - %s (%s)", customerDisplayName(customer), courseName, dateStr)

	var sb strings.Builder
	sb.WriteString("以下の予約がキャンセルされました。\n\n")
	fmt.Fprintf(&sb, "■ 予約番号: R-%06d\n", appt.ID)
	if customer != nil {
		fmt.Fprintf(&sb, "■ お名前: %s\n", customer.DisplayName)
	}
	fmt.Fprintf(&sb, "■ コース: %s\n", courseName)
	fmt.Fprintf(&sb, "■ 日時: %s〜%s\n",
		formatDateJPWithTime(appt.StartTime),
		appt.EndTime.Format("15:04"),
	)
	body = sb.String()
	return
}

// ---- SMTP送信 ----

func (s *reservationNotificationService) sendEmail(ctx context.Context, to, subject, body string) error {
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

	cfg := smtpConfig{Host: s.cfg.SMTPHost, Port: s.cfg.SMTPPort, User: s.cfg.SMTPUser, Pass: s.cfg.SMTPPass}
	if err := sendSMTPMail(ctx, cfg, from, to, msg); err != nil {
		return apperrors.Wrap(err, "smtp send")
	}
	return nil
}

// ---- ユーティリティ ----

func customerDisplayName(c *model.LineCustomer) string {
	if c == nil {
		return "不明"
	}
	if c.RealName != "" {
		return c.RealName
	}
	return c.DisplayName
}

var weekdaysJP = [...]string{"日", "月", "火", "水", "木", "金", "土"}

func formatDateJPWithTime(t time.Time) string {
	w := weekdaysJP[t.Weekday()]
	return fmt.Sprintf("%s(%s) %s", t.Format("2006年01月02日"), w, t.Format("15:04"))
}

func formatDateTimeJP(start, end time.Time) string {
	w := weekdaysJP[start.Weekday()]
	return fmt.Sprintf("%s(%s) %s〜%s",
		start.Format("2006年01月02日"),
		w,
		start.Format("15:04"),
		end.Format("15:04"),
	)
}

// extractPetNamesFromCustomerFields は customer_fields JSONB からペット名一覧を抽出する。
// appt.Pet (電カル予約) が nil の場合に customer_fields.pets (LINE予約) をフォールバックとして使用。
func extractPetNamesFromCustomerFields(appt *model.Reservation) string {
	if appt.Pet != nil {
		return appt.Pet.Name
	}
	if len(appt.CustomerFields) == 0 {
		return ""
	}
	var fields struct {
		Pets []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"pets"`
	}
	if err := json.Unmarshal(appt.CustomerFields, &fields); err != nil || len(fields.Pets) == 0 {
		return ""
	}
	names := make([]string, 0, len(fields.Pets))
	for _, p := range fields.Pets {
		if p.Type != "" {
			names = append(names, fmt.Sprintf("%s(%s)", p.Name, p.Type))
		} else {
			names = append(names, p.Name)
		}
	}
	return strings.Join(names, "、")
}
