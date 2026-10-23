package reservation

// liff_attach_savepoint_realdb_test.go — EMR-491 D6 review A-1 の real DB 回帰テスト（NOT_RUN: 実行環境未検証）。
//
// 死亡 reader が SQL レベルで失敗すると PostgreSQL は共有 tx 全体を中止状態にする
// （current transaction is aborted）。tryAttachReservationOwnerPet は reader の直前に savepoint を張り、
// reader エラー時に ROLLBACK TO SAVEPOINT で tx を回復させてから owner_id の attach を続行する。
// savepoint が無いと、後続の attach update が失敗して owner_id も付与されない。
//
// 注: テスト DB は AutoMigrate の double。savepoint / tx 中止は PostgreSQL 本体の挙動で、
// double でも本番と同じである。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// abortingLifeStateStore は実 ReservationStore の ReadPetLifeState だけを差し替え、ambient tx 上で
// 失敗する SQL を発行して tx を中止状態にしたうえでそのエラーを返す（lock_timeout 等の再現）。
type abortingLifeStateStore struct {
	ReservationStore
}

func (abortingLifeStateStore) ReadPetLifeState(ctx context.Context, _, _ uint64) (sharedkernel.PetLifeState, error) {
	return "", persistence.TxFromContext(ctx).WithContext(ctx).Exec("SELECT 1/0").Error
}

func TestLiffService_tryAttachReservationOwnerPet_ReaderSQLFailureStillAttachesOwner_RealDB(t *testing.T) {
	db := setupLiffPetPairTestDB(t)
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "savepoint飼主")
	appt := makeAdminReservationAt(t, db, clinicA, time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC), nil, nil, nil, nil)

	svc := &liffService{
		tx:              testNewTransactor(db),
		reservationRepo: abortingLifeStateStore{ReservationStore: NewReservationRepository(db)},
		customerRepo: &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:    1,
					Owner: &model.Owner{ID: owner.ID, ClinicID: clinicA, Pets: []model.Pet{{ID: 300}}},
				}, nil
			},
		},
	}

	svc.tryAttachReservationOwnerPet(context.Background(), clinicA, 1, appt, nil)

	var reloaded model.Reservation
	require.NoError(t, db.First(&reloaded, appt.ID).Error)
	if assert.NotNil(t, reloaded.OwnerID, "owner_id must be attached even though the life-state read aborted the transaction") {
		assert.Equal(t, owner.ID, *reloaded.OwnerID)
	}
	assert.Nil(t, reloaded.PetID, "pet_id must stay unset when the life state could not be read")
}
