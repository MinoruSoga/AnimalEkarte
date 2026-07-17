package repository

// preload_followup_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (a) 既修正だが専用テストの無かった3 repo の
// master Preload clinic 隔離回帰テスト（examination / trimming / reservation FindAllByCategory）。
//
// 各テストは別クリニックのマスタを指す FK を植え付け、clinic_id スコープ Preload で
// 別クリニックのマスタが応答に混入しないことを検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// --- (a1) examination: ExaminationType ---

func makeExamRec(t *testing.T, db *gorm.DB, clinicID, petID, examTypeID uint64) uint64 {
	t.Helper()
	pid := petID
	e := &model.Examination{ClinicID: clinicID, PetID: &pid, ExamTypeID: examTypeID, Date: time.Now()}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e.ID
}

func TestExaminationRepository_FindByID_ExaminationTypePreloadClinicIsolation(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.ExaminationType{}, &model.Examination{}, &model.ExamResult{}))
	db.Exec("TRUNCATE TABLE exam_results CASCADE")
	db.Exec("TRUNCATE TABLE exams CASCADE")
	db.Exec("TRUNCATE TABLE exam_types CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "検査飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "検査ポチA")
	typeB := &model.ExaminationType{ClinicID: clinicB, Name: "医院Bの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeB).Error)
	typeA := &model.ExaminationType{ClinicID: clinicA, Name: "医院Aの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeA).Error)

	crossID := makeExamRec(t, db, clinicA, petA.ID, typeB.ID)
	legitID := makeExamRec(t, db, clinicA, petA.ID, typeA.ID)

	gotCross, err := repo.FindByID(ctx, clinicA, crossID)
	require.NoError(t, err)
	assert.Nil(t, gotCross.ExaminationType, "別クリニックの検査種別マスタが Preload で混入してはならない")

	gotLegit, err := repo.FindByID(ctx, clinicA, legitID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.ExaminationType, "同一クリニックの検査種別は Preload されるべき")
	assert.Equal(t, typeA.ID, gotLegit.ExaminationType.ID)
}

// --- (a2) trimming: Course / Options ---

func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation(t *testing.T) {
	// setupAppointmentTrimmingDetailTestDB = setupIsolatedTestDB + trimming 系 TRUNCATE。
	// 共有プール上の並行 TRUNCATE 破壊を避ける（TEST-FLAKE-P2）。
	db := setupAppointmentTrimmingDetailTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	// appointment_trimming_details.appointment_id は appointments(=Reservation) への FK のため実予約を作る。
	appt := makeReservation(t, db, clinicA)
	apptID := appt.ID
	courseB := &model.TrimmingCourse{ClinicID: clinicB, Name: "医院Bのコース"}
	require.NoError(t, db.WithContext(ctx).Create(courseB).Error)
	optionB := &model.TrimmingOption{ClinicID: clinicB, Name: "医院Bのオプション"}
	require.NoError(t, db.WithContext(ctx).Create(optionB).Error)

	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: apptID, CourseID: &courseB.ID}
	require.NoError(t, db.WithContext(ctx).Create(detail).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.AppointmentTrimmingOption{AppointmentID: apptID, OptionID: optionB.ID}).Error)

	got, err := repo.FindByAppointmentID(ctx, clinicA, apptID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.Course, "別クリニックのトリミングコースマスタが Preload で混入してはならない")
	assert.Empty(t, got.Options, "別クリニックのトリミングオプションマスタが Preload で混入してはならない")
}

// --- (a3) reservation: FindAllByCategory ReservationType ---

func TestReservationRepository_FindAllByCategory_ReservationTypePreloadClinicIsolation(t *testing.T) {
	db := setupTestDB(t)
	// TRUNCATE first: 他テストが残した orphan 行を除去してから AutoMigrate（FK 検証を通すため）。
	db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE")
	db.Exec("TRUNCATE TABLE appointments CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	require.NoError(t, ensureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{},
	))
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtB := makeReservationType(t, db, clinicB) // category general
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicA,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: rtB.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(ctx).Create(res).Error)

	got, _, err := repo.FindAllByCategory(ctx, clinicA, model.ReservationTypeCategoryGeneral, nil, nil, nil, nil, 1, 100)
	require.NoError(t, err)

	var found *model.Reservation
	for i := range got {
		if got[i].ID == res.ID {
			found = &got[i]
			break
		}
	}
	require.NotNil(t, found, "clinic A の予約は取得できる")
	assert.Nil(t, found.ReservationType, "別クリニックの診療区分マスタが Preload で混入してはならない")
}
