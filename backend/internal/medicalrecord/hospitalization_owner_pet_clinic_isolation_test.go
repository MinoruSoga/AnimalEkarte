package medicalrecord

// hospitalization_owner_pet_clinic_isolation_test.go — AUD-004
// 入院 Create/Update の Owner/Pet clinic 所有確認・Owner-Pet 整合、および
// DischargeWithBilling が汚染 Owner/Pet を拒否し（会計あり/なし双方）、会計へ伝播しないことを検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func acceptMatchingOwnerPetReservationRepo(ownedOwnerID, ownedPetID uint64) *mockReservationRepository {
	return &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, clinicID, ownerID uint64) error {
			if clinicID != 1 {
				return apperrors.WrapNotFound("owner", "clinic")
			}
			if ownerID != ownedOwnerID {
				return apperrors.WrapNotFound("owner", "foreign")
			}
			return nil
		},
		findPetOwnerInClinicFn: func(_ context.Context, clinicID, petID uint64) (uint64, error) {
			if clinicID != 1 {
				return 0, apperrors.WrapNotFound("pet", "clinic")
			}
			if petID != ownedPetID {
				return 0, apperrors.WrapNotFound("pet", "foreign")
			}
			return ownedOwnerID, nil
		},
	}
}

func TestHospitalizationService_Create_RejectsCrossClinicOwnerPet(t *testing.T) {
	const clinicID = uint64(1)
	now := time.Now()

	tests := []struct {
		name          string
		ownerID       uint64
		petID         uint64
		assertOwnerFn func(context.Context, uint64, uint64) error
		findPetFn     func(context.Context, uint64, uint64) (uint64, error)
	}{
		{
			name:    "rejects_cross_clinic_owner",
			ownerID: 201,
			petID:   202,
			assertOwnerFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, uint64(201), ownerID)
				return apperrors.WrapNotFound("owner", "201")
			},
		},
		{
			name:    "rejects_cross_clinic_pet",
			ownerID: 10,
			petID:   202,
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, uint64(202), petID)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name:    "rejects_pet_belonging_to_different_owner",
			ownerID: 10,
			petID:   21,
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 999, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			hospRepo := &mockHospitalizationRepository{
				createFn: func(_ context.Context, _ *model.Hospitalization) error {
					created = true
					t.Fatal("hospitalization must not be created with invalid Owner/Pet links")
					return nil
				},
			}
			resRepo := &mockReservationRepository{
				assertOwnerInClinicFn:  tt.assertOwnerFn,
				findPetOwnerInClinicFn: tt.findPetFn,
			}
			svc := NewHospitalizationService(hospRepo, resRepo, nil, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

			got, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
				CageID:              func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:             tt.ownerID,
				PetID:               tt.petID,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
			assert.False(t, created)
		})
	}
}

func TestHospitalizationService_Create_AcceptsSameClinicOwnerPet(t *testing.T) {
	const clinicID, ownedOwnerID, ownedPetID = uint64(1), uint64(10), uint64(20)
	now := time.Now()
	created := false
	hospRepo := &mockHospitalizationRepository{
		createFn: func(_ context.Context, h *model.Hospitalization) error {
			created = true
			h.ID = 55
			assert.Equal(t, ownedOwnerID, h.OwnerID)
			assert.Equal(t, ownedPetID, h.PetID)
			return nil
		},
	}
	svc := NewHospitalizationService(hospRepo, acceptMatchingOwnerPetReservationRepo(ownedOwnerID, ownedPetID), &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id}, nil
		},
	}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

	got, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
		CageID:              func() *uint64 { v := uint64(10); return &v }(),
		OwnerID:             ownedOwnerID,
		PetID:               ownedPetID,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now.Add(24 * time.Hour),
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, created)
}

