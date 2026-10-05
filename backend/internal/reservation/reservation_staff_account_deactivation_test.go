package reservation

// reservation_staff_account_deactivation_test.go — PR#551 レビュー指摘の
// 予約経路カバレッジ: PatchStatus → repo.Update → staff.UpdateForReservation
// の deactivation が (4180445703) 共有アカウントを別の active スタッフ行が
// 使用中なら拒否し、(4178809086) 成功時の無効化を同一 tx の audit_logs に
// 実行者 actor で永続化することを service レベルで検証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationStaffDeactivationTestDB は共有スキーマに加えて audit_logs を
// 確実に存在させる（アカウント無効化は同一 tx で audit_logs に書く）。
func setupReservationStaffDeactivationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupReservationStaffTxAtomicityTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	return db
}

// makeReservationAccountLinkedStaff は is_active=true のスタッフ行を既存
// アカウントに紐付けて作成する（reservation 側 fixture は account を持たないため）。
func makeReservationAccountLinkedStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID, accountID uint64,
	name string,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		StaffType: model.StaffTypeDoctor,
		AccountID: &accountID,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	return staff
}

func reservationStaffDeactivationService(db *gorm.DB) ReservationStaffService {
	repo := NewReservationStaffRepository(db, staffpkg.NewRepository(db))
	return NewReservationStaffService(repo, testNewTransactor(db), nil, staffpkg.NewRepository(db))
}

func reservationStaffIsActive(t *testing.T, db *gorm.DB, staffID uint64) bool {
	t.Helper()
	var staff model.Staff
	require.NoError(t, db.WithContext(context.Background()).Unscoped().
		First(&staff, staffID).Error)
	return staff.IsActive
}

func reservationAccountIsActive(t *testing.T, db *gorm.DB, accountID uint64) bool {
	t.Helper()
	var account model.Account
	require.NoError(t, db.WithContext(context.Background()).Unscoped().
		First(&account, accountID).Error)
	return account.IsActive
}

func TestReservationStaffService_PatchStatus_RefusesSharedAccountDeactivation(t *testing.T) {
	db := setupReservationStaffDeactivationTestDB(t)
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	account := &model.Account{Email: "shared-reservation-svc@example.com", PasswordHash: "hash"}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)
	target := makeReservationAccountLinkedStaff(t, db, clinicA, account.ID, "予約対象スタッフ")
	makeReservationAccountLinkedStaff(t, db, clinicB, account.ID, "別医院の共有スタッフ")
	actor := makeDoctorAssignedToClinic(t, db, clinicA, "操作者スタッフ")

	svc := reservationStaffDeactivationService(db)
	updated, excluded, err := svc.PatchStatus(
		context.Background(), clinicA, target.ID, false, actor.ID,
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err),
		"共有アカウントの無効化拒否は conflict 型エラーでなければならない")
	assert.Nil(t, updated)
	assert.Nil(t, excluded)

	assert.True(t, reservationStaffIsActive(t, db, target.ID),
		"拒否された場合は対象スタッフの is_active も rollback される")
	assert.True(t, reservationAccountIsActive(t, db, account.ID),
		"別の active スタッフが使用中のアカウントは無効化されてはならない")
}

func TestReservationStaffService_PatchStatus_DeactivationAuditCarriesActor(t *testing.T) {
	db := setupReservationStaffDeactivationTestDB(t)
	const clinicA = uint64(1)
	seedClinicsForFK(t, db, clinicA)

	account := &model.Account{Email: "audit-reservation-svc@example.com", PasswordHash: "hash"}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)
	target := makeReservationAccountLinkedStaff(t, db, clinicA, account.ID, "監査対象スタッフ")
	actor := makeDoctorAssignedToClinic(t, db, clinicA, "監査操作者")

	svc := reservationStaffDeactivationService(db)
	updated, _, err := svc.PatchStatus(
		context.Background(), clinicA, target.ID, false, actor.ID,
	)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.False(t, updated.IsActive)
	assert.False(t, reservationAccountIsActive(t, db, account.ID))

	var entries []model.AuditLog
	require.NoError(t, db.WithContext(context.Background()).
		Where("action = ?", "staff.account.deactivate").
		Where("resource_id = ?", account.ID).
		Find(&entries).Error)
	require.Len(t, entries, 1, "アカウント無効化は audit_logs に1行永続化される")
	entry := entries[0]
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, actor.ID, *entry.ActorID,
		"PatchStatus の actorStaffID が監査行に引き継がれる")
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, clinicA, *entry.ClinicID)
}
