package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOwnerRepositoryForClinics is a minimal wrapper mock used only to exercise
// GetByIDForClinics, since mockOwnerRepository (owner_service_test.go) hard-codes
// FindByIDForClinics to always return (nil, nil) without a configurable hook.
type mockOwnerRepositoryForClinics struct {
	ServiceRepository
	findByIDForClinicsFn func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error)
}

func (m *mockOwnerRepositoryForClinics) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	return m.findByIDForClinicsFn(ctx, clinicIDs, id)
}

func TestOwnerService_GetByIDForClinics(t *testing.T) {
	tests := []struct {
		name      string
		repoOwner *model.Owner
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "returns owner when found across clinics",
			repoOwner: &model.Owner{ID: 10, ClinicID: 2, Name: "山田 太郎"},
			wantErr:   false,
		},
		{
			name:    "propagates repository error",
			repoErr: apperrors.WrapNotFound("owner", "10"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepositoryForClinics{
				findByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
					return tt.repoOwner, tt.repoErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.GetByIDForClinics(context.Background(), []uint64{1, 2}, 10)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, owner)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOwner, owner)
			}
		})
	}
}

// ---- CreateWithPets: email/phone uniqueness branches + tagSyncSvc nil branch ----

func TestOwnerService_CreateWithPets_UniquenessAndSyncBranches(t *testing.T) {
	baseInput := func() CreateOwnerInput {
		return CreateOwnerInput{
			OwnerName: "テスト 太郎",
			Email:     "taro@example.com",
			Phone:     "090-1234-5678",
		}
	}

	t.Run("email conflict returns already exists", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByEmailFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 999}, nil
			},
			createWithPetsFn: func(_ context.Context, _ *model.Owner, _ []model.Pet) error {
				return errors.New("should not be called")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		input := baseInput()
		owner, err := svc.CreateWithPets(context.Background(), 1, &input)
		assert.Error(t, err)
		assert.Nil(t, owner)
		assert.True(t, apperrors.IsAlreadyExists(err))
		// BUG-024: 日本語完成文のみ（英語テンプレ混在なし）
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "このメールアドレスはすでに登録されています", appErr.Message)
		assert.NotContains(t, appErr.Message, "already exists")
		assert.NotContains(t, appErr.Message, "owner '")
	})

	t.Run("email uniqueness check repository error propagates", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByEmailFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		input := baseInput()
		owner, err := svc.CreateWithPets(context.Background(), 1, &input)
		assert.Error(t, err)
		assert.Nil(t, owner)
	})

	t.Run("phone conflict returns already exists", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByPhoneFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 999}, nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		input := baseInput()
		owner, err := svc.CreateWithPets(context.Background(), 1, &input)
		assert.Error(t, err)
		assert.Nil(t, owner)
		assert.True(t, apperrors.IsAlreadyExists(err))
		// BUG-024 / BUG-019
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "この電話番号はすでに登録されています", appErr.Message)
		assert.NotContains(t, appErr.Message, "already exists")
	})

	t.Run("phone uniqueness check repository error propagates", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByPhoneFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		input := baseInput()
		owner, err := svc.CreateWithPets(context.Background(), 1, &input)
		assert.Error(t, err)
		assert.Nil(t, owner)
	})

	t.Run("nil tagSyncSvc skips animal classification sync", func(t *testing.T) {
		repo := &mockOwnerRepository{
			createWithPetsFn: func(_ context.Context, owner *model.Owner, _ []model.Pet) error {
				owner.ID = 42
				return nil
			},
		}
		svc := NewService(repo, nil, nil, nil)
		input := baseInput()
		owner, err := svc.CreateWithPets(context.Background(), 1, &input)
		assert.NoError(t, err)
		assert.NotNil(t, owner)
	})
}

// ---- Update: validation, email/phone uniqueness, repo errors, reload errors ----

func TestOwnerService_Update_AdditionalBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("find by id error returns wrapped error", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "1")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		name := "更新"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{OwnerName: &name})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("validation failure after successful find returns error", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		badRate := 150.0
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{DiscountRate: &badRate})
		assert.Error(t, err)
	})

	t.Run("empty fields returns invalid input", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("email conflict during update returns already exists", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			findByEmailFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 999}, nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		email := "conflict@example.com"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{Email: &email})
		assert.Error(t, err)
		assert.True(t, apperrors.IsAlreadyExists(err))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "このメールアドレスはすでに登録されています", appErr.Message)
	})

	t.Run("email matching same owner does not conflict", func(t *testing.T) {
		callCount := 0
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				callCount++
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			findByEmailFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		email := "same@example.com"
		owner, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{Email: &email})
		assert.NoError(t, err)
		assert.NotNil(t, owner)
	})

	t.Run("email uniqueness check repository error propagates", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			findByEmailFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		email := "err@example.com"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{Email: &email})
		assert.Error(t, err)
	})

	t.Run("phone conflict during update returns already exists", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			findByPhoneFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 999}, nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		phone := "090-0000-0000"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{Phone: &phone})
		assert.Error(t, err)
		assert.True(t, apperrors.IsAlreadyExists(err))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "この電話番号はすでに登録されています", appErr.Message)
	})

	t.Run("phone uniqueness check repository error propagates", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			findByPhoneFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		phone := "090-0000-0000"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{Phone: &phone})
		assert.Error(t, err)
	})

	t.Run("repository update error is wrapped", func(t *testing.T) {
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		name := "更新後"
		_, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{OwnerName: &name})
		assert.Error(t, err)
	})

	t.Run("reload after update error is wrapped", func(t *testing.T) {
		callCount := 0
		repo := &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				callCount++
				if callCount == 1 {
					return &model.Owner{ID: 1, ClinicID: 1}, nil
				}
				return nil, errors.New("reload error")
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return nil
			},
		}
		svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
		name := "更新後"
		owner, err := svc.Update(ctx, 1, 1, &UpdateOwnerInput{OwnerName: &name})
		assert.Error(t, err)
		assert.Nil(t, owner)
	})
}

// ---- Delete: initial FindByID error branch (not exercised by owner_service_test.go) ----

func TestOwnerService_Delete_FindByIDError(t *testing.T) {
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "999")
		},
	}
	svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
	err := svc.Delete(context.Background(), 1, 999)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// ---- reloadOwner: error branch exercised indirectly via UpdateDeliveryExclusion,
// since reloadOwner itself is only invoked from owner_service_delivery.go / owner_service_line.go
// (its success path is already covered there by owner_service_test.go). ----

func TestOwnerService_ReloadOwner_ErrorBranchViaUpdateDeliveryExclusion(t *testing.T) {
	callCount := 0
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			callCount++
			if callCount == 1 {
				return &model.Owner{ID: 10, ClinicID: 1}, nil
			}
			return nil, errors.New("reload failed")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)
	owner, err := svc.UpdateDeliveryExclusion(context.Background(), 1, 10, UpdateDeliveryExclusionInput{Excluded: true})
	assert.Error(t, err)
	assert.Nil(t, owner)
}
