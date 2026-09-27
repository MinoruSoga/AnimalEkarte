package medicalrecord

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CarePlanItem モック ----

type mockCarePlanItemRepository struct {
	listByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	findByIDFn                func(ctx context.Context, clinicID, itemID uint64) (*model.CarePlanItem, error)
	createFn                  func(ctx context.Context, item *model.CarePlanItem) error
	updateFn                  func(ctx context.Context, clinicID, itemID uint64, cmd UpdateCarePlanItemInput) error
	deleteFn                  func(ctx context.Context, clinicID, itemID uint64) error
}

func (m *mockCarePlanItemRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	return m.listByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockCarePlanItemRepository) FindByID(ctx context.Context, clinicID, itemID uint64) (*model.CarePlanItem, error) {
	return m.findByIDFn(ctx, clinicID, itemID)
}

func (m *mockCarePlanItemRepository) Create(ctx context.Context, item *model.CarePlanItem) error {
	return m.createFn(ctx, item)
}

func (m *mockCarePlanItemRepository) Update(ctx context.Context, clinicID, itemID uint64, cmd UpdateCarePlanItemInput) error {
	return m.updateFn(ctx, clinicID, itemID, cmd)
}

func (m *mockCarePlanItemRepository) Delete(ctx context.Context, clinicID, itemID uint64) error {
	return m.deleteFn(ctx, clinicID, itemID)
}

// okHospRepoForCarePlan は親入院の所有権検証が成功する（同一クリニック）モックを返す。
func okHospRepoForCarePlan() *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{}, nil
		},
	}
}

type passthroughCarePlanTransactor struct{}

