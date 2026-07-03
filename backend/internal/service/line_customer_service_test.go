package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockLineCustomerRepositoryFull は LineCustomerService テスト用のフル設定可能な
// LineCustomerRepository モック実装。
// 注意: medical_record_service_test.go に同名の mockLineCustomerRepository が
// 既に定義されている（findByIDFn のみ設定可能・AutoCreateFromReservation テスト専用）ため、
// 名前衝突を避けて別名で定義する。
type mockLineCustomerRepositoryFull struct {
	findAllFn         func(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error)
	findByIDFn        func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	updateOwnerLinkFn func(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
}

func (m *mockLineCustomerRepositoryFull) FindAll(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockLineCustomerRepositoryFull) FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockLineCustomerRepositoryFull) UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error {
	if m.updateOwnerLinkFn != nil {
		return m.updateOwnerLinkFn(ctx, clinicID, id, ownerID)
	}
	return nil
}

func (m *mockLineCustomerRepositoryFull) FindOrCreateByLineUserID(_ context.Context, _ uint64, _, _ string) (*model.LineCustomer, error) {
	return nil, nil
}

func (m *mockLineCustomerRepositoryFull) UpdateAdditionalFields(_ context.Context, _, _ uint64, _ []byte) error {
	return nil
}

func TestNewLineCustomerService(t *testing.T) {
	svc := NewLineCustomerService(&mockLineCustomerRepositoryFull{})
	assert.NotNil(t, svc)
}

func TestLineCustomerService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.LineCustomer
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns line customer list",
			repoData: []model.LineCustomer{
				{ID: 1, ClinicID: 1, LineUserID: "U1"},
				{ID: 2, ClinicID: 1, LineUserID: "U2"},
			},
			wantLen: 2,
		},
		{
			name:     "returns empty list when none exist",
			repoData: []model.LineCustomer{},
			wantLen:  0,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLineCustomerRepositoryFull{
				findAllFn: func(_ context.Context, _ uint64) ([]model.LineCustomer, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewLineCustomerService(repo)

			customers, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, customers, tt.wantLen)
			}
		})
	}
}

func TestLineCustomerService_LinkOwner(t *testing.T) {
	ownerID := uint64(5)

	tests := []struct {
		name               string
		id                 uint64
		ownerID            *uint64
		findByIDErr        error
		findByIDSecondErr  error
		updateOwnerLinkErr error
		wantErr            bool
		wantNF             bool
	}{
		{
			name:    "links owner successfully",
			id:      1,
			ownerID: &ownerID,
		},
		{
			name:    "unlinks owner when ownerID is nil",
			id:      1,
			ownerID: nil,
		},
		{
			name:        "returns not found error when line customer does not exist",
			id:          999,
			ownerID:     &ownerID,
			findByIDErr: apperrors.WrapNotFound("line_customer", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:               "returns error when UpdateOwnerLink fails",
			id:                 1,
			ownerID:            &ownerID,
			updateOwnerLinkErr: errors.New("db error"),
			wantErr:            true,
		},
		{
			name:              "returns error when refetch after link fails",
			id:                1,
			ownerID:           &ownerID,
			findByIDSecondErr: errors.New("db error on refetch"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			repo := &mockLineCustomerRepositoryFull{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
					callCount++
					if callCount == 1 {
						if tt.findByIDErr != nil {
							return nil, tt.findByIDErr
						}
						return &model.LineCustomer{ID: tt.id}, nil
					}
					if tt.findByIDSecondErr != nil {
						return nil, tt.findByIDSecondErr
					}
					return &model.LineCustomer{ID: tt.id, OwnerID: tt.ownerID}, nil
				},
				updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return tt.updateOwnerLinkErr
				},
			}
			svc := NewLineCustomerService(repo)

			result, err := svc.LinkOwner(context.Background(), 1, tt.id, tt.ownerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.ownerID, result.OwnerID)
			}
		})
	}
}
