package medicalrecord

// EMR-356: 問診画面/履歴の体重投影のために FindAll が vital_records を
// 1回のバッチ preload で取得することを検証する（行ごとの N+1 禁止）。
// 投影対象は「同一 clinic/pet/record に相関・weight 記録済み・daily_record
// （入院/日次）由来でない・論理削除されていない」行のみ。

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// vitalQueryCountingLogger は発行 SQL を収集して vital_records への
// クエリ回数を数える（N+1 検出用）。
type vitalQueryCountingLogger struct {
	gormlogger.Interface
	statements *[]string
}

func (l vitalQueryCountingLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	sql, _ := fc()
	*l.statements = append(*l.statements, sql)
}

func TestMedicalRecordRepository_FindAll_WeightMeasurementsPreload(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Cage{},
		&model.Hospitalization{},
		&model.DailyRecord{},
		&model.VitalRecord{},
	))
	db.Exec("TRUNCATE TABLE vital_records, daily_records, hospitalizations, medical_records CASCADE")

	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	ownerA := makeTestOwner(t, db, clinicA, "体重投影自院飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "体重投影自院ペット")
	ownerB := makeTestOwner(t, db, clinicB, "体重投影別院飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "体重投影別院ペット")

	older := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "WPROJ-OLDER",
		Date:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	newer := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "WPROJ-NEWER",
		Date:    time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	foreignRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "WPROJ-FOREIGN",
		Date:    time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerB.ID, PetID: &petB.ID,
	})

	kg, g := model.BodyWeightUnitKg, model.BodyWeightUnitG
	w42, w4300 := 4.2, 4300.0
	temp38 := 38.5
	recorded1 := time.Date(2026, 3, 1, 9, 30, 0, 0, time.UTC)
	recorded2 := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)
	olderID := older.ID
	newerID := newer.ID
	foreignID := foreignRecord.ID

	// 投影対象: 2件（recorded_at 昇順 → v1 が先）
	v1 := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &olderID,
		RecordedAt: recorded1, Weight: &w42, WeightUnit: kg,
	}
	v2 := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &olderID,
		RecordedAt: recorded2, Weight: &w4300, WeightUnit: g,
	}
	// 除外対象: 体重なし / 論理削除 / daily_record（入院・日次）由来 / 別院・別ペット相関
	tempOnly := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &olderID,
		RecordedAt: recorded1.Add(time.Hour), Temperature: &temp38,
	}
	deletedWeight := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &olderID,
		RecordedAt: recorded1.Add(2 * time.Hour), Weight: &w42, WeightUnit: kg,
	}
	hosp := makeHospitalizationFixture(t, db, clinicA, ownerA.ID, petA.ID, nil)
	daily := &model.DailyRecord{
		ClinicID: clinicA, HospitalizationID: hosp.ID, Date: recorded1,
	}
	require.NoError(t, db.WithContext(ctx).Create(daily).Error)
	dailyVital := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, DailyRecordID: &daily.ID,
		RecordedAt: recorded1.Add(3 * time.Hour), Weight: &w42, WeightUnit: kg,
	}
	foreignVital := &model.VitalRecord{
		ClinicID: clinicB, PetID: petB.ID, MedicalRecordID: &olderID,
		RecordedAt: recorded1.Add(4 * time.Hour), Weight: &w42, WeightUnit: kg,
	}
	otherRecordWeight := &model.VitalRecord{
		ClinicID: clinicB, PetID: petB.ID, MedicalRecordID: &foreignID,
		RecordedAt: recorded1, Weight: &w42, WeightUnit: kg,
	}
	// newer 側の「今回」測定が古いカルテ行へ漏れないことも確認する
	newerWeight := &model.VitalRecord{
		ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &newerID,
		RecordedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), Weight: &w42, WeightUnit: kg,
	}

	for _, v := range []*model.VitalRecord{
		v1, v2, tempOnly, deletedWeight, dailyVital, foreignVital, otherRecordWeight, newerWeight,
	} {
		require.NoError(t, db.WithContext(ctx).Create(v).Error)
	}
	require.NoError(t, db.WithContext(ctx).Delete(deletedWeight).Error)

	// N+1 計測: SQL を収集する logger を差した repo で FindAll を実行する。
	var statements []string
	countingDB := db.Session(&gorm.Session{
		Logger: vitalQueryCountingLogger{Interface: gormlogger.Discard, statements: &statements},
	})
	countingRepo := NewMedicalRecordRepository(countingDB)

	petAID := petA.ID
	items, total, err := countingRepo.FindAll(
		ctx, []uint64{clinicA}, MedicalRecordListFilters{PetID: &petAID}, 1, 50,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, items, 2)

	vitalQueries := 0
	for _, sql := range statements {
		if strings.Contains(sql, "vital_records") {
			vitalQueries++
		}
	}
	assert.Equal(t, 1, vitalQueries, "weight preload は行ごとの N+1 ではなく1回のバッチ取得")

	// 日付降順: newer が先。newer には newer 自身の測定だけが相関する。
	assert.Equal(t, newer.ID, items[0].ID)
	require.Len(t, items[0].Vitals, 1)
	assert.Equal(t, newerWeight.ID, items[0].Vitals[0].ID)

	assert.Equal(t, older.ID, items[1].ID)
	require.Len(t, items[1].Vitals, 2, "体重なし/削除/daily/別院相関は投影しない")
	assert.Equal(t, v1.ID, items[1].Vitals[0].ID, "recorded_at 昇順")
	assert.Equal(t, v2.ID, items[1].Vitals[1].ID)
	assert.Equal(t, 4.2, *items[1].Vitals[0].Weight)
	assert.Equal(t, 4300.0, *items[1].Vitals[1].Weight)
}
