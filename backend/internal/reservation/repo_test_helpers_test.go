package reservation

// repo_test_helpers_test.go — internal/repository の isolation_test_helpers_test.go 等からの
// 最小限の複製（BE9-2C R③: package 跨ぎで test helper import 不能のため。shiftentry 先例）。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var _ = time.Now // keep time import if unused by some subset

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

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

// makeShiftEntry はテスト用シフトエントリを1件作成する。
func makeShiftEntry(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time) {
	t.Helper()
	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(entry).Error)
}

// makeLineCustomerForAdmin はテスト用 LineCustomer を作成して返す。
func makeLineCustomerForAdmin(t *testing.T, db *gorm.DB, clinicID uint64, lineUserID string) *model.LineCustomer {
	t.Helper()
	lc := &model.LineCustomer{
		ClinicID:         clinicID,
		LineUserID:       lineUserID,
		AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(lc).Error)
	return lc
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

// testTransactorImpl は production Transactor.WithTx を persistence kernel で再現する。
// （repository import は facade 経由の循環になるため不可・prescription 先例）。
type testTransactorImpl struct{ db *gorm.DB }

func (t testTransactorImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

func testNewTransactor(db *gorm.DB) Transactor { return testTransactorImpl{db: db} }
