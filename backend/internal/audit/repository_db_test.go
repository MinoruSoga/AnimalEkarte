package audit

// audit_repository_test.go — AuditRepository の統合テスト（内部カバレッジ向上）。
//
// 対象: Create / CreateTx
// 検証観点: 正常系（永続化される）、CreateTx の ambient tx 参加（#211: 同一 tx の rollback で
//           監査ログも巻き戻る）。
//
// ambient tx 不在時の fail-closed contract と MarshalJSON の純粋関数 contract は
// internal/audit の unit tests が直接検証するため、legacy facade 経由の重複 test は削除した。

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

// setupAuditTestDB は audit_logs テーブルを整備する（依存する他テーブルなし）。
//
// X-3 (audit-ip-inet-model-drift): 従来は ensureAutoMigrated(db, &model.AuditLog{}) で
// audit_logs.ip_address(実DDL=inet) から `text` カラムを生成していたため、本番で発生する
// `”::inet` の 22P02 エラーをこのテストスイートが検出できなかった。setupAuditRealDDLTestDB
// （audit_real_ddl_test.go）に委譲し、001_init.sql の実 DDL でテーブルを再作成する。
func setupAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupAuditRealDDLTestDB(t)
}

func uint64Ptr(v uint64) *uint64 { return &v }

// TestAuditRepository_Create は監査ログの正常系永続化を検証する。
func TestAuditRepository_Create(t *testing.T) {
	db := setupAuditTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("clinic_id / actor_id ありで作成できる", func(t *testing.T) {
		log := &model.AuditLog{
			ClinicID:  uint64Ptr(1),
			ActorID:   uint64Ptr(10),
			ActorType: model.AuditActorTypeStaff,
			Action:    model.AuditActionAuthLoginSuccess,
			Resource:  "staff",
		}
		require.NoError(t, repo.Create(ctx, log))
		assert.NotZero(t, log.ID)

		var got model.AuditLog
		require.NoError(t, db.WithContext(ctx).First(&got, log.ID).Error)
		assert.Equal(t, model.AuditActionAuthLoginSuccess, got.Action)
	})

	t.Run("actor_id が nil のシステム操作でも作成できる（clinic_id は必須）", func(t *testing.T) {
		// X-3: audit_logs.clinic_id は実DDLで NOT NULL。actor_id=nil（システム操作）は
		// actor_consistency_check(actor_type='system' AND actor_id IS NULL) を満たすため許容される。
		// clinic_id=nil を許容しないことは TestAuditLogRealDDL_NilClinicIDFails（audit_real_ddl_test.go）
		// が実DDLレベルで検証する。
		log := &model.AuditLog{
			ClinicID:  uint64Ptr(1),
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionLstepTagSyncBulk,
			Resource:  "lstep",
		}
		require.NoError(t, repo.Create(ctx, log))
		assert.NotZero(t, log.ID)
		assert.Nil(t, log.ActorID)
	})
}

// TestAuditRepository_CreateTx_AmbientTxCommit は Transactor.WithTx 内から CreateTx を呼ぶと
// 同一トランザクションに参加し、正常終了（コミット）で永続化されることを検証する（#211）。
func TestAuditRepository_CreateTx_AmbientTxCommit(t *testing.T) {
	db := setupAuditTestDB(t)
	repo := NewRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()

	var createdID uint64
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		log := &model.AuditLog{
			ClinicID:  uint64Ptr(1),
			ActorID:   uint64Ptr(10),
			ActorType: model.AuditActorTypeStaff,
			Action:    model.AuditActionCheckupFieldResultReplace,
			Resource:  model.AuditResourceCheckupFieldResult,
		}
		if err := repo.CreateTx(txCtx, log); err != nil {
			return err
		}
		createdID = log.ID
		return nil
	})
	require.NoError(t, err)
	require.NotZero(t, createdID)

	var count int64
	require.NoError(t, db.WithContext(ctx).Model(&model.AuditLog{}).Where("id = ?", createdID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "コミットされたトランザクション内の監査ログは永続化される")
}

// TestAuditRepository_CreateTx_AmbientTxRollback は ambient tx 内で後続処理が失敗した場合、
// CreateTx で書き込んだ監査ログも tx のロールバックに巻き込まれて消えることを検証する（#211 の核心）。
func TestAuditRepository_CreateTx_AmbientTxRollback(t *testing.T) {
	db := setupAuditTestDB(t)
	repo := NewRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()

	var createdID uint64
	forcedErr := errors.New("simulated downstream failure after audit write")
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		log := &model.AuditLog{
			ClinicID:  uint64Ptr(1),
			ActorID:   uint64Ptr(10),
			ActorType: model.AuditActorTypeStaff,
			Action:    model.AuditActionBillingCancel,
			Resource:  "billing",
		}
		if err := repo.CreateTx(txCtx, log); err != nil {
			return err
		}
		createdID = log.ID
		return forcedErr
	})
	require.Error(t, err)
	require.NotZero(t, createdID, "ID はメモリ上には採番されるが、コミットはされない")

	var count int64
	require.NoError(t, db.WithContext(ctx).Model(&model.AuditLog{}).Where("id = ?", createdID).Count(&count).Error)
	assert.Equal(t, int64(0), count, "ロールバックされた tx 内の監査ログ書込は巻き戻る")
}
