package reservation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/timeutil"
)

// ReservationNotifier は予約確定・キャンセル時の通知インターフェース。
// 実装はすべてのエラーをログに記録し、呼び出し元には返さない（fire-and-forget）。
type ReservationNotifier interface {
	// NotifyCreated は予約確定後に LINE + メールを送信する。
	NotifyCreated(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
	// NotifyCancelled は予約キャンセル後に LINE + メールを送信する。
	NotifyCancelled(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
	// Wait は送信中の通知 goroutine（fire-and-forget）の完了を待つ。
	// graceful shutdown で server.Shutdown(ctx) の後に呼び出し、goroutine の孤児化を防ぐ（BE-refactor.md B-1）。
	Wait()
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

// SMTPConfig は SMTP 送信設定（service/smtp_sender.go の同名 unexported struct と同形）。
type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
}

type reservationNotificationService struct {
	cfg         ReservationNotificationConfig
	settingRepo lineReservationSettingFinder
	// sendMail は auth 側実装の closure 注入（topo で reservation からの import 禁止のため。
	// 具象はcomposition rootが包んで渡す）。
	sendMail func(ctx context.Context, cfg SMTPConfig, from, to string, msg []byte) error
	// wg は通知 goroutine（fire-and-forget）を追跡する（BE-refactor.md B-1）。
	wg sync.WaitGroup
}

// Wait は送信中の通知 goroutine の完了を待つ（BE-refactor.md B-1）。
func (s *reservationNotificationService) Wait() {
	s.wg.Wait()
}

// NewReservationNotificationService は通知サービスを初期化して返す。
// SMTP設定が空の場合はメール送信をスキップする。
// 通知先メールは予約のクリニック設定（DB）から都度取得する。
// EMR-504: 予約確認・変更の LINE Push は配送記録（dispatch intent / attempt / result）と目的別の
// 送信条件を経由する契約で、その writer が未接続の間は送らない（skipLineNotification）。
func NewReservationNotificationService(
	cfg *ReservationNotificationConfig,
	settingRepo lineReservationSettingFinder,
	sendMail func(ctx context.Context, cfg SMTPConfig, from, to string, msg []byte) error,
) ReservationNotifier {
	return &reservationNotificationService{
		cfg:         *cfg,
		settingRepo: settingRepo,
		sendMail:    sendMail,
	}
}

// skipLineNotification は LINE Push を行わない（fail-closed）。直接 Push は配送記録・本人/宛先・
// 独立停止・結果不明時の照合保留の条件を満たせないため、line_dispatch_* の writer が接続されるまで
// 停止する（apperrors.CodeLineDispatchPending）。メール通知は別経路で継続する。
func skipLineNotification(ctx context.Context, event string, reservationID uint64) {
	slog.ErrorContext(ctx, "LINE notify skipped: dispatch writer not connected",
		"event", event,
		"reservation_id", reservationID,
		"code", apperrors.CodeLineDispatchPending,
	)
}

// ---- NotifyCreated ----

func (s *reservationNotificationService) NotifyCreated(
	ctx context.Context,
	appt *model.Reservation,
	customer *model.LineCustomer,
) {
	emailSubject, emailBody := s.buildCreatedEmail(appt, customer)
	clinicID := appt.ClinicID

	s.wg.Add(1)
	sharedkernel.GoSafe("reservation notify created", func() { //nolint:contextcheck,gosec // 意図的: 通知はリクエスト完了後も継続するため独立した background context を使用
		defer s.wg.Done()
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// クリニックごとの設定を DB から取得（通知先メール）
		setting, err := s.settingRepo.FindByClinicID(bgCtx, clinicID)
		if err != nil {
			slog.ErrorContext(bgCtx, "notify created: failed to load clinic setting",
				"clinic_id", clinicID,
				"reservation_id", appt.ID,
				"error", err.Error(),
			)
			return
		}

		skipLineNotification(bgCtx, "created", appt.ID)

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
	emailSubject, emailBody := s.buildCancelledEmail(appt, customer)
	clinicID := appt.ClinicID

	s.wg.Add(1)
	sharedkernel.GoSafe("reservation notify cancelled", func() { //nolint:contextcheck,gosec // 意図的: 通知はリクエスト完了後も継続するため独立した background context を使用
		defer s.wg.Done()
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// クリニックごとの設定を DB から取得（通知先メール）
		setting, err := s.settingRepo.FindByClinicID(bgCtx, clinicID)
		if err != nil {
			slog.ErrorContext(bgCtx, "notify cancelled: failed to load clinic setting",
				"clinic_id", clinicID,
				"reservation_id", appt.ID,
				"error", err.Error(),
			)
			return
		}

		skipLineNotification(bgCtx, "cancelled", appt.ID)

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
			appendCustomerAdditionalFields(&sb, customer.AdditionalFields)
		}
	}
	if appt.Owner != nil {
		fmt.Fprintf(&sb, "■ 飼主名: %s\n", appt.Owner.Name)
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

	cfg := SMTPConfig{Host: s.cfg.SMTPHost, Port: s.cfg.SMTPPort, User: s.cfg.SMTPUser, Pass: s.cfg.SMTPPass}
	if err := s.sendMail(ctx, cfg, from, to, msg); err != nil {
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

func formatDateJPWithTime(t time.Time) string {
	w := timeutil.WeekdayJP(t)
	return fmt.Sprintf("%s(%s) %s", t.Format("2006年01月02日"), w, t.Format("15:04"))
}

func formatDateTimeJP(start, end time.Time) string {
	w := timeutil.WeekdayJP(start)
	return fmt.Sprintf("%s(%s) %s〜%s",
		start.Format("2006年01月02日"),
		w,
		start.Format("15:04"),
		end.Format("15:04"),
	)
}

func appendCustomerAdditionalFields(sb *strings.Builder, raw json.RawMessage) {
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return
	}
	phone, _ := fields["phone"].(string)
	if phone != "" {
		fmt.Fprintf(sb, "■ 電話番号: %s\n", phone)
	}
	note, _ := fields["note"].(string)
	if note != "" {
		fmt.Fprintf(sb, "■ 診察内容: %s\n", note)
	}
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
