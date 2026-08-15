package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockInventoryRepository は InventoryRepository のテスト用モック実装
type mockInventoryRepository struct {
	findAllFn        func(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	findByIDFn       func(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	createFn         func(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	updateFieldsFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
	countUsageByIDFn func(ctx context.Context, clinicID, inventoryID uint64) (int64, error)
	decreaseStockFn  func(ctx context.Context, clinicID, id uint64, quantity float64) error // INV-SEC P1 signature parity
}

func (m *mockInventoryRepository) FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	return m.findAllFn(ctx, clinicID, category, status, page, limit)
}

func (m *mockInventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockInventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	return m.createFn(ctx, clinicID, item)
}

func (m *mockInventoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockInventoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInventoryRepository) DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error {
	if m.decreaseStockFn != nil {
		return m.decreaseStockFn(ctx, clinicID, id, quantity)
	}
	return nil
}

func (m *mockInventoryRepository) CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error) {
	if m.countUsageByIDFn != nil {
		return m.countUsageByIDFn(ctx, clinicID, inventoryID)
	}
	return 0, nil
}

func (m *mockInventoryRepository) DeleteByNameAndMedicineCategory(_ context.Context, _ uint64, _ string) error {
	return nil
}

func (m *mockInventoryRepository) UpdateNameByMedicineCategory(_ context.Context, _ uint64, _, _ string) error {
	return nil
}

func TestInventoryService_List(t *testing.T) {
	category := string(model.InventoryCategoryMedicine)
	status := string(model.InventoryStatusLow)

	tests := []struct {
		name      string
		clinicID  uint64
		category  *string
		status    *string
		page      int
		limit     int
		repoItems []model.InventoryItem
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns inventory list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			repoItems: []model.InventoryItem{
				{ID: 1, ClinicID: 1, Name: "点滴セット", Category: model.InventoryCategoryMedicine, Quantity: 50, Status: model.InventoryStatusSufficient},
				{ID: 2, ClinicID: 1, Name: "消毒液", Category: model.InventoryCategoryConsumable, Quantity: 5, Status: model.InventoryStatusLow},
			},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no items exist",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: []model.InventoryItem{},
			repoTotal: 0,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by category",
			clinicID: 1,
			category: &category,
			page:     1,
			limit:    20,
			repoItems: []model.InventoryItem{
				{ID: 1, ClinicID: 1, Name: "点滴セット", Category: model.InventoryCategoryMedicine},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			status:   &status,
			page:     1,
			limit:    20,
			repoItems: []model.InventoryItem{
				{ID: 2, ClinicID: 1, Name: "消毒液", Status: model.InventoryStatusLow},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInventoryRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *string, _ *string, _, _ int) ([]model.InventoryItem, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewInventoryService(repo)

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.category, tt.status, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestInventoryService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoItem *model.InventoryItem
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns inventory item when found",
			clinicID: 1,
			id:       10,
			repoItem: &model.InventoryItem{
				ID:       10,
				ClinicID: 1,
				Name:     "点滴セット",
				Category: model.InventoryCategoryMedicine,
				Quantity: 100,
				Status:   model.InventoryStatusSufficient,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when item does not exist",
			clinicID: 1,
			id:       999,
			repoItem: nil,
			repoErr:  apperrors.WrapNotFound("inventory_item", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoItem: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInventoryRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.InventoryItem, error) {
					return tt.repoItem, tt.repoErr
				},
			}
			svc := NewInventoryService(repo)

			item, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoItem, item)
			}
		})
	}
}

func TestBuildInventoryUpdate(t *testing.T) {
	name := "点滴セット"
	category := model.InventoryCategoryMedicine
	quantity := 42
	unit := "本"
	minStock := 5
	location := "棚A"
	expiry := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	supplier := "サプライヤーA"
	lastRestocked := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		input      *UpdateInventoryInput
		wantFields map[string]any
	}{
		{
			name:       "nil-value input produces empty fields",
			input:      &UpdateInventoryInput{},
			wantFields: map[string]any{},
		},
		{
			name: "all writable fields set produces field map without status",
			input: &UpdateInventoryInput{
				Name:          &name,
				Category:      &category,
				Quantity:      &quantity,
				Unit:          &unit,
				MinStockLevel: &minStock,
				Location:      &location,
				ExpiryDate:    &expiry,
				Supplier:      &supplier,
				LastRestocked: &lastRestocked,
			},
			wantFields: map[string]any{
				"name":            name,
				"category":        category,
				"quantity":        quantity,
				"unit":            unit,
				"min_stock_level": minStock,
				"location":        location,
				"expiry_date":     expiry,
				"supplier":        supplier,
				"last_restocked":  lastRestocked,
			},
		},
		{
			name: "only name set produces single field",
			input: &UpdateInventoryInput{
				Name: &name,
			},
			wantFields: map[string]any{"name": name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildInventoryUpdate(tt.input)
			assert.Equal(t, tt.wantFields, got)
			// SD-4: client status write 経路 0 — fields["status"] を絶対に書かない
			_, hasStatus := got["status"]
			assert.False(t, hasStatus, "buildInventoryUpdate must not write status")
		})
	}
}

