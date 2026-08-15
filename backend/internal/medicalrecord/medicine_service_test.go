package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMedicineRepository は MedicineRepository のテスト用モック実装
type mockMedicineRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	createFn                  func(ctx context.Context, medicine *model.Medicine) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMedicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockMedicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn != nil {
		return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
	}
	return 0, nil
}

func (m *mockMedicineRepository) CountUsageByMedicineID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	return m.createFn(ctx, medicine)
}

func (m *mockMedicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockMedicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func newTestMedicineService(repo *mockMedicineRepository) MedicineService {
	// mockInventoryRepository は inventory_service_test.go で定義済み（パッケージスコープ共有）
	inventoryRepo := &mockInventoryRepository{
		createFn: func(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
			if item != nil && item.ID == 0 {
				item.ID = 9001 // simulate GORM auto-increment for inventory_id link (MRC-02)
			}
			return nil // デフォルト: 在庫作成成功
		},
	}
	// mockTransactor は trimming_service_test.go で定義済み（パッケージスコープ共有）
	// MRC-02: delete requires fail-closed audit dependency.
	return NewMedicineServiceWithAudit(repo, inventoryRepo, &mockTransactor{}, okCarePlanAuditTx{})
}

func TestMedicineService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Medicine
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns medicine list",
			repoData: []model.Medicine{
				{ID: 1, Name: "アモキシシリン"},
				{ID: 2, Name: "メトロニダゾール"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no medicines exist",
			repoData: []model.Medicine{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				findAllFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Medicine, int64, error) {
					return tt.repoData, int64(len(tt.repoData)), tt.repoErr
				},
			}
			svc := newTestMedicineService(repo)

			medicines, total, err := svc.List(context.Background(), 1, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, medicines, tt.wantLen)
				assert.Equal(t, int64(tt.wantLen), total)
			}
		})
	}
}

func TestMedicineService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoMedicine *model.Medicine
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns medicine when found",
			id:           1,
			repoMedicine: &model.Medicine{ID: 1, Name: "アモキシシリン"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when medicine does not exist",
			id:           999,
			repoMedicine: nil,
			repoErr:      apperrors.WrapNotFound("medicine", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoMedicine: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
					return tt.repoMedicine, tt.repoErr
				},
			}
			svc := newTestMedicineService(repo)

			medicine, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, medicine)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoMedicine, medicine)
			}
		})
	}
}

func TestMedicineService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateMedicineInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates medicine successfully",
			input: &CreateMedicineInput{
				Name:     "新規薬剤",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates medicine with is_non_insurance=true",
			input: &CreateMedicineInput{
				Name:           "保険対象外薬剤",
				IsActive:       true,
				IsNonInsurance: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when name is empty",
			input: &CreateMedicineInput{
				Name: "",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when medicine already exists",
			input: &CreateMedicineInput{
				Name: "重複薬剤",
			},
			repoErr: apperrors.WrapAlreadyExists("medicine", "重複薬剤"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateMedicineInput{
				Name: "エラー薬剤",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				createFn: func(_ context.Context, m *model.Medicine) error {
					if tt.repoErr == nil && m != nil && m.ID == 0 {
						m.ID = 42
					}
					return tt.repoErr
				},
				updateFieldsFn: func(_ context.Context, _, id uint64, fields map[string]any) (*model.Medicine, error) {
					// MRC-02 inventory_id link writeback after auto-create.
					med := &model.Medicine{ID: id, Name: tt.input.Name, ClinicID: 1}
					if inv, ok := fields[colMedicineInventoryID]; ok {
						if invID, ok := inv.(uint64); ok {
							med.InventoryID = &invID
						}
					}
					return med, nil
				},
			}
			svc := newTestMedicineService(repo)

			medicine, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, medicine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, medicine)
				if medicine != nil {
					assert.NotNil(t, medicine.InventoryID, "auto-created inventory must be linked (MRC-02)")
				}
			}
		})
	}
}