func (passthroughCarePlanTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type okCarePlanAuditTx struct{}

func (okCarePlanAuditTx) LogEntryTx(context.Context, *AuditEntry) error { return nil }

func newTestCarePlanItemService(repo CarePlanItemRepository, hospRepo HospitalizationRepository, medicineRepo medicineFinder, procedureRepo procedureFinder, hospPlanRepo HospitalizationPlanRepository) CarePlanItemService {
	return NewCarePlanItemService(repo, hospRepo, medicineRepo, procedureRepo, hospPlanRepo, passthroughCarePlanTransactor{}, okCarePlanAuditTx{})
}

// ---- Tests ----

func TestBuildCarePlanItemUpdate(t *testing.T) {
	typ := string(model.CarePlanTypeFood)
	name := "更新名"
	description := "更新説明"
	status := string(model.CarePlanStatusCompleted)
	notes := "メモ"
	medicineID := uint64(1)
	procedureID := uint64(2)
	hospPlanID := uint64(3)
	unitPrice := int64(500)
	category := "カテゴリ"
	sortOrder := 4
	manualTrue := true
	manualFalse := false
	reason := "持ち込み療養食のため"

	tests := []struct {
		name  string
		input *UpdateCarePlanItemInput
		want  map[string]any
	}{
		{
			name:  "no fields set returns empty map",
			input: &UpdateCarePlanItemInput{},
			want:  map[string]any{},
		},
		{
			name:  "only type set",
			input: &UpdateCarePlanItemInput{Type: &typ},
			want:  map[string]any{"type": typ},
		},
		{
			name:  "only name set",
			input: &UpdateCarePlanItemInput{Name: &name},
			want:  map[string]any{"name": name},
		},
		{
			name:  "only description set",
			input: &UpdateCarePlanItemInput{Description: &description},
			want:  map[string]any{"description": description},
		},
		{
			name:  "timing set writes pq.StringArray",
			input: &UpdateCarePlanItemInput{Timing: []string{"morning", "evening"}},
			want:  map[string]any{"timing": pq.StringArray{"morning", "evening"}},
		},
		{
			name:  "timing set to empty slice still writes (explicit clear)",
			input: &UpdateCarePlanItemInput{Timing: []string{}},
			want:  map[string]any{"timing": pq.StringArray{}},
		},
		{
			name:  "only status set",
			input: &UpdateCarePlanItemInput{Status: &status},
			want:  map[string]any{"status": status},
		},
		{
			name:  "only notes set",
			input: &UpdateCarePlanItemInput{Notes: &notes},
			want:  map[string]any{"notes": notes},
		},
		{
			name:  "only medicine_id set",
			input: &UpdateCarePlanItemInput{MedicineID: &medicineID},
			want:  map[string]any{"medicine_id": medicineID},
		},
		{
			name:  "only procedure_id set",
			input: &UpdateCarePlanItemInput{ProcedureID: &procedureID},
			want:  map[string]any{"procedure_id": procedureID},
		},
		{
			name:  "only hospitalization_plan_id set",
			input: &UpdateCarePlanItemInput{HospitalizationPlanID: &hospPlanID},
			want:  map[string]any{"hospitalization_plan_id": hospPlanID},
		},
		{
			name:  "only unit_price set",
			input: &UpdateCarePlanItemInput{UnitPrice: &unitPrice},
			want:  map[string]any{"unit_price": unitPrice},
		},
		{
			name:  "only category set",
			input: &UpdateCarePlanItemInput{Category: &category},
			want:  map[string]any{"category": category},
		},
		{
			name:  "only sort_order set",
			input: &UpdateCarePlanItemInput{SortOrder: &sortOrder},
			want:  map[string]any{"sort_order": sortOrder},
		},
		{
			name:  "manual=true clears hospitalization_plan_id and forces category=other",
			input: &UpdateCarePlanItemInput{Manual: &manualTrue},
			want:  map[string]any{"hospitalization_plan_id": nil, "category": "other"},
		},
		{
			name:  "manual=false alone writes nothing for the manual flag",
			input: &UpdateCarePlanItemInput{Manual: &manualFalse},
			want:  map[string]any{},
		},
		{
			name:  "other_reason set",
			input: &UpdateCarePlanItemInput{OtherReason: &reason},
			want:  map[string]any{"other_reason": reason},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCarePlanItemUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCarePlanItemService_List(t *testing.T) {
	tests := []struct {
		name              string
		hospitalizationID uint64
		repoItems         []model.CarePlanItem
		repoErr           error
		wantLen           int
		wantErr           bool
	}{
		{
			name:              "returns items for hospitalization",
			hospitalizationID: 1,
			repoItems: []model.CarePlanItem{
				{ID: 1, HospitalizationID: 1, Type: model.CarePlanTypeFood, Name: "Breakfast"},
				{ID: 2, HospitalizationID: 1, Type: model.CarePlanTypeMedicine, Name: "Medication"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:              "returns empty list when no items exist",
			hospitalizationID: 999,
			repoItems:         []model.CarePlanItem{},
			repoErr:           nil,
			wantLen:           0,
			wantErr:           false,
		},
		{
			name:              "propagates repository error",
			hospitalizationID: 1,
			repoItems:         nil,
			repoErr:           errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCarePlanItemRepository{
				listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
					return tt.repoItems, tt.repoErr
				},
			}
			svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

			items, err := svc.List(context.Background(), 1, tt.hospitalizationID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
			}
		})
	}
}

func TestCarePlanItemService_Create(t *testing.T) {
	hospitalizationID := uint64(1)
	medicineID := uint64(10)
	sortOrder := 1
	hospPlanID := uint64(30)

	tests := []struct {
		name              string
		hospitalizationID uint64
		input             *CreateCarePlanItemInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "creates care plan item with valid type",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeFood),
				Name:        "Breakfast",
				Description: "Regular diet",
				Status:      string(model.CarePlanStatusActive),
				SortOrder:   sortOrder,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "creates item with default status",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:       string(model.CarePlanTypeMedicine),
				Name:       "Pain relief",
				MedicineID: &medicineID,
				Status:     "", // Will default to Active
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error on invalid type",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type: "invalid_type",
				Name: "Test",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:              "returns error on invalid status",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:   string(model.CarePlanTypeFood),
				Name:   "Test",
				Status: "invalid_status",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:              "returns error when repository fails",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type: string(model.CarePlanTypeFood),
				Name: "Test",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:              "creates manual item with category=other and reason",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "持ち込みおもちゃ",
				Manual:      true,
				UnitPrice:   800,
				OtherReason: "  持ち込み品のため  ",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "manual=true on non-item type is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeMedicine),
				Name:        "x",
				Manual:      true,
				OtherReason: "理由",
			},
			wantErr: true,
		},
		{
			name:              "manual=true with hospitalization_plan_id is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:                  string(model.CarePlanTypeItem),
				Name:                  "x",
				Manual:                true,
				HospitalizationPlanID: &hospPlanID,
				OtherReason:           "理由",
			},
			wantErr: true,
		},
		{
			name:              "manual=true with missing reason is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:   string(model.CarePlanTypeItem),
				Name:   "x",
				Manual: true,
			},
			wantErr: true,
		},
		{
			name:              "manual=true with blank reason is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "x",
				Manual:      true,
				OtherReason: "   ",
			},
			wantErr: true,
		},
		{
			name:              "manual=true with over-500-rune reason is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "x",
				Manual:      true,
				OtherReason: strings.Repeat("あ", 501),
			},
			wantErr: true,
		},
		{
			name:              "manual=true with non-other category is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "x",
				Manual:      true,
				Category:    "goods",
				OtherReason: "理由",
			},
			wantErr: true,
		},
		{
			name:              "non-manual with other_reason is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "x",
				OtherReason: "理由",
			},
			wantErr: true,
		},
		{
			name:              "negative unit_price is rejected",
			hospitalizationID: hospitalizationID,
			input: &CreateCarePlanItemInput{
				Type:        string(model.CarePlanTypeItem),
				Name:        "x",
				Manual:      true,
				UnitPrice:   -1,
				OtherReason: "理由",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCarePlanItemRepository{
				createFn: func(_ context.Context, _ *model.CarePlanItem) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

			item, err := svc.Create(context.Background(), 1, tt.hospitalizationID, tt.input)

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

// TestCarePlanItemService_Create_ManualPersistsContractShape は有効な手入力作成時に
// 正規化された形（category='other' / ref NULL / trim 済み other_reason）で永続化されることを検証する。
func TestCarePlanItemService_Create_ManualPersistsContractShape(t *testing.T) {
	var persisted *model.CarePlanItem
	repo := &mockCarePlanItemRepository{
		createFn: func(_ context.Context, item *model.CarePlanItem) error {
			persisted = item
			return nil
		},
		findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
			return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
		},
	}
	svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

	item, err := svc.Create(context.Background(), 1, 1, &CreateCarePlanItemInput{
		Type:        string(model.CarePlanTypeItem),
		Name:        "持ち込みおもちゃ",
		Manual:      true,
		UnitPrice:   800,
		OtherReason: "  持ち込み品のため  ",
	})

	assert.NoError(t, err)
	assert.NotNil(t, item)
	require.NotNil(t, persisted)
	assert.Nil(t, persisted.HospitalizationPlanID, "手入力行はマスタ参照を持たない")
	assert.Equal(t, "other", persisted.Category, "手入力行は category=other に正規化される")
	assert.Equal(t, "持ち込み品のため", persisted.OtherReason, "other_reason は trim されて保存される")
}

// TestCarePlanItemService_ValidateMasterFKs_ProcedureOwnership は procedure_id の
// クロステナント所有権検証（validateMasterFKs の procedureID 分岐）を Create 経由で検証する。
// medicine_id 分岐は cross_tenant_master_fk_write_test.go で既にカバーされているため、
// ここでは procedure_id 分岐のみを対象にする。
func TestCarePlanItemService_ValidateMasterFKs_ProcedureOwnership(t *testing.T) {
	const clinicID = uint64(1)
	const ownedProcedureID = uint64(20)
	const foreignProcedureID = uint64(888)

	t.Run("rejects cross-clinic procedure_id and does not persist", func(t *testing.T) {
		created := false
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), rejectProcedureRepo(ownedProcedureID), okHospitalizationPlanRepo())

		foreign := foreignProcedureID
		item, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeTreatment), Name: "x", ProcedureID: &foreign,
		})

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.False(t, created, "care plan item must NOT be persisted referencing another clinic's procedure")
	})

	t.Run("accepts same-clinic procedure_id (no false-reject)", func(t *testing.T) {
		created := false
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), rejectProcedureRepo(ownedProcedureID), okHospitalizationPlanRepo())

		owned := ownedProcedureID
		item, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeTreatment), Name: "x", ProcedureID: &owned,
		})

		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.True(t, created)
	})
}

