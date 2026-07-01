package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockLineReservationSettingRepository struct {
	repository.LineReservationSettingRepository
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	findAllFn        func(ctx context.Context) ([]model.LineReservationSetting, error)
	saveFn           func(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error
}

func (m *mockLineReservationSettingRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{ClinicID: clinicID}, nil
}

func (m *mockLineReservationSettingRepository) FindAll(ctx context.Context) ([]model.LineReservationSetting, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []model.LineReservationSetting{}, nil
}

func (m *mockLineReservationSettingRepository) Save(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, setting)
	}
	return nil
}

func TestLineReservationSettingService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		res, err := svc.Get(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("not found returns nil without error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
		}
		svc := NewLineReservationSettingService(repo)
		res, err := svc.Get(ctx, 1)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, err := svc.Get(ctx, 1)
		assert.Error(t, err)
	})
}

func TestLineReservationSettingService_Save(t *testing.T) {
	ctx := context.Background()

	t.Run("success create new setting", func(t *testing.T) {
		var callCount int
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				callCount++
				if callCount == 1 {
					return nil, apperrors.WrapNotFound("setting", "")
				}
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			Status:            "active",
			ClosedWeekdays:    []byte("[1, 2]"),
			LineChannelSecret: "secret",
			LineAccessToken:   "token",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
	})

	t.Run("success update existing setting keeping sensitive keys", func(t *testing.T) {
		existing := &model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: "old_secret",
			LineAccessToken:   "old_token",
		}
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return existing, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			Status:            "active",
			ClosedWeekdays:    []byte("[1, 2]"),
			LineChannelSecret: "",
			LineAccessToken:   "",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.NotNil(t, res)
	})

	t.Run("invalid json fields", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil)
		input := &UpsertLineReservationSettingInput{
			ClosedWeekdays: []byte("bad-json"),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("db find error on save", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})

	t.Run("db save error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
				return errors.New("save error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})

	t.Run("db reload error after save", func(t *testing.T) {
		var findCount int
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				findCount++
				if findCount > 1 {
					return nil, errors.New("db load error")
				}
				return nil, apperrors.WrapNotFound("setting", "")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})
}