func TestMedicineService_Update(t *testing.T) {
	existingMedicine := &model.Medicine{ID: 1, Name: "更新後薬剤名", ClinicID: 1}

	tests := []struct {
		name      string
		id        uint64
		input     *UpdateMedicineInput
		updateErr error
		findErr   error
		wantErr   bool
	}{
		{
			name: "updates medicine successfully",
			id:   1,
			input: &UpdateMedicineInput{
				Name: strPtr("更新後薬剤名"),
			},
			updateErr: nil,
			findErr:   nil,
			wantErr:   false,
		},
		{
			name:  "returns 400 when no fields provided",
			id:    1,
			input: &UpdateMedicineInput{
				// 全フィールド nil
			},
			updateErr: nil,
			findErr:   nil,
			wantErr:   true,
		},
		{
			name: "returns not found error when medicine does not exist",
			id:   999,
			input: &UpdateMedicineInput{
				Name: strPtr("存在しない薬剤"),
			},
			updateErr: apperrors.WrapNotFound("medicine", "999"),
			findErr:   nil,
			wantErr:   true,
		},
		{
			name: "returns error on repository failure",
			id:   1,
			input: &UpdateMedicineInput{
				Name: strPtr("エラーケース"),
			},
			updateErr: errors.New("db error"),
			findErr:   nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return existingMedicine, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Medicine, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return existingMedicine, nil
				},
			}
			svc := newTestMedicineService(repo)

			medicine, err := svc.Update(context.Background(), 1, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, medicine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, medicine)
			}
		})
	}
}

func TestMedicineService_Update_NilInput(t *testing.T) {
	repo := &mockMedicineRepository{}
	svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, &mockTransactor{}, nil)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMedicineService_Update_IsNonInsurance(t *testing.T) {
	nonIns := true
	var capturedFields map[string]any
	existing := &model.Medicine{ID: 1, ClinicID: 1, IsNonInsurance: false}
	repo := &mockMedicineRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Medicine, error) {
			capturedFields = fields
			return &model.Medicine{ID: 1, ClinicID: 1, IsNonInsurance: true}, nil
		},
	}
	svc := newTestMedicineService(repo)
	input := &UpdateMedicineInput{IsNonInsurance: &nonIns}
	result, err := svc.Update(context.Background(), 1, 1, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, capturedFields[colMedicineIsNonInsurance])
}

func TestMedicineService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		medicine         *model.Medicine // FindByID が返すレコード
		findErr          error
		childrenCount    int64
		countChildrenErr error
		deleteErr        error
		wantErr          bool
		wantNotFound     bool
		wantInvalid      bool
		wantConflict     bool
	}{
		{
			name:      "deletes a leaf medicine item successfully",
			id:        10,
			medicine:  &model.Medicine{ID: 10, ClinicID: 1, ParentID: uint64Ptr(1)},
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name:          "deletes an empty category successfully",
			id:            1,
			medicine:      &model.Medicine{ID: 1, ClinicID: 1, ParentID: nil},
			childrenCount: 0,
			deleteErr:     nil,
			wantErr:       false,
		},
		{
			name:          "rejects deletion of category that has children",
			id:            1,
			medicine:      &model.Medicine{ID: 1, ClinicID: 1, ParentID: nil},
			childrenCount: 3,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:         "returns not found when medicine does not exist",
			id:           999,
			findErr:      apperrors.WrapNotFound("medicine", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "returns error on delete repository failure",
			id:        10,
			medicine:  &model.Medicine{ID: 10, ClinicID: 1, ParentID: uint64Ptr(1)},
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:             "returns error on count children failure",
			id:               1,
			medicine:         &model.Medicine{ID: 1, ClinicID: 1, ParentID: nil},
			countChildrenErr: errors.New("db error"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
					return tt.medicine, tt.findErr
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childrenCount, tt.countChildrenErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := newTestMedicineService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// strPtr はテスト用のヘルパー関数
func strPtr(s string) *string { return &s }