func TestCarePlanItemService_Update(t *testing.T) {
	newName := "Updated Food"
	newStatus := string(model.CarePlanStatusCompleted)

	tests := []struct {
		name                      string
		hospitalizationID         uint64
		itemID                    uint64
		input                     *UpdateCarePlanItemInput
		repoItemHospitalizationID uint64
		repoUpdateErr             error
		repoReturnItem            *model.CarePlanItem
		wantErr                   bool
	}{
		{
			name:              "updates item successfully",
			hospitalizationID: 1,
			itemID:            1,
			input: &UpdateCarePlanItemInput{
				Name:   &newName,
				Status: &newStatus,
			},
			repoItemHospitalizationID: 1,
			repoUpdateErr:             nil,
			repoReturnItem: &model.CarePlanItem{
				ID:                1,
				HospitalizationID: 1,
				Name:              newName,
				Status:            model.CarePlanStatusCompleted,
			},
			wantErr: false,
		},
		{
			name:                      "returns error when no fields provided",
			hospitalizationID:         1,
			itemID:                    1,
			input:                     &UpdateCarePlanItemInput{},
			repoItemHospitalizationID: 1,
			repoUpdateErr:             nil,
			repoReturnItem:            nil,
			wantErr:                   true,
		},
		{
			name:              "returns error when item doesn't belong to hospitalization",
			hospitalizationID: 1,
			itemID:            1,
			input: &UpdateCarePlanItemInput{
				Name: &newName,
			},
			repoItemHospitalizationID: 2,
			repoUpdateErr:             nil,
			repoReturnItem: &model.CarePlanItem{
				ID:                1,
				HospitalizationID: 2,
			},
			wantErr: true,
		},
		{
			name:              "returns error when update fails",
			hospitalizationID: 1,
			itemID:            1,
			input: &UpdateCarePlanItemInput{
				Name: &newName,
			},
			repoItemHospitalizationID: 1,
			repoUpdateErr:             errors.New("db error"),
			repoReturnItem:            nil,
			wantErr:                   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCarePlanItemRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{
						ID:                tt.itemID,
						HospitalizationID: tt.repoItemHospitalizationID,
					}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateCarePlanItemInput) error {
					return tt.repoUpdateErr
				},
			}
			svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

			item, err := svc.Update(context.Background(), 1, tt.hospitalizationID, tt.itemID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

func TestCarePlanItemService_Update_InvalidType(t *testing.T) {
	invalidType := "invalid_type"
	repo := &mockCarePlanItemRepository{
		findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
			return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
		},
	}
	svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

	item, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCarePlanItemInput{Type: &invalidType})

	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestCarePlanItemService_Update_InvalidStatus(t *testing.T) {
	invalidStatus := "invalid_status"
	repo := &mockCarePlanItemRepository{
		findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
			return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
		},
	}
	svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

	item, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCarePlanItemInput{Status: &invalidStatus})

	assert.Error(t, err)
	assert.Nil(t, item)
}

