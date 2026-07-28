package pet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockPetChronicConditionRepository struct {
	ChronicConditionRepository
	findByPetIDFn                    func(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	findByIDFn                       func(ctx context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error)
	createFn                         func(ctx context.Context, record *model.PetChronicCondition) error
	updateFn                         func(ctx context.Context, clinicID, petID, id uint64, fields map[string]any) error
	deleteFn                         func(ctx context.Context, clinicID, petID, id uint64) error
	getActiveConditionCodesByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) ([]string, error)
}

func (m *mockPetChronicConditionRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	if m.findByPetIDFn != nil {
		return m.findByPetIDFn(ctx, clinicID, petID)
	}
	return []model.PetChronicCondition{}, nil
}

func (m *mockPetChronicConditionRepository) FindByID(
	ctx context.Context,
	clinicID, petID, id uint64,
) (*model.PetChronicCondition, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, petID, id)
	}
	return &model.PetChronicCondition{ID: id, ClinicID: clinicID, PetID: petID}, nil
}

func (m *mockPetChronicConditionRepository) Create(ctx context.Context, record *model.PetChronicCondition) error {
	if m.createFn != nil {
		return m.createFn(ctx, record)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) Update(
	ctx context.Context,
	clinicID, petID, id uint64,
	fields map[string]any,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, petID, id, fields)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) Delete(ctx context.Context, clinicID, petID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, petID, id)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) FindActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error) {
	if m.getActiveConditionCodesByOwnerFn != nil {
		return m.getActiveConditionCodesByOwnerFn(ctx, clinicID, ownerID)
	}
	return []string{}, nil
}

// chronicTagSyncOverride は *mockLstepTagSyncService（lstep_lifecycle_service_test.go）を
// 埋め込み、SyncChronicConditionTags のみを差し替え可能にする小さなラッパー。
// mockLstepTagSyncService.SyncChronicConditionTags は常に nil を返す固定実装のため、
// エラー注入テストにはこのラッパーが必要（対象外ファイルを編集せずにテストする）。
type chronicTagSyncOverride struct {
	*mockLstepTagSyncService
	syncChronicConditionTagsFn func(ctx context.Context, clinicID, ownerID uint64, codes []string) error
}

func (m *chronicTagSyncOverride) SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, codes []string) error {
	if m.syncChronicConditionTagsFn != nil {
		return m.syncChronicConditionTagsFn(ctx, clinicID, ownerID, codes)
	}
	return nil
}

func TestBuildChronicConditionUpdateFields(t *testing.T) {
	code := "C02"
	name := "疾患名"
	diagnosedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	notes := "メモ"
	isActive := true

	tests := []struct {
		name  string
		input UpdateChronicConditionInput
		want  map[string]any
	}{
		{
			name:  "no fields set returns empty map",
			input: UpdateChronicConditionInput{},
			want:  map[string]any{},
		},
		{
			name:  "only condition_code set",
			input: UpdateChronicConditionInput{ConditionCode: &code},
			want:  map[string]any{"condition_code": code},
		},
		{
			name:  "only condition_name set",
			input: UpdateChronicConditionInput{ConditionName: &name},
			want:  map[string]any{"condition_name": name},
		},
		{
			name:  "only diagnosed_at set",
			input: UpdateChronicConditionInput{DiagnosedAt: &diagnosedAt},
			want:  map[string]any{"diagnosed_at": diagnosedAt},
		},
		{
			name:  "only notes set",
			input: UpdateChronicConditionInput{Notes: &notes},
			want:  map[string]any{"notes": notes},
		},
		{
			name:  "only is_active set",
			input: UpdateChronicConditionInput{IsActive: &isActive},
			want:  map[string]any{"is_active": isActive},
		},
		{
			name: "all fields set",
			input: UpdateChronicConditionInput{
				ConditionCode: &code, ConditionName: &name, DiagnosedAt: &diagnosedAt, Notes: &notes, IsActive: &isActive,
			},
			want: map[string]any{
				"condition_code": code, "condition_name": name, "diagnosed_at": diagnosedAt, "notes": notes, "is_active": isActive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildChronicConditionUpdateFields(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChronicConditionService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockPetChronicConditionRepository{
			findByPetIDFn: func(_ context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
				return []model.PetChronicCondition{{ID: 1, ClinicID: clinicID, PetID: petID}}, nil
			},
		}
		svc := NewChronicConditionService(repo, nil, nil)
		res, err := svc.List(ctx, 1, 100)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockPetChronicConditionRepository{
			findByPetIDFn: func(_ context.Context, _, _ uint64) ([]model.PetChronicCondition, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewChronicConditionService(repo, nil, nil)
		res, err := svc.List(ctx, 1, 100)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestChronicConditionService_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			createFn: func(_ context.Context, record *model.PetChronicCondition) error {
				record.ID = 10
				return nil
			},
			getActiveConditionCodesByOwnerFn: func(_ context.Context, _, _ uint64) ([]string, error) {
				return []string{"C01"}, nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		notes := "some notes"
		input := CreateChronicConditionInput{
			ConditionCode: "C01",
			ConditionName: "Disease 1",
			DiagnosedAt:   now,
			Notes:         &notes,
			IsActive:      true,
		}
		res, err := svc.Create(ctx, 1, 100, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), res.ID)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		_, err := svc.Create(ctx, 1, 100, CreateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("db create error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			createFn: func(_ context.Context, _ *model.PetChronicCondition) error {
				return errors.New("db error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		_, err := svc.Create(ctx, 1, 100, CreateChronicConditionInput{})
		assert.Error(t, err)
	})
}

func TestChronicConditionService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, clinicID, petID, id uint64, fields map[string]any) error {
				assert.Equal(t, "C02", fields["condition_code"])
				return nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		code := "C02"
		input := UpdateChronicConditionInput{
			ConditionCode: &code,
		}
		res, err := svc.Update(ctx, 1, 100, 10, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("condition not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, _, _, _ uint64) (*model.PetChronicCondition, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, _, _, _ uint64, _ map[string]any) error {
				return errors.New("update error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		code := "C02"
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{ConditionCode: &code})
		assert.Error(t, err)
	})

	t.Run("returns error when no fields provided", func(t *testing.T) {
		updateCalled := false
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID, PetID: petID}, nil
			},
			updateFn: func(_ context.Context, _, _, _ uint64, _ map[string]any) error {
				updateCalled = true
				return nil
			},
		}
		svc := NewChronicConditionService(repo, petRepo, &mockLstepTagSyncService{})
		res, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.False(t, updateCalled, "empty PATCH must not reach repository.Update")
		assert.True(t, apperrors.IsInvalidInput(err), "empty PATCH must be client-correctable invalid input, not not-found")
	})
}

// TestChronicConditionService_Update_ReloadError は Update 成功後の再取得
// （s.repo.FindByID の2回目呼び出し）が失敗した場合にエラーが返ることを検証する。
func TestChronicConditionService_Update_ReloadError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
		},
	}
	repo := &mockPetChronicConditionRepository{
		findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
			callCount++
			if callCount == 1 {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			}
			return nil, errors.New("db error on reload")
		},
		updateFn: func(_ context.Context, _, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	svc := NewChronicConditionService(repo, petRepo, &mockLstepTagSyncService{})
	code := "C03"
	res, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{ConditionCode: &code})

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, 2, callCount)
}

