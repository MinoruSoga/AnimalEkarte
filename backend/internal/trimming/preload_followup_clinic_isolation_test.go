package trimming

// preload_followup_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — master Preload clinic 隔離回帰テスト。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// --- (a2) trimming: Course / Options ---

func setupB10PreloadTrimmingDetailTestDB(t *testing.T) *gorm.DB {
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
	// setupB10PreloadTrimmingDetailTestDB = setupIsolatedTestDB + trimming 系 TRUNCATE。
	// 共有プール上の並行 TRUNCATE 破壊を避ける（TEST-FLAKE-P2）。
	db := setupB10PreloadTrimmingDetailTestDB(t)
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
