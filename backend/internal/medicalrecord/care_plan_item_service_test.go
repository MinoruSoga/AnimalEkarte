package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

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
