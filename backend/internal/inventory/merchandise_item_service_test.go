package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// mockMerchandiseItemRepository は MerchandiseItemRepository のテスト用モック実装
type mockMerchandiseItemRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	countUsageByMerchandiseItemFn func(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	createFn                      func(ctx context.Context, item *model.MerchandiseItem) error
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMerchandiseItemRepository) FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, category)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MerchandiseItem{ID: id, ClinicID: clinicID}, nil
}

func (m *mockMerchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
	if m.countUsageByMerchandiseItemFn != nil {
		return m.countUsageByMerchandiseItemFn(ctx, clinicID, merchandiseItemID)
	}
	return 0, nil
}

func (m *mockMerchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	return m.createFn(ctx, item)
}

func (m *mockMerchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMerchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// mockMerchandiseItemTransactor runs WithTx by invoking fn on the same ctx (unit tests).
type mockMerchandiseItemTransactor struct {
	withTxFn func(ctx context.Context, fn func(context.Context) error) error
}

func (m *mockMerchandiseItemTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m != nil && m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

func newTestMerchandiseItemService(repo *mockMerchandiseItemRepository) MerchandiseItemService {
	return NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
}

func newTestMerchandiseItemServiceWithTx(repo *mockMerchandiseItemRepository, tx Transactor) MerchandiseItemService {
	return NewMerchandiseItemService(repo, tx)
}

// ---- List テスト ----

func TestMerchandiseItemService_List(t *testing.T) {
	tests := []struct {
		name     string
		category string
		items    []model.MerchandiseItem
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns items for category",
			category: "food",
			items: []model.MerchandiseItem{
				{ID: 1, Name: "ドッグフード", Category: "food"},
				{ID: 2, Name: "キャットフード", Category: "food"},
			},
			wantLen: 2,
		},
		{
			name:     "returns empty list when no items exist",
			category: "toy",
			items:    []model.MerchandiseItem{},
			wantLen:  0,
		},
		{
			name:     "propagates repository error",
			category: "food",
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{
				findAllFn: func(_ context.Context, _ uint64, _ string) ([]model.MerchandiseItem, error) {
					return tt.items, tt.repoErr
				},
			}
			svc := newTestMerchandiseItemService(repo)

			items, err := svc.List(context.Background(), 1, tt.category)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, items, tt.wantLen)
		})
	}
}

// ---- GetByID テスト ----

func TestMerchandiseItemService_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
		wantErr    bool
	}{
		{
			name: "returns item successfully",
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
				return &model.MerchandiseItem{ID: id, ClinicID: clinicID}, nil
			},
		},
		{
			name: "returns error when item does not exist",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
				return nil, apperrors.WrapNotFound("merchandise_item", "999")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{findByIDFn: tt.findByIDFn}
			svc := newTestMerchandiseItemService(repo)

			item, err := svc.GetByID(context.Background(), 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, item)
			assert.Equal(t, uint64(5), item.ID)
		})
	}
}

// ---- Reorder テスト ----

