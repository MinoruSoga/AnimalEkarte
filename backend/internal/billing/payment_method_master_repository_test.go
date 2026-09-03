package billing

// payment_method_master_repository_test.go
// payment_method_master_repository.go の実 DB 結合テスト。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupPaymentMethodMasterRepoTestDB は payment_method_master_repository のテストに必要なテーブルを整備する。
// CountUsageByPaymentMethodID は payments を billings 経由でJOINするため両方 migrate する。
func setupPaymentMethodMasterRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.PaymentMethodMaster{}, &model.Billing{}, &model.Payment{},
	))
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE payment_methods CASCADE")
	return db
}

func makePaymentMethodMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.PaymentMethodMaster {
	t.Helper()
	// SystemKey must match Payment.Method for the TASK-ADR003 DB boundary check.
	systemKey := string(model.PaymentMethodCash)
	m := &model.PaymentMethodMaster{
		ClinicID: clinicID, Name: name, IsActive: true, SystemKey: &systemKey,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makeCustomPaymentMethodMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.PaymentMethodMaster {
	t.Helper()
	m := &model.PaymentMethodMaster{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makePaymentMethodBilling(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{ClinicID: clinicID, ScheduledDate: time.Now(), Status: model.BillingStatusWaiting}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

func makePaymentForBilling(t *testing.T, db *gorm.DB, billingID, paymentMethodID uint64) *model.Payment {
	t.Helper()
	pmID := paymentMethodID
	// clinic_id + payment_method system_key alignment required by ADR-003 check constraint.
	var clinicID uint64
	require.NoError(t, db.WithContext(context.Background()).
		Model(&model.Billing{}).Select("clinic_id").Where("id = ?", billingID).Scan(&clinicID).Error)
	var systemKey *string
	require.NoError(t, db.WithContext(context.Background()).
		Model(&model.PaymentMethodMaster{}).Select("system_key").Where("id = ?", paymentMethodID).Scan(&systemKey).Error)
	method := model.PaymentMethodCash
	if systemKey != nil && *systemKey != "" {
		method = model.PaymentMethod(*systemKey)
	}
	p := &model.Payment{
		ClinicID: clinicID, BillingID: billingID, PaymentMethodID: &pmID,
		Method: method, TotalAmount: 1000, BillingAmount: 1000,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func TestPaymentMethodMasterRepository_Create_FindByID_LegacyCoverage(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("作成した支払方法を同一クリニックで取得できる", func(t *testing.T) {
		m := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "現金", IsActive: true}
		created, err := repo.Create(ctx, m)
		require.NoError(t, err)
		require.NotZero(t, created.ID)

		got, err := repo.FindByID(ctx, clinicA, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "現金", got.Name)
	})

	t.Run("別クリニックからはNotFound", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "クレジットカード")
		_, err := repo.FindByID(ctx, clinicB, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPaymentMethodMasterRepository_FindAll(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mB := makePaymentMethodMaster(t, db, clinicB, "医院Bの支払方法")
	mA2 := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "B支払方法", DisplayOrder: 2, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(mA2).Error)
	mA1 := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "A支払方法", DisplayOrder: 1, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(mA1).Error)

	t.Run("クリニックで隔離されdisplay_order/nameの昇順で返る", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, mA1.ID, got[0].ID)
		assert.Equal(t, mA2.ID, got[1].ID)
		for _, m := range got {
			assert.NotEqual(t, mB.ID, m.ID)
		}
	})

	t.Run("ソフトデリート済みは一覧から除外されるがレコードは残る", func(t *testing.T) {
		db2 := setupPaymentMethodMasterRepoTestDB(t)
		repo2 := NewPaymentMethodMasterRepository(db2)
		m := makeCustomPaymentMethodMaster(t, db2, clinicA, "削除予定支払方法")

		require.NoError(t, repo2.Delete(ctx, clinicA, m.ID))

		got, err := repo2.FindAll(ctx, clinicA)
		require.NoError(t, err)
		assert.Len(t, got, 0)

		var raw model.PaymentMethodMaster
		require.NoError(t, db2.WithContext(ctx).Unscoped().First(&raw, m.ID).Error)
		assert.NotNil(t, raw.DeletedAt)
	})
}

func TestPaymentMethodMasterRepository_Update(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの更新は反映される", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "更新前支払方法")
		got, err := repo.Update(ctx, clinicA, m.ID, map[string]any{"name": "更新後支払方法"})
		require.NoError(t, err)
		assert.Equal(t, "更新後支払方法", got.Name)
	})

	t.Run("別クリニックの更新はNotFound", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "越境更新対象")
		_, err := repo.Update(ctx, clinicB, m.ID, map[string]any{"name": "越境更新"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPaymentMethodMasterRepository_Update_ReloadFailureRollsBackUpdate(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	record := makeCustomPaymentMethodMaster(t, db, clinicID, "更新前支払方法")
	const callbackName = "payment_method:update_and_find_reload_failure"
	reloadErr := errors.New("forced reload failure")
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement == nil {
			return
		}
		table := query.Statement.Table
		if table == "" && query.Statement.Schema != nil {
			table = query.Statement.Schema.Table
		}
		if table == "payment_methods" {
			query.AddError(reloadErr)
		}
	}))
	callbackRegistered := true
	t.Cleanup(func() {
		if callbackRegistered {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	got, err := repo.Update(ctx, clinicID, record.ID, map[string]any{"name": "更新後支払方法"})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, reloadErr)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRegistered = false

	var persisted model.PaymentMethodMaster
	require.NoError(t, db.WithContext(ctx).First(&persisted, record.ID).Error)
	assert.Equal(t, "更新前支払方法", persisted.Name)
}

func TestPaymentMethodMasterRepository_Delete(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの未使用カスタム行は削除に成功しその後取得できない", func(t *testing.T) {
		m := makeCustomPaymentMethodMaster(t, db, clinicA, "削除対象支払方法")
		require.NoError(t, repo.Delete(ctx, clinicA, m.ID))

		_, err := repo.FindByID(ctx, clinicA, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの削除はNotFoundで対象データは残る", func(t *testing.T) {
		m := makeCustomPaymentMethodMaster(t, db, clinicA, "越境削除対象支払方法")
		err := repo.Delete(ctx, clinicB, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		assert.Equal(t, m.ID, got.ID)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("system_key 空文字の未使用行は削除できる", func(t *testing.T) {
		empty := ""
		m := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "空キー支払方法", IsActive: true, SystemKey: &empty}
		require.NoError(t, db.WithContext(ctx).Create(m).Error)
		require.NoError(t, repo.Delete(ctx, clinicA, m.ID))
		_, err := repo.FindByID(ctx, clinicA, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("システム標準行はConflictで削除されない", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "システム現金")
		err := repo.Delete(ctx, clinicA, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "システム標準の支払方法は削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, findErr)
		assert.Equal(t, m.ID, got.ID)
	})

	t.Run("使用中のカスタム行はConflictで削除されない", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "使用中カスタム化対象")
		billing := makePaymentMethodBilling(t, db, clinicA)
		makePaymentForBilling(t, db, billing.ID, m.ID)
		require.NoError(t, db.WithContext(ctx).
			Model(&model.PaymentMethodMaster{}).
			Where("id = ?", m.ID).
			Updates(map[string]any{"system_key": nil}).Error)

		err := repo.Delete(ctx, clinicA, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この支払方法は使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, findErr)
		assert.Equal(t, m.ID, got.ID)
	})

	t.Run("CountUsage==0 直後に参照が追加されても削除は失敗する", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "TOCTOUカスタム化対象")
		count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		billing := makePaymentMethodBilling(t, db, clinicA)
		makePaymentForBilling(t, db, billing.ID, m.ID)
		require.NoError(t, db.WithContext(ctx).
			Model(&model.PaymentMethodMaster{}).
			Where("id = ?", m.ID).
			Updates(map[string]any{"system_key": nil}).Error)

		err = repo.Delete(ctx, clinicA, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この支払方法は使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, findErr)
		assert.Equal(t, m.ID, got.ID)
	})
}

func TestPaymentMethodMasterRepository_CountUsageByPaymentMethodID(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("billingsを経由するpaymentsの件数をカウントする", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "使用中支払方法")
		billing1 := makePaymentMethodBilling(t, db, clinicA)
		makePaymentForBilling(t, db, billing1.ID, m.ID)
		billing2 := makePaymentMethodBilling(t, db, clinicA)
		makePaymentForBilling(t, db, billing2.ID, m.ID)

		count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("未使用の支払方法は0を返す", func(t *testing.T) {
		unused := makePaymentMethodMaster(t, db, clinicA, "未使用支払方法")
		count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済みのpaymentはカウントされない", func(t *testing.T) {
		m := makePaymentMethodMaster(t, db, clinicA, "削除支払対象方法")
		billing := makePaymentMethodBilling(t, db, clinicA)
		p := makePaymentForBilling(t, db, billing.ID, m.ID)
		require.NoError(t, db.WithContext(ctx).Delete(&model.Payment{}, p.ID).Error)

		count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("別クリニックの請求に紐づく支払はカウントされない", func(t *testing.T) {
		mA := makePaymentMethodMaster(t, db, clinicA, "越境参照対象支払方法")
		// ADR-003 DB boundary rejects cross-clinic payment_method_id pollution.
		// Prove isolation via a valid clinic-B payment that must not inflate clinic-A usage.
		mB := makePaymentMethodMaster(t, db, clinicB, "医院B支払方法")
		billingB := makePaymentMethodBilling(t, db, clinicB)
		makePaymentForBilling(t, db, billingB.ID, mB.ID)

		count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, mA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "別クリニックのbillingに紐づく支払はJOINで除外される")
	})
}

// TestPaymentMethodMasterRepository_Reorder は Reorder の期待される正しい挙動を検証する。
func TestPaymentMethodMasterRepository_Reorder(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("指定した順序でdisplay_orderが1始まりに更新される", func(t *testing.T) {
		m1 := makePaymentMethodMaster(t, db, clinicA, "支払方法1")
		m2 := makePaymentMethodMaster(t, db, clinicA, "支払方法2")
		m3 := makePaymentMethodMaster(t, db, clinicA, "支払方法3")

		err := repo.Reorder(ctx, clinicA, []uint64{m3.ID, m1.ID, m2.ID})
		require.NoError(t, err, "Reorderが成功すること")

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		byID := make(map[uint64]model.PaymentMethodMaster, 3)
		for _, m := range got {
			byID[m.ID] = m
		}
		assert.Equal(t, 1, byID[m3.ID].DisplayOrder)
		assert.Equal(t, 2, byID[m1.ID].DisplayOrder)
		assert.Equal(t, 3, byID[m2.ID].DisplayOrder)
	})

	t.Run("別クリニックのIDが混ざるとエラーになる", func(t *testing.T) {
		mA := makePaymentMethodMaster(t, db, clinicA, "医院A支払方法")
		mB := makePaymentMethodMaster(t, db, clinicB, "医院B支払方法")

		err := repo.Reorder(ctx, clinicA, []uint64{mA.ID, mB.ID})
		require.Error(t, err)
	})

	t.Run("ambient WithTx ロールバックで並び替えが取り消される", func(t *testing.T) {
		m1 := makePaymentMethodMaster(t, db, clinicA, "tx並び1")
		m2 := makePaymentMethodMaster(t, db, clinicA, "tx並び2")
		before, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		orderBefore := map[uint64]int{}
		for _, m := range before {
			if m.ID == m1.ID || m.ID == m2.ID {
				orderBefore[m.ID] = m.DisplayOrder
			}
		}

		tx := persistence.NewTransactor(db)
		err = tx.WithTx(ctx, func(txCtx context.Context) error {
			if err := repo.Reorder(txCtx, clinicA, []uint64{m2.ID, m1.ID}); err != nil {
				return err
			}
			return errors.New("force ambient rollback")
		})
		require.Error(t, err)

		after, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, m := range after {
			if want, ok := orderBefore[m.ID]; ok {
				assert.Equal(t, want, m.DisplayOrder, "id=%d must not commit reorder after ambient rollback", m.ID)
			}
		}
	})
}