func TestInventoryService_Create(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateInventoryInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates inventory item successfully",
			clinicID: 1,
			input: &CreateInventoryInput{
				Name:          "新規薬剤",
				Category:      string(model.InventoryCategoryMedicine),
				Quantity:      100,
				Unit:          "本",
				MinStockLevel: 10,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when item already exists",
			clinicID: 1,
			input:    &CreateInventoryInput{Name: "既存薬剤", Category: string(model.InventoryCategoryMedicine)},
			repoErr:  apperrors.WrapAlreadyExists("inventory_item", "既存薬剤"),
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input:    &CreateInventoryInput{Name: "エラー薬剤", Category: string(model.InventoryCategoryOther)},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
		{
			// SD-4: Create 入力に Status は存在しない。高在庫でも dead column は default sufficient。
			// JSON 応答の status は DeriveInventoryStatus で sufficient になる（response 側テスト）。
			name:     "create uses dead-column default sufficient; client cannot set status",
			clinicID: 1,
			input: &CreateInventoryInput{
				Name:          "高在庫アイテム",
				Category:      string(model.InventoryCategoryMedicine),
				Quantity:      100,
				MinStockLevel: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *model.InventoryItem
			repo := &mockInventoryRepository{
				createFn: func(_ context.Context, _ uint64, item *model.InventoryItem) error {
					created = item
					return tt.repoErr
				},
			}
			svc := NewInventoryService(repo)

			item, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				// SD-4: client status write 経路 0 — 常に default sufficient（dead column）
				assert.Equal(t, model.InventoryStatusSufficient, item.Status)
				// 公開 status は quantity/min 導出と一致する
				assert.Equal(t, model.DeriveInventoryStatus(tt.input.Quantity, tt.input.MinStockLevel),
					model.DeriveInventoryStatus(item.Quantity, item.MinStockLevel))
				if created != nil {
					assert.Equal(t, model.InventoryStatusSufficient, created.Status)
					assert.NotEqual(t, model.InventoryStatusLow, created.Status)
				}
			}
		})
	}
}

// TestInventoryService_Create_ClientStatusCannotPoisonDerivedRead は SD-4 回帰:
// Create 後の公開 status は quantity/min 導出のみ。高 qty でも low にならない。
func TestInventoryService_Create_ClientStatusCannotPoisonDerivedRead(t *testing.T) {
	repo := &mockInventoryRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.InventoryItem) error {
			return nil
		},
	}
	svc := NewInventoryService(repo)

	item, err := svc.Create(context.Background(), 1, &CreateInventoryInput{
		Name:          "高在庫",
		Category:      string(model.InventoryCategoryMedicine),
		Quantity:      50,
		MinStockLevel: 5,
		Unit:          "本",
	})
	assert.NoError(t, err)
	if !assert.NotNil(t, item) {
		return
	}

	// dead column default
	assert.Equal(t, model.InventoryStatusSufficient, item.Status)
	// 公開 SoT: Derive が sufficient（client が low を送っても受け付けない入力型）
	derived := model.DeriveInventoryStatus(item.Quantity, item.MinStockLevel)
	assert.Equal(t, model.InventoryStatusSufficient, derived)
	assert.Equal(t, string(derived), toInventoryResponse(item).Status)
}

func TestInventoryService_Update_NilInput(t *testing.T) {
	repo := &mockInventoryRepository{}
	svc := NewInventoryService(repo)
	item, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, item)
}

func TestInventoryService_Update(t *testing.T) {
	name := "更新後薬剤"
	quantity := 200
	tests := []struct {
		name        string
		input       UpdateInventoryInput
		repoErr     error
		findByIDErr error
		wantErr     bool
		wantNF      bool
	}{
		{
			name: "updates inventory item successfully",
			input: UpdateInventoryInput{
				Name:     &name,
				Quantity: &quantity,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateInventoryInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when item does not exist",
			input: UpdateInventoryInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("inventory_item", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateInventoryInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns error when FindByID fails before update",
			input: UpdateInventoryInput{
				Name: &name,
			},
			findByIDErr: apperrors.WrapNotFound("inventory_item", "999"),
			wantErr:     true,
			wantNF:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInventoryRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.InventoryItem{ID: id, ClinicID: clinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.InventoryItem, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.InventoryItem{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewInventoryService(repo)

			item, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

func TestInventoryService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		usageCount   int64
		usageErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:       "deletes inventory item successfully",
			clinicID:   1,
			id:         10,
			usageCount: 0,
			repoErr:    nil,
			wantErr:    false,
		},
		{
			name:         "returns conflict error when inventory item is in use",
			clinicID:     1,
			id:           10,
			usageCount:   3,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when usage count check fails",
			clinicID: 1,
			id:       10,
			usageErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:       "returns not found error when item does not exist",
			clinicID:   1,
			id:         999,
			usageCount: 0,
			repoErr:    apperrors.WrapNotFound("inventory_item", "999"),
			wantErr:    true,
			wantNF:     true,
		},
		{
			name:       "returns error on repository failure",
			clinicID:   1,
			id:         10,
			usageCount: 0,
			repoErr:    errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInventoryRepository{
				countUsageByIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.usageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewInventoryService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
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
