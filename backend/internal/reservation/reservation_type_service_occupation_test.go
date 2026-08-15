package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func newOccupationLinkTestService(
	repo *mockReservationTypeRepository,
	occRepo *mockReservationTypeOccupationRepository,
	baseOccRepo *mockOccupationRepository,
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
			occRepo := &mockReservationTypeOccupationRepository{findAllFn: tt.findAllFn}
			svc := newOccupationLinkTestService(repo, occRepo, &mockOccupationRepository{})

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
			baseOccRepo := &mockOccupationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
					if tt.occupationErr != nil {
						return nil, tt.occupationErr
					}
					return &model.Occupation{ID: id, ClinicID: clinicID}, nil
				},
			}
			occRepo := &mockReservationTypeOccupationRepository{
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
			occRepo := &mockReservationTypeOccupationRepository{
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
			svc := newOccupationLinkTestService(repo, occRepo, &mockOccupationRepository{})

			err := svc.UnlinkOccupation(context.Background(), 10, 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
