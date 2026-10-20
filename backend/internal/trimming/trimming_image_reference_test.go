package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// setupImageReferenceTestDB は ImageKeyInUse 用に reservation → detail の実 FK 構成を整備する。
func setupImageReferenceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupIsolatedTestDB(t)
	if db.Migrator().HasTable(&model.AppointmentTrimmingOption{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.AppointmentTrimmingDetail{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE").Error)
	}
	if db.Migrator().HasTable(&model.ReservationType{}) {
		require.NoError(t, db.Exec("TRUNCATE TABLE reservation_types CASCADE").Error)
	}
	require.NoError(t, ensureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.AppointmentTrimmingDetail{}, &model.AppointmentTrimmingOption{}, &model.TrimmingCourse{}, &model.AuditLog{}))
	return db
}

func insertTrimmingDetailWithImages(t *testing.T, db *gorm.DB, clinicID uint64, style, completed string) uint64 {
	t.Helper()
	appointment := makeReservation(t, db, clinicID)
	require.NoError(t, db.Create(&model.AppointmentTrimmingDetail{
		ClinicID:       clinicID,
		AppointmentID:  appointment.ID,
		StyleImage:     style,
		CompletedImage: completed,
		BWUnit:         model.BodyWeightUnitKg,
	}).Error)
	return appointment.ID
}

// TestImageKeyInUse_CountsEveryResolvableReferenceForm は canonical key だけでなく
// resolver が同一 key と解決する legacy 形式も参照として数えることを検証する
// （DELETE /trimmings/images の 409 ガードが legacy 参照を見落とさないため）。
func TestImageKeyInUse_CountsEveryResolvableReferenceForm(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const key = "trimming-images/1/style.jpg"

	cases := []struct {
		name      string
		style     string
		completed string
		want      bool
	}{
		{name: "canonical key", style: key, want: true},
		{name: "canonical in completed", completed: key, want: true},
		{name: "uploads-wrapped legacy", style: "/uploads/" + key, want: true},
		{name: "uploads-wrapped with query", style: "/uploads/" + key + "?sig=x", want: true},
		{name: "http url no longer resolves", style: "https://cdn.example/x/" + key, want: false},
		{name: "foreign clinic key is not ours", style: "trimming-images/2/style.jpg", want: false},
		{name: "unrelated value", style: "before.jpg", want: false},
		{name: "empty", style: "", want: false},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 各行を別 clinic に入れて subtest 間を隔離する（行の clinic_id が
			// 参照チェックのスコープ。値の医院 prefix は resolver 側で検証済み）。
			clinicID := uint64(1000 + i)
			insertTrimmingDetailWithImages(t, db, clinicID, tc.style, tc.completed)
			got, err := repo.ImageKeyInUse(ctx, clinicID, key, 0)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestImageKeyInUse_ExclusionOnlySkipsThatAppointment は除外が指定 appointment の
// 行だけに効き、別レコードの同一キー参照は依然として守られることを検証する。
func TestImageKeyInUse_ExclusionOnlySkipsThatAppointment(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	const key = "trimming-images/1/shared.jpg"

	selfID := insertTrimmingDetailWithImages(t, db, clinicID, key, "")
	insertTrimmingDetailWithImages(t, db, clinicID, "", key)

	got, err := repo.ImageKeyInUse(ctx, clinicID, key, selfID)
	require.NoError(t, err)
	assert.True(t, got, "other record's reference must still count")
}

// TestImageKeyInUse_ObservesAmbientTx は repository が DBOrTx 経由で呼び出し側の
// ambient transaction を見ることを検証する（rollback 後は参照なし）。
func TestImageKeyInUse_ObservesAmbientTx(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	repo := NewAppointmentTrimmingDetailRepository(db)
	transactor := newTestTransactor(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	const key = "trimming-images/1/ambient.jpg"
	sentinel := errors.New("force rollback after reference probe")

	appointment := makeReservation(t, db, clinicID)
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(&model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appointment.ID,
			StyleImage:    key,
			BWUnit:        model.BodyWeightUnitKg,
		}).Error)
		inUse, err := repo.ImageKeyInUse(txCtx, clinicID, key, 0)
		require.NoError(t, err)
		assert.True(t, inUse, "probe must see the caller's uncommitted write")
		return sentinel
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)

	inUse, err := repo.ImageKeyInUse(ctx, clinicID, key, 0)
	require.NoError(t, err)
	assert.False(t, inUse, "rolled-back rows must not count as references")
}

// This DB regression verifies that a cleared/replaced course image remains a
// protected original even after no current course row references its key.
func TestCourseImageHistoryInUse_RetainsOriginalReferences(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	const key = "trimming-images/1/original.jpg"
	for _, tc := range []struct {
		name, resource, oldValue, newValue string
		clinicID                           uint64
		want                               bool
	}{
		{"before reference", courseImageAuditResource, `{"image_object_key":"trimming-images/1/original.jpg"}`, `{"image_object_key":null}`, 1, true},
		{"after reference", courseImageAuditResource, `{"image_object_key":null}`, `{"image_object_key":"trimming-images/1/original.jpg"}`, 1, true},
		{"legacy wrapped reference", courseImageAuditResource, `{"image_object_key":"/uploads/trimming-images/1/original.jpg"}`, `null`, 1, true},
		{"other clinic", courseImageAuditResource, `{"image_object_key":"trimming-images/1/original.jpg"}`, `null`, 2, false},
		{"unrelated resource", "unrelated", `{"image_object_key":"trimming-images/1/original.jpg"}`, `null`, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Use an isolated per-subtest rollback so one retained reference
			// cannot make a later negative case pass accidentally.
			tx := db.Begin()
			require.NoError(t, tx.Error)
			t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
			require.NoError(t, tx.Exec(`INSERT INTO audit_logs (clinic_id, actor_type, action, resource, resource_id, old_value, new_value)
				VALUES (?, 'system', 'trimming_course.image_update', ?, 10, ?::jsonb, ?::jsonb)`, tc.clinicID, tc.resource, tc.oldValue, tc.newValue).Error)
			got, err := trimmingCourseImageHistoryInUse(t.Context(), tx, 1, key)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCourseImageHistoryInUse_ReadFailureDoesNotPermitDeletion(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	got, err := trimmingCourseImageHistoryInUse(ctx, db, 1, "trimming-images/1/original.jpg")
	require.Error(t, err)
	assert.False(t, got, "caller must reject deletion on error; false is not a proof")
}