func TestMerchandiseItemService_Reorder(t *testing.T) {
	tests := []struct {
		name      string
		ids       []uint64
		reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
		wantErr   bool
	}{
		{
			name: "reorders items successfully",
			ids:  []uint64{3, 1, 2},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return nil
			},
		},
		{
			name:    "returns error when ids is empty",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name:    "returns error when ids is nil",
			ids:     nil,
			wantErr: true,
		},
		{
			name: "propagates repository error",
			ids:  []uint64{1, 2},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{reorderFn: tt.reorderFn}
			svc := newTestMerchandiseItemService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ---- Create テスト ----

func TestMerchandiseItemService_Create(t *testing.T) {
	price100 := int64(100)
	priceNeg := int64(-1)

	tests := []struct {
		name    string
		input   *CreateMerchandiseItemInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates merchandise item successfully",
			input: &CreateMerchandiseItemInput{
				Name:      "ドッグフード",
				Category:  "food",
				UnitPrice: price100,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "applies default tax type and rate when not specified",
			input: &CreateMerchandiseItemInput{
				Name:      "デフォルト税率商品",
				UnitPrice: price100,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when name is empty",
			input: &CreateMerchandiseItemInput{
				Name: "",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when unit_price is negative",
			input: &CreateMerchandiseItemInput{
				Name:      "マイナス商品",
				UnitPrice: priceNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateMerchandiseItemInput{
				Name:      "エラー商品",
				UnitPrice: price100,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{
				createFn: func(_ context.Context, _ *model.MerchandiseItem) error {
					return tt.repoErr
				},
			}
			svc := newTestMerchandiseItemService(repo)

			item, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

// TestMerchandiseItemService_Create_CustomTaxTypeAndRate は TaxType/TaxRate が明示指定された場合、
// デフォルト値ではなく指定値がそのまま採用されることを検証する。
func TestMerchandiseItemService_Create_CustomTaxTypeAndRate(t *testing.T) {
	customRate := 0.08
	var created *model.MerchandiseItem
	repo := &mockMerchandiseItemRepository{
		createFn: func(_ context.Context, item *model.MerchandiseItem) error {
			created = item
			return nil
		},
	}
	svc := newTestMerchandiseItemService(repo)

	item, err := svc.Create(context.Background(), 1, &CreateMerchandiseItemInput{
		Name:      "軽減税率商品",
		UnitPrice: 500,
		TaxType:   "included",
		TaxRate:   &customRate,
	})

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.NotNil(t, created)
	assert.Equal(t, model.TaxType("included"), created.TaxType)
	assert.Equal(t, customRate, created.TaxRate)
}

// ---- Update テスト ----

func TestMerchandiseItemService_Update(t *testing.T) {
	name := "更新後商品名"
	isActive := false
	price200 := int64(200)
	priceNeg := int64(-1)
	category := "food"

	tests := []struct {
		name    string
		input   *UpdateMerchandiseItemInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates merchandise item successfully",
			input: &UpdateMerchandiseItemInput{
				Name:      &name,
				IsActive:  &isActive,
				UnitPrice: &price200,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "updates category only",
			input: &UpdateMerchandiseItemInput{
				Category: &category,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateMerchandiseItemInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when unit_price is negative",
			input: &UpdateMerchandiseItemInput{
				UnitPrice: &priceNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &UpdateMerchandiseItemInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns not found error when item does not exist",
			input: &UpdateMerchandiseItemInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("merchandise_item", "999"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MerchandiseItem, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.MerchandiseItem{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := newTestMerchandiseItemService(repo)

			item, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

// ---- Delete テスト ----

func TestMerchandiseItemService_Delete_Success(t *testing.T) {
	// Soft-delete first, then usage re-check under the same WithTx.
	var deleted, counted bool
	inTx := false
	tx := &mockMerchandiseItemTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		inTx = true
		defer func() { inTx = false }()
		return fn(ctx)
	}}
	repo := &mockMerchandiseItemRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			assert.True(t, inTx, "Delete must run inside WithTx")
			deleted = true
			return nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			assert.True(t, inTx, "CountUsage must run inside WithTx")
			assert.True(t, deleted, "soft-delete must precede usage re-check")
			counted = true
			return 0, nil
		},
	}
	svc := newTestMerchandiseItemServiceWithTx(repo, tx)

	err := svc.Delete(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.True(t, deleted)
	assert.True(t, counted)
}

func TestMerchandiseItemService_Delete_NotFound(t *testing.T) {
	// Soft-delete miss surfaces as NotFound (distinct from Conflict when usage > 0).
	repo := &mockMerchandiseItemRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return apperrors.WrapNotFound("merchandise_item", "999")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestMerchandiseItemService_Delete_ConflictWhenInUse(t *testing.T) {
	// Soft-delete succeeds inside the tx; usage re-check returns Conflict and rolls back.
	repo := &mockMerchandiseItemRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 2, nil // billing / estimate / campaign target references
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, apperrors.IsNotFound(err), "usage conflict must stay distinct from NotFound")
}

func TestMerchandiseItemService_Delete_CountUsageError(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db connection error")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	// 依存チェック失敗はNotFoundでもConflictでもなく一般エラー
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestMerchandiseItemService_Delete_RepositoryError(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestMerchandiseItemService_Delete_RequiresTransactor(t *testing.T) {
	svc := NewMerchandiseItemService(&mockMerchandiseItemRepository{}, nil)
	err := svc.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

// ---- Update 追加分岐テスト ----

func TestMerchandiseItemService_Update_NilInput(t *testing.T) {
	svc := newTestMerchandiseItemService(&mockMerchandiseItemRepository{})

	item, err := svc.Update(context.Background(), 1, 1, nil)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestMerchandiseItemService_Update_FindByIDError(t *testing.T) {
	name := "更新後商品名"
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return nil, apperrors.WrapNotFound("merchandise_item", "1")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	item, err := svc.Update(context.Background(), 1, 1, &UpdateMerchandiseItemInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestMerchandiseItemService_Update_InvalidOptionalName(t *testing.T) {
	blank := "   "
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			return &model.MerchandiseItem{ID: id}, nil
		},
	}
	svc := newTestMerchandiseItemService(repo)

	item, err := svc.Update(context.Background(), 1, 1, &UpdateMerchandiseItemInput{Name: &blank})

	assert.Error(t, err)
	assert.Nil(t, item)
}

// ---- buildMerchandiseItemUpdate 直接テスト ----

func TestBuildMerchandiseItemUpdate(t *testing.T) {
	t.Run("nilフィールドはmapに含まれない（ゼロ値入力）", func(t *testing.T) {
		fields := buildMerchandiseItemUpdate(&UpdateMerchandiseItemInput{})
		assert.Empty(t, fields)
	})

	t.Run("全フィールド指定時はすべてmapに含まれる", func(t *testing.T) {
		name := "商品A"
		category := "food"
		price := int64(500)
		taxType := "included"
		taxRate := 0.08
		isActive := true
		sortOrder := 3

		fields := buildMerchandiseItemUpdate(&UpdateMerchandiseItemInput{
			Name:      &name,
			Category:  &category,
			UnitPrice: &price,
			TaxType:   &taxType,
			TaxRate:   &taxRate,
			IsActive:  &isActive,
			SortOrder: &sortOrder,
		})

		assert.Equal(t, name, fields[colMerchandiseItemName])
		assert.Equal(t, category, fields[colMerchandiseItemCategory])
		assert.Equal(t, price, fields[colMerchandiseItemUnitPrice])
		assert.Equal(t, taxType, fields[colMerchandiseItemTaxType])
		assert.Equal(t, taxRate, fields[colMerchandiseItemTaxRate])
		assert.Equal(t, isActive, fields[colMerchandiseItemIsActive])
		assert.Equal(t, sortOrder, fields[colMerchandiseItemSortOrder])
	})

	t.Run("IsActive=falseも明示的にmapへ含まれる（GORMゼロ値スキップ回避）", func(t *testing.T) {
		isActive := false
		fields := buildMerchandiseItemUpdate(&UpdateMerchandiseItemInput{IsActive: &isActive})
		val, ok := fields[colMerchandiseItemIsActive]
		assert.True(t, ok, "IsActive=false でもキーが存在するべき")
		assert.Equal(t, false, val)
	})
}

// ---- Atomic delete concurrent interleavings (BE-ACT-MERCHANDISE-ATOMIC-DELETE) ----

// TestMerchandiseItemService_Delete_ConcurrentAttachFirstYieldsConflict proves attach-first:
// campaign target attachment holds FOR SHARE, concurrent Delete waits, then after attach
// commits CountUsage sees the target and returns Conflict (soft-delete rolls back).
func TestMerchandiseItemService_Delete_ConcurrentAttachFirstYieldsConflict(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	svc := NewMerchandiseItemService(repo, persistence.NewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now().UTC().Truncate(24 * time.Hour)

	item := makeMerchItem(t, db, clinicID, "attach-first merch", model.ItemCategoryGoods)
	camp := makeMerchCampaign(t, db, clinicID, "attach-first camp", true, now, now.Add(24*time.Hour))

	// Hold FOR SHARE on merchandise (same path as campaign target validation) and insert target.
	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			if _, err := repo.FindByID(txCtx, clinicID, item.ID); err != nil {
				return err
			}
			if err := tx.Create(&model.CampaignTargetItem{
				CampaignID:        camp.ID,
				MerchandiseItemID: item.ID,
			}).Error; err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- svc.Delete(ctx, clinicID, item.ID)
	}()

	// Soft-delete exclusive lock must wait behind the attacher's FOR SHARE.
	select {
	case err := <-deleteDone:
		close(release)
		require.Failf(t, "merchandise delete was not serialized behind campaign attach share-lock", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		// still waiting — good
	}

	close(release)
	require.NoError(t, <-holderDone)

	deleteErr := <-deleteDone
	require.Error(t, deleteErr)
	assert.True(t, apperrors.IsConflict(deleteErr), "attach-first must yield Conflict after serialization, got %v", deleteErr)
	assert.False(t, apperrors.IsNotFound(deleteErr))

	// Merchandise row must remain (soft-delete rolled back).
	got, err := repo.FindByID(ctx, clinicID, item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, got.ID)
}

// TestMerchandiseItemService_Delete_ConcurrentDeleteFirstRejectsLaterAttach proves delete-first:
// atomic Delete soft-deletes with zero usage; later attach FindByID rejects the inactive row.
func TestMerchandiseItemService_Delete_ConcurrentDeleteFirstRejectsLaterAttach(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	svc := NewMerchandiseItemService(repo, persistence.NewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)

	item := makeMerchItem(t, db, clinicID, "delete-first merch", model.ItemCategoryGoods)
	require.NoError(t, svc.Delete(ctx, clinicID, item.ID))

	// Later attach path: ambient FindByID (FOR SHARE) must NotFound soft-deleted merchandise.
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		_, findErr := repo.FindByID(txCtx, clinicID, item.ID)
		return findErr
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "delete-first must make later attach reject inactive row as NotFound")
}

// TestMerchandiseItemService_Delete_InactiveCampaignTargetBlocksWithConflict is the fixture
// that inactive / out-of-date campaign targets still Conflict on Delete.
func TestMerchandiseItemService_Delete_InactiveCampaignTargetBlocksWithConflict(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	svc := NewMerchandiseItemService(repo, persistence.NewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now().UTC().Truncate(24 * time.Hour)

	item := makeMerchItem(t, db, clinicID, "inactive camp block merch", model.ItemCategoryGoods)
	// Inactive + expired date window — still blocks.
	camp := makeMerchCampaign(t, db, clinicID, "inactive expired camp", false, now.Add(-90*24*time.Hour), now.Add(-60*24*time.Hour))
	makeMerchCampaignTarget(t, db, camp.ID, item.ID)

	err := svc.Delete(ctx, clinicID, item.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "inactive/out-of-date campaign target must Conflict, got %v", err)
	assert.False(t, apperrors.IsNotFound(err))

	got, findErr := repo.FindByID(ctx, clinicID, item.ID)
	require.NoError(t, findErr)
	assert.Equal(t, item.ID, got.ID, "Conflict must leave the merchandise row active")
}

// TestMerchandiseItemService_Delete_NotFoundDistinctFromConflict ensures missing rows stay NotFound.
func TestMerchandiseItemService_Delete_NotFoundDistinctFromConflict_RealDB(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	svc := NewMerchandiseItemService(repo, persistence.NewTransactor(db))
	ctx := context.Background()

	err := svc.Delete(ctx, 1, 999999)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}
