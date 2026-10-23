package medicalrecord

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type nullWeightUnitFixture struct {
	db      *gorm.DB
	clinicA uint64
	clinicB uint64

	petA   *model.Pet
	staffA *model.Staff

	older   *model.MedicalRecord
	newer   *model.MedicalRecord
	foreign *model.MedicalRecord
	deleted *model.MedicalRecord

	daily *model.DailyRecord

	nullVital          *model.VitalRecord
	kgVital            *model.VitalRecord
	gVital             *model.VitalRecord
	tempOnlyVital      *model.VitalRecord
	deletedVital       *model.VitalRecord
	wrongClinicVital   *model.VitalRecord
	wrongPetVital      *model.VitalRecord
	foreignVital       *model.VitalRecord
	newerNullVital     *model.VitalRecord
	deletedParentVital *model.VitalRecord
	dailyOnlyVital     *model.VitalRecord
	dualDailyVital     *model.VitalRecord
	standaloneVital    *model.VitalRecord
}

func setupNullWeightUnitFixture(t *testing.T) *nullWeightUnitFixture {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Cage{},
		&model.Hospitalization{},
		&model.DailyRecord{},
		&model.VitalRecord{},
	), "fixture setup")

	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	ownerA := makeTestOwner(t, db, clinicA, "NULL単位自院飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "NULL単位自院ペット")
	otherPetA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "NULL単位別ペット")
	ownerB := makeTestOwner(t, db, clinicB, "NULL単位別院飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "NULL単位別院ペット")
	staffA := makeMedicalRecordListStaff(t, db, clinicA, "NULL単位看護師", model.StaffTypeNurse)

	older := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "NULLUNIT-OLDER",
		Date:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	newer := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "NULLUNIT-NEWER",
		Date:    time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	foreign := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "NULLUNIT-FOREIGN",
		Date:    time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerB.ID, PetID: &petB.ID,
	})
	deleted := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "NULLUNIT-DELETED",
		Date:    time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})

	kgUnit, gUnit := model.BodyWeightUnitKg, model.BodyWeightUnitG
	wNull, wKg, wG, wForeign, wNewer := 4.2, 4.5, 4300.0, 6.8, 5.6
	tempNull, tempKg, tempOnly := 38.5, 38.7, 37.8
	heartNull, heartKg := 90, 88
	respNull, respKg := 24, 22
	recordedMorning := time.Date(2026, 3, 1, 9, 30, 0, 0, time.UTC)
	recordedAfternoon := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)

	nullVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning, StaffID: &staffA.ID,
		Temperature: &tempNull, HeartRate: &heartNull, RespirationRate: &respNull,
		Weight: &wNull, WeightUnit: kgUnit,
		Notes: "historic-null-unit", Version: 7,
		CreatedAt: time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 1, 8, 30, 0, 0, time.UTC),
	}
	kgVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning, StaffID: &staffA.ID,
		Temperature: &tempKg, HeartRate: &heartKg, RespirationRate: &respKg,
		Weight: &wKg, WeightUnit: kgUnit,
		Notes: "explicit-kg-full", Version: 3,
		CreatedAt: time.Date(2026, 3, 1, 8, 10, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 1, 8, 40, 0, 0, time.UTC),
	}
	gVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedAfternoon, StaffID: &staffA.ID,
		Weight: &wG, WeightUnit: gUnit,
		Notes: "explicit-g", Version: 2,
		CreatedAt: time.Date(2026, 3, 1, 8, 20, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 1, 8, 50, 0, 0, time.UTC),
	}
	tempOnlyVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning.Add(time.Hour), Temperature: &tempOnly,
		WeightUnit: kgUnit,
	}
	deletedVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning.Add(2 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}
	wrongClinicVital := &model.VitalRecord{
		ClinicID: clinicB, PetID: petA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning.Add(3 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}
	wrongPetVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: otherPetA.ID, MedicalRecordID: &older.ID,
		RecordedAt: recordedMorning.Add(4 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}
	foreignVital := &model.VitalRecord{
		ClinicID: clinicB, PetID: petB.ID, MedicalRecordID: &foreign.ID,
		RecordedAt: time.Date(2026, 3, 3, 11, 0, 0, 0, time.UTC),
		Weight:     &wForeign, WeightUnit: kgUnit,
	}
	newerNullVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &newer.ID,
		RecordedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
		Weight:     &wNewer, WeightUnit: kgUnit,
		Notes: "newer-null-unit",
	}
	deletedParentVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &deleted.ID,
		RecordedAt: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC),
		Weight:     &wNull, WeightUnit: kgUnit,
	}
	standaloneVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID,
		RecordedAt: recordedMorning.Add(5 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}

	hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)
	daily := &model.DailyRecord{
		ClinicID: clinicA, HospitalizationID: hosp.ID,
		Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(ctx).Create(daily).Error, "fixture setup")

	dailyOnlyVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, DailyRecordID: &daily.ID,
		RecordedAt: recordedMorning.Add(6 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}
	dualDailyVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &older.ID, DailyRecordID: &daily.ID,
		RecordedAt: recordedMorning.Add(7 * time.Hour), Weight: &wNull, WeightUnit: kgUnit,
	}

	for _, v := range []*model.VitalRecord{
		nullVital, kgVital, gVital, tempOnlyVital, deletedVital, wrongClinicVital,
		wrongPetVital, foreignVital, newerNullVital, deletedParentVital,
		standaloneVital, dailyOnlyVital, dualDailyVital,
	} {
		require.NoError(t, db.WithContext(ctx).Create(v).Error, "fixture setup")
	}
	require.NoError(t, db.WithContext(ctx).Delete(deletedVital).Error, "fixture setup")
	require.NoError(t, db.WithContext(ctx).Delete(deleted).Error, "fixture setup")

	for _, v := range []*model.VitalRecord{nullVital, newerNullVital} {
		res := db.WithContext(ctx).
			Model(&model.VitalRecord{}).
			Where("id = ?", v.ID).
			Update("weight_unit", gorm.Expr("NULL"))
		require.NoError(t, res.Error, "fixture setup")
		require.Equal(t, int64(1), res.RowsAffected, "fixture setup")
	}

	fixture := &nullWeightUnitFixture{
		db: db, clinicA: clinicA, clinicB: clinicB,
		petA: petA, staffA: staffA,
		older: older, newer: newer, foreign: foreign, deleted: deleted,
		daily:              daily,
		nullVital:          nullVital,
		kgVital:            kgVital,
		gVital:             gVital,
		tempOnlyVital:      tempOnlyVital,
		deletedVital:       deletedVital,
		wrongClinicVital:   wrongClinicVital,
		wrongPetVital:      wrongPetVital,
		foreignVital:       foreignVital,
		newerNullVital:     newerNullVital,
		deletedParentVital: deletedParentVital,
		dailyOnlyVital:     dailyOnlyVital,
		dualDailyVital:     dualDailyVital,
		standaloneVital:    standaloneVital,
	}
	fixture.assertStoredWeightUnitsNULL(t, ctx)
	return fixture
}

