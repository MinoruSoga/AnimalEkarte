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
		svc := NewReservationNotificationService(cfg, repo)

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
		svc := NewReservationNotificationService(cfg, repo)

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
		svc := NewReservationNotificationService(cfg, repo)

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
		svc := NewReservationNotificationService(cfg, repo)

		appt := &model.Reservation{
			ID:       123,
			ClinicID: 1,
		}

		svc.NotifyCancelled(ctx, appt, nil)
		time.Sleep(50 * time.Millisecond)
	})
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
