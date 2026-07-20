package medicalrecord

// Cross-tenant master FK write isolation tests (#201 follow-up) for the services migrated to
// internal/medicalrecord (BE9-2D), ported from internal/service/cross_tenant_master_fk_write_test.go.
// Each fix keeps its dual regression: (i) clinic A referencing clinic B's master is rejected and
// NOT persisted; (ii) same-clinic master still succeeds (no false-reject).
//
// okVaccineRepo/rejectVaccineRepo and okCheckupTypeRepo/rejectCheckupTypeRepo were *moved* here
// (their only consumers — the vaccination/checkup service tests and these sections — all migrated).
// okChiefComplaintTypeRepo/rejectChiefComplaintTypeRepo are *replicated*: the residual
// MedicalRecordService cross-tenant section still uses the internal/service copies, so both exist.
// The mock repositories (mockVaccineRepository, mockCheckupTypeRepository, mockChiefComplaintTypeRepository,
// mockMedicalRecordRepository) are the medicalrecord-package copies declared in the sibling test files.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ── shared permissive / rejecting master-repo builders (clinic-scoped FindByID semantics) ──

func okVaccineRepo() VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		return &model.Vaccine{ID: id}, nil
	}}
}

func rejectVaccineRepo(ownedID uint64) VaccineRepository {
	return &mockVaccineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("vaccine", "foreign")
		}
		return &model.Vaccine{ID: id}, nil
	}}
}

func okCheckupTypeRepo() CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		return &model.CheckupType{ID: id}, nil
	}}
}

func rejectCheckupTypeRepo(ownedID uint64) CheckupTypeRepository {
	return &mockCheckupTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("checkup_type", "foreign")
		}
		return &model.CheckupType{ID: id}, nil
	}}
}

func okChiefComplaintTypeRepo() ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

func rejectChiefComplaintTypeRepo(ownedID uint64) ChiefComplaintTypeRepository {
	return &mockChiefComplaintTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ChiefComplaintType, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("chief_complaint_type", "foreign")
		}
		return &model.ChiefComplaintType{ID: id}, nil
	}}
}

// ── vaccination (CRITICAL #125): vaccine_id ──

func TestVaccinationService_Create_RejectsCrossClinicVaccine(t *testing.T) {
	const clinicID = uint64(1)
	const ownedVaccineID = uint64(10)
	const foreignVaccineID = uint64(999)

	newSvc := func(created *bool) VaccinationService {
		repo := &mockVaccinationRepository{
			createFn: func(_ context.Context, _ *model.Vaccination) error { *created = true; return nil },
		}
		return NewVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
	}

	t.Run("rejects cross-clinic vaccine_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{VaccineID: foreignVaccineID})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "vaccination must NOT be persisted referencing another clinic's vaccine")
	})

	t.Run("accepts same-clinic vaccine_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{VaccineID: ownedVaccineID})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestVaccinationService_Update_RejectsCrossClinicVaccine(t *testing.T) {
	const clinicID = uint64(1)
	const ownedVaccineID = uint64(10)
	const foreignVaccineID = uint64(999)

	newSvc := func(updated *bool) VaccinationService {
		repo := &mockVaccinationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
				return &model.Vaccination{ID: id, ClinicID: clinicID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
				*updated = true
				return &model.Vaccination{ID: 1}, nil
			},
		}
		return NewVaccinationService(repo, rejectVaccineRepo(ownedVaccineID), nil)
	}

	t.Run("rejects cross-clinic vaccine_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignVaccineID
		out, err := svc.Update(context.Background(), clinicID, 1, &UpdateVaccinationInput{VaccineID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "vaccination must NOT be updated to reference another clinic's vaccine")
	})
}

// ── checkup (MEDIUM, clinical screening record): checkup_type_id ──

func TestCheckupService_Create_RejectsCrossClinicCheckupType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedCheckupTypeID = uint64(10)
	const foreignCheckupTypeID = uint64(999)

	newSvc := func(created *bool) CheckupService {
		repo := &mockCheckupRepository{
			createFn: func(_ context.Context, _ *model.Checkup) error { *created = true; return nil },
			findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
				return &model.Checkup{ID: checkupID}, nil
			},
		}
		medRec := &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		}}
		return NewCheckupService(repo, medRec, rejectCheckupTypeRepo(ownedCheckupTypeID), nil, nil)
	}

	t.Run("rejects cross-clinic checkup_type_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), 1, &CreateCheckupInput{ClinicID: clinicID, CheckupTypeID: foreignCheckupTypeID})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "checkup must NOT be persisted referencing another clinic's checkup_type")
	})

	t.Run("accepts same-clinic checkup_type_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), 1, &CreateCheckupInput{ClinicID: clinicID, CheckupTypeID: ownedCheckupTypeID})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

// ── checkup_type (self-ref parent_id, X-14 batch3) ──

func TestCheckupTypeService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) CheckupTypeService {
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("checkup_type", "foreign")
				}
				return &model.CheckupType{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.CheckupType) error { *created = true; return nil },
		}
		return NewCheckupTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateCheckupTypeInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "checkup type must NOT be persisted referencing another clinic's parent checkup type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateCheckupTypeInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCheckupTypeService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) CheckupTypeService {
		repo := &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
				if id == entityID || id == ownedParentID {
					return &model.CheckupType{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("checkup_type", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.CheckupType, error) {
				*updated = true
				return &model.CheckupType{ID: id}, nil
			},
		}
		return NewCheckupTypeService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateCheckupTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "checkup type must NOT be updated to reference another clinic's parent checkup type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateCheckupTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── vaccine (self-ref parent_id, X-14 batch3) ──

func TestVaccineService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) VaccineService {
		repo := &mockVaccineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("vaccine", "foreign")
				}
				return &model.Vaccine{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.Vaccine) error { *created = true; return nil },
		}
		return NewVaccineService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccineInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "vaccine must NOT be persisted referencing another clinic's parent vaccine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateVaccineInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestVaccineService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) VaccineService {
		repo := &mockVaccineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
				if id == entityID || id == ownedParentID {
					return &model.Vaccine{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("vaccine", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Vaccine, error) {
				*updated = true
				return &model.Vaccine{ID: id}, nil
			},
		}
		return NewVaccineService(repo)
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateVaccineInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "vaccine must NOT be updated to reference another clinic's parent vaccine")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateVaccineInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}

// ── inquiry (X-14 batch U4): chief_complaint_type_id ──

func TestInquiryService_Save_RejectsCrossClinicChiefComplaintType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedTypeID = uint64(10)
	const foreignTypeID = uint64(999)

	newSvc := func(saved *bool) InquiryService {
		repo := &mockInquiryRepository{
			upsertFn: func(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
				*saved = true
				return inquiry, nil
			},
		}
		return NewInquiryService(repo, rejectChiefComplaintTypeRepo(ownedTypeID))
	}

	t.Run("rejects cross-clinic chief_complaint_type_id and does not persist", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := foreignTypeID
		out, err := svc.Save(context.Background(), UpsertInquiryInput{
			ClinicID: clinicID, MedicalRecordID: 1, ChiefComplaintTypeID: &foreign,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, saved, "inquiry must NOT be persisted referencing another clinic's chief complaint type")
	})

	t.Run("accepts same-clinic chief_complaint_type_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := ownedTypeID
		out, err := svc.Save(context.Background(), UpsertInquiryInput{
			ClinicID: clinicID, MedicalRecordID: 1, ChiefComplaintTypeID: &owned,
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, saved)
	})
}
