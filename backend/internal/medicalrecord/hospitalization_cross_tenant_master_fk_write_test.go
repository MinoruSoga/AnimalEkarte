package medicalrecord

// hospitalization_cross_tenant_master_fk_write_test.go — BE9-2D ⑤: internal/service
// cross_tenant_master_fk_write_test.go の care_plan_item（MasterFK/HospitalizationPlanFK）+
// hospitalization（CageFK）節を同名のまま縦移動。builder は narrow-view 版
// （treatment_mocks_test.go / hospitalization_mocks_test.go）を使う。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// okHospitalizationPlanRepo / rejectHospitalizationPlanRepo は internal/service 側の同名 builder
// の移設（consumer が本 package の care_plan_item テストのみになったため）。
func okHospitalizationPlanRepo() HospitalizationPlanRepository {
	return &mockHospitalizationPlanRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.HospitalizationPlan, error) {
		return &model.HospitalizationPlan{ID: id}, nil
	}}
}

func rejectHospitalizationPlanRepo(ownedID uint64) HospitalizationPlanRepository {
	return &mockHospitalizationPlanRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.HospitalizationPlan, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("hospitalization_plan", "foreign")
		}
		return &model.HospitalizationPlan{ID: id}, nil
	}}
}

// rejectMedicineRepo は internal/service 側の同名 builder の narrow-view 版。
func rejectMedicineRepo(ownedID uint64) medicineFinder {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("medicine", "foreign")
		}
		return &model.Medicine{ID: id}, nil
	}}
}

// ── care_plan_item (HIGH): medicine_id / procedure_id ──

func TestCarePlanItemService_Create_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMedicineID = uint64(10)
	const foreignMedicineID = uint64(999)

	newSvc := func(created *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo(), okHospitalizationPlanRepo(), passthroughCarePlanTransactor{}, okCarePlanAuditTx{})
	}

	t.Run("rejects cross-clinic medicine_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignMedicineID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeMedicine), Name: "x", MedicineID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "care plan item must NOT be persisted referencing another clinic's medicine")
	})

	t.Run("accepts same-clinic medicine_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedMedicineID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeMedicine), Name: "x", MedicineID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCarePlanItemService_Update_RejectsCrossClinicMasterFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMedicineID = uint64(10)
	const foreignMedicineID = uint64(999)

	newSvc := func(updated *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateCarePlanItemInput) error { *updated = true; return nil },
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), rejectMedicineRepo(ownedMedicineID), okProcedureRepo(), okHospitalizationPlanRepo(), passthroughCarePlanTransactor{}, okCarePlanAuditTx{})
	}

	t.Run("rejects cross-clinic medicine_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignMedicineID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{MedicineID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "care plan item must NOT be updated to reference another clinic's medicine")
	})
}

// ── care_plan_item (X-14): hospitalization_plan_id ──

