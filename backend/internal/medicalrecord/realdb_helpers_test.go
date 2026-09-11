package medicalrecord

// realdb_*_isolation_test.go — D3 medicalrecord package (61 clinic-fixed + 2 cross-clinic)
//
// Proves selected-clinic grant isolation through real repository + service + HTTP
// handler paths. Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBMRPrefix = "D3mr-realdb"

	realDBChiefA        = realDBMRPrefix + "-chief-A"
	realDBChiefB        = realDBMRPrefix + "-chief-B"
	realDBConsultA      = realDBMRPrefix + "-consult-A"
	realDBConsultB      = realDBMRPrefix + "-consult-B"
	realDBDiagTypeA     = realDBMRPrefix + "-diagtype-A"
	realDBDiagTypeB     = realDBMRPrefix + "-diagtype-B"
	realDBDiagNameA     = realDBMRPrefix + "-diagname-A"
	realDBDiagNameB     = realDBMRPrefix + "-diagname-B"
	realDBExamTypeA     = realDBMRPrefix + "-examtype-A"
	realDBExamTypeB     = realDBMRPrefix + "-examtype-B"
	realDBInquiryA      = realDBMRPrefix + "-inquiry-A"
	realDBInquiryB      = realDBMRPrefix + "-inquiry-B"
	realDBMedicineA     = realDBMRPrefix + "-medicine-A"
	realDBMedicineB     = realDBMRPrefix + "-medicine-B"
	realDBProcedureA    = realDBMRPrefix + "-procedure-A"
	realDBProcedureB    = realDBMRPrefix + "-procedure-B"
	realDBVaccineA      = realDBMRPrefix + "-vaccine-A"
	realDBVaccineB      = realDBMRPrefix + "-vaccine-B"
	realDBCageA         = realDBMRPrefix + "-cage-A"
	realDBCageB         = realDBMRPrefix + "-cage-B"
	realDBHospPlanA     = realDBMRPrefix + "-hospplan-A"
	realDBHospPlanB     = realDBMRPrefix + "-hospplan-B"
	realDBCheckupTypeA  = realDBMRPrefix + "-checkuptype-A"
	realDBCheckupTypeB  = realDBMRPrefix + "-checkuptype-B"
	realDBCheckupFieldA = realDBMRPrefix + "-checkupfield-A"
	realDBCheckupFieldB = realDBMRPrefix + "-checkupfield-B"

	realDBRecordNoA        = realDBMRPrefix + "-MR-A"
	realDBRecordNoB        = realDBMRPrefix + "-MR-B"
	realDBPetNameA         = realDBMRPrefix + "-pet-A"
	realDBPetNameB         = realDBMRPrefix + "-pet-B"
	realDBOwnerNameA       = realDBMRPrefix + "-owner-A"
	realDBOwnerNameB       = realDBMRPrefix + "-owner-B"
	realDBClinicalPlanA    = realDBMRPrefix + "-plan-A"
	realDBClinicalPlanB    = realDBMRPrefix + "-plan-B"
	realDBAddendumA        = realDBMRPrefix + "-addendum-A"
	realDBAddendumB        = realDBMRPrefix + "-addendum-B"
	realDBTreatmentA       = realDBMRPrefix + "-treatment-A"
	realDBTreatmentB       = realDBMRPrefix + "-treatment-B"
	realDBTreatmentPlanA   = realDBMRPrefix + "-tpl-A"
	realDBTreatmentPlanB   = realDBMRPrefix + "-tpl-B"
	realDBVitalNotesA      = realDBMRPrefix + "-vital-A"
	realDBVitalNotesB      = realDBMRPrefix + "-vital-B"
	realDBImageURLA        = "https://example.test/" + realDBMRPrefix + "-img-A.png"
	realDBImageURLB        = "https://example.test/" + realDBMRPrefix + "-img-B.png"
	realDBVaccLotA         = realDBMRPrefix + "-lot-A"
	realDBVaccLotB         = realDBMRPrefix + "-lot-B"
	realDBCarePlanA        = realDBMRPrefix + "-care-A"
	realDBCarePlanB        = realDBMRPrefix + "-care-B"
	realDBExamMachineA     = realDBMRPrefix + "-exam-A"
	realDBExamMachineB     = realDBMRPrefix + "-exam-B"
	realDBLabDeviceA       = realDBMRPrefix + "-labdev-A"
	realDBLabDeviceB       = realDBMRPrefix + "-labdev-B"
	realDBLabItemCodeA     = "D3A-CODE"
	realDBLabItemCodeB     = "D3B-CODE"
	realDBLabStationSlotsB = `[{"key":"d3b","source_type":"fuji_nx600","device_hint":"D3B"}]`
	realDBAgentToken       = "d3-mr-agent-consumer-token"
	realDBDailyDate        = "2026-09-15"
)

