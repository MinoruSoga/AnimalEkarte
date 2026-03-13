package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockTrimmingRepository は TrimmingRepository のテスト用モック実装
type mockTrimmingRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64, petID *uint64, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
	createFn   func(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	updateFn   func(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockTrimmingRepository) FindAll(ctx context.Context, clinicID uint64, petID *uint64, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, page, limit)
}

func (m *mockTrimmingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingRepository) Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	return m.createFn(ctx, clinicID, trimming)
}

func (m *mockTrimmingRepository) Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	return m.updateFn(ctx, clinicID, trimming)
}

func (m *mockTrimmingRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func TestTrimmingService_List(t *testing.T) {
	petID := uint64(10)
	ownerID := uint64(5)

	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		page      int
		limit     int
		repoData  []model.TrimmingRecord
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns trimming list with total count",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoData: []model.TrimmingRecord{
				{ID: 1, ClinicID: 1, Date: time.Now()},
				{ID: 2, ClinicID: 1, Date: time.Now()},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "filters by pet ID",
			clinicID:  1,
			petID:     &petID,
			ownerID:   nil,
			page:      1,
			limit:     20,
			repoData:  []model.TrimmingRecord{{ID: 1, ClinicID: 1, PetID: &petID}},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "filters by owner ID",
			clinicID:  1,
			petID:     nil,
			ownerID:   &ownerID,
			page:      1,
			limit:     20,
			repoData:  []model.TrimmingRecord{{ID: 1, ClinicID: 1}},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no records exist",
			clinicID:  1,
			petID:     nil,
			ownerID:   nil,
			page:      1,
			limit:     20,
			repoData:  []model.TrimmingRecord{},
			repoTotal: 0,
			repoErr:   nil,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			petID:     nil,
			ownerID:   nil,
			page:      1,
			limit:     20,
			repoData:  nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _ int) ([]model.TrimmingRecord, int64, error) {
					return tt.repoData, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewTrimmingService(repo)

			trimmings, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, trimmings, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestTrimmingService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		repoRecord   *model.TrimmingRecord
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:       "returns trimming record when found",
			clinicID:   1,
			id:         10,
			repoRecord: &model.TrimmingRecord{ID: 10, ClinicID: 1},
			repoErr:    nil,
			wantErr:    false,
			wantNotFound: false,
		},
		{
			name:       "returns not found error when record does not exist",
			clinicID:   1,
			id:         999,
			repoRecord: nil,
			repoErr:    apperrors.WrapNotFound("trimming_record", "999"),
			wantErr:    true,
			wantNotFound: true,
		},
		{
			name:       "returns error on repository failure",
			clinicID:   1,
			id:         10,
			repoRecord: nil,
			repoErr:    errors.New("db error"),
			wantErr:    true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingRecord, error) {
					return tt.repoRecord, tt.repoErr
				},
			}
			svc := NewTrimmingService(repo)

			record, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoRecord, record)
			}
		})
	}
}

func TestTrimmingService_Create(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		trimming *model.TrimmingRecord
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates trimming record successfully",
			clinicID: 1,
			trimming: &model.TrimmingRecord{
				Date:     time.Now(),
				StaffID:  1,
				CourseID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			trimming: &model.TrimmingRecord{
				Date:     time.Now(),
				StaffID:  1,
				CourseID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingRepository{
				createFn: func(_ context.Context, _ uint64, _ *model.TrimmingRecord) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingService(repo)

			err := svc.Create(context.Background(), tt.clinicID, tt.trimming)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingService_Update(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		trimming *model.TrimmingRecord
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "updates trimming record successfully",
			clinicID: 1,
			trimming: &model.TrimmingRecord{
				ID:       1,
				StaffID:  2,
				CourseID: 1,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:     "returns not found error when record does not exist",
			clinicID: 1,
			trimming: &model.TrimmingRecord{
				ID:       999,
				StaffID:  1,
				CourseID: 1,
			},
			repoErr: apperrors.WrapNotFound("trimming_record", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			trimming: &model.TrimmingRecord{
				ID:       1,
				StaffID:  1,
				CourseID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingRepository{
				updateFn: func(_ context.Context, _ uint64, _ *model.TrimmingRecord) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingService(repo)

			err := svc.Update(context.Background(), tt.clinicID, tt.trimming)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes trimming record successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when record does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("trimming_record", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
