package billing

// repo_test_helpers_test.go — internal/repository の test helper の最小限の複製
// （package 跨ぎ import 不能のため・shiftentry/reservation 先例）。

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func seedBillingClinicForFK(t *testing.T, db *gorm.DB, clinicID uint64) {
	t.Helper()
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.StaffClinicAssignment{},
	))

	ctx := context.Background()
	var existing model.Clinic
	err := db.WithContext(ctx).First(&existing, clinicID).Error
	if err == nil {
		return
	}
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	company := &model.Company{
		Name: fmt.Sprintf("billing fixture company %d", clinicID),
	}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)
	clinic := &model.Clinic{
		ID:        clinicID,
		CompanyID: company.ID,
		Name:      fmt.Sprintf("billing fixture clinic %d", clinicID),
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(ctx).Create(clinic).Error)
}

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func ptrU64(v uint64) *uint64 { return &v }

// testTransactorImpl は repository.Transactor.WithTx を repohelpers ベースで再現する
// （repository import は facade 経由の循環になるため不可・reservation/prescription 先例）。
type testTransactorImpl struct{ db *gorm.DB }

func (t testTransactorImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

func testNewTransactor(db *gorm.DB) Transactor { return testTransactorImpl{db: db} }

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

func ptrUint64(v uint64) *uint64 { return &v }

func ptrFloat64(v float64) *float64 { return &v }

func ptrBool(v bool) *bool { return &v }
