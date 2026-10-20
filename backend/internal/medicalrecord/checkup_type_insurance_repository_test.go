package medicalrecord

// EMR-456: checkup_types.is_non_insurance / version の実 DB 往復（PR727 の nullable 列を再利用）。
// 旧欠落 NULL を別項目更新で補完しない・版不一致と別医院は書き換えない・service の tx に参加する。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestCheckupTypeRepository_InsuranceSettingRoundTrip(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("create persists explicit value and starts at version 1", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicA, Name: "新規健診", IsActive: true, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		got, err := repo.FindByID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		require.NotNil(t, got.IsNonInsurance)
		assert.True(t, *got.IsNonInsurance)
		assert.Equal(t, int64(1), got.Version)
	})

	t.Run("inactive create returns the storage version so an immediate edit succeeds", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicA, Name: "無効健診", IsActive: false, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		var stored model.CheckupType
		require.NoError(t, db.First(&stored, ct.ID).Error)
		assert.False(t, stored.IsActive)
		assert.Equal(t, stored.Version, ct.Version, "returned version must match the row after the is_active follow-up update")

		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(false), expectedVersion: int64Ptr(ct.Version),
		})
		require.NoError(t, err)
		assert.False(t, *got.IsNonInsurance)
	})

	t.Run("legacy NULL survives unrelated update and explicit false is kept", func(t *testing.T) {
		legacy := &model.CheckupType{ClinicID: clinicA, Name: "旧健診", IsActive: true}
		require.NoError(t, db.WithContext(ctx).Create(legacy).Error)
		require.NoError(t, db.Exec("UPDATE checkup_types SET is_non_insurance = NULL WHERE id = ?", legacy.ID).Error)

		name := "旧健診（改）"
		got, err := repo.Update(ctx, clinicA, legacy.ID, UpdateCheckupTypeInput{
			Name: &name, expectedVersion: int64Ptr(1),
		})
		require.NoError(t, err)
		assert.Nil(t, got.IsNonInsurance, "unrelated update must not fill a legacy missing value")
		assert.Equal(t, int64(2), got.Version)

		got, err = repo.Update(ctx, clinicA, legacy.ID, UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(false), expectedVersion: int64Ptr(2),
		})
		require.NoError(t, err)
		require.NotNil(t, got.IsNonInsurance)
		assert.False(t, *got.IsNonInsurance)
		assert.Equal(t, int64(3), got.Version)
	})

	t.Run("same-value re-save succeeds and keeps the storage version", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicA, Name: "再保存健診", IsActive: true, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(true), expectedVersion: int64Ptr(1),
		})
		require.NoError(t, err, "entity_version_guard must not reject an unchanged save")
		assert.True(t, *got.IsNonInsurance)
		assert.Equal(t, int64(1), got.Version)
	})

	t.Run("stale version is a conflict and writes nothing", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicA, Name: "競合健診", IsActive: true, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(false), expectedVersion: int64Ptr(7),
		})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err))
		reread, err := repo.FindByID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.True(t, *reread.IsNonInsurance)
		assert.Equal(t, int64(1), reread.Version)
	})

	t.Run("other clinic is not found and writes nothing", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicB, Name: "B院健診", IsActive: true, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(false), expectedVersion: int64Ptr(1),
		})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
		reread, err := repo.FindByID(ctx, clinicB, ct.ID)
		require.NoError(t, err)
		assert.True(t, *reread.IsNonInsurance)
		assert.Equal(t, int64(1), reread.Version)
	})

	t.Run("update joins the ambient transaction and rolls back with it", func(t *testing.T) {
		ct := &model.CheckupType{ClinicID: clinicA, Name: "tx健診", IsActive: true, IsNonInsurance: boolPtr(true)}
		require.NoError(t, repo.Create(ctx, ct))
		rollback := errors.New("rollback")
		err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.Update(txCtx, clinicA, ct.ID, UpdateCheckupTypeInput{
				IsNonInsurance: boolPtr(false), expectedVersion: int64Ptr(1),
			}); err != nil {
				return err
			}
			return rollback
		})
		require.ErrorIs(t, err, rollback)
		reread, err := repo.FindByID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.True(t, *reread.IsNonInsurance)
		assert.Equal(t, int64(1), reread.Version)
	})
}

// TestCheckupTypeService_InsuranceAuditFailureRollsBackRealDB は実 DB の transactor で、
// 保険設定の監査書込が失敗したとき業務行（作成・更新）も残らないことを確認する。
func TestCheckupTypeService_InsuranceAuditFailureRollsBackRealDB(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	tx := persistence.NewTransactor(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	auditDown := &recordingCheckupTypeAudit{err: errors.New("audit down")}

	t.Run("create is not persisted", func(t *testing.T) {
		svc := NewCheckupTypeService(repo, tx, auditDown)
		got, err := svc.Create(ctx, clinicID, &CreateCheckupTypeInput{
			Name: "監査失敗健診", IsActive: true, ActorID: &testCheckupTypeActorID,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		var count int64
		require.NoError(t, db.Model(&model.CheckupType{}).Where("name = ?", "監査失敗健診").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("insurance change is rolled back", func(t *testing.T) {
		ok := NewCheckupTypeService(repo, tx, &recordingCheckupTypeAudit{})
		created, err := ok.Create(ctx, clinicID, &CreateCheckupTypeInput{
			Name: "監査ロールバック健診", IsActive: true, ActorID: &testCheckupTypeActorID,
		})
		require.NoError(t, err)

		svc := NewCheckupTypeService(repo, tx, auditDown)
		got, err := svc.Update(ctx, clinicID, created.ID, &UpdateCheckupTypeInput{
			IsNonInsurance: boolPtr(false), Version: int64Ptr(created.Version), ActorID: &testCheckupTypeActorID,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		reread, err := repo.FindByID(ctx, clinicID, created.ID)
		require.NoError(t, err)
		assert.True(t, *reread.IsNonInsurance)
		assert.Equal(t, created.Version, reread.Version)
	})
}
