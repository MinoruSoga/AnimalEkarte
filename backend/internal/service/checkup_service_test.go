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

// ---- Checkup モック ----

type mockCheckupRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn          func(ctx context.Context, clinicID uint64, filters repository.CheckupFilters) ([]model.Checkup, error)
	findByOwnerIDFn         func(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	findByIDFn              func(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error)
	findAlertsFn            func(ctx context.Context, clinicID uint64, withinDays int) ([]model.Checkup, error)
	createFn                func(ctx context.Context, checkup *model.Checkup) error
	updateFn                func(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, checkupID uint64) error
}

func (m *mockCheckupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockCheckupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters repository.CheckupFilters) ([]model.Checkup, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, filters)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindByID(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, checkupID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	if m.findByOwnerIDFn != nil {
		return m.findByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) FindAlerts(ctx context.Context, clinicID uint64, withinDays int) ([]model.Checkup, error) {
	if m.findAlertsFn != nil {
		return m.findAlertsFn(ctx, clinicID, withinDays)
	}
	return nil, nil
}

func (m *mockCheckupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	return m.createFn(ctx, checkup)
}

func (m *mockCheckupRepository) Update(ctx context.Context, clinicID, checkupID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, checkupID, fields)
}

func (m *mockCheckupRepository) Delete(ctx context.Context, clinicID, checkupID uint64) error {
	return m.deleteFn(ctx, clinicID, checkupID)
}

// ---- Tests ----

func TestCheckupService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoCheckups    []model.Checkup
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns checkups for medical record",
			medicalRecordID: 1,
			repoCheckups: []model.Checkup{
				{ID: 1, MedicalRecordID: 1, CheckupTypeID: 1, Result: "Normal"},
				{ID: 2, MedicalRecordID: 1, CheckupTypeID: 2, Result: "Abnormal"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no checkups exist",
			medicalRecordID: 999,
			repoCheckups:    []model.Checkup{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoCheckups:    nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return tt.repoCheckups, tt.repoErr
				},
			}
			svc := NewCheckupService(repo, nil)

			checkups, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, checkups, tt.wantLen)
			}
		})
	}
}

