package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TreatmentPlan モック ----

type passthroughTreatmentPlanTransactor struct{}

func (passthroughTreatmentPlanTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockTreatmentPlanRepository struct {
	listByMedicalRecordIDFn   func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	listByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	lockByIDForUpdateFn       func(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	createFn                  func(ctx context.Context, plan *model.TreatmentPlan) error
	updateFn                  func(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, cmd UpdateTreatmentPlanInput) error
	deleteFn                  func(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error
}

func (m *mockTreatmentPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockTreatmentPlanRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	return m.listByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockTreatmentPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockTreatmentPlanRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, id)
	}
	// Default: same snapshot as FindByID so existing unit tests keep working.
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockTreatmentPlanRepository) Create(ctx context.Context, plan *model.TreatmentPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockTreatmentPlanRepository) Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, cmd UpdateTreatmentPlanInput) error {
	return m.updateFn(ctx, clinicID, id, medicalRecordID, hospitalizationID, cmd)
}

func (m *mockTreatmentPlanRepository) Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
	return m.deleteFn(ctx, clinicID, id, medicalRecordID, hospitalizationID)
}

// ---- Tests ----

const testClinicIDTP = uint64(1)

// boolPtr は旧 internal/service 共有 helper の最小複製（⑥移動）。
func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

func TestTreatmentPlanService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name:    "returns plan when found",
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTreatmentPlanRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.TreatmentPlan{ID: id, ClinicID: clinicID, TreatmentContent: "Surgery"}, nil
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			plan, err := svc.GetByID(context.Background(), testClinicIDTP, 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, "Surgery", plan.TreatmentContent)
			}
		})
	}
}

func TestBuildTreatmentPlanUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input *UpdateTreatmentPlanInput
		want  map[string]any
	}{
		{
			name:  "全フィールド nil → 空マップ",
			input: &UpdateTreatmentPlanInput{},
			want:  map[string]any{},
		},
		{
			name: "TreatmentContent のみ設定",
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: strPtr("Updated content"),
			},
			want: map[string]any{"treatment_content": "Updated content"},
		},
		{
			name: "全フィールド設定",
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: strPtr("Surgery"),
				Memo:             strPtr("memo"),
				IsInsurance:      boolPtr(true),
				UnitPrice:        int64Ptr(1000),
				Quantity:         float64Ptr(2),
				DiscountRate:     float64Ptr(10),
				DiscountAmount:   int64Ptr(100),
				Subtotal:         int64Ptr(1700),
				SortOrder:        intPtr(3),
			},
			want: map[string]any{
				"treatment_content": "Surgery",
				"memo":              "memo",
				"is_insurance":      true,
				"unit_price":        int64(1000),
				"quantity":          float64(2),
				"discount_rate":     float64(10),
				"discount_amount":   int64(100),
				// subtotal is server-computed in Update, not via buildTreatmentPlanUpdate (MRD-04)
				"sort_order": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTreatmentPlanUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTreatmentPlanService_ListByMedicalRecord(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoPlans       []model.TreatmentPlan
		repoErr         error
		wantLen         int
		wantErr         bool
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
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return tt.repoPlans, tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			plans, err := svc.ListByMedicalRecord(context.Background(), testClinicIDTP, tt.medicalRecordID)

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
		name              string
		hospitalizationID uint64
		repoPlans         []model.TreatmentPlan
		repoErr           error
		wantLen           int
		wantErr           bool
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
				listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return tt.repoPlans, tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			plans, err := svc.ListByHospitalization(context.Background(), testClinicIDTP, tt.hospitalizationID)

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
		name              string
		medicalRecordID   *uint64
		hospitalizationID *uint64
		input             *CreateTreatmentPlanInput
		repoErr           error
		reloadErr         error
		wantErr           bool
	}{
		{
			name:              "creates plan for medical record",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Office visit",
				UnitPrice:        100,
				Quantity:         1,
				IsInsurance:      true,
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
				Quantity:         1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:              "returns error when reload after create fails",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Checkup",
				UnitPrice:        100,
				Quantity:         1,
			},
			reloadErr: errors.New("reload failed"),
			wantErr:   true,
		},
		{
			name:              "rejects non-positive quantity (MRD-04)",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Bad qty",
				UnitPrice:        100,
				Quantity:         0,
			},
			wantErr: true,
		},
		{
			name:              "rejects discount rate above 100 (MRD-04)",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Bad rate",
				UnitPrice:        100,
				Quantity:         1,
				DiscountRate:     1000,
			},
			wantErr: true,
		},
		{
			name:              "ignores client subtotal and recomputes server-side (MRD-04)",
			medicalRecordID:   &medicalRecordID,
			hospitalizationID: nil,
			input: &CreateTreatmentPlanInput{
				TreatmentContent: "Subtotal ignore",
				UnitPrice:        1000,
				Quantity:         2,
				DiscountRate:     10,
				DiscountAmount:   50,
				Subtotal:         -999999,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *model.TreatmentPlan
			repo := &mockTreatmentPlanRepository{
				createFn: func(_ context.Context, plan *model.TreatmentPlan) error {
					if tt.repoErr == nil {
						assert.Equal(t, testClinicIDTP, plan.ClinicID)
						created = plan
					}
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					if tt.reloadErr != nil {
						return nil, tt.reloadErr
					}
					if created != nil {
						return created, nil
					}
					return &model.TreatmentPlan{
						ClinicID:          testClinicIDTP,
						MedicalRecordID:   tt.medicalRecordID,
						HospitalizationID: tt.hospitalizationID,
					}, nil
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			plan, err := svc.Create(context.Background(), testClinicIDTP, tt.medicalRecordID, tt.hospitalizationID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, testClinicIDTP, plan.ClinicID)
				if tt.name == "ignores client subtotal and recomputes server-side (MRD-04)" {
					// 1000*2*(1-0.1)-50 = 1750
					assert.Equal(t, int64(1750), plan.Subtotal)
				}
			}
		})
	}
}