func TestCarePlanItemService_Create_RejectsCrossClinicHospitalizationPlanFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedPlanID = uint64(10)
	const foreignPlanID = uint64(999)

	newSvc := func(created *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			createFn: func(_ context.Context, _ *model.CarePlanItem) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID}, nil
			},
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), rejectHospitalizationPlanRepo(ownedPlanID), passthroughCarePlanTransactor{}, okCarePlanAuditTx{})
	}

	t.Run("rejects cross-clinic hospitalization_plan_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignPlanID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeFood), Name: "x", HospitalizationPlanID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "care plan item must NOT be persisted referencing another clinic's hospitalization plan")
	})

	t.Run("accepts same-clinic hospitalization_plan_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedPlanID
		out, err := svc.Create(context.Background(), clinicID, 1, &CreateCarePlanItemInput{
			Type: string(model.CarePlanTypeFood), Name: "x", HospitalizationPlanID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCarePlanItemService_Update_RejectsCrossClinicHospitalizationPlanFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedPlanID = uint64(10)
	const foreignPlanID = uint64(999)

	newSvc := func(updated *bool) CarePlanItemService {
		repo := &mockCarePlanItemRepository{
			findByIDFn: func(_ context.Context, _, itemID uint64) (*model.CarePlanItem, error) {
				return &model.CarePlanItem{ID: itemID, HospitalizationID: 1}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateCarePlanItemInput) error { *updated = true; return nil },
		}
		return NewCarePlanItemService(repo, okHospRepoForCarePlan(), okMedicineRepo(), okProcedureRepo(), rejectHospitalizationPlanRepo(ownedPlanID), passthroughCarePlanTransactor{}, okCarePlanAuditTx{})
	}

	t.Run("rejects cross-clinic hospitalization_plan_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignPlanID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{HospitalizationPlanID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "care plan item must NOT be updated to reference another clinic's hospitalization plan")
	})

	t.Run("accepts same-clinic hospitalization_plan_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedPlanID
		out, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateCarePlanItemInput{HospitalizationPlanID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── hospitalization (X-14): cage_id ──

func TestHospitalizationService_Create_RejectsCrossClinicCageFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCageID = uint64(10)
	const foreignCageID = uint64(999)

	newSvc := func(created *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, _ *model.Hospitalization) error { *created = true; return nil },
		}
		return NewHospitalizationService(repo, &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
			findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 2, nil
			},
		}, &mockPetRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id}, nil
			},
		}, rejectCageRepo(ownedCageID), nil, nil, nil, &mockTransactor{})
	}

	t.Run("rejects cross-clinic cage_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignCageID
		out, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID: 2, PetID: 5, CageID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "hospitalization must NOT be persisted referencing another clinic's cage")
	})

	t.Run("accepts same-clinic cage_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedCageID
		out, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			OwnerID: 2, PetID: 5, CageID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestHospitalizationService_Update_RejectsCrossClinicCageFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCageID = uint64(10)
	const foreignCageID = uint64(999)

	newSvc := func(updated *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				*updated = true
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		}
		return NewHospitalizationService(repo, nil, nil, rejectCageRepo(ownedCageID), nil, nil, nil, &mockTransactor{})
	}

	t.Run("rejects cross-clinic cage_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignCageID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "hospitalization must NOT be updated to reference another clinic's cage")
	})

	t.Run("accepts same-clinic cage_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedCageID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{CageID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestHospitalizationService_Create_ValidatesDoctorInWriteTransaction(t *testing.T) {
	type txContextKey struct{}
	const clinicID = uint64(1)
	const ownedDoctorID = uint64(10)
	const foreignDoctorID = uint64(999)

	newSvc := func(created *bool, doctorCheckedInTx *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			createFn: func(_ context.Context, _ *model.Hospitalization) error {
				*created = true
				return nil
			},
		}
		relations := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
			findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 2, nil
			},
			assertMedicalRecordDoctorInClinic: func(ctx context.Context, _, doctorID uint64) error {
				*doctorCheckedInTx = ctx.Value(txContextKey{}) == true
				if doctorID != ownedDoctorID {
					return apperrors.WrapNotFound("staff", "foreign")
				}
				return nil
			},
		}
		tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txContextKey{}, true))
		}}
		return NewHospitalizationService(repo, relations, &mockPetRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id}, nil
			},
		}, acceptAnyCageRepo(), nil, nil, nil, tx)
	}

	t.Run("rejects doctor outside clinic without persisting", func(t *testing.T) {
		created := false
		checkedInTx := false
		svc := newSvc(&created, &checkedInTx)
		doctorID := foreignDoctorID

		got, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			CageID:  func() *uint64 { v := uint64(10); return &v }(),
			OwnerID: 2, PetID: 5, DoctorID: &doctorID,
		})

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.False(t, created)
		assert.True(t, checkedInTx)
	})

	t.Run("accepts assigned active doctor", func(t *testing.T) {
		created := false
		checkedInTx := false
		svc := newSvc(&created, &checkedInTx)
		doctorID := ownedDoctorID

		got, err := svc.Create(context.Background(), clinicID, &CreateHospitalizationInput{
			CageID:  func() *uint64 { v := uint64(10); return &v }(),
			OwnerID: 2, PetID: 5, DoctorID: &doctorID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.True(t, created)
		assert.True(t, checkedInTx)
	})
}

func TestHospitalizationService_Update_ValidatesDoctorInWriteTransaction(t *testing.T) {
	type txContextKey struct{}
	const clinicID = uint64(1)
	const ownedDoctorID = uint64(10)
	const foreignDoctorID = uint64(999)

	newSvc := func(updated *bool, doctorCheckedInTx *bool) HospitalizationService {
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				*updated = true
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		}
		relations := &mockReservationRepository{
			assertMedicalRecordDoctorInClinic: func(ctx context.Context, _, doctorID uint64) error {
				*doctorCheckedInTx = ctx.Value(txContextKey{}) == true
				if doctorID != ownedDoctorID {
					return apperrors.WrapNotFound("staff", "foreign")
				}
				return nil
			},
		}
		tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txContextKey{}, true))
		}}
		return NewHospitalizationService(repo, relations, nil, nil, nil, nil, nil, tx)
	}

	t.Run("rejects doctor outside clinic without persisting", func(t *testing.T) {
		updated := false
		checkedInTx := false
		svc := newSvc(&updated, &checkedInTx)
		doctorID := foreignDoctorID

		got, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{DoctorID: &doctorID})

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.False(t, updated)
		assert.True(t, checkedInTx)
	})

	t.Run("accepts assigned active doctor", func(t *testing.T) {
		updated := false
		checkedInTx := false
		svc := newSvc(&updated, &checkedInTx)
		doctorID := ownedDoctorID

		got, err := svc.Update(context.Background(), clinicID, 1, &UpdateHospitalizationInput{DoctorID: &doctorID})

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.True(t, updated)
		assert.True(t, checkedInTx)
	})
}