// TestCarePlanItemService_Update_ManualTransitions は manual<->master の双方向遷移と
// 拒否マトリクスを検証する（repo へ渡る正規化済み input を捕捉して判定）。
func TestCarePlanItemService_Update_ManualTransitions(t *testing.T) {
	const itemID = uint64(7)
	hospPlanID := uint64(50)
	manualTrue := true
	manualFalse := false
	reason := "持ち込み品のため"
	itemType := string(model.CarePlanTypeItem)
	foodType := string(model.CarePlanTypeFood)
	negativePrice := int64(-1)

	manualExisting := &model.CarePlanItem{
		ID: itemID, HospitalizationID: 1, Type: model.CarePlanTypeItem,
		Category: "other", OtherReason: "既存理由",
	}
	masterExisting := &model.CarePlanItem{
		ID: itemID, HospitalizationID: 1, Type: model.CarePlanTypeItem,
		HospitalizationPlanID: &hospPlanID, Category: "goods",
	}

	tests := []struct {
		name         string
		existing     *model.CarePlanItem
		input        *UpdateCarePlanItemInput
		wantErr      bool
		wantCmdCheck func(t *testing.T, cmd UpdateCarePlanItemInput)
	}{
		{
			name:     "master->manual clears ref and forces category=other with reason",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{Manual: &manualTrue, OtherReason: &reason},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.Manual)
				assert.True(t, *cmd.Manual)
				require.NotNil(t, cmd.Category)
				assert.Equal(t, "other", *cmd.Category)
				require.NotNil(t, cmd.OtherReason)
				assert.Equal(t, reason, *cmd.OtherReason)
			},
		},
		{
			name:     "master->manual without reason is rejected",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{Manual: &manualTrue},
			wantErr:  true,
		},
		{
			name:     "manual->master sets ref and clears category/other_reason",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{HospitalizationPlanID: &hospPlanID},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.HospitalizationPlanID)
				assert.Equal(t, hospPlanID, *cmd.HospitalizationPlanID)
				require.NotNil(t, cmd.Category, "category=other は手入力終了時にクリアされる")
				assert.Equal(t, "", *cmd.Category)
				require.NotNil(t, cmd.OtherReason, "other_reason は手入力終了時にクリアされる")
				assert.Equal(t, "", *cmd.OtherReason)
			},
		},
		{
			name:     "manual item keeps manual shape when editing reason only",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{OtherReason: ptrStr("  新しい理由  ")},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.OtherReason)
				assert.Equal(t, "新しい理由", *cmd.OtherReason, "other_reason は trim されて保存される")
			},
		},
		{
			name:     "manual item rejects blank reason edit",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{OtherReason: ptrStr("   ")},
			wantErr:  true,
		},
		{
			name:     "manual item rejects non-other category while staying manual",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{Category: ptrStr("goods")},
			wantErr:  true,
		},
		{
			name:     "manual item rejects manual=false without a master ref",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{Manual: &manualFalse},
			wantErr:  true,
		},
		{
			name:     "manual=true conflicts with hospitalization_plan_id",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{Manual: &manualTrue, HospitalizationPlanID: &hospPlanID, OtherReason: &reason},
			wantErr:  true,
		},
		{
			name:     "manual=true on non-item merged type is rejected",
			existing: &model.CarePlanItem{ID: itemID, HospitalizationID: 1, Type: model.CarePlanTypeFood},
			input:    &UpdateCarePlanItemInput{Manual: &manualTrue, OtherReason: &reason},
			wantErr:  true,
		},
		{
			name:     "non-item to item without ref or manual is rejected",
			existing: &model.CarePlanItem{ID: itemID, HospitalizationID: 1, Type: model.CarePlanTypeFood},
			input:    &UpdateCarePlanItemInput{Type: &itemType},
			wantErr:  true,
		},
		{
			name:     "other_reason on master-referenced item is rejected",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{OtherReason: &reason},
			wantErr:  true,
		},
		{
			name:     "manual->master with explicit manual=false and new ref clears metadata",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{Manual: &manualFalse, HospitalizationPlanID: &hospPlanID},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.Category)
				assert.Equal(t, "", *cmd.Category)
				require.NotNil(t, cmd.OtherReason)
				assert.Equal(t, "", *cmd.OtherReason)
			},
		},
		{
			name:     "manual->non-item type switch clears manual metadata",
			existing: manualExisting,
			input:    &UpdateCarePlanItemInput{Type: &foodType},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.Type)
				assert.Equal(t, foodType, *cmd.Type)
				require.NotNil(t, cmd.Category)
				assert.Equal(t, "", *cmd.Category)
				require.NotNil(t, cmd.OtherReason)
				assert.Equal(t, "", *cmd.OtherReason)
			},
		},
		{
			name:     "whitespace-only other_reason on non-manual item is normalized to empty",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{OtherReason: ptrStr("   ")},
			wantErr:  false,
			wantCmdCheck: func(t *testing.T, cmd UpdateCarePlanItemInput) {
				require.NotNil(t, cmd.OtherReason)
				assert.Equal(t, "", *cmd.OtherReason)
			},
		},
		{
			name:     "negative unit_price is rejected",
			existing: masterExisting,
			input:    &UpdateCarePlanItemInput{UnitPrice: &negativePrice},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *UpdateCarePlanItemInput
			repo := &mockCarePlanItemRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CarePlanItem, error) {
					return tt.existing, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, cmd UpdateCarePlanItemInput) error {
					captured = &cmd
					return nil
				},
			}
			svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

			item, err := svc.Update(context.Background(), 1, 1, itemID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "想定外のエラー種別: %v", err)
				assert.Nil(t, captured, "拒否時は repo.Update が呼ばれない")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				require.NotNil(t, captured, "受理時は repo.Update が呼ばれる")
				if tt.wantCmdCheck != nil {
					tt.wantCmdCheck(t, *captured)
				}
			}
		})
	}
}

