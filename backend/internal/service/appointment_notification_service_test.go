package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockLineSettingRepo struct {
	repository.LineReservationSettingRepository
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

func (m *mockLineSettingRepo) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{
		ClinicID:          clinicID,
		LineAccessToken:   "dummy-token",
		NotificationEmail: "notify@example.com",
	}, nil
}

func TestReservationNotificationService_NotifyCreated(t *testing.T) {
	ctx := context.Background()

	t.Run("success notifying created", func(t *testing.T) {
		repo := &mockLineSettingRepo{}
		cfg := &ReservationNotificationConfig{
			SMTPHost: "",
		}
		svc := NewReservationNotificationService(cfg, repo, nil)

		appt := &model.Reservation{
			ID:        123,
			ClinicID:  1,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Pet:       &model.Pet{Name: "ポチ"},
			Doctor:    &model.Staff{Name: "獣医師A"},
			ReservationType: &model.ReservationType{
				Name: "一般診療",
			},
			CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}]}`),
		}
		customer := &model.LineCustomer{
			LineUserID:       "U12345",
			DisplayName:      "山田太郎",
			RealName:         "山田太郎本名",
			AdditionalFields: []byte(`{"phone": "090-0000-0000", "note": "かゆみ"}`),
		}

		svc.NotifyCreated(ctx, appt, customer)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("db setting load error", func(t *testing.T) {
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db load error")
			},
		}
		cfg := &ReservationNotificationConfig{}
		svc := NewReservationNotificationService(cfg, repo, nil)

		appt := &model.Reservation{
			ID:       123,
			ClinicID: 1,
		}

		svc.NotifyCreated(ctx, appt, nil)
		time.Sleep(50 * time.Millisecond)
	})
}

func TestReservationNotificationService_NotifyCancelled(t *testing.T) {
	ctx := context.Background()

	t.Run("success notifying cancelled", func(t *testing.T) {
		repo := &mockLineSettingRepo{}
		cfg := &ReservationNotificationConfig{
			SMTPHost: "",
		}
		svc := NewReservationNotificationService(cfg, repo, nil)

		appt := &model.Reservation{
			ID:        123,
			ClinicID:  1,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
		}
		customer := &model.LineCustomer{
			LineUserID:  "U12345",
			DisplayName: "山田太郎",
		}

		svc.NotifyCancelled(ctx, appt, customer)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("db setting load error", func(t *testing.T) {
		repo := &mockLineSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db load error")
			},
		}
		cfg := &ReservationNotificationConfig{}
		svc := NewReservationNotificationService(cfg, repo, nil)

		appt := &model.Reservation{
			ID:       123,
			ClinicID: 1,
		}

		svc.NotifyCancelled(ctx, appt, nil)
		time.Sleep(50 * time.Millisecond)
	})
}

// TestReservationNotificationService_BuildCancelledLineMessage は buildCancelledLineMessage の
// ReservationType の有無による分岐（メニュー行を含めるか）を直接検証する。
func TestReservationNotificationService_BuildCancelledLineMessage(t *testing.T) {
	svc := &reservationNotificationService{}
	start := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	t.Run("includes menu line when ReservationType is set", func(t *testing.T) {
		appt := &model.Reservation{
			ID: 42, StartTime: start, EndTime: end,
			ReservationType: &model.ReservationType{Name: "一般診療"},
		}
		msg := svc.buildCancelledLineMessage(appt)
		assert.Contains(t, msg, "R-000042")
		assert.Contains(t, msg, "■ メニュー: 一般診療")
		assert.Contains(t, msg, "再度のご予約はLINEメニューの")
	})

	t.Run("omits menu line when ReservationType is nil", func(t *testing.T) {
		appt := &model.Reservation{ID: 43, StartTime: start, EndTime: end}
		msg := svc.buildCancelledLineMessage(appt)
		assert.Contains(t, msg, "R-000043")
		assert.NotContains(t, msg, "■ メニュー:")
	})
}

// TestReservationNotificationService_SendEmail は SMTP 未設定時のスキップ、
// auth 分岐（SMTPUser 設定時）、接続失敗時のラップされたエラー返却を検証する。
// smtp.SendMail は実ネットワークの成功応答をモックできないため、成功パスは
// SMTPHost="" のスキップ分岐で代替する（本番コード変更なしでのテスト可能な範囲）。
func TestReservationNotificationService_SendEmail(t *testing.T) {
	t.Run("skips send when SMTPHost is empty", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{SMTPHost: ""}}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.NoError(t, err)
	})

	t.Run("wraps error when SMTP connection fails without auth", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: "65533",
			SMTPFrom: "from@example.com",
		}}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.Error(t, err)
	})

	t.Run("wraps error when SMTP connection fails with auth configured", func(t *testing.T) {
		svc := &reservationNotificationService{cfg: ReservationNotificationConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: "65533",
			SMTPUser: "user",
			SMTPPass: "pass",
			SMTPFrom: "from@example.com",
		}}
		err := svc.sendEmail(context.Background(), "to@example.com", "subject", "body")
		assert.Error(t, err)
	})
}

// TestExtractPetNamesFromCustomerFields は appt.Pet 優先・customer_fields フォールバック・
// 不正JSON・空配列・type有無混在の各分岐を検証する。
func TestExtractPetNamesFromCustomerFields(t *testing.T) {
	tests := []struct {
		name string
		appt *model.Reservation
		want string
	}{
		{
			name: "appt.Pet takes priority over customer_fields",
			appt: &model.Reservation{
				Pet:            &model.Pet{Name: "ポチ"},
				CustomerFields: []byte(`{"pets": [{"name": "タマ", "type": "猫"}]}`),
			},
			want: "ポチ",
		},
		{
			name: "empty customer_fields returns empty string",
			appt: &model.Reservation{CustomerFields: nil},
			want: "",
		},
		{
			name: "invalid JSON returns empty string",
			appt: &model.Reservation{CustomerFields: []byte(`not-json`)},
			want: "",
		},
		{
			name: "empty pets array returns empty string",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": []}`)},
			want: "",
		},
		{
			name: "single pet with type",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}]}`)},
			want: "ポチ(犬)",
		},
		{
			name: "single pet without type",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ"}]}`)},
			want: "ポチ",
		},
		{
			name: "multiple pets joined with japanese comma",
			appt: &model.Reservation{CustomerFields: []byte(`{"pets": [{"name": "ポチ", "type": "犬"}, {"name": "タマ"}]}`)},
			want: "ポチ(犬)、タマ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPetNamesFromCustomerFields(tt.appt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationHelpers(t *testing.T) {
	t.Run("reservationTypeDisplayName", func(t *testing.T) {
		assert.Equal(t, "", reservationTypeDisplayName(nil))
		assert.Equal(t, "disp", reservationTypeDisplayName(&model.ReservationType{ReservationDisplayName: "disp"}))
		assert.Equal(t, "short", reservationTypeDisplayName(&model.ReservationType{ShowShortName: true, ShortName: "short"}))
		assert.Equal(t, "name", reservationTypeDisplayName(&model.ReservationType{Name: "name"}))
	})

	t.Run("staffDisplayName", func(t *testing.T) {
		assert.Equal(t, "", staffDisplayName(nil))
		assert.Equal(t, "disp", staffDisplayName(&model.Staff{ReservationDisplayName: "disp"}))
		assert.Equal(t, "name", staffDisplayName(&model.Staff{Name: "name"}))
	})

	t.Run("customerDisplayName", func(t *testing.T) {
		assert.Equal(t, "不明", customerDisplayName(nil))
		assert.Equal(t, "real", customerDisplayName(&model.LineCustomer{RealName: "real", DisplayName: "disp"}))
		assert.Equal(t, "disp", customerDisplayName(&model.LineCustomer{DisplayName: "disp"}))
	})
}