// TestChronicConditionService_SyncTags_BestEffort は syncTags が best-effort であり、
// FindActiveConditionCodesByOwner / SyncChronicConditionTags のエラーが呼び出し元の
// Create/Update/Delete の成功結果に伝播しないことを検証する（患者記録操作を失敗させない設計）。
func TestChronicConditionService_SyncTags_BestEffort(t *testing.T) {
	ctx := context.Background()

	t.Run("FindActiveConditionCodesByOwner error does not fail Create", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		syncCalled := false
		repo := &mockPetChronicConditionRepository{
			createFn: func(_ context.Context, record *model.PetChronicCondition) error {
				record.ID = 10
				return nil
			},
			getActiveConditionCodesByOwnerFn: func(_ context.Context, _, _ uint64) ([]string, error) {
				return nil, errors.New("db error")
			},
		}
		tagSync := &chronicTagSyncOverride{
			mockLstepTagSyncService: &mockLstepTagSyncService{},
			syncChronicConditionTagsFn: func(_ context.Context, _, _ uint64, _ []string) error {
				syncCalled = true
				return nil
			},
		}
		svc := NewChronicConditionService(repo, petRepo, tagSync)

		res, err := svc.Create(ctx, 1, 100, CreateChronicConditionInput{ConditionCode: "C01"})

		assert.NoError(t, err, "codes fetch failure must not fail the create operation")
		assert.NotNil(t, res)
		assert.False(t, syncCalled, "tag sync must not be attempted when codes fetch failed")
	})

	t.Run("SyncChronicConditionTags error does not fail Delete", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _, _ uint64) error { return nil },
			getActiveConditionCodesByOwnerFn: func(_ context.Context, _, _ uint64) ([]string, error) {
				return []string{"C01"}, nil
			},
		}
		tagSync := &chronicTagSyncOverride{
			mockLstepTagSyncService: &mockLstepTagSyncService{},
			syncChronicConditionTagsFn: func(_ context.Context, _, _ uint64, _ []string) error {
				return errors.New("lstep sync failed")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, tagSync)

		err := svc.Delete(ctx, 1, 100, 10)

		assert.NoError(t, err, "tag sync failure must not fail the delete operation (best-effort)")
	})
}

func TestChronicConditionService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _, _ uint64) error {
				return nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.NoError(t, err)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})

	t.Run("condition not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, _, _, _ uint64) (*model.PetChronicCondition, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _, _ uint64) error {
				return errors.New("delete error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})
}