func ptrStr(s string) *string { return &s }

// TestCarePlanItemService_Update_ReloadError は Update 成功後の再取得
// （s.repo.FindByID の2回目呼び出し）が失敗した場合にラップされたエラーを返すことを検証する。
func TestCarePlanItemService_Update_ReloadError(t *testing.T) {
	newName := "更新後"
	callCount := 0
	repo := &mockCarePlanItemRepository{
		findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
			callCount++
			if callCount == 1 {
				return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
			}
			return nil, errors.New("db error on reload")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateCarePlanItemInput) error {
			return nil
		},
	}
	svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

	item, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCarePlanItemInput{Name: &newName})

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Equal(t, 2, callCount)
}

func TestCarePlanItemService_Delete(t *testing.T) {
	tests := []struct {
		name                      string
		hospitalizationID         uint64
		itemID                    uint64
		repoItemHospitalizationID uint64
		repoDeleteErr             error
		wantErr                   bool
	}{
		{
			name:                      "deletes item successfully",
			hospitalizationID:         1,
			itemID:                    1,
			repoItemHospitalizationID: 1,
			repoDeleteErr:             nil,
			wantErr:                   false,
		},
		{
			name:                      "returns error when item doesn't belong to hospitalization",
			hospitalizationID:         1,
			itemID:                    1,
			repoItemHospitalizationID: 2,
			repoDeleteErr:             nil,
			wantErr:                   true,
		},
		{
			name:                      "returns error when delete fails",
			hospitalizationID:         1,
			itemID:                    1,
			repoItemHospitalizationID: 1,
			repoDeleteErr:             errors.New("db error"),
			wantErr:                   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCarePlanItemRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{
						ID:                tt.itemID,
						HospitalizationID: tt.repoItemHospitalizationID,
					}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoDeleteErr
				},
			}
			svc := newTestCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

			err := svc.Delete(context.Background(), 1, tt.hospitalizationID, tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCarePlanItemService_Create_CrossTenantParentRejected は
// clinic A の呼び出しが clinic B の入院を参照する Create を拒否し、
// 子レコード（care_plan_items は自前 clinic_id を持たない）が永続化されない
// （repo.Create が呼ばれない）ことを検証する。
// 親所有権検証（hospRepo.FindByID(clinicID, hospID)）を削除すると必ず失敗する回帰テスト。
func TestCarePlanItemService_Create_CrossTenantParentRejected(t *testing.T) {
	const (
		clinicA       = uint64(1)
		clinicBHospID = uint64(99) // clinic B の入院 ID（clinic A は所有しない）
	)
	createCalled := false
	repo := &mockCarePlanItemRepository{
		createFn: func(_ context.Context, _ *model.CarePlanItem) error {
			createCalled = true
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CarePlanItem, error) {
			return &model.CarePlanItem{}, nil
		},
	}
	// clinic A のスコープからは clinic B の入院は見えない → NotFound（クロステナント）。
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "99")
		},
	}
	svc := newTestCarePlanItemService(repo, hospRepo, okMedicineRepo(), okProcedureRepo(), okHospitalizationPlanRepo())

	item, err := svc.Create(context.Background(), clinicA, clinicBHospID, &CreateCarePlanItemInput{
		Type: string(model.CarePlanTypeMedicine),
		Name: "cross-tenant injection",
	})

	assert.Error(t, err, "clinic A から clinic B の入院へのケアプラン write は拒否されるべき")
	assert.True(t, apperrors.IsNotFound(err), "拒否は NotFound(404) にマップされるべき: %v", err)
	assert.Nil(t, item)
	assert.False(t, createCalled, "親所有権検証に失敗した場合、子レコードを永続化してはならない")
}
