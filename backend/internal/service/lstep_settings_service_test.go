package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LstepSettingsRepository ----

type mockLstepSettingsRepository struct {
	findByClinicAndServiceFn   func(ctx context.Context, clinicID uint64, svc string) ([]*model.ClinicIntegration, error)
	upsertFn                   func(ctx context.Context, integration *model.ClinicIntegration) error
	deleteByClinicAndServiceFn func(ctx context.Context, clinicID uint64, svc string) error
}

func (m *mockLstepSettingsRepository) FindByClinicAndService(ctx context.Context, clinicID uint64, svc string) ([]*model.ClinicIntegration, error) {
	if m.findByClinicAndServiceFn != nil {
		return m.findByClinicAndServiceFn(ctx, clinicID, svc)
	}
	return []*model.ClinicIntegration{}, nil
}
func (m *mockLstepSettingsRepository) Upsert(ctx context.Context, integration *model.ClinicIntegration) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, integration)
	}
	return nil
}
func (m *mockLstepSettingsRepository) DeleteByClinicAndService(ctx context.Context, clinicID uint64, svc string) error {
	if m.deleteByClinicAndServiceFn != nil {
		return m.deleteByClinicAndServiceFn(ctx, clinicID, svc)
	}
	return nil
}

// ---- tests ----

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockLstepSettingsRepository
		wantErr bool
	}{
		{
			name: "success empty records",
			repo: &mockLstepSettingsRepository{},
		},
		{
			name: "repo error",
			repo: &mockLstepSettingsRepository{
				findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepSettingsService(tt.repo, nil, nil)
			res, err := svc.GetSettings(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.False(t, res.IsConfigured)
			}
		})
	}
}

func TestDeleteSettings(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockLstepSettingsRepository
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockLstepSettingsRepository{},
		},
		{
			name: "repo error",
			repo: &mockLstepSettingsRepository{
				deleteByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepSettingsService(tt.repo, nil, nil)
			err := svc.DeleteSettings(context.Background(), 1, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
