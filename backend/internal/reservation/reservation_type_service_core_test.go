package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// このファイルは reservation_type_service_core.go の GetByID / Create / Update / Delete /
// validateReservationTypeParent の低カバレッジ分岐を対象とする追加テスト。
// mockReservationTypeRepository / newTestReservationTypeService は reservation_type_service_test.go で
// 定義済みのものを再利用する（同一パッケージ内のため再定義しない）。

// ---- GetByID ----

func TestReservationTypeServiceCore_GetByID_RepositoryError(t *testing.T) {
	repo := &mockReservationTypeRepository{
		findByIDWithChildrenFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	got, err := svc.GetByID(context.Background(), 1, 5)

	assert.Error(t, err)
	assert.Nil(t, got)
}

// ---- Create ----

func TestReservationTypeServiceCore_Create_EmptyNameError(t *testing.T) {
	svc := newTestReservationTypeService(&mockReservationTypeRepository{})

	got, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{Name: ""})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestReservationTypeServiceCore_Create_GetAfterCreateError(t *testing.T) {
	repo := &mockReservationTypeRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			st.ID = 100
			return nil
		},
		findByIDWithChildrenFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	got, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{Name: "診察"})

	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestReservationTypeServiceCore_Create_Defaults(t *testing.T) {
	var captured *model.ReservationType
	repo := &mockReservationTypeRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			captured = st
			st.ID = 1
			return nil
		},
	}
	svc := newTestReservationTypeService(repo)

	got, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{Name: "診察"})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, model.DayOptionNone, captured.ReservationDayOption)
	assert.Equal(t, 15, captured.DurationMinutes)
	assert.True(t, captured.ReservationVisible)
	assert.Equal(t, model.ReservationTypeCategoryGeneral, captured.Category)
}

func TestReservationTypeServiceCore_Create_ExplicitOverrides(t *testing.T) {
	var captured *model.ReservationType
	durationMinutes := 30
	reservationVisible := false
	repo := &mockReservationTypeRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			captured = st
			st.ID = 1
			return nil
		},
	}
	svc := newTestReservationTypeService(repo)

	_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{
		Name:                 "トリミング",
		ReservationDayOption: string(model.DayOptionWeekday),
		DurationMinutes:      &durationMinutes,
		ReservationVisible:   &reservationVisible,
		Category:             string(model.ReservationTypeCategoryTrimming),
	})

	assert.NoError(t, err)
	assert.Equal(t, model.DayOptionWeekday, captured.ReservationDayOption)
	assert.Equal(t, 30, captured.DurationMinutes)
	assert.False(t, captured.ReservationVisible)
	assert.Equal(t, model.ReservationTypeCategoryTrimming, captured.Category)
}

// ---- Update ----

func TestReservationTypeServiceCore_Update_NilInput(t *testing.T) {
	svc := newTestReservationTypeService(&mockReservationTypeRepository{})

	got, err := svc.Update(context.Background(), 1, 5, nil)

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestReservationTypeServiceCore_Update_FindByIDError(t *testing.T) {
	name := "新しい名前"
	repo := &mockReservationTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type", "5")
		},
	}
	svc := newTestReservationTypeService(repo)

	got, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReservationTypeServiceCore_Update_InvalidName(t *testing.T) {
	invalidName := "無効\x00文字"
	svc := newTestReservationTypeService(&mockReservationTypeRepository{})

	got, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{Name: &invalidName})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestReservationTypeServiceCore_Update_GetAfterUpdateError(t *testing.T) {
	name := "新しい名前"
	repo := &mockReservationTypeRepository{
		findByIDWithChildrenFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	got, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, got)
}

// ---- Delete ----

func TestReservationTypeServiceCore_Delete_CountChildrenError(t *testing.T) {
	repo := &mockReservationTypeRepository{
		countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	err := svc.Delete(context.Background(), 1, 5)

	assert.Error(t, err)
}

func TestReservationTypeServiceCore_Delete_CountUsageError(t *testing.T) {
	repo := &mockReservationTypeRepository{
		countUsageByReservationTypeFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	err := svc.Delete(context.Background(), 1, 5)

	assert.Error(t, err)
}

func TestReservationTypeServiceCore_Delete_RepoDeleteError(t *testing.T) {
	repo := &mockReservationTypeRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	svc := newTestReservationTypeService(repo)

	err := svc.Delete(context.Background(), 1, 5)

	assert.Error(t, err)
}

// ---- validateReservationTypeParent ----

func TestReservationTypeServiceCore_ValidateReservationTypeParent(t *testing.T) {
	t.Run("nil parentID は許可", func(t *testing.T) {
		svc := &reservationTypeService{repo: &mockReservationTypeRepository{}}
		err := svc.validateReservationTypeParent(context.Background(), 1, nil)
		assert.NoError(t, err)
	})

	t.Run("ルートノードの親は許可", func(t *testing.T) {
		parentID := uint64(2)
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id}, nil // ParentID が nil = ルート
			},
		}
		svc := &reservationTypeService{repo: repo}
		err := svc.validateReservationTypeParent(context.Background(), 1, &parentID)
		assert.NoError(t, err)
	})

	t.Run("非ルートノードの親は拒否", func(t *testing.T) {
		parentID := uint64(2)
		grandParentID := uint64(1)
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ParentID: &grandParentID}, nil
			},
		}
		svc := &reservationTypeService{repo: repo}
		err := svc.validateReservationTypeParent(context.Background(), 1, &parentID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("親が見つからない場合はエラーを伝播", func(t *testing.T) {
		parentID := uint64(99)
		repo := &mockReservationTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "99")
			},
		}
		svc := &reservationTypeService{repo: repo}
		err := svc.validateReservationTypeParent(context.Background(), 1, &parentID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