func assertStoredWeightUnitNULL(t *testing.T, ctx context.Context, db *gorm.DB, vitalID uint64) {
	t.Helper()
	var stored struct {
		WeightUnit *string `gorm:"column:weight_unit"`
	}
	require.NoError(t, db.WithContext(ctx).
		Model(&model.VitalRecord{}).
		Select("weight_unit").
		Where("id = ?", vitalID).
		Take(&stored).Error)
	assert.Nil(t, stored.WeightUnit,
		"stored weight_unit must stay NULL; normalization is query-only")
}

func (f *nullWeightUnitFixture) assertStoredWeightUnitsNULL(t *testing.T, ctx context.Context) {
	t.Helper()
	assertStoredWeightUnitNULL(t, ctx, f.db, f.nullVital.ID)
	assertStoredWeightUnitNULL(t, ctx, f.db, f.newerNullVital.ID)
}

func TestMedicalRecordRepository_NullWeightUnit_FindAll(t *testing.T) {
	f := setupNullWeightUnitFixture(t)
	ctx := context.Background()

	for read := 1; read <= 2; read++ {
		var statements []string
		countingDB := f.db.Session(&gorm.Session{
			Logger: vitalQueryCountingLogger{Interface: gormlogger.Discard, statements: &statements},
		})
		repo := NewMedicalRecordRepository(countingDB)

		items, total, err := repo.FindAll(
			ctx, []uint64{f.clinicA, f.clinicB}, MedicalRecordListFilters{}, 1, 50,
		)
		t.Logf("FindAll read %d: err=%v total=%d items=%d", read, err, total, len(items))
		assert.NoError(t, err)
		if err != nil {
			f.assertStoredWeightUnitsNULL(t, ctx)
			continue
		}

		assert.Equal(t, int64(3), total)
		require.Len(t, items, 3)
		assert.Equal(t, f.newer.ID, items[0].ID)
		assert.Equal(t, f.foreign.ID, items[1].ID)
		assert.Equal(t, f.older.ID, items[2].ID)

		vitalQueries := 0
		vitalSQL := ""
		for _, sql := range statements {
			if strings.Contains(sql, "vital_records") {
				vitalQueries++
				vitalSQL = sql
			}
		}
		assert.Equal(t, 1, vitalQueries, "vitals must load as one batched preload")
		t.Logf("FindAll read %d vital preload SQL: %s", read, vitalSQL)
		inIdx := strings.Index(vitalSQL, " IN (")
		if assert.GreaterOrEqual(t, inIdx, 0, "vital preload must use a parent IN batch") {
			endIdx := strings.Index(vitalSQL[inIdx:], ")")
			require.Greater(t, endIdx, 0)
			inList := "," + vitalSQL[inIdx+len(" IN ("):inIdx+endIdx] + ","
			for _, parentID := range []uint64{f.older.ID, f.newer.ID, f.foreign.ID} {
				assert.Contains(t, inList, ","+strconv.FormatUint(parentID, 10)+",",
					"parent %d must be inside the batched vital preload IN list", parentID)
			}
		}

		require.Len(t, items[0].Vitals, 1)
		assert.Equal(t, f.newerNullVital.ID, items[0].Vitals[0].ID)
		assert.Equal(t, model.BodyWeightUnitKg, items[0].Vitals[0].WeightUnit,
			"newer NULL weight_unit must read as Kg")
		require.Len(t, items[1].Vitals, 1)
		assert.Equal(t, f.foreignVital.ID, items[1].Vitals[0].ID)
		assert.Equal(t, model.BodyWeightUnitKg, items[1].Vitals[0].WeightUnit)

		olderVitals := items[2].Vitals
		olderIDs := make([]uint64, 0, len(olderVitals))
		for _, v := range olderVitals {
			olderIDs = append(olderIDs, v.ID)
		}
		assert.Equal(t,
			[]uint64{f.nullVital.ID, f.kgVital.ID, f.gVital.ID},
			olderIDs,
			"only correlated non-deleted non-daily weight rows, recorded_at then id order",
		)

		var nullLoaded *model.VitalRecord
		for i := range olderVitals {
			if olderVitals[i].ID == f.nullVital.ID {
				nullLoaded = &olderVitals[i]
			}
		}
		if nullLoaded != nil {
			t.Logf("FindAll read %d: null vital id=%d loaded unit=%q", read, nullLoaded.ID, nullLoaded.WeightUnit)
		} else {
			t.Logf("FindAll read %d: null vital id=%d absent from projection", read, f.nullVital.ID)
		}
		require.Len(t, olderVitals, 3)
		require.NotNil(t, olderVitals[0].Weight)
		assert.Equal(t, 4.2, *olderVitals[0].Weight)
		assert.Equal(t, model.BodyWeightUnitKg, olderVitals[0].WeightUnit,
			"imported NULL weight_unit must normalize to Kg on read")
		require.NotNil(t, olderVitals[1].Weight)
		assert.Equal(t, 4.5, *olderVitals[1].Weight)
		assert.Equal(t, model.BodyWeightUnitKg, olderVitals[1].WeightUnit)
		require.NotNil(t, olderVitals[2].Weight)
		assert.Equal(t, 4300.0, *olderVitals[2].Weight)
		assert.Equal(t, model.BodyWeightUnitG, olderVitals[2].WeightUnit)

		_, probe := marshalResponseForWeightProbe(t, &items[2])
		require.Len(t, probe.WeightMeasurements, 3)
		wantWire := []struct {
			id    uint64
			value float64
			unit  string
			at    time.Time
		}{
			{f.nullVital.ID, 4.2, "Kg", f.nullVital.RecordedAt},
			{f.kgVital.ID, 4.5, "Kg", f.kgVital.RecordedAt},
			{f.gVital.ID, 4300.0, "g", f.gVital.RecordedAt},
		}
		for i, want := range wantWire {
			got := probe.WeightMeasurements[i]
			assert.Equal(t, strconv.FormatUint(want.id, 10), string(got.ID))
			require.NotNil(t, got.Value)
			assert.Equal(t, want.value, *got.Value)
			assert.Equal(t, want.unit, got.Unit)
			assert.True(t, got.RecordedAt.Equal(want.at),
				"wire recorded_at must keep the measured instant: got %s want %s", got.RecordedAt, want.at)
		}

		_, newerProbe := marshalResponseForWeightProbe(t, &items[0])
		require.Len(t, newerProbe.WeightMeasurements, 1)
		assert.Equal(t, "Kg", newerProbe.WeightMeasurements[0].Unit)
		_, foreignProbe := marshalResponseForWeightProbe(t, &items[1])
		require.Len(t, foreignProbe.WeightMeasurements, 1)
		assert.Equal(t, "Kg", foreignProbe.WeightMeasurements[0].Unit)

		f.assertStoredWeightUnitsNULL(t, ctx)
	}
}

