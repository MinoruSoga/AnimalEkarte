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
			nil,
			&mockCarePlanItemRepository{
				createFn: func(_ context.Context, _ *model.CarePlanItem) error {
					t.Fatal("care plan create must not run after plan failure")
					return nil
				},
			},
			nil, nil,
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
		var createdCareItems []*model.CarePlanItem
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
		careRepo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, item *model.CarePlanItem) error {
				created := *item
				created.ID = uint64(len(createdCareItems) + 1)
				createdCareItems = append(createdCareItems, &created)
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
			nil, careRepo, nil, nil,
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
		// BUG-032: nested treatment rows seed care-plan items visible on detail.
		require.Len(t, createdCareItems, 2)
		assert.Equal(t, "adm", createdCareItems[0].Name)
		assert.Equal(t, "monitor", createdCareItems[1].Name)
		assert.Equal(t, model.CarePlanTypeInstruction, createdCareItems[0].Type)
		assert.Equal(t, model.CarePlanStatusActive, createdCareItems[0].Status)
		assert.Equal(t, uint64(42), createdCareItems[0].HospitalizationID)
		assert.Equal(t, int64(0), createdCareItems[0].UnitPrice)
		assert.Equal(t, 0, createdCareItems[0].SortOrder)
		assert.Equal(t, 1, createdCareItems[1].SortOrder)
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
			nil,
			&mockCarePlanItemRepository{
				createFn: func(_ context.Context, _ *model.CarePlanItem) error {
					t.Fatal("care plan create must not run without plan repo")
					return nil
				},
			},
			nil, nil,
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

	// BUG-032: create→read path — nested treatment_plans seed care_plan_items for detail tab.
	t.Run("BUG-032 nested treatment plans seed care plan items", func(t *testing.T) {
		var createdCare []*model.CarePlanItem
		careCreateCalls := 0
		hospRepo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, h *model.Hospitalization) error {
				h.ID = 77
				h.ClinicID = clinicID
				return nil
			},
		}
		planRepo := &mockTreatmentPlanRepository{
			createFn: func(_ context.Context, plan *model.TreatmentPlan) error {
				plan.ID = 1
				return nil
			},
		}
		careRepo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, item *model.CarePlanItem) error {
				careCreateCalls++
				cp := *item
				cp.ID = uint64(careCreateCalls)
				createdCare = append(createdCare, &cp)
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
			nil, careRepo, nil, nil,
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
				{TreatmentContent: "給餌指示", Memo: "朝夕", UnitPrice: 0, Quantity: 1, SortOrder: 0},
				{TreatmentContent: "投薬指示", Memo: "抗菌薬", UnitPrice: 0, Quantity: 1, SortOrder: 1},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, hosp)
		require.Len(t, createdCare, 2)
		assert.Equal(t, "給餌指示", createdCare[0].Name)
		assert.Equal(t, "朝夕", createdCare[0].Notes)
		assert.Equal(t, "投薬指示", createdCare[1].Name)
		assert.Equal(t, "抗菌薬", createdCare[1].Notes)
		assert.Equal(t, model.CarePlanTypeInstruction, createdCare[0].Type)
		assert.Equal(t, uint64(77), createdCare[0].HospitalizationID)
	})

	t.Run("nested plans without care plan repo fails closed", func(t *testing.T) {
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
			WithTreatmentPlanRepository(&mockTreatmentPlanRepository{
				createFn: func(_ context.Context, _ *model.TreatmentPlan) error { return nil },
			}),
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

	t.Run("care plan seed failure rolls back", func(t *testing.T) {
		planCalls := 0
		svc := NewHospitalizationService(
			&mockHospitalizationRepository{
				createFn: func(_ context.Context, h *model.Hospitalization) error {
					h.ID = 3
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
			nil,
			&mockCarePlanItemRepository{
				createFn: func(_ context.Context, _ *model.CarePlanItem) error {
					return errors.New("forced care plan failure")
				},
			},
			nil, nil,
			&mockTransactor{},
			WithTreatmentPlanRepository(&mockTreatmentPlanRepository{
				createFn: func(_ context.Context, _ *model.TreatmentPlan) error {
					planCalls++
					return nil
				},
			}),
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
		assert.Equal(t, 1, planCalls)
	})
}
