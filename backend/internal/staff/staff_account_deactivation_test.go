package staff_test

// staff_account_deactivation_test.go — EMR-289: スタッフ無効化/削除時にリンク済み
// ログインアカウント（accounts.is_active）を同一トランザクションで無効化する
// 不変条件の統合テスト。対象は staffs の lint 強制 write funnel
// （staffRepository.Update / UpdateForReservation / Delete）の全無効化経路と、
// staffService 経由の管理画面経路。再活性化が account を復活させない非対称
// セマンティクスもここで固定する。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
)

// makeAccountLinkedStaff は is_active=true の account と、それを紐付けた
// is_active=true のスタッフ + clinic assignment を作成する。
func makeAccountLinkedStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	email string,
) (*model.Account, *model.Staff) {
	t.Helper()
	acc := makeStaffAccount(t, db, email)
	staff := &model.Staff{
		ClinicID:  clinicID,
		Name:      "アカウント紐付きスタッフ",
		StaffType: model.StaffTypeDoctor,
		AccountID: &acc.ID,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	return acc, staff
}

func accountIsActive(t *testing.T, db *gorm.DB, accountID uint64) bool {
	t.Helper()
	var acc model.Account
	require.NoError(t, db.WithContext(context.Background()).Unscoped().
		First(&acc, accountID).Error)
	return acc.IsActive
}

func staffRowIsActive(t *testing.T, db *gorm.DB, staffID uint64) bool {
	t.Helper()
	var staff model.Staff
	require.NoError(t, db.WithContext(context.Background()).Unscoped().
		First(&staff, staffID).Error)
	return staff.IsActive
}

func staffRowIsDeleted(t *testing.T, db *gorm.DB, staffID uint64) bool {
	t.Helper()
	var staff model.Staff
	require.NoError(t, db.WithContext(context.Background()).Unscoped().
		First(&staff, staffID).Error)
	return staff.DeletedAt.Valid
}

// ---- Update（管理画面 PATCH）----

func TestStaffRepository_Update_DeactivateDisablesLinkedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "deactivate@example.com")
	isActive := false

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))

	assert.False(t, staffRowIsActive(t, db, staff.ID), "staffs.is_active が false になる")
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: スタッフ無効化と同一トランザクションで accounts.is_active も false になる")
}

func TestStaffRepository_Update_DeactivateRollsBackWithAmbientTransaction(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "rollback@example.com")
	isActive := false
	sentinel := errors.New("abort after staff update")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.Update(txCtx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	assert.True(t, staffRowIsActive(t, db, staff.ID),
		"ambient tx rollback で staffs.is_active は true のまま")
	assert.True(t, accountIsActive(t, db, acc.ID),
		"account 無効化も ambient tx と一緒に rollback される（部分コミットしない）")
}

func TestStaffRepository_Update_ReactivateDoesNotReviveAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "reactivate@example.com")
	// 事前状態: staff/account ともに無効（EMR-289 で観測された一貫状態）。
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", staff.ID).Update("is_active", false).Error)
	require.NoError(t, db.Model(&model.Account{}).Where("id = ?", acc.ID).Update("is_active", false).Error)

	isActive := true
	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))

	assert.True(t, staffRowIsActive(t, db, staff.ID), "staffs.is_active は true になる")
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: スタッフ再有効化は account を復活させない（account 再有効化は別操作）")
}

func TestStaffRepository_Update_DeactivateStaffWithoutAccountSucceeds(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	staff := makeDoctor(t, db, clinicID, "アカウント未紐付けスタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	isActive := false

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))
	assert.False(t, staffRowIsActive(t, db, staff.ID))
}

func TestStaffRepository_Update_ProfileOnlyLeavesAccountActive(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "profile-only@example.com")
	name := "氏名変更のみ"

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{Name: &name}))

	assert.True(t, accountIsActive(t, db, acc.ID),
		"is_active 非対象の更新は account に触れない")
}

// ---- UpdateForReservation（予約画面 PatchStatus）----

func TestStaffRepository_UpdateForReservation_DeactivateDisablesLinkedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "reservation-deactivate@example.com")
	isActive := false

	require.NoError(t, repo.UpdateForReservation(
		ctx, clinicID, staff.ID, ReservationStaffUpdate{IsActive: &isActive},
	))

	assert.False(t, staffRowIsActive(t, db, staff.ID))
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: 予約経路の無効化も同一トランザクションで account を無効化する")
}

func TestStaffRepository_UpdateForReservation_ReactivateDoesNotReviveAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "reservation-reactivate@example.com")
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", staff.ID).Update("is_active", false).Error)
	require.NoError(t, db.Model(&model.Account{}).Where("id = ?", acc.ID).Update("is_active", false).Error)

	isActive := true
	require.NoError(t, repo.UpdateForReservation(
		ctx, clinicID, staff.ID, ReservationStaffUpdate{IsActive: &isActive},
	))

	assert.True(t, staffRowIsActive(t, db, staff.ID))
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: 予約経路の再有効化も account を復活させない")
}

// ---- Delete（管理画面 / 予約経由は service 経由で同一 repo.Delete を呼ぶ）----

func TestStaffRepository_Delete_DisablesLinkedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "delete@example.com")

	require.NoError(t, repo.Delete(ctx, clinicID, staff.ID))

	assert.True(t, staffRowIsDeleted(t, db, staff.ID), "staffs は soft delete される")
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: スタッフ削除と同一トランザクションで accounts.is_active も false になる")
}

func TestStaffRepository_Delete_RollsBackWithAmbientTransaction(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "delete-rollback@example.com")
	sentinel := errors.New("abort after staff delete")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.Delete(txCtx, clinicID, staff.ID))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	assert.False(t, staffRowIsDeleted(t, db, staff.ID),
		"ambient tx rollback で staffs は未削除のまま")
	assert.True(t, accountIsActive(t, db, acc.ID),
		"account 無効化も ambient tx と一緒に rollback される")
}

// ---- staffService 経由（管理画面 Update）の wiring 確認 ----

func TestStaffService_Update_DeactivateDisablesLinkedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "service-deactivate@example.com")
	actor := makeDoctor(t, db, clinicID, "操作者スタッフ")

	svc := NewService(
		NewRepository(db),
		nil,
		NewStaffClinicAssignmentRepository(db),
		nil, nil, nil, nil, nil, nil,
		persistence.NewTransactor(db),
	)
	isActive := false

	updated, err := svc.Update(ctx, clinicID, staff.ID, &UpdateStaffInput{
		IsActive:            &isActive,
		AuthorizedClinicIDs: []uint64{clinicID},
		ActorStaffID:        actor.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.False(t, updated.IsActive)
	assert.False(t, accountIsActive(t, db, acc.ID),
		"EMR-289: 管理画面 Update 経路でも account は同一トランザクションで無効化される")
}
