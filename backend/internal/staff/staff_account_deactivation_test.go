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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// ---- PR#551: 共有アカウントの無効化ガード + 無効化の durable 監査 ----
//
// レビュー指摘 (4180445703): 同じログインアカウントを共有する別の active スタッフ
// 行が残っている間はアカウントを無効化してはならない（共有行のログインを奪う）。
// レビュー指摘 (4178809086): アカウント無効化は変更と同一トランザクション内で
// audit_logs に永続化されなければならない（監査失敗は全体を rollback する）。

// makeSharedAccountStaff は既存アカウントを参照する active スタッフ行を追加する。
func makeSharedAccountStaff(
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

// findDeactivationAudits は対象アカウントの無効化監査行を返す。
func findDeactivationAudits(t *testing.T, db *gorm.DB, accountID uint64) []model.AuditLog {
	t.Helper()
	var entries []model.AuditLog
	require.NoError(t, db.WithContext(context.Background()).
		Where("action = ?", "staff.account.deactivate").
		Where("resource_id = ?", accountID).
		Find(&entries).Error)
	return entries
}

func TestStaffRepository_Update_DeactivateRefusesSharedAccountAcrossClinics(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	acc, staffA := makeAccountLinkedStaff(t, db, clinicA, "shared-across@example.com")
	staffB := makeSharedAccountStaff(t, db, clinicB, acc.ID, "別医院の共有スタッフ")

	isActive := false
	err := repo.Update(ctx, clinicA, staffA.ID, UpdateStaffInput{IsActive: &isActive})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err),
		"共有アカウントの無効化拒否は conflict 型エラーでなければならない")

	assert.True(t, staffRowIsActive(t, db, staffA.ID),
		"拒否された場合は対象スタッフの is_active も rollback される")
	assert.True(t, staffRowIsActive(t, db, staffB.ID))
	assert.True(t, accountIsActive(t, db, acc.ID),
		"別の active スタッフが使用中のアカウントは無効化されてはならない")
	assert.Empty(t, findDeactivationAudits(t, db, acc.ID),
		"拒否された deactivation は監査行を残してはならない")
}

func TestStaffRepository_Update_DeactivateSharedAccountAllowedWhenOtherRowInactive(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	acc, staffA := makeAccountLinkedStaff(t, db, clinicA, "shared-inactive@example.com")
	other := makeSharedAccountStaff(t, db, clinicB, acc.ID, "無効化済みの共有スタッフ")
	require.NoError(t, db.Model(&model.Staff{}).
		Where("id = ?", other.ID).Update("is_active", false).Error)

	isActive := false
	require.NoError(t, repo.Update(ctx, clinicA, staffA.ID, UpdateStaffInput{IsActive: &isActive}))

	assert.False(t, staffRowIsActive(t, db, staffA.ID))
	assert.False(t, accountIsActive(t, db, acc.ID),
		"共有行が inactive なら最後の利用者の無効化でアカウントも無効化される")
}

func TestStaffRepository_Update_DeactivateSharedAccountAllowedWhenOtherRowDeleted(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	acc, staffA := makeAccountLinkedStaff(t, db, clinicA, "shared-deleted@example.com")
	other := makeSharedAccountStaff(t, db, clinicB, acc.ID, "削除済みの共有スタッフ")
	require.NoError(t, db.WithContext(ctx).Delete(other).Error)
	require.True(t, staffRowIsDeleted(t, db, other.ID))

	isActive := false
	require.NoError(t, repo.Update(ctx, clinicA, staffA.ID, UpdateStaffInput{IsActive: &isActive}))

	assert.False(t, staffRowIsActive(t, db, staffA.ID))
	assert.False(t, accountIsActive(t, db, acc.ID),
		"共有行が削除済みならアカウントは無効化できる")
}

func TestStaffRepository_UpdateForReservation_DeactivateRefusesSharedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	acc, staffA := makeAccountLinkedStaff(t, db, clinicA, "shared-reservation@example.com")
	makeSharedAccountStaff(t, db, clinicB, acc.ID, "共有スタッフB")

	isActive := false
	err := repo.UpdateForReservation(
		ctx, clinicA, staffA.ID, ReservationStaffUpdate{IsActive: &isActive},
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))

	assert.True(t, staffRowIsActive(t, db, staffA.ID))
	assert.True(t, accountIsActive(t, db, acc.ID),
		"予約経路でも共有アカウントは無効化されてはならない")
}

