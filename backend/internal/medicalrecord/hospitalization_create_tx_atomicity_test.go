package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestHospitalizationService_Create_NestedPlansAtomicity verifies that nested
// treatment plans share the parent create transaction: plan failure aborts parent
// create, and success creates all plans with the same clinic and hospitalization ID.
func TestHospitalizationService_Create_NestedPlansAtomicity(t *testing.T) {
	now := time.Now()
	const clinicID uint64 = 1

	t.Run("plan failure rolls back parent create", func(t *testing.T) {
		parentCreated := false
		planCreateCalls := 0
		hospRepo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, h *model.Hospitalization) error {
				parentCreated = true
				h.ID = 99
				return nil
			},
		}
		planRepo := &mockTreatmentPlanRepository{
			createFn: func(_ context.Context, _ *model.TreatmentPlan) error {
				planCreateCalls++
				return errors.New("forced plan failure")
			},
		}
		tx := &mockTransactor{}
		svc := NewHospitalizationService(
			hospRepo,
			&mockReservationRepository{
				assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return 2, nil
				},
			},
			&mockPetRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
					return &model.Pet{ID: id}, nil
				},
			},
			nil, nil, nil, nil,
			tx,
			WithTreatmentPlanRepository(planRepo),
		)

		hosp, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID:             2,
			PetID:               5,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.Add(24 * time.Hour),
			TreatmentPlans: []CreateTreatmentPlanInput{
				{TreatmentContent: "IV fluids", UnitPrice: 1000, Quantity: 1},
			},
		})

		require.Error(t, err)
		assert.Nil(t, hosp)
		assert.True(t, parentCreated, "parent create should run before plan failure")
		assert.Equal(t, 1, planCreateCalls)
		// mockTransactor surfaces callback error as WithTx error → caller must not treat success.
		assert.Error(t, err)
	})

	t.Run("success creates all plans same clinic and hospitalization", func(t *testing.T) {
		var createdPlans []*model.TreatmentPlan
		hospRepo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, h *model.Hospitalization) error {
				h.ID = 42
				h.ClinicID = clinicID
				return nil
			},
		}
		planRepo := &mockTreatmentPlanRepository{
			createFn: func(_ context.Context, plan *model.TreatmentPlan) error {
				created := *plan
				created.ID = uint64(len(createdPlans) + 1)
				createdPlans = append(createdPlans, &created)
				return nil
			},
		}
		svc := NewHospitalizationService(
			hospRepo,
			&mockReservationRepository{
				assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return 2, nil
				},
			},
			&mockPetRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
					return &model.Pet{ID: id}, nil
				},
			},
			nil, nil, nil, nil,
			&mockTransactor{},
			WithTreatmentPlanRepository(planRepo),
		)

		hosp, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID:             2,
			PetID:               5,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.Add(24 * time.Hour),
			TreatmentPlans: []CreateTreatmentPlanInput{
				{TreatmentContent: "adm", UnitPrice: 990, Quantity: 1, SortOrder: 0},
				{TreatmentContent: "monitor", UnitPrice: 500, Quantity: 2, SortOrder: 1},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, hosp)
		assert.Equal(t, uint64(42), hosp.ID)
		require.Len(t, createdPlans, 2)
		for i, plan := range createdPlans {
			assert.Equal(t, clinicID, plan.ClinicID)
			require.NotNil(t, plan.HospitalizationID)
			assert.Equal(t, uint64(42), *plan.HospitalizationID)
			assert.Equal(t, i, plan.SortOrder)
		}
	})

	t.Run("nested plans without plan repo fails closed", func(t *testing.T) {
		svc := NewHospitalizationService(
			&mockHospitalizationRepository{
				createFn: func(_ context.Context, h *model.Hospitalization) error {
					h.ID = 1
					return nil
				},
			},
			&mockReservationRepository{
				assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return 2, nil
				},
			},
			&mockPetRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
					return &model.Pet{ID: id}, nil
				},
			},
			nil, nil, nil, nil,
			&mockTransactor{},
		)

		hosp, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID:             2,
			PetID:               5,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.Add(24 * time.Hour),
			TreatmentPlans: []CreateTreatmentPlanInput{
				{TreatmentContent: "adm", UnitPrice: 100, Quantity: 1},
			},
		})
		require.Error(t, err)
		assert.Nil(t, hosp)
	})
}
