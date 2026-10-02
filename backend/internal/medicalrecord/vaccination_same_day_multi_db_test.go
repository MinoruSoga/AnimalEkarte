package medicalrecord

// vaccination_same_day_multi_db_test.go — EMR-105 real-DB pin for the SAVED-chart
// in-chart vaccination path. Registering 3 vaccinations on one saved (draft)
// medical record on the same day — each a sequential single Create call exactly
// as the in-chart form's useCreateVaccination POST issues it — must persist as
// 3 independent rows, and a later pet-scoped fetch must re-display them without
// mixing lot / next_date / vaccine (price) / record linkage. A 4th control row
// without medical_record_id models the standalone form path and proves linkage
// is not absorbed across entry paths.
//
// Deps mirror the production wiring: real repository + real vaccine repo +
// real reservation verifier + real medical-record locker inside a real tx.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestVaccinationService_SameDayMultiVaccineInChartRealDB(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.MedicalRecord{}))
	// medical_records CASCADE also empties vaccinations; all fixtures are created after this.
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	ctx := context.Background()
	const clinicID = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicID)

	owner := makeTestOwner(t, db, clinicID, "同日複数接種 飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicID, owner.ID, "同日複数接種 ペット")

	mkVaccine := func(name string, price int64) *model.Vaccine {
		t.Helper()
		v := &model.Vaccine{ClinicID: clinicID, Name: name, Price: &price, IsActive: true}
		require.NoError(t, db.WithContext(ctx).Create(v).Error)
		return v
	}
	vaccineA := mkVaccine("同日ワクチンA", 3000)
	vaccineB := mkVaccine("同日ワクチンB", 5500)
	vaccineC := mkVaccine("同日ワクチンC", 8000)
	vaccineStandalone := mkVaccine("同日ワクチン単独経路", 1200)

	// 保存済み（draft 確定前）カルテ — カルテ内フォーム経路の入口。
	record := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: "EMR105-SAMEDAY", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)

	svc := NewVaccinationService(
		NewVaccinationRepository(db),
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
	)

	nowJST := time.Now().In(config.JST)
	today := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, config.JST)
	todayStr := today.Format(time.DateOnly)
	next := func(days int) *time.Time { d := today.AddDate(0, 0, days); return &d }
	nst := func(s model.NextScheduleType) *model.NextScheduleType { return &s }

	// In-chart 経路: 同一保存済みカルテ・同一日・逐次 1 接種ずつ Create。
	inChartInputs := []CreateVaccinationInput{
		{MedicalRecordID: &record.ID, PetID: &pet.ID, VaccineID: vaccineA.ID, Date: today,
			Lot1: "LOT-A", NextDate: next(28), NextScheduleType: nst(model.NextScheduleType4Weeks), Remarks: "1回目"},
		{MedicalRecordID: &record.ID, PetID: &pet.ID, VaccineID: vaccineB.ID, Date: today,
			Lot1: "LOT-B", NextDate: next(35), NextScheduleType: nst(model.NextScheduleTypeOther), Remarks: "2回目"},
		{MedicalRecordID: &record.ID, PetID: &pet.ID, VaccineID: vaccineC.ID, Date: today,
			Lot1: "LOT-C", NextDate: next(365), NextScheduleType: nst(model.NextScheduleType1Year), Remarks: "3回目"},
	}
	// 単独フォーム経路の対照行（medical_record_id なし）— linkage が他行へ漏れないことの検証用。
	standaloneInput := CreateVaccinationInput{
		PetID: &pet.ID, VaccineID: vaccineStandalone.ID, Date: today,
		Lot1: "LOT-S", NextDate: next(14), Remarks: "単独フォーム対照",
	}

	wantLot := map[uint64]string{
		vaccineA.ID: "LOT-A", vaccineB.ID: "LOT-B", vaccineC.ID: "LOT-C", vaccineStandalone.ID: "LOT-S",
	}
	wantPrice := map[uint64]int64{
		vaccineA.ID: 3000, vaccineB.ID: 5500, vaccineC.ID: 8000, vaccineStandalone.ID: 1200,
	}
	wantRemarks := map[uint64]string{
		vaccineA.ID: "1回目", vaccineB.ID: "2回目", vaccineC.ID: "3回目", vaccineStandalone.ID: "単独フォーム対照",
	}

	createdIDs := make([]uint64, 0, 4)
	for i := range inChartInputs {
		got, err := svc.Create(ctx, clinicID, &inChartInputs[i])
		require.NoError(t, err, "in-chart registration %d must succeed", i+1)
		require.NotNil(t, got)
		createdIDs = append(createdIDs, got.ID)
	}
	standalone, err := svc.Create(ctx, clinicID, &standaloneInput)
	require.NoError(t, err, "standalone-path same-day registration must succeed")
	createdIDs = append(createdIDs, standalone.ID)

	// ── Persisted as 4 independent rows on the same day ──
	assert.Len(t, createdIDs, 4)
	assert.NotEqual(t, createdIDs[0], createdIDs[1])
	assert.NotEqual(t, createdIDs[1], createdIDs[2])

	assertRow := func(t *testing.T, row *model.Vaccination) {
		t.Helper()
		require.NotNil(t, row.Vaccine, "vaccine preload must resolve per row")
		assert.Equal(t, row.VaccineID, row.Vaccine.ID, "vaccine linkage must be the row's own")
		assert.Equal(t, wantLot[row.VaccineID], row.Lot1, "lot must be the row's own")
		assert.Equal(t, wantRemarks[row.VaccineID], row.Remarks, "remarks must be the row's own")
		require.NotNil(t, row.Vaccine.Price, "amount (vaccine price) must resolve per row")
		assert.Equal(t, wantPrice[row.VaccineID], *row.Vaccine.Price, "amount must be the row's own vaccine price")
		require.NotNil(t, row.PetID)
		assert.Equal(t, pet.ID, *row.PetID)
	}
	wantNextByVaccine := map[uint64]time.Time{
		vaccineA.ID: *next(28), vaccineB.ID: *next(35), vaccineC.ID: *next(365), vaccineStandalone.ID: *next(14),
	}

	// ── Later fetch (list re-display) returns all 4 same-day rows unmixed ──
	listed, total, err := svc.List(ctx, clinicID, &pet.ID, nil, &todayStr, &todayStr, "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total, "4 same-day registrations must persist as 4 rows")
	require.Len(t, listed, 4)
	seenIDs := make(map[uint64]bool, len(listed))
	for i := range listed {
		row := listed[i]
		seenIDs[row.ID] = true
		assertRow(t, &row)
		require.NotNil(t, row.NextDate)
		// next_date is a `date` column: compare calendar dates, not instants
		// (DB rows scan back as UTC midnight while inputs are JST-midnight).
		assert.Equal(t, wantNextByVaccine[row.VaccineID].Format(time.DateOnly),
			row.NextDate.Format(time.DateOnly),
			"next_due must be the row's own (vaccine=%d)", row.VaccineID)
		if row.VaccineID == vaccineStandalone.ID {
			assert.Nil(t, row.MedicalRecordID, "standalone row must keep nil record linkage")
		} else {
			require.NotNil(t, row.MedicalRecordID)
			assert.Equal(t, record.ID, *row.MedicalRecordID, "in-chart row must link the saved record")
		}
	}
	for _, id := range createdIDs {
		assert.True(t, seenIDs[id], "created row %d must be visible in the pet-scoped re-display", id)
	}

	// ── Detail re-fetch (FindByID) returns each row with its own fields ──
	for i, input := range inChartInputs {
		got, err := svc.GetByID(ctx, clinicID, createdIDs[i])
		require.NoError(t, err)
		assert.Equal(t, input.Lot1, got.Lot1)
		assert.Equal(t, input.VaccineID, got.VaccineID)
		require.NotNil(t, got.MedicalRecordID)
		assert.Equal(t, record.ID, *got.MedicalRecordID)
		require.NotNil(t, got.NextDate)
		assert.Equal(t, input.NextDate.Format(time.DateOnly), got.NextDate.Format(time.DateOnly))
	}
	gotStandalone, err := svc.GetByID(ctx, clinicID, standalone.ID)
	require.NoError(t, err)
	assert.Nil(t, gotStandalone.MedicalRecordID)
	assert.Equal(t, "LOT-S", gotStandalone.Lot1)
}