func TestStaffRepository_Delete_RefusesSharedAccount(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	acc, staffA := makeAccountLinkedStaff(t, db, clinicA, "shared-delete@example.com")
	makeSharedAccountStaff(t, db, clinicB, acc.ID, "共有スタッフB")

	err := repo.Delete(ctx, clinicA, staffA.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))

	assert.False(t, staffRowIsDeleted(t, db, staffA.ID),
		"共有アカウント保護で拒否された場合、staffs の soft delete も rollback される")
	assert.True(t, accountIsActive(t, db, acc.ID))
}

func TestStaffRepository_Update_DeactivateWritesDurableAudit(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "audit@example.com")
	actor := makeDoctor(t, db, clinicID, "監査実行者")
	isActive := false

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{
		IsActive:     &isActive,
		ActorStaffID: actor.ID,
	}))

	entries := findDeactivationAudits(t, db, acc.ID)
	require.Len(t, entries, 1, "アカウント無効化は audit_logs に1行永続化される")
	entry := entries[0]
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, clinicID, *entry.ClinicID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, actor.ID, *entry.ActorID)
	assert.Equal(t, model.AuditResourceAccount, entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, acc.ID, *entry.ResourceID)
	assert.Contains(t, string(entry.NewValue), "staff_id")
}

func TestStaffRepository_Update_DeactivateAuditFailureRollsBackMutation(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "audit-fail@example.com")

	// 監査 insert を決定的に失敗させる CHECK 制約を一時的に張る（テスト終了時に除去）。
	// 対象アカウントの行だけを弾く条件にして、共有テスト DB 上の他テストが
	// 別アカウントの無効化監査を書く経路を巻き込まないようにする。
	const constraint = "ck_test_block_deactivation_audit"
	require.NoError(t, db.Exec(
		"ALTER TABLE audit_logs ADD CONSTRAINT "+constraint+
			" CHECK (action <> 'staff.account.deactivate' OR resource_id IS DISTINCT FROM "+
			strconv.FormatUint(acc.ID, 10)+")",
	).Error)
	t.Cleanup(func() {
		_ = db.Exec("ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS " + constraint)
	})

	isActive := false
	err := repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive})
	require.Error(t, err, "監査書き込み失敗は deactivation 全体を失敗させる")

	assert.True(t, staffRowIsActive(t, db, staff.ID),
		"監査失敗時に staffs.is_active=false は rollback されなければならない")
	assert.True(t, accountIsActive(t, db, acc.ID),
		"監査失敗時に accounts.is_active=false も rollback されなければならない")
}

func TestStaffRepository_Update_DeactivateAuditRollsBackWithAmbientTransaction(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "audit-rollback@example.com")
	isActive := false
	sentinel := errors.New("abort after deactivation audit")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.Update(txCtx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	assert.True(t, staffRowIsActive(t, db, staff.ID))
	assert.True(t, accountIsActive(t, db, acc.ID))
	assert.Empty(t, findDeactivationAudits(t, db, acc.ID),
		"ambient tx の rollback で監査行も残ってはならない")
}

func TestStaffRepository_Delete_WritesDeactivationAudit(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc, staff := makeAccountLinkedStaff(t, db, clinicID, "delete-audit@example.com")

	require.NoError(t, repo.Delete(ctx, clinicID, staff.ID))

	entries := findDeactivationAudits(t, db, acc.ID)
	require.Len(t, entries, 1,
		"削除経路のアカウント無効化も audit_logs に永続化される")
	entry := entries[0]
	assert.Equal(t, model.AuditResourceAccount, entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, acc.ID, *entry.ResourceID)
	// Delete は actor を受け取らない経路のため system actor を記録する。
	assert.Equal(t, model.AuditActorTypeSystem, entry.ActorType)
	assert.Nil(t, entry.ActorID)
}

func TestStaffRepository_Update_DeactivateWithoutAccountWritesNoAudit(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	staff := makeDoctor(t, db, clinicID, "アカウント未紐付けスタッフ(監査)")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	isActive := false

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, UpdateStaffInput{IsActive: &isActive}))
	assert.False(t, staffRowIsActive(t, db, staff.ID))

	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("action = ?", "staff.account.deactivate").
		Count(&auditCount).Error)
	assert.Zero(t, auditCount, "アカウント未紐付けの無効化は監査行を書かない")
}
