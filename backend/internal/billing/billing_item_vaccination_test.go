package billing

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const billingItemVaccinationProvenanceMigrationSQL = `ALTER TABLE billing_items
    ADD COLUMN vaccination_id bigint,
    ADD COLUMN clinic_id bigint,
    ADD CONSTRAINT chk_billing_items_vaccination_clinic_pair
        CHECK (
            (vaccination_id IS NULL AND clinic_id IS NULL)
            OR (vaccination_id IS NOT NULL AND clinic_id IS NOT NULL)
        );

ALTER TABLE vaccinations
    ADD CONSTRAINT uq_vaccinations_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE billings
    ADD CONSTRAINT uq_billings_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE billing_items
    ADD CONSTRAINT fk_billing_items_billing_clinic
        FOREIGN KEY (billing_id, clinic_id)
        REFERENCES billings (id, clinic_id),
    ADD CONSTRAINT fk_billing_items_vaccination_clinic
        FOREIGN KEY (vaccination_id, clinic_id)
        REFERENCES vaccinations (id, clinic_id)
        ON DELETE RESTRICT;

CREATE INDEX idx_vaccinations_clinic_pet_date_active
    ON vaccinations(clinic_id, pet_id, date, id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_billing_items_vaccination_lifetime
    ON billing_items(vaccination_id)
    WHERE vaccination_id IS NOT NULL;

COMMENT ON COLUMN billing_items.vaccination_id IS
    '予防接種イベント由来の会計明細を識別するprovenance';

COMMENT ON COLUMN billing_items.clinic_id IS
    'vaccination または exam provenance がある明細だけに保持する内部tenant scope';
`

type unbilledVaccinationFinder interface {
	FindUnbilledVaccinationItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, int, error)
}

type vaccinationClaimAttemptRepository struct {
	BillingItemRepository
	attempts chan struct{}
}

func (r *vaccinationClaimAttemptRepository) ValidateVaccinationCreateReference(
	ctx context.Context,
	clinicID, billingID uint64,
	vaccinationID uint64,
) (*vaccinationBillingValues, error) {
	select {
	case r.attempts <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return r.BillingItemRepository.
		ValidateVaccinationCreateReference(ctx, clinicID, billingID, vaccinationID)
}

func makeBillingVaccination(
	t *testing.T,
	f billingItemReferenceFixture,
	name string,
	price *int64,
) (*model.Vaccine, *model.Vaccination) {
	t.Helper()
	require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.Vaccine{}, &model.Vaccination{}, &model.BillingItem{}))
	vaccine := &model.Vaccine{
		ClinicID: f.clinicID,
		Name:     name,
		Price:    price,
		IsActive: true,
	}
	require.NoError(t, f.db.Create(vaccine).Error)
	ensureConfirmedMedicalRecord(t, f)
	vaccination := &model.Vaccination{
		ClinicID:        f.clinicID,
		MedicalRecordID: &f.medicalRecord.ID,
		PetID:           &f.pet.ID,
		VaccineID:       vaccine.ID,
		Date:            time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, f.db.Create(vaccination).Error)
	return vaccine, vaccination
}

func ensureConfirmedMedicalRecord(t *testing.T, f billingItemReferenceFixture) {
	t.Helper()
	require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.BillingConfirmation{}))
	var existing model.BillingConfirmation
	err := f.db.Where("medical_record_id = ?", f.medicalRecord.ID).First(&existing).Error
	if err == nil {
		if existing.Status != model.ConfirmationStatusConfirmed {
			require.NoError(t, f.db.Model(&existing).Update("status", model.ConfirmationStatusConfirmed).Error)
		}
		return
	}
	require.NoError(t, f.db.Create(&model.BillingConfirmation{
		MedicalRecordID: f.medicalRecord.ID,
		Status:          model.ConfirmationStatusConfirmed,
	}).Error)
}

