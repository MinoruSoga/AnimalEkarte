package repository

// be9_2c_r3_test_helper_carrier_test.go — R③でreservation_clinic_isolation_test.goが
// internal/reservationへ移動した際、同居していた共有helperの残留consumer
// （accounting_complete_appointments_test.go等）向けcarrier複製。当該domain移行時に解消。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// makeReservationType はテスト用予約区分を1件作成する。
// ekarte_db_test は本番 migration 適用済みのため reservation_types テーブルが存在する。
// clinic_id・name のみ DB default がない必須フィールド。category は ENUM のため明示的に設定する。
func makeReservationType(t *testing.T, db *gorm.DB, clinicID uint64) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "テスト診療区分",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

// parseJSTDate は reservation.ParseJSTDate への delegate（appointment 系残留 test 用・R④で解消）。
func parseJSTDate(value string) (time.Time, error) {
	return reservation.ParseJSTDate(value)
}

// appointmentDayRange は reservation.AppointmentDayRange への delegate（同上・R④で解消）。
func appointmentDayRange(date time.Time) (start, end time.Time) {
	return reservation.AppointmentDayRange(date)
}

// setupReservationIsolationTestDB — R③移動元と同一実装のcarrier複製（count_clinic_scope_isolation用）。
func setupReservationIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

// makeReservation はテスト用予約を1件作成する。
// 先に makeReservationType で有効な区分を用意し、FK 制約を満たす。
func makeReservation(t *testing.T, db *gorm.DB, clinicID uint64) *model.Reservation {
	t.Helper()
	rt := makeReservationType(t, db, clinicID)
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: rt.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}