func realDBNow() time.Time {
	return time.Date(2026, 9, 15, 0, 0, 0, 0, time.Local)
}

func configureGrant(fx testdb.ClinicGrantFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx, resource, grantClinicID)
}

func configureGrantAction(fx testdb.ClinicGrantFixture, resource, action string, grantClinicIDs ...uint64) func(*gin.Context) {
	granted := make(map[uint64]struct{}, len(grantClinicIDs))
	for _, id := range grantClinicIDs {
		granted[id] = struct{}{}
	}
	return func(c *gin.Context) {
		c.Set("clinic_id", fmt.Sprintf("%d", fx.ClinicB))
		c.Set("clinic_ids", []uint64{fx.ClinicA, fx.ClinicB})
		c.Set("user_id", fmt.Sprintf("%d", fx.StaffID))
		c.Set("is_system_admin", false)
		httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, res, act string) bool {
			if res != resource || act != action {
				return false
			}
			_, ok := granted[clinicID]
			return ok
		})
	}
}

func withIDParam(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%d", id)})
	}
}

func withParams(configure func(*gin.Context), params ...gin.Param) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = append(c.Params, params...)
	}
}

func withUUIDParam(configure func(*gin.Context), key string, id uuid.UUID) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = append(c.Params, gin.Param{Key: key, Value: id.String()})
	}
}

func fatalIfForbiddenClinic(t *testing.T, clinicID, forbidden uint64, op string) {
	t.Helper()
	if forbidden != 0 && clinicID == forbidden {
		t.Fatalf("%s must not query clinic %d without selected-clinic grant", op, clinicID)
	}
}

type clinicIDGuard struct {
	t                 *testing.T
	forbiddenClinicID uint64
}

func (g clinicIDGuard) check(clinicID uint64, op string) {
	fatalIfForbiddenClinic(g.t, clinicID, g.forbiddenClinicID, op)
}

func newMedicalRecordReadService(db *gorm.DB) MedicalRecordService {
	return NewMedicalRecordServiceWithTxAudit(
		NewMedicalRecordRepository(db),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
}

func seedOwnerPetRecord(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	ownerName, petName, recordNo string,
) (*model.Owner, *model.Pet, *model.MedicalRecord) {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, ownerName)
	pet := testdb.MakeSpeciesAndPet(t, db, clinicID, owner.ID, petName)
	require.NotZero(t, pet.ID)
	require.NotZero(t, pet.AnimalSpeciesID)
	record := testdb.MakeHistoryMedicalRecord(t, db, clinicID, pet.ID, recordNo, realDBNow())
	require.NoError(t, db.WithContext(context.Background()).Model(record).Update("owner_id", owner.ID).Error)
	record.OwnerID = &owner.ID
	return owner, pet, record
}

func ensureLabImportEnums(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, stmt := range []string{
		`DO $$ BEGIN CREATE TYPE lab_import_job_status AS ENUM ('received','validated','mapped','persisted','duplicate','needs_review','failed','reverted'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN CREATE TYPE lab_import_source_type AS ENUM ('fixture','drwan','manual'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}
	for _, value := range []string{"fuji_nx600", "fuji_au10v", "arkray_pu4010"} {
		require.NoError(t, db.Exec(`ALTER TYPE lab_import_source_type ADD VALUE IF NOT EXISTS '`+value+`'`).Error)
	}
}

func ensureCheckupFieldOptions(f *model.CheckupTypeField) *model.CheckupTypeField {
	if len(f.Options) == 0 {
		f.Options = datatypes.JSON([]byte("[]"))
	}
	return f
}

func reservationOwnerPet(db *gorm.DB) reservation.ReservationStore {
	return reservation.NewReservationRepository(db)
}

func persistenceTx(db *gorm.DB) Transactor {
	return persistence.NewTransactor(db)
}