func TestBillingItemVaccinationProvenance_UnbilledCandidates(t *testing.T) {
	t.Run("derives authoritative vaccine values and preserves event provenance", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(5500)
		_, vaccination := makeBillingVaccination(t, f, "犬6種混合", &price)
		finder := f.repo.(unbilledVaccinationFinder)

		items, _, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, model.ItemCategoryVaccine, items[0].Category)
		assert.Equal(t, "犬6種混合", items[0].Name)
		assert.Equal(t, price, items[0].UnitPrice)
		assert.Equal(t, float64(1), items[0].Quantity)
		assert.Equal(t, model.ItemSourceMedicalRecord, items[0].Source)
		require.NotNil(t, items[0].VaccinationID)
		assert.Equal(t, vaccination.ID, *items[0].VaccinationID)
	})

	t.Run("linked event remains hidden after billing cancellation", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(5500)
		_, vaccination := makeBillingVaccination(t, f, "猫3種混合", &price)
		clinicID := f.clinicID
		linkedItem := &model.BillingItem{
			BillingID:     f.billing.ID,
			ClinicID:      &clinicID,
			Category:      model.ItemCategoryVaccine,
			Name:          "猫3種混合",
			UnitPrice:     price,
			Quantity:      1,
			Source:        model.ItemSourceMedicalRecord,
			VaccinationID: &vaccination.ID,
		}
		require.NoError(t, f.db.Create(linkedItem).Error)
		finder := f.repo.(unbilledVaccinationFinder)

		items, _, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)

		require.NoError(t, f.db.Model(&model.Billing{}).
			Where("id = ?", f.billing.ID).
			Update("status", model.BillingStatusCancelled).Error)
		items, _, err = finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("skips unpriced vaccine master as unbillable count (BUG-013)", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		vaccine, _ := makeBillingVaccination(t, f, "価格未設定", nil)
		finder := f.repo.(unbilledVaccinationFinder)

		items, unbillable, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 1, unbillable)

		require.NoError(t, f.db.Delete(&model.Vaccine{}, vaccine.ID).Error)
		items, unbillable, err = finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 1, unbillable)
	})

	t.Run("foreign-clinic vaccine master counts as unbillable skip (BUG-013)", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.Vaccine{}, &model.Vaccination{}, &model.BillingItem{}))
		foreignOwner := testdb.MakeTestOwner(t, f.db, 2, "foreign vaccine owner")
		_ = makeSpeciesAndPet(t, f.db, 2, foreignOwner.ID, "foreign vaccine pet")
		price := int64(5000)
		foreignVaccine := &model.Vaccine{ClinicID: 2, Name: "foreign vaccine", Price: &price, IsActive: true}
		require.NoError(t, f.db.Create(foreignVaccine).Error)
		ensureConfirmedMedicalRecord(t, f)
		corruptVaccination := &model.Vaccination{
			ClinicID:        f.clinicID,
			MedicalRecordID: &f.medicalRecord.ID,
			PetID:           &f.pet.ID,
			VaccineID:       foreignVaccine.ID,
			Date:            time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(corruptVaccination).Error)
		finder := f.repo.(unbilledVaccinationFinder)

		items, unbillable, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)

		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 1, unbillable)
	})

	t.Run("foreign or mismatched medical-record graph produces no candidate", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(5000)
		vaccine, seedVaccination := makeBillingVaccination(t, f, "graph vaccine", &price)
		require.NoError(t, f.db.Delete(&model.Vaccination{}, seedVaccination.ID).Error)
		foreignOwner := testdb.MakeTestOwner(t, f.db, 2, "foreign graph owner")
		foreignPet := makeSpeciesAndPet(t, f.db, 2, foreignOwner.ID, "foreign graph pet")
		foreignRecord := &model.MedicalRecord{
			ClinicID: 2,
			RecordNo: "foreign-graph-record",
			Date:     time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
			OwnerID:  &foreignOwner.ID,
			PetID:    &foreignPet.ID,
		}
		require.NoError(t, f.db.Create(foreignRecord).Error)
		corruptVaccination := &model.Vaccination{
			ClinicID:        f.clinicID,
			MedicalRecordID: &foreignRecord.ID,
			PetID:           &f.pet.ID,
			VaccineID:       vaccine.ID,
			Date:            time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(corruptVaccination).Error)
		finder := f.repo.(unbilledVaccinationFinder)

		items, _, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)

		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("chartless vaccination is excluded until attached to a confirmed chart", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(5500)
		require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.Vaccine{}, &model.Vaccination{}, &model.BillingItem{}))
		vaccine := &model.Vaccine{ClinicID: f.clinicID, Name: "カルテなし", Price: &price, IsActive: true}
		require.NoError(t, f.db.Create(vaccine).Error)
		vaccination := &model.Vaccination{
			ClinicID:  f.clinicID,
			PetID:     &f.pet.ID,
			VaccineID: vaccine.ID,
			Date:      time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(vaccination).Error)
		finder := f.repo.(unbilledVaccinationFinder)
		items, unbillable, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Zero(t, unbillable)
	})

	// DEC-27: after pet transfer, MR.owner_id (snapshot) may differ from
	// pets.owner_id (current). Unbilled candidates must still appear when
	// clinic + pet identity match.
	t.Run("unbilled candidates remain after pet owner transfer", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(5200)
		vaccine, vaccination := makeBillingVaccination(t, f, "transfer vaccine", &price)
		_ = vaccine
		require.NoError(t, f.db.Model(&model.Vaccination{}).
			Where("id = ?", vaccination.ID).
			Update("medical_record_id", f.medicalRecord.ID).Error)

		newOwner := testdb.MakeTestOwner(t, f.db, f.clinicID, "post-transfer owner")
		require.NoError(t, f.db.Model(&model.Pet{}).
			Where("id = ?", f.pet.ID).
			Update("owner_id", newOwner.ID).Error)

		finder := f.repo.(unbilledVaccinationFinder)
		items, _, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)

		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].VaccinationID)
		assert.Equal(t, vaccination.ID, *items[0].VaccinationID)
		assert.Equal(t, "transfer vaccine", items[0].Name)
		assert.Equal(t, price, items[0].UnitPrice)
	})
}

