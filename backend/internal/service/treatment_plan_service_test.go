package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TreatmentPlan モック ----

type mockTreatmentPlanRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	listByHospitalizationIDFn func(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	findByIDFn              func(ctx context.Context, id uint64) (*model.TreatmentPlan, error)
	createFn                func(ctx context.Context, plan *model.TreatmentPlan) error
	updateFn                func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, id uint64) error
}

func (m *mockTreatmentPlanRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	return m.listByMedicalRecordIDFn(ctx, medicalRecordID)
}

func (m *mockTreatmentPlanRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	return m.listByHospitalizationIDFn(ctx, hospitalizationID)
}

func (m *mockTreatmentPlanRepository) FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockTreatmentPlanRepository) Create(ctx context.Context, plan *model.TreatmentPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockTreatmentPlanRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, id, fields)
}

func (m *mockTreatmentPlanRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

// ---- Tests ----

func TestTreatmentPlanService_ListByMedicalRecord(t *testing.T) {
	tests := []struct {
		name                 string
		medicalRecordID      uint64
		repoPlans            []model.TreatmentPlan
		repoErr              error
		wantLen              int
		wantErr              bool
	}{
		{
			name:            "returns plans for medical record",
			medicalRecordID: 1,
			repoPlans: []model.TreatmentPlan{
				{ID: 1, MedicalRecordID: uint64P(1), TreatmentContent: "Surgery", UnitPrice: 500, Quantity: 1},
				{ID: 2, MedicalRecordID: uint64P(1), TreatmentContent: "Medication", UnitPrice: 50, Quantity: 5},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no plans exist",
			medicalRecordID: 999,
			repoPlans:       []model.TreatmentPlan{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoPlans:       nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _ uint64) ([]model.TreatmentPlan, error) {
					return tt.repoPlans, tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo)

			plans, err := svc.ListByMedicalRecord(context.Background(), tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, plans, tt.wantLen)
			}
		})
	}
}

func TestTreatmentPlanService_ListByHospitalization(t *testing.T) {
	tests := []struct {
		name                 string
		hospitalizationID    uint64
		repoPlans            []model.TreatmentPlan
		repoErr              error
		wantLen              int
		wantErr              bool
	}{
		{
			name:              "returns plans for hospitalization",
			hospitalizationID: 1,
			repoPlans: []model.TreatmentPlan{
				{ID: 1, HospitalizationID: uint64P(1), TreatmentContent: "Daily care", UnitPrice: 100, Quantity: 10},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:              "returns empty list when no plans exist",
			hospitalizationID: 999,
			repoPlans:         []model.TreatmentPlan{},
			repoErr:           nil,
			wantLen:           0,
			wantErr:           false,
		},
		{
			name:              "propagates repository error",
			hospitalizationID: 1,
			repoPlans:         nil,
			repoErr:           errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				listByHospitalizationIDFn: func(_ context.Context, _ uint64) ([]model.TreatmentPlan, error) {
					return tt.repoPlans, tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo)

			plans, err := svc.ListByHospitalization(context.Background(), tt.hospitalizationID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, plans, tt.wantLen)
			}
		})
	}
}

func TestTreatmentPlanService_Create(t *testing.T) {
	medicalRecordID := uint64(1)
	hospitalizationID := uint64(2)

	tests := []struct {
		name                  string
		medicalRecordID       *uint64
		hospitalizationID     *uint64
		input                 *CreateTreatmentPlanInput
		repoErr               error
		wantErr               bool
	}{
		{
			name:              "creates plan for medical record",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Office visit",
				UnitPrice:        100,
				Quantity:         1,
				Insurance:        true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "creates plan for hospitalization",
			medicalRecordID:   nil,
			hospitalizationID: &hospitalizationID,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Daily monitoring",
				UnitPrice:        50,
				Quantity:         14,
				DiscountRate:     10,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "calculates subtotal when not provided",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Surgery",
				UnitPrice:        1000,
				Quantity:         1,
				Subtotal:         0, // Will be calculated
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error when repository fails",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Test",
				UnitPrice:        100,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				createFn: func(_ context.Context, _ *model.TreatmentPlan) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _ uint64) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{
						MedicalRecordID:   tt.medicalRecordID,
						HospitalizationID: tt.hospitalizationID,
					}, nil
				},
			}
			svc := NewTreatmentPlanService(repo)

			plan, err := svc.Create(context.Background(), tt.medicalRecordID, tt.hospitalizationID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestTreatmentPlanService_Update(t *testing.T) {
	newContent := "Updated content"
	newPrice := 200.0

	tests := []struct {
		name              string
		id                uint64
		input             *UpdateTreatmentPlanInput
		repoUpdateErr     error
		repoReturnPlan    *model.TreatmentPlan
		wantErr           bool
	}{
		{
			name: "updates plan successfully",
			id:   1,
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: &newContent,
				UnitPrice:        &newPrice,
			},
			repoUpdateErr: nil,
			repoReturnPlan: &model.TreatmentPlan{
				ID:               1,
				TreatmentContent: newContent,
				UnitPrice:        newPrice,
			},
			wantErr: false,
		},
		{
			name:              "returns error when no fields provided",
			id:                1,
			input:             &UpdateTreatmentPlanInput{},
			repoUpdateErr:     nil,
			repoReturnPlan:    nil,
			wantErr:           true,
		},
		{
			name: "returns error when update fails",
			id:   1,
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: &newContent,
			},
			repoUpdateErr:  errors.New("db error"),
			repoReturnPlan: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
				findByIDFn: func(_ context.Context, _ uint64) (*model.TreatmentPlan, error) {
					return tt.repoReturnPlan, nil
				},
			}
			svc := NewTreatmentPlanService(repo)

			plan, err := svc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestTreatmentPlanService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "deletes plan successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when plan not found",
			id:      999,
			repoErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "returns error when delete fails",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper
func uint64P(v uint64) *uint64 {
	return &v
}
