package billing

// repo_test_helpers_test.go — internal/repository の test helper の最小限の複製
// （package 跨ぎ import 不能のため・shiftentry/reservation 先例）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func ptrU64(v uint64) *uint64 { return &v }

// makeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

// testTransactorImpl は repository.Transactor.WithTx を repohelpers ベースで再現する
// （repository import は facade 経由の循環になるため不可・reservation/prescription 先例）。
type testTransactorImpl struct{ db *gorm.DB }

func (t testTransactorImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(repohelpers.WithTxValue(ctx, tx))
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

func ptrInt64(v int64) *int64 { return &v }

func ptrFloat64(v float64) *float64 { return &v }

func ptrBool(v bool) *bool { return &v }