func TestBillingItemVaccinationProvenance_CreatePreservesEditedCandidateValues(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(6600)
	_, vaccination := makeBillingVaccination(t, f, "狂犬病予防", &price)
	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.Category = string(model.ItemCategoryGoods)
	input.Name = "spoofed"
	input.UnitPrice = 1
	input.Quantity = 9
	input.Source = string(model.ItemSourceManual)
	input.VaccinationID = &vaccination.ID

	item, err := svc.CreateItem(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, model.ItemCategoryVaccine, item.Category)
	assert.Equal(t, "spoofed", item.Name)
	assert.Equal(t, int64(1), item.UnitPrice)
	assert.Equal(t, float64(9), item.Quantity)
	assert.Equal(t, model.ItemSourceMedicalRecord, item.Source)
	require.NotNil(t, item.VaccinationID)
	assert.Equal(t, vaccination.ID, *item.VaccinationID)

	var stored model.BillingItem
	require.NoError(t, f.db.First(&stored, item.ID).Error)
	require.NotNil(t, stored.VaccinationID)
	assert.Equal(t, vaccination.ID, *stored.VaccinationID)
}

func TestBillingItemVaccinationProvenance_SkipsServerAutoDiscount(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(6600)
	_, vaccination := makeBillingVaccination(t, f, "割引対象ワクチン", &price)
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.Billing, error) {
		return f.billing, nil
	}
	var campaignCalled atomic.Bool
	var ownerCalled atomic.Bool
	svc := NewBillingItemServiceWithCampaign(
		f.repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		&mockCampaignRepository{
			findApplicableForItemFn: func(
				_ context.Context,
				_ uint64,
				_ time.Time,
				_ model.ItemCategory,
				_ *uint64,
			) (*model.Campaign, error) {
				campaignCalled.Store(true)
				return nil, nil
			},
		},
		&mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				ownerCalled.Store(true)
				return &model.Owner{ID: f.owner.ID, DiscountRate: 10}, nil
			},
		},
	)
	input := billingItemReferenceCreateInput(f)
	input.Name = "spoofed"
	input.UnitPrice = 1
	input.Quantity = 9
	input.VaccinationID = &vaccination.ID

	item, err := svc.CreateItem(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, int64(0), item.DiscountAmount)
	assert.False(t, campaignCalled.Load())
	assert.False(t, ownerCalled.Load())
}

