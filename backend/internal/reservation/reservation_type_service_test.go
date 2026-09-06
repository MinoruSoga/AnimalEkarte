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

// mockReservationTypeRepository は ReservationTypeRepository のテスト用モック実装
type mockReservationTypeRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findAllWithChildrenFn         func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	findByIDWithChildrenFn        func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	countUsageByReservationTypeFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	countChildrenByParentIDFn     func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	createFn                      func(ctx context.Context, st *model.ReservationType) error
	updateFn                      func(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeInput) (*model.ReservationType, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockReservationTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.ReservationType{}, nil
}

func (m *mockReservationTypeRepository) FindAllWithChildren(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.findAllWithChildrenFn != nil {
		return m.findAllWithChildrenFn(ctx, clinicID)
	}
	return m.FindAll(ctx, clinicID)
}

func (m *mockReservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockReservationTypeRepository) FindByIDWithChildren(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDWithChildrenFn != nil {
		return m.findByIDWithChildrenFn(ctx, clinicID, id)
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockReservationTypeRepository) Create(ctx context.Context, st *model.ReservationType) error {
	if m.createFn != nil {
		return m.createFn(ctx, st)
	}
	return nil
}

func (m *mockReservationTypeRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeInput) (*model.ReservationType, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, cmd)
	}
	return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockReservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeRepository) CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByReservationTypeFn != nil {
		return m.countUsageByReservationTypeFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockReservationTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn != nil {
		return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
	}
	return 0, nil
}

func (m *mockReservationTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// mockUnavailableTimeRepository は ReservationTypeUnavailableTimeRepository のテスト用モック
type mockUnavailableTimeRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error)
	createFn   func(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockUnavailableTimeRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeUnavailableTime{}, nil
}
func (m *mockUnavailableTimeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeUnavailableTime{ID: id}, nil
}
func (m *mockUnavailableTimeRepository) Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}
func (m *mockUnavailableTimeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func newTestReservationTypeService(repo *mockReservationTypeRepository) ReservationTypeService {
	return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)
}

// mockAvailableSlotRepoForRType is a minimal ReservationTypeAvailableSlotRepository
// implementation used to exercise NewReservationTypeService's variadic
// availableSlotRepo parameter (only set when a caller explicitly passes one).
type mockAvailableSlotRepoForRType struct {
	findAllFn func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
}

func (m *mockAvailableSlotRepoForRType) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeAvailableSlot{}, nil
}
func (m *mockAvailableSlotRepoForRType) FindByID(_ context.Context, _, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	return &model.ReservationTypeAvailableSlot{ID: id}, nil
}
func (m *mockAvailableSlotRepoForRType) Create(_ context.Context, _ *model.ReservationTypeAvailableSlot) error {
	return nil
}
func (m *mockAvailableSlotRepoForRType) Delete(_ context.Context, _, _ uint64) error { return nil }

func TestNewReservationTypeService_WithAvailableSlotRepo(t *testing.T) {
	t.Run("variadic availableSlotRepo is wired when provided", func(t *testing.T) {
		called := false
		slotRepo := &mockAvailableSlotRepoForRType{
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeAvailableSlot, error) {
				called = true
				return []model.ReservationTypeAvailableSlot{{ID: 1}}, nil
			},
		}
		svc := NewReservationTypeService(
			&mockReservationTypeRepository{},
			&mockUnavailableTimeRepository{},
			&mockReservationTypeOccupationRepository{},
			&mockOccupationRepository{},
			nil,
			slotRepo,
		)

		slots, err := svc.ListAvailableSlots(context.Background(), 1, 5)

		assert.NoError(t, err)
		assert.True(t, called, "the explicitly provided availableSlotRepo should be used")
		assert.Len(t, slots, 1)
	})

	t.Run("omitting availableSlotRepo leaves it nil", func(t *testing.T) {
		svc := NewReservationTypeService(
			&mockReservationTypeRepository{},
			&mockUnavailableTimeRepository{},
			&mockReservationTypeOccupationRepository{},
			&mockOccupationRepository{},
			nil,
		)
		concrete, ok := svc.(*reservationTypeService)
		require.True(t, ok)
		assert.Nil(t, concrete.availableSlotRepo)
	})
}

// ---- List ----

