package repository

// preload_followup_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (a) 既修正だが専用テストの無かった3 repo の
// master Preload clinic 隔離回帰テスト（examination / trimming / reservation FindAllByCategory）。
//
// 各テストは別クリニックのマスタを指す FK を植え付け、clinic_id スコープで
// 別クリニックのマスタが応答に混入しないことを検証する。必須マスタを指す
// Examination は、現在の relation scope により汚染行そのものを fail-closed で除外する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/trimming"
)

// --- (a1) examination: ExaminationType ---

func makeExamRec(t *testing.T, db *gorm.DB, clinicID, medicalRecordID, petID, doctorID, examTypeID uint64) uint64 {
	t.Helper()
	mrid := medicalRecordID
	pid := petID
	did := doctorID
	e := &model.Examination{
		ClinicID:        clinicID,
		MedicalRecordID: &mrid,
		PetID:           &pid,
		DoctorID:        &did,
		ExamTypeID:      examTypeID,
		Date:            time.Now(),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e.ID
}

func TestExaminationRepository_FindByID_ExaminationTypePreloadClinicIsolation(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ExaminationType{},
		&model.Examination{},
		&model.ExamResult{},
	))
	db.Exec("TRUNCATE TABLE exam_results CASCADE")
	db.Exec("TRUNCATE TABLE exams CASCADE")
	db.Exec("TRUNCATE TABLE exam_types CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	petA, mrA, doctorA := makeClinicScopedClinicalReadParents(t, db, clinicA, "検査")
	typeB := &model.ExaminationType{ClinicID: clinicB, Name: "医院Bの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeB).Error)
	typeA := &model.ExaminationType{ClinicID: clinicA, Name: "医院Aの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeA).Error)

	crossID := makeExamRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeB.ID)
	legitID := makeExamRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeA.ID)

	tests := []struct {
		name         string
		id           uint64
		wantNotFound bool
		wantTypeID   uint64
	}{
		{
			name:         "別クリニックの必須検査種別を指す行は取得対象外",
			id:           crossID,
			wantNotFound: true,
		},
		{
			name:       "同一クリニックの検査種別と整合した患者医師関係を取得",
			id:         legitID,
			wantTypeID: typeA.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, clinicA, tt.id)
			if tt.wantNotFound {
				require.Error(t, err)
				assert.True(t, apperrors.IsNotFound(err))
				assert.Nil(t, got, "別クリニックの検査種別を参照する行を返してはならない")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got.ExaminationType, "同一クリニックの検査種別は Preload されるべき")
			assert.Equal(t, tt.wantTypeID, got.ExaminationType.ID)
			require.NotNil(t, got.Pet)
			require.NotNil(t, got.Pet.Owner)
			require.NotNil(t, got.Doctor)
			assert.Equal(t, doctorA.ID, got.Doctor.ID)
		})
	}
}

// --- (a2) trimming: Course / Options ---

func setupPreloadTrimmingDetailTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupIsolatedTestDB(t)
	if db.Migrator().HasTable(&model.AppointmentTrimmingOption{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.AppointmentTrimmingDetail{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.TrimmingCourse{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE trimming_courses CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.TrimmingOption{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE trimming_options CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.ReservationType{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE reservation_types CASCADE").Error)
	}
	require.NoError(t, ensureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{},
	))
	return db
}

func TestAppointmentTrimmingDetailRepository_FindByAppointmentID_MasterPreloadClinicIsolation(t *testing.T) {
	// setupPreloadTrimmingDetailTestDB = setupIsolatedTestDB + trimming 系 TRUNCATE。
	// 共有プール上の並行 TRUNCATE 破壊を避ける（TEST-FLAKE-P2）。
	db := setupPreloadTrimmingDetailTestDB(t)
	repo := trimming.NewAppointmentTrimmingDetailRepository(db)
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

func TestReservationRepository_FindAllByCategory_ReservationTypeJoinClinicIsolation(t *testing.T) {
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
	assert.Nil(t, found, "別クリニックの診療区分マスタで clinic A の予約をカテゴリ分類してはならない")
}