func TestBillingItemVaccinationProvenance_CreateRejectsSmuggledReferences(t *testing.T) {
	t.Run("nonexistent and cross-clinic event IDs are not found", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.Vaccine{}, &model.Vaccination{}, &model.BillingItem{}))
		svc := newBillingItemReferenceService(f, f.repo)

		input := billingItemReferenceCreateInput(f)
		missingID := uint64(9_999_999)
		input.VaccinationID = &missingID
		_, err := svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "missing event must be indistinguishable from foreign event: %v", err)

		foreignOwner := testdb.MakeTestOwner(t, f.db, 2, "foreign vaccination owner")
		foreignPet := makeSpeciesAndPet(t, f.db, 2, foreignOwner.ID, "foreign vaccination pet")
		foreignPrice := int64(3000)
		foreignVaccine := &model.Vaccine{ClinicID: 2, Name: "foreign vaccine", Price: &foreignPrice, IsActive: true}
		require.NoError(t, f.db.Create(foreignVaccine).Error)
		foreignVaccination := &model.Vaccination{
			ClinicID:  2,
			PetID:     &foreignPet.ID,
			VaccineID: foreignVaccine.ID,
			Date:      time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(foreignVaccination).Error)
		input.VaccinationID = &foreignVaccination.ID
		_, err = svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "foreign event must be hidden: %v", err)
	})

	t.Run("same-clinic other-pet event and mixed provenance are invalid", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(3000)
		vaccine, _ := makeBillingVaccination(t, f, "local vaccine", &price)
		otherPet := makeSpeciesAndPet(t, f.db, f.clinicID, f.owner.ID, "other vaccination pet")
		otherVaccination := &model.Vaccination{
			ClinicID:  f.clinicID,
			PetID:     &otherPet.ID,
			VaccineID: vaccine.ID,
			Date:      time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(otherVaccination).Error)
		svc := newBillingItemReferenceService(f, f.repo)
		input := billingItemReferenceCreateInput(f)
		input.VaccinationID = &otherVaccination.ID

		_, err := svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "same-clinic pet mismatch must be invalid: %v", err)

		input.VaccinationID = nil
		_, localVaccination := makeBillingVaccination(t, f, "second local vaccine", &price)
		input.VaccinationID = &localVaccination.ID
		input.AppointmentID = &f.appointment.ID
		_, err = svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "mixed provenance must be rejected: %v", err)
	})

	t.Run("nil authoritative price fails closed", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		_, vaccination := makeBillingVaccination(t, f, "unpriced vaccine", nil)
		svc := newBillingItemReferenceService(f, f.repo)
		input := billingItemReferenceCreateInput(f)
		input.VaccinationID = &vaccination.ID

		item, err := svc.CreateItem(context.Background(), input)

		require.Error(t, err)
		assert.Nil(t, item)
	})

	// DEC-27: pets.owner_id is current owner; billing.owner_id is a snapshot.
	// Claim create must succeed after pet transfer when pet_id + clinic match.
	t.Run("claim create still succeeds after pet owner transfer", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(3000)
		_, vaccination := makeBillingVaccination(t, f, "owner correlation vaccine", &price)
		otherOwner := testdb.MakeTestOwner(t, f.db, f.clinicID, "other local owner")
		require.NoError(t, f.db.Model(&model.Pet{}).
			Where("id = ?", f.pet.ID).
			Update("owner_id", otherOwner.ID).Error)
		svc := newBillingItemReferenceService(f, f.repo)
		input := billingItemReferenceCreateInput(f)
		input.VaccinationID = &vaccination.ID

		item, err := svc.CreateItem(context.Background(), input)

		require.NoError(t, err)
		require.NotNil(t, item)
		require.NotNil(t, item.VaccinationID)
		assert.Equal(t, vaccination.ID, *item.VaccinationID)
		assert.Equal(t, model.ItemCategoryVaccine, item.Category)
	})

	t.Run("vaccination medical record must match clinic and locked billing graph", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(3000)
		_, vaccination := makeBillingVaccination(t, f, "medical-record correlation vaccine", &price)
		svc := newBillingItemReferenceService(f, f.repo)
		input := billingItemReferenceCreateInput(f)
		input.VaccinationID = &vaccination.ID

		foreignOwner := testdb.MakeTestOwner(t, f.db, 2, "foreign record owner")
		foreignPet := makeSpeciesAndPet(t, f.db, 2, foreignOwner.ID, "foreign record pet")
		foreignRecord := &model.MedicalRecord{
			ClinicID: 2,
			RecordNo: "foreign-vaccination-record",
			Date:     time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
			OwnerID:  &foreignOwner.ID,
			PetID:    &foreignPet.ID,
		}
		require.NoError(t, f.db.Create(foreignRecord).Error)
		require.NoError(t, f.db.Model(&model.Vaccination{}).
			Where("id = ?", vaccination.ID).
			Update("medical_record_id", foreignRecord.ID).Error)
		item, err := svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.True(t, apperrors.IsNotFound(err))

		otherLocalRecord := &model.MedicalRecord{
			ClinicID: f.clinicID,
			RecordNo: "other-local-vaccination-record",
			Date:     time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
			OwnerID:  &f.owner.ID,
			PetID:    &f.pet.ID,
		}
		require.NoError(t, f.db.Create(otherLocalRecord).Error)
		require.NoError(t, f.db.Model(&model.Vaccination{}).
			Where("id = ?", vaccination.ID).
			Update("medical_record_id", otherLocalRecord.ID).Error)
		item, err = svc.CreateItem(context.Background(), input)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

func TestBillingItemVaccinationProvenance_PollutedClaimClinicIsolation(t *testing.T) {
	foreignClinicID := uint64(2)
	tests := []struct {
		name     string
		clinicID *uint64
	}{
		{name: "foreign clinic claim", clinicID: &foreignClinicID},
		{name: "missing clinic claim", clinicID: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			price := int64(4400)
			_, vaccination := makeBillingVaccination(t, f, "汚染行隔離ワクチン", &price)
			polluted := &model.BillingItem{
				BillingID:     f.billing.ID,
				ClinicID:      tt.clinicID,
				Category:      model.ItemCategoryVaccine,
				Name:          "汚染行",
				UnitPrice:     price,
				Quantity:      1,
				Source:        model.ItemSourceMedicalRecord,
				VaccinationID: &vaccination.ID,
			}
			require.NoError(t, f.db.Create(polluted).Error)

			finder := f.repo.(unbilledVaccinationFinder)
			items, _, err := finder.FindUnbilledVaccinationItemsByPetID(
				context.Background(),
				f.clinicID,
				f.pet.ID,
			)
			require.NoError(t, err)
			assert.Len(t, items, 1, "foreign or unscoped claim must not suppress this clinic's candidate")

			svc := newBillingItemReferenceService(f, f.repo)
			input := billingItemReferenceCreateInput(f)
			input.VaccinationID = &vaccination.ID
			created, err := svc.CreateItem(context.Background(), input)
			require.NoError(t, err, "foreign or unscoped claim must not block this clinic's claim")
			require.NotNil(t, created.ClinicID)
			assert.Equal(t, f.clinicID, *created.ClinicID)
		})
	}
}

func TestBillingItemVaccinationProvenance_CreateStatusGuard(t *testing.T) {
	tests := []struct {
		name         string
		status       model.BillingStatus
		wantConflict bool
	}{
		{name: "waiting billing accepts claim", status: model.BillingStatusWaiting},
		{name: "pending billing accepts claim", status: model.BillingStatusPending},
		{name: "completed billing rejects claim", status: model.BillingStatusCompleted, wantConflict: true},
		{name: "cancelled billing rejects claim", status: model.BillingStatusCancelled, wantConflict: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			price := int64(4400)
			_, vaccination := makeBillingVaccination(t, f, "会計状態作成ガードワクチン", &price)
			require.NoError(t, f.db.Model(&model.Billing{}).
				Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
				Update("status", tt.status).Error)
			f.billing.Status = tt.status

			finder := f.repo.(unbilledVaccinationFinder)
			candidates, _, err := finder.FindUnbilledVaccinationItemsByPetID(
				context.Background(),
				f.clinicID,
				f.pet.ID,
			)
			require.NoError(t, err)
			require.Len(t, candidates, 1)

			svc := newBillingItemReferenceService(f, f.repo)
			input := billingItemReferenceCreateInput(f)
			input.VaccinationID = &vaccination.ID
			created, err := svc.CreateItem(context.Background(), input)

			var persistedCount int64
			require.NoError(t, f.db.Model(&model.BillingItem{}).
				Where("billing_id = ? AND vaccination_id = ?", f.billing.ID, vaccination.ID).
				Count(&persistedCount).Error)
			candidates, _, candidateErr := finder.FindUnbilledVaccinationItemsByPetID(
				context.Background(),
				f.clinicID,
				f.pet.ID,
			)
			require.NoError(t, candidateErr)

			if tt.wantConflict {
				assert.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "finalized billing create must return conflict: %v", err)
				assert.Nil(t, created)
				assert.Zero(t, persistedCount)
				require.Len(t, candidates, 1, "rejected claim must remain available")
				require.NotNil(t, candidates[0].VaccinationID)
				assert.Equal(t, vaccination.ID, *candidates[0].VaccinationID)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, int64(1), persistedCount)
			assert.Empty(t, candidates)
		})
	}
}

func TestBillingItemVaccinationProvenance_RequestAndResponseRoundTrip(t *testing.T) {
	vaccinationID := uint64(42)
	input := (&createBillingItemRequest{VaccinationID: &vaccinationID}).toServiceInput(7)
	require.NotNil(t, input.VaccinationID)
	assert.Equal(t, vaccinationID, *input.VaccinationID)

	response := ToBillingItemResponse(&model.BillingItem{VaccinationID: &vaccinationID})
	require.NotNil(t, response.VaccinationID)
	assert.Equal(t, vaccinationID, *response.VaccinationID)
}

func TestBillingItemVaccinationProvenance_ConcurrentClaim(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(4400)
	_, vaccination := makeBillingVaccination(t, f, "フェレットワクチン", &price)
	secondBilling := &model.Billing{
		ClinicID:      f.clinicID,
		OwnerID:       &f.owner.ID,
		PetID:         &f.pet.ID,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: f.billing.ScheduledDate,
	}
	require.NoError(t, f.db.Create(secondBilling).Error)

	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, clinicID, billingID uint64) (*model.Billing, error) {
		if clinicID != f.clinicID {
			return nil, apperrors.WrapNotFound("billing", "test")
		}
		switch billingID {
		case f.billing.ID:
			return f.billing, nil
		case secondBilling.ID:
			return secondBilling, nil
		default:
			return nil, apperrors.WrapNotFound("billing", "test")
		}
	}
	controllerTx := f.db.Begin()
	require.NoError(t, controllerTx.Error)
	controllerCommitted := false
	defer func() {
		if !controllerCommitted {
			_ = controllerTx.Rollback().Error
		}
	}()
	var lockedVaccinationID uint64
	require.NoError(t, controllerTx.Raw(
		"SELECT id FROM vaccinations WHERE id = ? FOR UPDATE",
		vaccination.ID,
	).Scan(&lockedVaccinationID).Error)
	require.Equal(t, vaccination.ID, lockedVaccinationID)

	attempts := make(chan struct{}, 2)
	repo := &vaccinationClaimAttemptRepository{
		BillingItemRepository: f.repo,
		attempts:              attempts,
	}
	svc := NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		testNewTransactor(f.db),
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
	)
	inputs := []*CreateBillingItemInput{
		billingItemReferenceCreateInput(f),
		billingItemReferenceCreateInput(f),
	}
	inputs[0].VaccinationID = &vaccination.ID
	inputs[1].BillingID = secondBilling.ID
	inputs[1].VaccinationID = &vaccination.ID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, in := range inputs {
		input := in
		go func() {
			<-start
			_, err := svc.CreateItem(ctx, input)
			results <- err
		}()
	}
	close(start)

	for range inputs {
		select {
		case <-attempts:
		case <-ctx.Done():
			t.Fatal("concurrent vaccination claim did not reach the locked-event attempt")
		}
	}
	require.NoError(t, controllerTx.Commit().Error)
	controllerCommitted = true

	var successCount, conflictCount int
	for range inputs {
		var err error
		select {
		case err = <-results:
		case <-ctx.Done():
			t.Fatal("concurrent vaccination claim did not complete before deadline")
		}
		switch {
		case err == nil:
			successCount++
		case apperrors.IsConflict(err):
			conflictCount++
		default:
			t.Fatalf("unexpected concurrent claim result: %v", err)
		}
	}
	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, conflictCount)
}

func TestBillingItemVaccinationProvenance_DeleteReleasesClaim(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(4400)
	_, vaccination := makeBillingVaccination(t, f, "再取込ワクチン", &price)
	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.VaccinationID = &vaccination.ID

	created, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, created.ClinicID)
	assert.Equal(t, f.clinicID, *created.ClinicID)

	finder := f.repo.(unbilledVaccinationFinder)
	items, _, err := finder.FindUnbilledVaccinationItemsByPetID(
		context.Background(),
		f.clinicID,
		f.pet.ID,
	)
	require.NoError(t, err)
	assert.Empty(t, items)

	staffID := uint64(42)
	require.NoError(t, svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{StaffID: &staffID}))

	var released struct {
		VaccinationID *uint64
		ClinicID      *uint64
	}
	require.NoError(t, f.db.Unscoped().
		Table("billing_items").
		Select("vaccination_id", "clinic_id").
		Where("id = ?", created.ID).
		Take(&released).Error)
	assert.Nil(t, released.VaccinationID)
	assert.Nil(t, released.ClinicID)

	items, _, err = finder.FindUnbilledVaccinationItemsByPetID(
		context.Background(),
		f.clinicID,
		f.pet.ID,
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].VaccinationID)
	assert.Equal(t, vaccination.ID, *items[0].VaccinationID)

	reclaimed, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, reclaimed.VaccinationID)
	assert.Equal(t, vaccination.ID, *reclaimed.VaccinationID)
}

func TestBillingItemVaccinationProvenance_DeleteStatusGuard(t *testing.T) {
	tests := []struct {
		name         string
		status       model.BillingStatus
		wantConflict bool
	}{
		{name: "pending billing releases claim", status: model.BillingStatusPending},
		{name: "completed billing preserves claim", status: model.BillingStatusCompleted, wantConflict: true},
		{name: "cancelled billing preserves claim", status: model.BillingStatusCancelled, wantConflict: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupBillingItemReferenceFixture(t)
			price := int64(4400)
			_, vaccination := makeBillingVaccination(t, f, "会計状態ガードワクチン", &price)
			svc := newBillingItemReferenceService(f, f.repo)
			input := billingItemReferenceCreateInput(f)
			input.VaccinationID = &vaccination.ID

			created, err := svc.CreateItem(context.Background(), input)
			require.NoError(t, err)
			require.NoError(t, f.db.Model(&model.Billing{}).
				Where("id = ? AND clinic_id = ?", f.billing.ID, f.clinicID).
				Update("status", tt.status).Error)
			f.billing.Status = tt.status

			err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{})
			if tt.wantConflict {
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "finalized billing delete must return conflict: %v", err)
			} else {
				require.NoError(t, err)
			}

			var stored struct {
				VaccinationID *uint64
				ClinicID      *uint64
				DeletedAt     *time.Time
			}
			require.NoError(t, f.db.Unscoped().
				Table("billing_items").
				Select("vaccination_id", "clinic_id", "deleted_at").
				Where("id = ?", created.ID).
				Take(&stored).Error)

			finder := f.repo.(unbilledVaccinationFinder)
			items, _, err := finder.FindUnbilledVaccinationItemsByPetID(
				context.Background(),
				f.clinicID,
				f.pet.ID,
			)
			require.NoError(t, err)
			if tt.wantConflict {
				require.NotNil(t, stored.VaccinationID)
				assert.Equal(t, vaccination.ID, *stored.VaccinationID)
				require.NotNil(t, stored.ClinicID)
				assert.Equal(t, f.clinicID, *stored.ClinicID)
				assert.Nil(t, stored.DeletedAt)
				assert.Empty(t, items, "finalized billing must keep the vaccination candidate claimed")
				return
			}

			assert.Nil(t, stored.VaccinationID)
			assert.Nil(t, stored.ClinicID)
			assert.NotNil(t, stored.DeletedAt)
			require.Len(t, items, 1)
			require.NotNil(t, items[0].VaccinationID)
			assert.Equal(t, vaccination.ID, *items[0].VaccinationID)
		})
	}
}

