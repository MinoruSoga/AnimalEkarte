package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CarePlanItem モック ----

type mockCarePlanItemRepository struct {
	listByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	findByIDFn                func(ctx context.Context, clinicID, itemID uint64) (*model.CarePlanItem, error)
	createFn                  func(ctx context.Context, item *model.CarePlanItem) error
	updateFn                  func(ctx context.Context, clinicID, itemID uint64, fields map[string]any) error
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

func (m *mockCarePlanItemRepository) Update(ctx context.Context, clinicID, itemID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, itemID, fields)
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

// ---- Tests ----

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
			svc := NewCarePlanItemService(repo, okHospRepoForCarePlan())

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
			svc := NewCarePlanItemService(repo, okHospRepoForCarePlan())

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
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewCarePlanItemService(repo, okHospRepoForCarePlan())

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
			svc := NewCarePlanItemService(repo, okHospRepoForCarePlan())

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
	svc := NewCarePlanItemService(repo, hospRepo)

	item, err := svc.Create(context.Background(), clinicA, clinicBHospID, &CreateCarePlanItemInput{
		Type: string(model.CarePlanTypeMedicine),
		Name: "cross-tenant injection",
	})

	assert.Error(t, err, "clinic A から clinic B の入院へのケアプラン write は拒否されるべき")
	assert.True(t, apperrors.IsNotFound(err), "拒否は NotFound(404) にマップされるべき: %v", err)
	assert.Nil(t, item)
	assert.False(t, createCalled, "親所有権検証に失敗した場合、子レコードを永続化してはならない")
}