func TestTreatmentPlanService_Update(t *testing.T) {
	newContent := "Updated content"
	newPrice := int64(200)

	tests := []struct {
		name                   string
		id                     uint64
		input                  *UpdateTreatmentPlanInput
		repoUpdateErr          error
		repoReturnPlan         *model.TreatmentPlan
		findByIDErrOnFirstCall bool
		findByIDErrOnReload    bool
		wantErr                bool
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
				Quantity:         1,
			},
			wantErr: false,
		},
		{
			name:           "returns error when no fields provided",
			id:             1,
			input:          &UpdateTreatmentPlanInput{},
			repoUpdateErr:  nil,
			repoReturnPlan: &model.TreatmentPlan{ID: 1, Quantity: 1},
			wantErr:        true,
		},
		{
			name: "returns error when update fails",
			id:   1,
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: &newContent,
			},
			repoUpdateErr:  errors.New("db error"),
			repoReturnPlan: &model.TreatmentPlan{ID: 1, Quantity: 1},
			wantErr:        true,
		},
		{
			name: "returns error when treatment plan not found",
			id:   999,
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: &newContent,
			},
			findByIDErrOnFirstCall: true,
			wantErr:                true,
		},
		{
			name: "returns error when reload after update fails",
			id:   1,
			input: &UpdateTreatmentPlanInput{
				TreatmentContent: &newContent,
			},
			repoUpdateErr:       nil,
			repoReturnPlan:      &model.TreatmentPlan{ID: 1, Quantity: 1},
			findByIDErrOnReload: true,
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			repo := &mockTreatmentPlanRepository{
				updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ UpdateTreatmentPlanInput) error {
					return tt.repoUpdateErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					callCount++
					if callCount == 1 && tt.findByIDErrOnFirstCall {
						return nil, errors.New("not found")
					}
					if callCount == 2 && tt.findByIDErrOnReload {
						return nil, errors.New("reload failed")
					}
					return tt.repoReturnPlan, nil
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			plan, err := svc.Update(context.Background(), testClinicIDTP, tt.id, nil, nil, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}

	t.Run("rejects parent mismatch on update (MRD-03)", func(t *testing.T) {
		mrOwned := uint64(10)
		wrongMR := uint64(99)
		repo := &mockTreatmentPlanRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
				return &model.TreatmentPlan{ID: id, MedicalRecordID: &mrOwned, Quantity: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ UpdateTreatmentPlanInput) error {
				t.Fatal("update must not run on parent mismatch")
				return nil
			},
		}
		svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
		content := "x"
		plan, err := svc.Update(context.Background(), testClinicIDTP, 1, &wrongMR, nil, &UpdateTreatmentPlanInput{TreatmentContent: &content})
		assert.Error(t, err)
		assert.Nil(t, plan)
	})

	t.Run("recomputes subtotal on money update and ignores client subtotal (MRD-04)", func(t *testing.T) {
		price := int64(500)
		qty := float64(3)
		clientSub := int64(-1)
		var captured UpdateTreatmentPlanInput
		repo := &mockTreatmentPlanRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
				return &model.TreatmentPlan{
					ID: id, UnitPrice: 100, Quantity: 1, DiscountRate: 0, DiscountAmount: 0, Subtotal: 100,
				}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, cmd UpdateTreatmentPlanInput) error {
				captured = cmd
				return nil
			},
		}
		svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
		_, err := svc.Update(context.Background(), testClinicIDTP, 1, nil, nil, &UpdateTreatmentPlanInput{
			UnitPrice: &price,
			Quantity:  &qty,
			Subtotal:  &clientSub,
		})
		assert.NoError(t, err)
		assert.Nil(t, captured.Subtotal)
		if assert.NotNil(t, captured.persistSubtotal) {
			assert.Equal(t, int64(1500), *captured.persistSubtotal)
		}
	})
}

