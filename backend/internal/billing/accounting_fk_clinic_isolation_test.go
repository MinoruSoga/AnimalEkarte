package billing

// accounting_fk_clinic_isolation_test.go — AUD-002 Mode 3 write-path
// Create/Update の medical_record_id / hospitalization_id / owner_id / pet_id
// clinic 所有確認・相互整合・拒否時非副作用・completed Create の tx 内検証。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestAccountingService_Create_RejectsCrossClinicRelatedFKs(t *testing.T) {
	const clinicID = uint64(1)
	const ownedMRID = uint64(100)
	const foreignMRID = uint64(901)
	const ownedHospID = uint64(200)
	const foreignHospID = uint64(902)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const ownedPetID = uint64(20)
	const foreignPetID = uint64(998)
	const otherOwnerSameClinicPetID = uint64(21)
	const otherOwnerID = uint64(11)
	scheduled := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	linkAware := func(created *bool) AccountingService {
		repo := &mockAccountingRepository{
			createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
				*created = true
				return nil
			},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.MedicalRecord, error) {
				if gotClinicID != clinicID || id == foreignMRID {
					return nil, apperrors.WrapNotFound("medical_record", "foreign")
				}
				return &model.MedicalRecord{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(ownedPetID),
				}, nil
			},
		}
		hospRepo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Hospitalization, error) {
				if gotClinicID != clinicID || id == foreignHospID {
					return nil, apperrors.WrapNotFound("hospitalization", "foreign")
				}
				return &model.Hospitalization{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ownedOwnerID, PetID: ownedPetID,
				}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				if gotClinicID != clinicID {
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
				switch petID {
				case ownedPetID:
					return ownedOwnerID, nil
				case otherOwnerSameClinicPetID:
					return otherOwnerID, nil
				default:
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
			},
		}
		return NewAccountingService(repo, mrRepo, hospRepo, resRepo, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	}

	base := func() *CreateAccountingInput {
		return &CreateAccountingInput{
			ClinicID: clinicID, Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
			Status: model.BillingStatusWaiting, ScheduledDate: scheduled,
		}
	}

	t.Run("rejects cross-clinic medical_record_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.MedicalRecordID = ptrU64(foreignMRID)
		out, err := svc.Create(context.Background(), in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic hospitalization_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.HospitalizationID = ptrU64(foreignHospID)
		out, err := svc.Create(context.Background(), in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic owner_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.OwnerID = ptrU64(foreignOwnerID)
		out, err := svc.Create(context.Background(), in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic pet_id and does not persist", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.PetID = ptrU64(foreignPetID)
		out, err := svc.Create(context.Background(), in)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects billing owner mismatch with medical_record owner", func(t *testing.T) {
		created := false
		repo := &mockAccountingRepository{
			createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
				created = true
				return nil
			},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(ownedPetID),
				}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID != ownedOwnerID && ownerID != otherOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, petID uint64) (uint64, error) {
				if petID == ownedPetID {
					return ownedOwnerID, nil
				}
				if petID == otherOwnerSameClinicPetID {
					return otherOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "foreign")
			},
		}
		svc := NewAccountingService(repo, mrRepo, &mockHospitalizationRepository{}, resRepo, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
		in := base()
		in.MedicalRecordID = ptrU64(ownedMRID)
		in.OwnerID = ptrU64(otherOwnerID)
		in.PetID = ptrU64(otherOwnerSameClinicPetID)
		out, err := svc.Create(context.Background(), in)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic matching related FKs", func(t *testing.T) {
		created := false
		svc := linkAware(&created)
		in := base()
		in.MedicalRecordID = ptrU64(ownedMRID)
		in.HospitalizationID = ptrU64(ownedHospID)
		in.OwnerID = ptrU64(ownedOwnerID)
		in.PetID = ptrU64(ownedPetID)
		out, err := svc.Create(context.Background(), in)
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestAccountingService_Update_RejectsCrossClinicRelatedFKs(t *testing.T) {
	const clinicID = uint64(1)
	const billingID = uint64(7)
	const ownedMRID = uint64(100)
	const foreignMRID = uint64(901)
	const ownedHospID = uint64(200)
	const foreignHospID = uint64(902)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const ownedPetID = uint64(20)
	const foreignPetID = uint64(998)
	scheduled := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	current := &model.Billing{
		ID: billingID, ClinicID: clinicID,
		MedicalRecordID: ptrU64(ownedMRID), HospitalizationID: ptrU64(ownedHospID),
		OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(ownedPetID),
		Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
		Status: model.BillingStatusWaiting, ScheduledDate: scheduled,
	}

	linkAware := func(updated *bool, snapshot **model.Billing) AccountingService {
		cur := *current
		*snapshot = &cur
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Billing, error) {
				c := **snapshot
				c.ID, c.ClinicID = id, gotClinicID
				return &c, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Billing, error) {
				*updated = true
				c := **snapshot
				if v, ok := fields["medical_record_id"]; ok {
					id := v.(uint64)
					c.MedicalRecordID = &id
				}
				if v, ok := fields["hospitalization_id"]; ok {
					id := v.(uint64)
					c.HospitalizationID = &id
				}
				if v, ok := fields["owner_id"]; ok {
					id := v.(uint64)
					c.OwnerID = &id
				}
				if v, ok := fields["pet_id"]; ok {
					id := v.(uint64)
					c.PetID = &id
				}
				*snapshot = &c
				return &c, nil
			},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.MedicalRecord, error) {
				if gotClinicID != clinicID || id == foreignMRID {
					return nil, apperrors.WrapNotFound("medical_record", "foreign")
				}
				return &model.MedicalRecord{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(ownedPetID),
				}, nil
			},
		}
		hospRepo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Hospitalization, error) {
				if gotClinicID != clinicID || id == foreignHospID {
					return nil, apperrors.WrapNotFound("hospitalization", "foreign")
				}
				return &model.Hospitalization{
					ID: id, ClinicID: gotClinicID,
					OwnerID: ownedOwnerID, PetID: ownedPetID,
				}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				if gotClinicID != clinicID || petID != ownedPetID {
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
				return ownedOwnerID, nil
			},
		}
		return NewAccountingService(repo, mrRepo, hospRepo, resRepo, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	}

	assertUnchanged := func(t *testing.T, snap *model.Billing) {
		t.Helper()
		require.NotNil(t, snap)
		assert.Equal(t, ownedMRID, *snap.MedicalRecordID)
		assert.Equal(t, ownedHospID, *snap.HospitalizationID)
		assert.Equal(t, ownedOwnerID, *snap.OwnerID)
		assert.Equal(t, ownedPetID, *snap.PetID)
		assert.Equal(t, int64(1100), snap.TotalAmount)
		assert.Equal(t, model.BillingStatusWaiting, snap.Status)
	}

	t.Run("rejects cross-clinic medical_record_id and leaves billing unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Billing
		svc := linkAware(&updated, &snap)
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: billingID, ClinicID: clinicID, MedicalRecordID: ptrU64(foreignMRID),
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
		assertUnchanged(t, snap)
	})

	t.Run("rejects cross-clinic hospitalization_id and leaves billing unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Billing
		svc := linkAware(&updated, &snap)
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: billingID, ClinicID: clinicID, HospitalizationID: ptrU64(foreignHospID),
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
		assertUnchanged(t, snap)
	})

	t.Run("rejects cross-clinic owner_id and leaves billing unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Billing
		svc := linkAware(&updated, &snap)
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: billingID, ClinicID: clinicID, OwnerID: ptrU64(foreignOwnerID),
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
		assertUnchanged(t, snap)
	})

	t.Run("rejects cross-clinic pet_id and leaves billing unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Billing
		svc := linkAware(&updated, &snap)
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: billingID, ClinicID: clinicID, PetID: ptrU64(foreignPetID),
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
		assertUnchanged(t, snap)
	})

	t.Run("accepts same-clinic matching FK update", func(t *testing.T) {
		updated := false
		var snap *model.Billing
		svc := linkAware(&updated, &snap)
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: billingID, ClinicID: clinicID,
			OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(ownedPetID),
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestAccountingService_Create_CompletedValidatesInsideTx(t *testing.T) {
	const clinicID = uint64(1)
	const foreignOwnerID = uint64(999)
	scheduled := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	created := false
	txEntered := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			created = true
			return nil
		},
	}
	resRepo := &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
			if ownerID == foreignOwnerID {
				return apperrors.WrapNotFound("owner", "foreign")
			}
			return nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			txEntered = true
			return fn(ctx)
		},
	}
	svc := NewAccountingService(repo, &mockMedicalRecordRepository{}, &mockHospitalizationRepository{}, resRepo, nil, tx, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	out, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID: clinicID, OwnerID: ptrU64(foreignOwnerID),
		Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
		Status: model.BillingStatusCompleted, ScheduledDate: scheduled,
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, out)
	assert.False(t, created)
	assert.True(t, txEntered, "completed Create must enter WithTx so validation runs in same tx")
}