func TestMedicalRecordRepository_NullWeightUnit_Detail(t *testing.T) {
	f := setupNullWeightUnitFixture(t)
	ctx := context.Background()
	repo := NewMedicalRecordRepository(f.db)
	const missingID = uint64(9000000000)

	readers := []struct {
		name string
		read func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error)
	}{
		{
			name: "FindByID",
			read: func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
				return repo.FindByID(ctx, clinicIDs[0], id)
			},
		},
		{
			name: "FindByIDForClinics",
			read: func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
				return repo.FindByIDForClinics(ctx, clinicIDs, id)
			},
		},
	}

	for _, tc := range readers {
		t.Run(tc.name, func(t *testing.T) {
			for read := 1; read <= 2; read++ {
				record, err := tc.read(ctx, []uint64{f.clinicA, f.clinicB}, f.older.ID)
				t.Logf("%s read %d: err=%v", tc.name, read, err)
				assert.NoError(t, err)
				if err != nil {
					f.assertStoredWeightUnitsNULL(t, ctx)
					continue
				}
				require.NotNil(t, record)

				ids := make([]uint64, 0, len(record.Vitals))
				vitalByID := make(map[uint64]model.VitalRecord, len(record.Vitals))
				for _, v := range record.Vitals {
					ids = append(ids, v.ID)
					vitalByID[v.ID] = v
				}
				assert.ElementsMatch(t,
					[]uint64{
						f.nullVital.ID, f.kgVital.ID, f.gVital.ID,
						f.tempOnlyVital.ID, f.dualDailyVital.ID,
					},
					ids,
					"detail vitals keep non-weight and daily-linked rows but never deleted/wrong-clinic/wrong-pet/daily-only/standalone/other-record rows",
				)

				nullLoaded, nullOK := vitalByID[f.nullVital.ID]
				if nullOK {
					t.Logf("%s read %d: null vital id=%d loaded unit=%q", tc.name, read, nullLoaded.ID, nullLoaded.WeightUnit)
				} else {
					t.Logf("%s read %d: null vital id=%d absent from detail vitals", tc.name, read, f.nullVital.ID)
				}
				if assert.True(t, nullOK) {
					assert.Equal(t, f.clinicA, nullLoaded.ClinicID)
					assert.Equal(t, f.petA.ID, nullLoaded.PetID)
					require.NotNil(t, nullLoaded.MedicalRecordID)
					assert.Equal(t, f.older.ID, *nullLoaded.MedicalRecordID)
					assert.Nil(t, nullLoaded.DailyRecordID)
					assert.True(t, nullLoaded.RecordedAt.Equal(f.nullVital.RecordedAt))
					require.NotNil(t, nullLoaded.StaffID)
					assert.Equal(t, f.staffA.ID, *nullLoaded.StaffID)
					require.NotNil(t, nullLoaded.Temperature)
					assert.Equal(t, 38.5, *nullLoaded.Temperature)
					require.NotNil(t, nullLoaded.HeartRate)
					assert.Equal(t, 90, *nullLoaded.HeartRate)
					require.NotNil(t, nullLoaded.RespirationRate)
					assert.Equal(t, 24, *nullLoaded.RespirationRate)
					require.NotNil(t, nullLoaded.Weight)
					assert.Equal(t, 4.2, *nullLoaded.Weight)
					assert.Equal(t, model.BodyWeightUnitKg, nullLoaded.WeightUnit,
						"stored NULL weight_unit must read as Kg")
					assert.Equal(t, "historic-null-unit", nullLoaded.Notes)
					assert.Equal(t, 7, nullLoaded.Version)
					assert.True(t, nullLoaded.CreatedAt.Equal(f.nullVital.CreatedAt))
					assert.False(t, nullLoaded.UpdatedAt.IsZero())
					assert.False(t, nullLoaded.DeletedAt.Valid)
				}

				kgLoaded, kgOK := vitalByID[f.kgVital.ID]
				if assert.True(t, kgOK) {
					assert.Equal(t, f.kgVital.ID, kgLoaded.ID)
					assert.Equal(t, f.clinicA, kgLoaded.ClinicID)
					assert.Equal(t, f.petA.ID, kgLoaded.PetID)
					require.NotNil(t, kgLoaded.MedicalRecordID)
					assert.Equal(t, f.older.ID, *kgLoaded.MedicalRecordID)
					assert.Nil(t, kgLoaded.DailyRecordID)
					assert.True(t, kgLoaded.RecordedAt.Equal(f.kgVital.RecordedAt))
					require.NotNil(t, kgLoaded.StaffID)
					assert.Equal(t, f.staffA.ID, *kgLoaded.StaffID)
					require.NotNil(t, kgLoaded.Temperature)
					assert.Equal(t, 38.7, *kgLoaded.Temperature)
					require.NotNil(t, kgLoaded.HeartRate)
					assert.Equal(t, 88, *kgLoaded.HeartRate)
					require.NotNil(t, kgLoaded.RespirationRate)
					assert.Equal(t, 22, *kgLoaded.RespirationRate)
					require.NotNil(t, kgLoaded.Weight)
					assert.Equal(t, 4.5, *kgLoaded.Weight)
					assert.Equal(t, model.BodyWeightUnitKg, kgLoaded.WeightUnit)
					assert.Equal(t, "explicit-kg-full", kgLoaded.Notes)
					assert.Equal(t, 3, kgLoaded.Version)
					assert.True(t, kgLoaded.CreatedAt.Equal(f.kgVital.CreatedAt))
					assert.True(t, kgLoaded.UpdatedAt.Equal(f.kgVital.UpdatedAt))
					assert.False(t, kgLoaded.DeletedAt.Valid)
				}

				gLoaded, gOK := vitalByID[f.gVital.ID]
				if assert.True(t, gOK) {
					require.NotNil(t, gLoaded.Weight)
					assert.Equal(t, 4300.0, *gLoaded.Weight)
					assert.Equal(t, model.BodyWeightUnitG, gLoaded.WeightUnit)
					assert.True(t, gLoaded.RecordedAt.Equal(f.gVital.RecordedAt))
				}

				tempLoaded, tempOK := vitalByID[f.tempOnlyVital.ID]
				if assert.True(t, tempOK) {
					assert.Nil(t, tempLoaded.Weight)
					require.NotNil(t, tempLoaded.Temperature)
					assert.Equal(t, 37.8, *tempLoaded.Temperature)
					assert.Nil(t, tempLoaded.HeartRate)
					assert.Nil(t, tempLoaded.RespirationRate)
					assert.Nil(t, tempLoaded.StaffID)
					assert.Nil(t, tempLoaded.DailyRecordID)
				}

				dualLoaded, dualOK := vitalByID[f.dualDailyVital.ID]
				if assert.True(t, dualOK) {
					require.NotNil(t, dualLoaded.MedicalRecordID)
					assert.Equal(t, f.older.ID, *dualLoaded.MedicalRecordID)
					require.NotNil(t, dualLoaded.DailyRecordID)
					assert.Equal(t, f.daily.ID, *dualLoaded.DailyRecordID)
					require.NotNil(t, dualLoaded.Weight)
					assert.Equal(t, 4.2, *dualLoaded.Weight)
				}

				_, probe := marshalResponseForWeightProbe(t, record)
				require.Len(t, probe.WeightMeasurements, 3)
				wantWire := []struct {
					id    uint64
					value float64
					unit  string
					at    time.Time
				}{
					{f.nullVital.ID, 4.2, "Kg", f.nullVital.RecordedAt},
					{f.kgVital.ID, 4.5, "Kg", f.kgVital.RecordedAt},
					{f.gVital.ID, 4300.0, "g", f.gVital.RecordedAt},
				}
				for i, want := range wantWire {
					got := probe.WeightMeasurements[i]
					assert.Equal(t, strconv.FormatUint(want.id, 10), string(got.ID))
					require.NotNil(t, got.Value)
					assert.Equal(t, want.value, *got.Value)
					assert.Equal(t, want.unit, got.Unit)
					assert.True(t, got.RecordedAt.Equal(want.at))
				}

				f.assertStoredWeightUnitsNULL(t, ctx)
			}

			cross, err := tc.read(ctx, []uint64{f.clinicB}, f.older.ID)
			assert.Nil(t, cross)
			assert.True(t, apperrors.IsNotFound(err),
				"cross-clinic detail read must fail closed, got %v", err)

			deletedRecord, err := tc.read(ctx, []uint64{f.clinicA, f.clinicB}, f.deleted.ID)
			assert.Nil(t, deletedRecord)
			assert.True(t, apperrors.IsNotFound(err),
				"deleted parent detail read must fail closed, got %v", err)

			missing, err := tc.read(ctx, []uint64{f.clinicA, f.clinicB}, missingID)
			assert.Nil(t, missing)
			assert.True(t, apperrors.IsNotFound(err),
				"missing record detail read must fail closed, got %v", err)
		})
	}
}