func TestHospitalizationService_Update_RejectsCrossClinicOwnerPetAndMismatch(t *testing.T) {
	const clinicID = uint64(1)
	existing := &model.Hospitalization{
		ID: 1, ClinicID: clinicID,
		OwnerID: 10, PetID: 20,
		Status: model.HospitalizationStatusAdmitted,
	}

	tests := []struct {
		name          string
		input         *UpdateHospitalizationInput
		assertOwnerFn func(context.Context, uint64, uint64) error
		findPetFn     func(context.Context, uint64, uint64) (uint64, error)
	}{
		{
			name:  "rejects_cross_clinic_owner",
			input: &UpdateHospitalizationInput{OwnerID: uint64PtrHosp(201)},
			assertOwnerFn: func(_ context.Context, _, ownerID uint64) error {
				assert.Equal(t, uint64(201), ownerID)
				return apperrors.WrapNotFound("owner", "201")
			},
		},
		{
			name:  "rejects_cross_clinic_pet",
			input: &UpdateHospitalizationInput{PetID: uint64PtrHosp(202)},
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, _, petID uint64) (uint64, error) {
				assert.Equal(t, uint64(202), petID)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name:  "rejects_pet_only_change_that_breaks_final_owner_pet_consistency",
			input: &UpdateHospitalizationInput{PetID: uint64PtrHosp(21)},
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 999, nil
			},
		},
		{
			name:  "rejects_owner_only_change_that_breaks_final_owner_pet_consistency",
			input: &UpdateHospitalizationInput{OwnerID: uint64PtrHosp(11)},
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, _, petID uint64) (uint64, error) {
				assert.Equal(t, uint64(20), petID)
				return 10, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
					assert.Equal(t, uint64(1), id)
					return existing, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
					updated = true
					t.Fatal("hospitalization must not be updated with invalid Owner/Pet links")
					return nil, nil
				},
			}
			resRepo := &mockReservationRepository{
				assertOwnerInClinicFn:  tt.assertOwnerFn,
				findPetOwnerInClinicFn: tt.findPetFn,
			}
			svc := NewHospitalizationService(hospRepo, resRepo, nil, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

			got, err := svc.Update(context.Background(), clinicID, 1, tt.input)

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
			assert.False(t, updated)
		})
	}
}

func TestHospitalizationService_Update_AcceptsSameClinicFinalOwnerPet(t *testing.T) {
	const clinicID, ownedOwnerID, ownedPetID = uint64(1), uint64(10), uint64(20)
	existing := &model.Hospitalization{
		ID: 1, ClinicID: clinicID,
		OwnerID: ownedOwnerID, PetID: ownedPetID,
		Status: model.HospitalizationStatusAdmitted,
	}
	newPetID := uint64(20)
	updated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Hospitalization, error) {
			updated = true
			assert.Equal(t, newPetID, fields["pet_id"])
			return &model.Hospitalization{ID: 1, ClinicID: clinicID, OwnerID: ownedOwnerID, PetID: newPetID}, nil
		},
	}
	svc := NewHospitalizationService(hospRepo, acceptMatchingOwnerPetReservationRepo(ownedOwnerID, ownedPetID), &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id}, nil
		},
	}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

	got, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{PetID: &newPetID})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, updated)
}

func TestHospitalizationService_DischargeWithBilling_DoesNotPropagateForeignOwnerPet(t *testing.T) {
	const clinicID = uint64(1)
	const foreignOwnerID, foreignPetID = uint64(201), uint64(202)

	updated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID: id, ClinicID: clinicID,
				OwnerID: foreignOwnerID, PetID: foreignPetID,
				Status: model.HospitalizationStatusAdmitted,
			}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
			updated = true
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return nil, nil
		},
	}
	accountingCreated := false
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			accountingCreated = true
			t.Fatalf("Accounting.Create must not receive foreign Owner/Pet (owner=%v pet=%v)", billing.OwnerID, billing.PetID)
			return nil
		},
	}
	resRepo := &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
			assert.Equal(t, foreignOwnerID, ownerID)
			return apperrors.WrapNotFound("owner", "201")
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, nil)
	deps.reservation = resRepo
	svc := deps.svc()

	result, err := svc.DischargeWithBilling(context.Background(), clinicID, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
	assert.False(t, accountingCreated)
	assert.False(t, updated, "contaminated Owner/Pet discharge must not persist status update")
}