func TestReservationTypeService_List(t *testing.T) {
	t.Run("returns service types with clinicID filter", func(t *testing.T) {
		const clinicID uint64 = 42
		want := []model.ReservationType{
			{ID: 1, ClinicID: clinicID, Name: "診察"},
			{ID: 2, ClinicID: clinicID, Name: "トリミング"},
		}
		repo := &mockReservationTypeRepository{
			findAllFn: func(_ context.Context, cid uint64) ([]model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				return want, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.List(context.Background(), clinicID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.List(context.Background(), 1)

		assert.Error(t, err)
	})
}

// ---- GetByID ----

func TestReservationTypeService_GetByID(t *testing.T) {
	t.Run("returns reservation type with children", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		childParentID := id
		want := &model.ReservationType{
			ID:       id,
			ClinicID: clinicID,
			Name:     "親区分",
			Children: []model.ReservationType{
				{ID: 6, ClinicID: clinicID, ParentID: &childParentID, Name: "子区分"},
			},
		}
		repo := &mockReservationTypeRepository{
			findByIDWithChildrenFn: func(_ context.Context, cid, rid uint64) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				return want, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.GetByID(context.Background(), clinicID, id)

		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Len(t, got.Children, 1)
	})
}

// ---- Create ----

func TestReservationTypeService_Create(t *testing.T) {
	t.Run("creates service type successfully", func(t *testing.T) {
		const clinicID uint64 = 10
		input := &CreateReservationTypeInput{
			Name:        "診察",
			Color:       "#FF0000",
			IsActive:    true,
			Description: "通常診察",
			SortOrder:   1,
		}
		var captured *model.ReservationType
		childParentID := uint64(99)
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, st *model.ReservationType) error {
				captured = st
				st.ID = 99
				return nil
			},
			findByIDWithChildrenFn: func(_ context.Context, cid, id uint64) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, uint64(99), id)
				captured.Children = []model.ReservationType{{ID: 100, ClinicID: clinicID, ParentID: &childParentID, Name: "子区分"}}
				return captured, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.Create(context.Background(), clinicID, input)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicID, got.ClinicID)
		assert.Equal(t, input.Name, got.Name)
		assert.Equal(t, input.Color, got.Color)
		assert.Equal(t, input.IsActive, got.IsActive)
		assert.Equal(t, input.Description, got.Description)
		assert.Equal(t, input.SortOrder, got.SortOrder)
		assert.Same(t, captured, got)
		assert.Len(t, got.Children, 1)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, _ *model.ReservationType) error {
				return errors.New("db error")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{Name: "診察"})

		assert.Error(t, err)
	})

	t.Run("returns 400 when max_concurrent is not positive", func(t *testing.T) {
		maxConcurrent := 0
		svc := newTestReservationTypeService(&mockReservationTypeRepository{})

		_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{
			Name:          "診察",
			MaxConcurrent: &maxConcurrent,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("rejects non-root parent", func(t *testing.T) {
		parentID := uint64(2)
		grandParentID := uint64(1)
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: clinicID, ParentID: &grandParentID}, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{
			Name:     "子区分",
			ParentID: &parentID,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("rejects parent outside clinic as not found", func(t *testing.T) {
		parentID := uint64(99)
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _ uint64, id uint64) (*model.ReservationType, error) {
				if id == parentID {
					return nil, apperrors.WrapNotFound("reservation_type", "99")
				}
				return &model.ReservationType{ID: id}, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{
			Name:     "子区分",
			ParentID: &parentID,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- Update ----

func TestReservationTypeService_Update(t *testing.T) {
	t.Run("returns 400 when no fields provided", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		repo := &mockReservationTypeRepository{}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Update(context.Background(), clinicID, id, &UpdateReservationTypeInput{})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("updates fields and returns updated record", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		newName := "新しい名前"
		newActive := false
		childParentID := id
		updated := &model.ReservationType{
			ID:       id,
			ClinicID: clinicID,
			Name:     newName,
			IsActive: newActive,
			Children: []model.ReservationType{
				{ID: 6, ClinicID: clinicID, ParentID: &childParentID, Name: "子区分"},
			},
		}

		var captured UpdateReservationTypeInput
		repo := &mockReservationTypeRepository{
			updateFn: func(_ context.Context, cid, rid uint64, cmd UpdateReservationTypeInput) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				captured = cmd
				return updated, nil
			},
			findByIDWithChildrenFn: func(_ context.Context, cid, rid uint64) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				return updated, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.Update(context.Background(), clinicID, id, &UpdateReservationTypeInput{
			Name:     &newName,
			IsActive: &newActive,
		})

		require.NoError(t, err)
		assert.Equal(t, updated, got)
		require.NotNil(t, captured.Name)
		require.NotNil(t, captured.IsActive)
		assert.Equal(t, newName, *captured.Name)
		assert.Equal(t, newActive, *captured.IsActive)
		assert.Nil(t, captured.Color)
	})

	t.Run("rejects clear_parent_id and parent_id together", func(t *testing.T) {
		parentID := uint64(10)
		svc := newTestReservationTypeService(&mockReservationTypeRepository{})

		_, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{
			ParentID:      &parentID,
			ClearParentID: true,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("returns not found when record does not exist", func(t *testing.T) {
		name := "test"
		repo := &mockReservationTypeRepository{
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateReservationTypeInput) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "5")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{Name: &name})

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("rejects self parent", func(t *testing.T) {
		const id uint64 = 5
		parentID := id
		svc := newTestReservationTypeService(&mockReservationTypeRepository{})

		_, err := svc.Update(context.Background(), 1, id, &UpdateReservationTypeInput{ParentID: &parentID})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("returns 400 when update max_concurrent is not positive", func(t *testing.T) {
		maxConcurrent := 0
		svc := newTestReservationTypeService(&mockReservationTypeRepository{})

		_, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{MaxConcurrent: &maxConcurrent})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("rejects moving parent with children under another parent", func(t *testing.T) {
		parentID := uint64(10)
		repo := &mockReservationTypeRepository{
			countChildrenByParentIDFn: func(_ context.Context, _ uint64, parentID uint64) (int64, error) {
				if parentID == 5 {
					return 1, nil
				}
				return 0, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{ParentID: &parentID})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

// ---- Delete ----

func TestReservationTypeService_Delete(t *testing.T) {
	t.Run("deletes successfully", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		repo := &mockReservationTypeRepository{
			deleteFn: func(_ context.Context, cid, rid uint64) error {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				return nil
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), clinicID, id)

		assert.NoError(t, err)
	})

	t.Run("returns not found when record does not exist", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "5")
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), 1, 5)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("子予約区分がある親は削除できない", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), 1, 2)

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("使用中の予約種別は削除できない", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			countUsageByReservationTypeFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), 1, 2)

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})
}

// ---- Reorder ----

func TestReservationTypeService_Reorder(t *testing.T) {
	tests := []struct {
		name        string
		ids         []uint64
		repoErr     error
		wantErr     bool
		wantInvalid bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns error when ids is empty",
			ids:         []uint64{},
			repoErr:     nil,
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:        "returns error when ids is nil",
			ids:         nil,
			repoErr:     nil,
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2, 3},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns invalid error when id not found in clinic",
			ids:     []uint64{1, 999},
			repoErr: apperrors.WrapInvalidInput("reservation_type id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestReservationTypeService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- DeleteUnavailableTime ----

func TestReservationTypeService_DeleteUnavailableTime(t *testing.T) {
	t.Run("正常: FindByID → Delete が呼ばれる", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		err := svc.DeleteUnavailableTime(context.Background(), 1, 10, 5)

		assert.NoError(t, err)
	})

	t.Run("エラー: 親 ReservationType が存在しない → 404", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
		}
		svc := NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		err := svc.DeleteUnavailableTime(context.Background(), 1, 10, 5)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("エラー: unavailableTime が存在しない → 404", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationTypeUnavailableTime, error) {
				return nil, apperrors.WrapNotFound("reservation_type_unavailable_time", "5")
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		err := svc.DeleteUnavailableTime(context.Background(), 1, 10, 5)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("エラー: repo.Delete がエラー → error を返す", func(t *testing.T) {
		unavailableRepo := &mockUnavailableTimeRepository{
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
		}
		svc := NewReservationTypeService(&mockReservationTypeRepository{}, unavailableRepo, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil)

		err := svc.DeleteUnavailableTime(context.Background(), 1, 10, 5)

		assert.Error(t, err)
	})
}
