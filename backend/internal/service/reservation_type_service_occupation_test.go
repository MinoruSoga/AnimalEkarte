package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOccupationLinkRepo は ReservationTypeOccupationRepository のテスト用モック。
// reservation_type_service_test.go の mockOccupationRepoForRType とは別に、
// FindByID を独立して差し替えられるようこのファイル専用で定義する。
type mockOccupationLinkRepo struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	findByIDFn func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	createFn   func(ctx context.Context, o *model.ReservationTypeOccupation) error
	deleteFn   func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
}

func (m *mockOccupationLinkRepo) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeOccupation{}, nil
}

func (m *mockOccupationLinkRepo) FindByID(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return &model.ReservationTypeOccupation{ID: 1, ClinicID: clinicID, ReservationTypeID: reservationTypeID, OccupationID: occupationID}, nil
}

func (m *mockOccupationLinkRepo) Create(ctx context.Context, o *model.ReservationTypeOccupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, o)
	}
	return nil
}

func (m *mockOccupationLinkRepo) Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return nil
}

func (m *mockOccupationLinkRepo) CountWorkingStaffByReservationTypeID(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 1, nil
}

func (m *mockOccupationLinkRepo) CountWorkingStaffByReservationTypeIDs(_ context.Context, _, _ uint64, dates []time.Time) (map[string]int64, error) {
	result := make(map[string]int64, len(dates))
	for _, d := range dates {
		result[d.Format("2006-01-02")] = 1
	}
	return result, nil
}

func newOccupationLinkTestService(
	repo *mockReservationTypeRepository,
	occRepo *mockOccupationLinkRepo,
	baseOccRepo *mockBaseOccupationRepo,
) ReservationTypeService {
	return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, occRepo, baseOccRepo, nil)
}

// ---- ListOccupations ----

func TestReservationTypeService_ListOccupations(t *testing.T) {
	tests := []struct {
		name        string
		findByIDErr error
		findAllFn   func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
		wantLen     int
		wantErr     bool
	}{
		{
			name: "正常: 紐付け一覧を返す",
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeOccupation, error) {
				return []model.ReservationTypeOccupation{{ID: 1}, {ID: 2}}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "エラー: 予約種別が見つからない",
			findByIDErr: errors.New("not found"),
			wantErr:     true,
		},
		{
			name: "エラー: occupationRepo.FindAll がエラー",
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeOccupation, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
				},
			}
			occRepo := &mockOccupationLinkRepo{findAllFn: tt.findAllFn}
			svc := newOccupationLinkTestService(repo, occRepo, &mockBaseOccupationRepo{})

			items, err := svc.ListOccupations(context.Background(), 10, 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
			}
		})
	}
}

// ---- LinkOccupation ----

func TestReservationTypeService_LinkOccupation(t *testing.T) {
	tests := []struct {
		name           string
		reservationErr error
		occupationErr  error
		createErr      error
		findLinkedErr  error
		wantErr        bool
	}{
		{
			name:    "正常: 紐付けを作成して返す",
			wantErr: false,
		},
		{
			name:           "エラー: 予約種別が見つからない",
			reservationErr: errors.New("not found"),
			wantErr:        true,
		},
		{
			name:          "エラー: 職種が見つからない",
			occupationErr: errors.New("not found"),
			wantErr:       true,
		},
		{
			name:      "エラー: repo.Create がエラー",
			createErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:          "エラー: 作成後の再取得がエラー",
			findLinkedErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
					if tt.reservationErr != nil {
						return nil, tt.reservationErr
					}
					return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
				},
			}
			baseOccRepo := &mockBaseOccupationRepo{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
					if tt.occupationErr != nil {
						return nil, tt.occupationErr
					}
					return &model.Occupation{ID: id, ClinicID: clinicID}, nil
				},
			}
			occRepo := &mockOccupationLinkRepo{
				createFn: func(_ context.Context, o *model.ReservationTypeOccupation) error {
					if tt.createErr != nil {
						return tt.createErr
					}
					o.ID = 100
					return nil
				},
				findByIDFn: func(_ context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
					if tt.findLinkedErr != nil {
						return nil, tt.findLinkedErr
					}
					return &model.ReservationTypeOccupation{ID: 100, ClinicID: clinicID, ReservationTypeID: reservationTypeID, OccupationID: occupationID}, nil
				},
			}

			svc := newOccupationLinkTestService(repo, occRepo, baseOccRepo)

			result, err := svc.LinkOccupation(context.Background(), 10, 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, uint64(100), result.ID)
			}
		})
	}
}

// ---- UnlinkOccupation ----

func TestReservationTypeService_UnlinkOccupation(t *testing.T) {
	tests := []struct {
		name        string
		findByIDErr error
		deleteErr   error
		wantErr     bool
	}{
		{
			name:    "正常: 紐付けを削除する",
			wantErr: false,
		},
		{
			name:        "エラー: 紐付けが見つからない",
			findByIDErr: apperrors.WrapNotFound("reservation_type_occupation", "1"),
			wantErr:     true,
		},
		{
			name:      "エラー: repo.Delete がエラー",
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{}
			occRepo := &mockOccupationLinkRepo{
				findByIDFn: func(_ context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.ReservationTypeOccupation{ID: 1, ClinicID: clinicID, ReservationTypeID: reservationTypeID, OccupationID: occupationID}, nil
				},
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := newOccupationLinkTestService(repo, occRepo, &mockBaseOccupationRepo{})

			err := svc.UnlinkOccupation(context.Background(), 10, 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