// W-002: hospitalization-nested treatment plans are create-time snapshots.
func TestTreatmentPlanService_Update_HospitalizationNestedRejected(t *testing.T) {
	hospID := uint64(5)
	content := "must not apply"
	repo := &mockTreatmentPlanRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
			t.Fatal("repo must not be called for hospitalization-nested update")
			return nil, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ UpdateTreatmentPlanInput) error {
			t.Fatal("update must not run for hospitalization-nested plan")
			return nil
		},
	}
	svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
	plan, err := svc.Update(context.Background(), testClinicIDTP, 1, nil, &hospID, &UpdateTreatmentPlanInput{
		TreatmentContent: &content,
	})
	assert.Nil(t, plan)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	assert.Contains(t, err.Error(), "登録時スナップショット")
}

func TestTreatmentPlanService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          uint64
		findByIDErr error
		repoErr     error
		wantErr     bool
	}{
		{
			name:    "deletes plan successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns error when plan not found",
			id:          999,
			findByIDErr: errors.New("not found"),
			wantErr:     true,
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
				findByIDFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.TreatmentPlan{ID: id}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64, _, _ *uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})

			err := svc.Delete(context.Background(), testClinicIDTP, tt.id, nil, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("rejects parent mismatch on delete (MRD-03)", func(t *testing.T) {
		// Parent-mismatch for hospitalization is unreachable after W-002 snapshot reject
		// (hospitalizationID != nil short-circuits). Keep medical-record parent mismatch.
		mrOwned := uint64(10)
		wrongMR := uint64(99)
		repo := &mockTreatmentPlanRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.TreatmentPlan, error) {
				return &model.TreatmentPlan{ID: id, MedicalRecordID: &mrOwned}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64, _, _ *uint64) error {
				t.Fatal("delete must not run on parent mismatch")
				return nil
			},
		}
		svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
		err := svc.Delete(context.Background(), testClinicIDTP, 1, &wrongMR, nil)
		assert.Error(t, err)
	})
}

// W-002: hospitalization-nested treatment plans are create-time snapshots.
func TestTreatmentPlanService_Delete_HospitalizationNestedRejected(t *testing.T) {
	hospID := uint64(5)
	repo := &mockTreatmentPlanRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
			t.Fatal("repo must not be called for hospitalization-nested delete")
			return nil, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64, _, _ *uint64) error {
			t.Fatal("delete must not run for hospitalization-nested plan")
			return nil
		},
	}
	svc := NewTreatmentPlanService(repo, passthroughTreatmentPlanTransactor{})
	err := svc.Delete(context.Background(), testClinicIDTP, 1, nil, &hospID)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	assert.Contains(t, err.Error(), "登録時スナップショット")
}

// Helper
func uint64P(v uint64) *uint64 {
	return &v
}

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }
