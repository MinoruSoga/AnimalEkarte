package lstep

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func reviewAliveFixture(t *testing.T) (*gorm.DB, *model.Owner, *model.Pet) {
	t.Helper()
	db := setupDeliveryOwnerLockDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.CheckupType{}, &model.Checkup{}, &model.StaffClinicAssignment{}))
	testdb.Truncate(t, db, "checkups", "checkup_types", "medical_records", "pets", "owners", "animal_species")
	uid := "U_synthetic_review657"
	owner := &model.Owner{ClinicID: 1, Name: "synthetic review owner", LineUserID: &uid}
	require.NoError(t, db.Create(owner).Error)
	species := &model.AnimalSpecies{Name: "synthetic species"}
	require.NoError(t, db.Create(species).Error)
	pet := &model.Pet{ClinicID: 1, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "synthetic pet", Status: model.PetStatusAlive}
	require.NoError(t, db.Create(pet).Error)
	return db, owner, pet
}

func TestCheckupCreate_NilPetIDRealDeliveryEligibility(t *testing.T) {
	for _, tt := range []struct {
		name     string
		status   model.PetStatus
		wantTags int
	}{
		{"alive", model.PetStatusAlive, 1}, {"deceased", model.PetStatusDeceased, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, owner, pet := reviewAliveFixture(t)
			require.NoError(t, db.Model(pet).Update("status", tt.status).Error)
			record := &model.MedicalRecord{ClinicID: 1, RecordNo: "REVIEW657-CHECKUP", OwnerID: &owner.ID, PetID: &pet.ID, Date: time.Now(), Status: model.MedicalRecordStatusDraft}
			require.NoError(t, db.Create(record).Error)
			checkupType := &model.CheckupType{ClinicID: 1, Name: "synthetic checkup"}
			require.NoError(t, db.Create(checkupType).Error)
			added := 0
			client := &mockLstepClientForDelivery{addTagFn: func(_ context.Context, _, tag string) error {
				assert.Equal(t, model.TriggerTypeCheckupFollowUp, tag)
				added++
				return nil
			}}
			trigger := buildDeliverySvc(ownerdomain.NewRepository(db, nil), nil, nil, petdomain.NewRepository(db), &mockTagCacheRepoForDelivery{}, &mockDeliveryTriggerLogRepository{}, enabledSettings())
			injectTestClient(trigger, client)
			records := medicalrecord.NewMedicalRecordRepository(db)
			svc := medicalrecord.NewCheckupService(medicalrecord.NewCheckupRepository(db), records, medicalrecord.NewCheckupTypeRepository(db), trigger, nil, medicalrecord.CheckupWriteDependencies{RelationVerifier: reservation.NewReservationRepository(db), Transactor: persistence.NewTransactor(db)})
			got, err := svc.Create(context.Background(), record.ID, &medicalrecord.CreateCheckupInput{ClinicID: 1, CheckupTypeID: checkupType.ID, Date: time.Now()})
			require.NoError(t, err)
			require.Nil(t, got.PetID, "normal UI leaves the persisted optional checkup pet unset")
			svc.Wait()
			assert.Equal(t, tt.wantTags, added)
		})
	}
}

func TestDormantBatch_MixedPetHistoryKeepsCPMExclusive(t *testing.T) {
	for _, sequence := range []string{"CPM_then_batch", "batch_then_CPM"} {
		t.Run(sequence, func(t *testing.T) {
			db, owner, alive := reviewAliveFixture(t)
			now := time.Now()
			dead := &model.Pet{ClinicID: 1, OwnerID: owner.ID, AnimalSpeciesID: alive.AnimalSpeciesID, Name: "synthetic deceased", Status: model.PetStatusDeceased, DeceasedAt: &now}
			require.NoError(t, db.Create(dead).Error)
			for _, item := range []struct {
				petID  uint64
				days   int
				number string
			}{{alive.ID, 250, "REVIEW657-OLD"}, {dead.ID, 5, "REVIEW657-RECENT"}} {
				pid := item.petID
				require.NoError(t, db.Create(&model.MedicalRecord{ClinicID: 1, RecordNo: item.number, OwnerID: &owner.ID, PetID: &pid, Date: now.AddDate(0, 0, -item.days)}).Error)
			}
			records := medicalrecord.NewMedicalRecordRepository(db)
			// Notification candidates retain their alive-only boundary.
			candidates, err := records.FindDormantOwnerEntriesCursorAt(context.Background(), 1, 180, 0, 100, now)
			require.NoError(t, err)
			require.Len(t, candidates, 1)
			assert.Equal(t, owner.ID, candidates[0].OwnerID)
			tags := map[string]bool{}
			client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tag string) error { tags[tag] = true; return nil }, removeTagFn: func(_ context.Context, _, tag string) error { delete(tags, tag); return nil }}
			settings := newCPMSettingsService("v1", nil)
			account := newCPMAccountingRepository()
			account.sumPaidByOwnerFn = func(context.Context, uint64, uint64) (int64, error) { return 60000, nil }
			account.maxSingleVisitAmountByOwnerFn = func(context.Context, uint64, uint64) (int64, error) { return 30000, nil }
			tagSync := &lstepTagSyncService{settingsSvc: settings, ownerRepo: ownerdomain.NewRepository(db, nil), medRecordRepo: records, accountRepo: account, tagCacheRepo: NewLstepTagCacheRepository(db), buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil }}
			batch := newDormantBatchService(records, tagSync, nil, nil, settings)
			if sequence == "CPM_then_batch" {
				require.NoError(t, tagSync.SyncCPMStageTag(context.Background(), 1, owner.ID))
			}
			count, errs := batch.detectDormantOwnersAt(context.Background(), 1, now)
			require.Empty(t, errs)
			assert.Zero(t, count, "recent retained history makes this owner non-dormant")
			if sequence == "batch_then_CPM" {
				require.NoError(t, tagSync.SyncCPMStageTag(context.Background(), 1, owner.ID))
			}
			assert.False(t, tags["cpm_dormant"])
			activeCPM := 0
			for _, stage := range allCPMStages {
				if tags[string(stage)] {
					activeCPM++
				}
			}
			assert.Equal(t, 1, activeCPM, "CPM stages remain mutually exclusive after both sync orders")
		})
	}
}