func TestCheckupService_Create(t *testing.T) {
	now := time.Now()
	petID := uint64(5)
	doctorID := uint64(10)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateCheckupInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates checkup successfully",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
				PetID:         &petID,
				DoctorID:      &doctorID,
				Result:        "Normal findings",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates checkup with minimal fields",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				createFn: func(_ context.Context, _ *model.Checkup) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{ID: 1, MedicalRecordID: tt.medicalRecordID}, nil
				},
			}
			svc := NewCheckupService(repo, nil)

			checkup, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checkup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Update(t *testing.T) {
	newResult := "Updated result"
	newCheckupTypeID := uint64(2)

	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		input                      *UpdateCheckupInput
		repoCheckupMedicalRecordID uint64
		repoUpdateErr              error
		repoReturnCheckup          *model.Checkup
		wantErr                    bool
	}{
		{
			name:            "updates checkup successfully",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result:        &newResult,
				CheckupTypeID: &newCheckupTypeID,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 1,
				Result:          newResult,
			},
			wantErr: false,
		},
		{
			name:                       "returns error when no fields provided",
			medicalRecordID:            1,
			checkupID:                  1,
			input:                      &UpdateCheckupInput{},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
		{
			name:            "returns error when checkup doesn't belong to medical record",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 2,
			},
			wantErr: true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              errors.New("db error"),
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewCheckupService(repo, nil)

			checkup, err := svc.Update(context.Background(), 1, tt.medicalRecordID, tt.checkupID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Delete(t *testing.T) {
	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		repoCheckupMedicalRecordID uint64
		repoDeleteErr              error
		wantErr                    bool
	}{
		{
			name:                       "deletes checkup successfully",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              nil,
			wantErr:                    false,
		},
		{
			name:                       "returns error when checkup doesn't belong to medical record",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoDeleteErr:              nil,
			wantErr:                    true,
		},
		{
			name:                       "returns error when delete fails",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              errors.New("db error"),
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoDeleteErr
				},
			}
			svc := NewCheckupService(repo, nil)

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, tt.checkupID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAlerts_RejectsInvalidWithinDays(t *testing.T) {
	svc := NewCheckupService(&mockCheckupRepository{}, nil)

	for _, days := range []int{0, -1, 366} {
		result, err := svc.GetAlerts(context.Background(), 1, days)
		assert.Error(t, err, "days=%d should be rejected", days)
		assert.Nil(t, result, "days=%d should return nil result", days)
	}
}

func TestGetAlerts_SeparatesOverdueAndUpcoming(t *testing.T) {
	fixedNow := time.Date(2026, 5, 6, 12, 0, 0, 0, jstLocation())
	yesterdayJST := time.Date(2026, 5, 5, 0, 0, 0, 0, jstLocation())
	todayJST := time.Date(2026, 5, 6, 0, 0, 0, 0, jstLocation())
	inThirtyJST := time.Date(2026, 6, 5, 0, 0, 0, 0, jstLocation())

	repo := &mockCheckupRepository{
		findAlertsFn: func(_ context.Context, _ uint64, _ int) ([]model.Checkup, error) {
			return []model.Checkup{
				{ID: 1, NextDate: &yesterdayJST},
				{ID: 2, NextDate: &todayJST},
				{ID: 3, NextDate: &inThirtyJST},
			}, nil
		},
	}
	svc := NewCheckupService(repo, nil)
	svc.(*checkupService).nowFn = func() time.Time { return fixedNow }

	result, err := svc.GetAlerts(context.Background(), 1, 30)
	assert.NoError(t, err)
	assert.Len(t, result.Overdue, 1, "yesterday should be overdue")
	assert.Len(t, result.Upcoming, 2, "today and inThirty should be upcoming")
	assert.Equal(t, uint64(1), result.Overdue[0].ID)
}

func TestGetAlerts_NilNextDateExcluded(t *testing.T) {
	repo := &mockCheckupRepository{
		findAlertsFn: func(_ context.Context, _ uint64, _ int) ([]model.Checkup, error) {
			return []model.Checkup{
				{ID: 1, NextDate: nil},
			}, nil
		},
	}
	svc := NewCheckupService(repo, nil)
	svc.(*checkupService).nowFn = func() time.Time {
		return time.Date(2026, 5, 6, 12, 0, 0, 0, jstLocation())
	}

	result, err := svc.GetAlerts(context.Background(), 1, 30)
	assert.NoError(t, err)
	assert.Empty(t, result.Overdue)
	assert.Empty(t, result.Upcoming)
}

func TestGetAlerts_PropagatesRepositoryError(t *testing.T) {
	repo := &mockCheckupRepository{
		findAlertsFn: func(_ context.Context, _ uint64, _ int) ([]model.Checkup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(repo, nil)
	svc.(*checkupService).nowFn = func() time.Time {
		return time.Date(2026, 5, 6, 12, 0, 0, 0, jstLocation())
	}

	result, err := svc.GetAlerts(context.Background(), 1, 30)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAlerts_JSTLateNight_TodayIsCorrectJSTDate(t *testing.T) {
	// 2026-05-06 03:00 JST = 2026-05-05 18:00 UTC
	// 旧実装 Truncate(24h): today = 2026-05-05 00:00 UTC
	//   → next_date (2026-05-05 00:00 UTC) は today と等しいので Before = false → Upcoming (誤)
	// 新実装 (JST 基準): today = 2026-05-06 00:00 JST = 2026-05-05 15:00 UTC
	//   → next_date (2026-05-05 00:00 UTC) < today → Overdue (正)
	jstLateNight := time.Date(2026, 5, 6, 3, 0, 0, 0, jstLocation())
	may5UTC := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	repo := &mockCheckupRepository{
		findAlertsFn: func(_ context.Context, _ uint64, _ int) ([]model.Checkup, error) {
			return []model.Checkup{
				{ID: 1, NextDate: &may5UTC},
			}, nil
		},
	}
	svc := NewCheckupService(repo, nil)
	svc.(*checkupService).nowFn = func() time.Time { return jstLateNight }

	result, err := svc.GetAlerts(context.Background(), 1, 30)
	assert.NoError(t, err)
	assert.Len(t, result.Overdue, 1, "2026-05-05 00:00 UTC should be overdue when JST date is 2026-05-06")
	assert.Empty(t, result.Upcoming)
	assert.Equal(t, uint64(1), result.Overdue[0].ID)
}