// BUG-440 / DEC-28 A: vaccination claim 解放は audit_logs へ同 tx fail-closed で actor 監査する。
func TestBillingItemVaccinationClaimRelease_DeleteWritesAudit(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(4400)
	_, vaccination := makeBillingVaccination(t, f, "監査対象ワクチン", &price)
	audit := &mockAuditService{}
	svc := newBillingItemReferenceService(f, f.repo, WithBillingItemAuditTx(audit))
	input := billingItemReferenceCreateInput(f)
	input.VaccinationID = &vaccination.ID

	created, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)

	staffID := uint64(11)
	require.NoError(t, svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{StaffID: &staffID}))

	require.True(t, audit.logEntryTxCalled, "vaccination claim release must write audit in same tx")
	require.NotNil(t, audit.logEntryTxInput)
	assert.Equal(t, model.AuditActionBillingVaccinationClaimRelease, audit.logEntryTxInput.Action)
	assert.Equal(t, "billing_item", audit.logEntryTxInput.Resource)
	require.NotNil(t, audit.logEntryTxInput.ResourceID)
	assert.Equal(t, created.ID, *audit.logEntryTxInput.ResourceID)
	require.NotNil(t, audit.logEntryTxInput.ClinicID)
	assert.Equal(t, f.clinicID, *audit.logEntryTxInput.ClinicID)
	require.NotNil(t, audit.logEntryTxInput.ActorID)
	assert.Equal(t, staffID, *audit.logEntryTxInput.ActorID)
	assert.Equal(t, sharedkernel.AuditActorTypeFor(&staffID), audit.logEntryTxInput.ActorType)

	meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, f.billing.ID, meta["billing_id"])
	assert.Equal(t, created.ID, meta["item_id"])
	assert.Equal(t, vaccination.ID, meta["vaccination_id"])
	assert.Equal(t, "billing_item_delete", meta["reason"])
}

func TestBillingItemVaccinationClaimRelease_AuditFailureRollsBack(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(4400)
	_, vaccination := makeBillingVaccination(t, f, "監査失敗ロールバックワクチン", &price)
	baseSvc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.VaccinationID = &vaccination.ID

	created, err := baseSvc.CreateItem(context.Background(), input)
	require.NoError(t, err)

	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	svc := newBillingItemReferenceService(f, f.repo, WithBillingItemAuditTx(audit))
	staffID := uint64(12)
	err = svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{StaffID: &staffID})
	require.Error(t, err)
	assert.True(t, audit.logEntryTxCalled, "audit must be attempted before commit")

	var stored struct {
		VaccinationID *uint64
		ClinicID      *uint64
		DeletedAt     *time.Time
	}
	require.NoError(t, f.db.Unscoped().
		Table("billing_items").
		Select("vaccination_id", "clinic_id", "deleted_at").
		Where("id = ?", created.ID).
		Take(&stored).Error)
	require.NotNil(t, stored.VaccinationID, "audit failure must keep vaccination claim")
	assert.Equal(t, vaccination.ID, *stored.VaccinationID)
	require.NotNil(t, stored.ClinicID)
	assert.Equal(t, f.clinicID, *stored.ClinicID)
	assert.Nil(t, stored.DeletedAt, "audit failure must roll back soft-delete")

	finder := f.repo.(unbilledVaccinationFinder)
	items, _, err := finder.FindUnbilledVaccinationItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
	require.NoError(t, err)
	assert.Empty(t, items, "failed delete must not re-open vaccination candidacy")
}