func TestHospitalizationService_DischargeWithBilling_RejectsContaminatedOwnerPetAfterOuterFind(t *testing.T) {
	const clinicID = uint64(1)
	const foreignOwnerID, foreignPetID = uint64(201), uint64(202)

	findByIDCallCount := 0
	updated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			findByIDCallCount++
			if findByIDCallCount == 1 {
				return &model.Hospitalization{
					ID: id, ClinicID: clinicID,
					OwnerID: 2, PetID: 5,
					Status: model.HospitalizationStatusAdmitted,
				}, nil
			}
			return &model.Hospitalization{
				ID: id, ClinicID: clinicID,
				OwnerID: foreignOwnerID, PetID: foreignPetID,
				Status: model.HospitalizationStatusAdmitted,
			}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
			updated = true
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return nil, nil
		},
	}
	accountingCreated := false
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			accountingCreated = true
			t.Fatalf("Accounting.Create must not receive contaminated Owner/Pet (owner=%v pet=%v)", billing.OwnerID, billing.PetID)
			return nil
		},
	}
	resRepo := &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
			assert.Equal(t, foreignOwnerID, ownerID)
			return apperrors.WrapNotFound("owner", "201")
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, nil)
	deps.reservation = resRepo
	svc := deps.svc()

	result, err := svc.DischargeWithBilling(context.Background(), clinicID, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
	assert.False(t, accountingCreated)
	assert.False(t, updated, "TOCTOU-contaminated Owner/Pet discharge must not persist status update")
	assert.Equal(t, 2, findByIDCallCount, "outer + tx re-fetch FindByID must both run")
}

func TestHospitalizationService_DischargeWithBilling_WithoutAccounting_RejectsForeignOwnerPet(t *testing.T) {
	const clinicID = uint64(1)
	const foreignOwnerID, foreignPetID = uint64(201), uint64(202)

	updated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID: id, ClinicID: clinicID,
				OwnerID: foreignOwnerID, PetID: foreignPetID,
				Status: model.HospitalizationStatusAdmitted,
			}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
			updated = true
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			t.Fatal("care plan items must not be fetched when CreateAccounting is false")
			return nil, nil
		},
	}
	resRepo := &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
			assert.Equal(t, foreignOwnerID, ownerID)
			return apperrors.WrapNotFound("owner", "201")
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, nil, nil)
	deps.reservation = resRepo
	svc := deps.svc()

	result, err := svc.DischargeWithBilling(context.Background(), clinicID, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: false,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
	assert.False(t, updated, "contaminated Owner/Pet discharge must not persist status update")
}

func TestHospitalizationService_DischargeWithBilling_RejectsInvalidOwnerPetLinks(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOwnerID = uint64(10)

	tests := []struct {
		name                string
		ownerID             uint64
		petID               uint64
		createAccounting    bool
		assertOwnerFn       func(context.Context, uint64, uint64) error
		findPetFn           func(context.Context, uint64, uint64) (uint64, error)
		expectCarePlanFatal bool
	}{
		{
			name:             "rejects_foreign_pet_with_accounting",
			ownerID:          ownedOwnerID,
			petID:            202,
			createAccounting: true,
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, uint64(202), petID)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name:             "rejects_foreign_pet_without_accounting",
			ownerID:          ownedOwnerID,
			petID:            202,
			createAccounting: false,
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
			findPetFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, uint64(202), petID)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
			expectCarePlanFatal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID: id, ClinicID: clinicID,
						OwnerID: tt.ownerID, PetID: tt.petID,
						Status: model.HospitalizationStatusAdmitted,
					}, nil
				},
				updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
					updated = true
					t.Fatal("hospitalization must not be updated with invalid Owner/Pet links")
					return nil, nil
				},
			}

			var carePlanRepo *mockCarePlanItemRepository
			if tt.expectCarePlanFatal {
				carePlanRepo = &mockCarePlanItemRepository{
					listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
						t.Fatal("care plan items must not be fetched when CreateAccounting is false")
						return nil, nil
					},
				}
			} else {
				carePlanRepo = &mockCarePlanItemRepository{
					listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
						return nil, nil
					},
				}
			}

			var accountingRepo accountingCreator
			accountingCreated := false
			if tt.createAccounting {
				accountingRepo = &mockAccountingRepository{
					createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
						accountingCreated = true
						t.Fatalf("Accounting.Create must not receive invalid Owner/Pet (owner=%v pet=%v)", billing.OwnerID, billing.PetID)
						return nil
					},
				}
			}

			resRepo := &mockReservationRepository{
				assertOwnerInClinicFn:  tt.assertOwnerFn,
				findPetOwnerInClinicFn: tt.findPetFn,
			}
			deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, nil)
			deps.reservation = resRepo
			svc := deps.svc()

			result, err := svc.DischargeWithBilling(context.Background(), clinicID, 10, DischargeWithBillingInput{
				DischargeDate:    time.Now(),
				CreateAccounting: tt.createAccounting,
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, result)
			assert.False(t, updated, "contaminated Owner/Pet discharge must not persist status update")
			if tt.createAccounting {
				assert.False(t, accountingCreated)
			}
		})
	}
}

func uint64PtrHosp(v uint64) *uint64 { return &v }
