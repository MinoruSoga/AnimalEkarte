package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CarePlanItem モック ----

type mockCarePlanItemRepository struct {
	listByHospitalizationIDFn func(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error)
	findByIDFn                func(ctx context.Context, itemID uint64) (*model.CarePlanItem, error)
	createFn                  func(ctx context.Context, item *model.CarePlanItem) error
	updateFn                  func(ctx context.Context, itemID uint64, fields map[string]any) error
	deleteFn                  func(ctx context.Context, itemID uint64) error
}

func (m *mockCarePlanItemRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	return m.listByHospitalizationIDFn(ctx, hospitalizationID)
}

func (m *mockCarePlanItemRepository) FindByID(ctx context.Context, itemID uint64) (*model.CarePlanItem, error) {
	return m.findByIDFn(ctx, itemID)
}

func (m *mockCarePlanItemRepository) Create(ctx context.Context, item *model.CarePlanItem) error {
	return m.createFn(ctx, item)
}

func (m *mockCarePlanItemRepository) Update(ctx context.Context, itemID uint64, fields map[string]any) error {
	return m.updateFn(ctx, itemID, fields)
}

func (m *mockCarePlanItemRepository) Delete(ctx context.Context, itemID uint64) error {
	return m.deleteFn(ctx, itemID)
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
				listByHospitalizationIDFn: func(_ context.Context, _ uint64) ([]model.CarePlanItem, error) {
					return tt.repoItems, tt.repoErr
				},
			}
			svc := NewCarePlanItemService(repo)

			items, err := svc.List(context.Background(), tt.hospitalizationID)

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
				findByIDFn: func(_ context.Context, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewCarePlanItemService(repo)

			item, err := svc.Create(context.Background(), tt.hospitalizationID, tt.input)

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
			repoItemHospitalizationID: 2, // Different hospitalization
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
				findByIDFn: func(_ context.Context, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{
						ID:                1,
						HospitalizationID: tt.repoItemHospitalizationID,
					}, nil
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			repo.findByIDFn = func(_ context.Context, _ uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{
					ID:                tt.itemID,
					HospitalizationID: tt.repoItemHospitalizationID,
				}, nil
			}
			svc := NewCarePlanItemService(repo)

			item, err := svc.Update(context.Background(), tt.hospitalizationID, tt.itemID, tt.input)

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
			repoItemHospitalizationID: 2, // Different hospitalization
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
				findByIDFn: func(_ context.Context, _ uint64) (*model.CarePlanItem, error) {
					return &model.CarePlanItem{
						ID:                tt.itemID,
						HospitalizationID: tt.repoItemHospitalizationID,
					}, nil
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoDeleteErr
				},
			}
			svc := NewCarePlanItemService(repo)

			err := svc.Delete(context.Background(), tt.hospitalizationID, tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