func TestBillingItemDelete_NonVaccinationItemSkipsClaimReleaseAudit(t *testing.T) {
	// Policy (BUG-440 ledger): audit only when vaccination_id is present.
	// Non-vaccination soft-delete does not emit vaccination claim release audit.
	f := setupBillingItemReferenceFixture(t)
	audit := &mockAuditService{}
	svc := newBillingItemReferenceService(f, f.repo, WithBillingItemAuditTx(audit))

	created, err := svc.CreateItem(context.Background(), billingItemReferenceCreateInput(f))
	require.NoError(t, err)
	require.Nil(t, created.VaccinationID)

	staffID := uint64(13)
	require.NoError(t, svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{StaffID: &staffID}))
	assert.False(t, audit.logEntryTxCalled, "non-vaccination delete must not emit claim-release audit")

	var stored struct {
		DeletedAt *time.Time
	}
	require.NoError(t, f.db.Unscoped().
		Table("billing_items").
		Select("deleted_at").
		Where("id = ?", created.ID).
		Take(&stored).Error)
	assert.NotNil(t, stored.DeletedAt, "non-vaccination delete still soft-deletes")
}

func TestBillingItemVaccinationProvenance_MigrationContract(t *testing.T) {
	sql := strings.Join(strings.Fields(strings.ToUpper(billingItemVaccinationProvenanceMigrationSQL)), " ")

	assert.Contains(t, sql, "ADD COLUMN VACCINATION_ID BIGINT")
	assert.Contains(t, sql, "ADD COLUMN CLINIC_ID BIGINT")
	assert.Contains(t, sql, "ADD CONSTRAINT CHK_BILLING_ITEMS_VACCINATION_CLINIC_PAIR CHECK")
	assert.Contains(t, sql, "(VACCINATION_ID IS NULL AND CLINIC_ID IS NULL)")
	assert.Contains(t, sql, "(VACCINATION_ID IS NOT NULL AND CLINIC_ID IS NOT NULL)")
	assert.Contains(t, sql, "ADD CONSTRAINT UQ_VACCINATIONS_ID_CLINIC UNIQUE (ID, CLINIC_ID)")
	assert.Contains(t, sql, "ADD CONSTRAINT UQ_BILLINGS_ID_CLINIC UNIQUE (ID, CLINIC_ID)")
	assert.Contains(t, sql, "FOREIGN KEY (BILLING_ID, CLINIC_ID) REFERENCES BILLINGS (ID, CLINIC_ID)")
	assert.Contains(t, sql, "FOREIGN KEY (VACCINATION_ID, CLINIC_ID) REFERENCES VACCINATIONS (ID, CLINIC_ID) ON DELETE RESTRICT")
	assert.Contains(t, sql, "CREATE UNIQUE INDEX UQ_BILLING_ITEMS_VACCINATION_LIFETIME ON BILLING_ITEMS(VACCINATION_ID) WHERE VACCINATION_ID IS NOT NULL")
	assert.Contains(t, sql, "CREATE INDEX IDX_VACCINATIONS_CLINIC_PET_DATE_ACTIVE ON VACCINATIONS(CLINIC_ID, PET_ID, DATE, ID) WHERE DELETED_AT IS NULL")
	assert.NotContains(t, sql, "FUNCTION")
	assert.NotContains(t, sql, "TRIGGER")
	assert.NotContains(t, sql, "CASCADE")
}

func TestBillingItemVaccinationProvenance_MatchesArchivedInitialMigration(t *testing.T) {
	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	initial := string(raw)

	const sourceMarker = "-- Source file: 008_add_billing_item_vaccination_provenance.sql"
	const nextSourceMarker = "-- Source file: 009_add_billing_items_other_reason.sql"
	start := strings.Index(initial, sourceMarker)
	require.GreaterOrEqual(t, start, 0, "001_init.sql must contain the archived billing vaccination migration")

	endOffset := strings.Index(initial[start:], "\n"+nextSourceMarker)
	require.Greater(t, endOffset, 0, "archived billing vaccination migration must end at the 009 source marker")
	block := initial[start : start+endOffset]

	const sourceSHA = "251ebb5b09edce09104c0b4da938175a088d57962efa9566bc222d3c13bc251f"
	const shaHeader = "-- Source SHA-256: " + sourceSHA + "\n"
	shaOffset := strings.Index(block, shaHeader)
	require.GreaterOrEqual(t, shaOffset, 0, "archived billing vaccination migration must contain its exact SHA-256 header")

	body := block[shaOffset+len(shaHeader):]
	require.Equal(t, billingItemVaccinationProvenanceMigrationSQL, body)

	checksum := sha256.Sum256([]byte(body))
	require.Equal(t, sourceSHA, fmt.Sprintf("%x", checksum))
}
