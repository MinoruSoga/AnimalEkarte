package billing

// billing_item_repository_tx_atomicity_test.go — BE-refactor.md R1-1 follow-up（go-reviewer 指摘）
//
// 背景: billing_item_service の CreateItem/UpdateItem/DeleteItem は Create/Update/Delete と
// recalculateTotals（FindByBillingID + UpdateBillingTotals）を WithTx 内で txCtx 付きで呼ぶが、
// billing_item_repository.go は dbOrTx を一切使用していなかった（accounting_repository.go の
// SavePaymentSplits と同型の部分コミットバグ）。明細の書込は独立 tx で即コミットされるため、
// 直後の合計再計算だけが失敗しても ambient tx の rollback で明細は戻らず、billing_items と
// billings.subtotal/tax_total/total_amount が不整合になりうる。
//
// temp-revert RED の手順:
//   - billing_item_repository.go の対象メソッドを dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//   - 本ファイルのテストを実行 → Rollback 系が RED（書込が残ってしまう）
//   - 元に戻す → GREEN

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeBillingForItemTx は billing_item tx-atomicity テスト用の最小 Billing を作成する。
// makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
func makeBillingForItemTx(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		TotalAmount:   0,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
}

var errSentinelBillingItemTx = errors.New("simulated post-write failure in ambient tx")

func TestBillingItemRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForItemTx(t, db, clinicA)
	repo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)

	item := &model.BillingItem{
		BillingID: billing.ID,
		Category:  model.ItemCategoryOther,
		Name:      "原子性テスト明細",
		UnitPrice: 3000,
		Quantity:  1,
	}
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, item); err != nil {
			return err
		}
		return errSentinelBillingItemTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelBillingItemTx)

	var count int64
	db.Model(&model.BillingItem{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 0, count, "ambient tx 失敗時、明細の新規作成はロールバックされる（fail-closed）")
}

func TestBillingItemRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForItemTx(t, db, clinicA)
	repo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)

	item := &model.BillingItem{
		BillingID: billing.ID,
		Category:  model.ItemCategoryOther,
		Name:      "コミットテスト明細",
		UnitPrice: 3000,
		Quantity:  1,
	}
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, item)
	}))

	var count int64
	db.Model(&model.BillingItem{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 1, count, "commit 後は billing_items に行が永続化される")
}

// TestBillingItemRepository_CreateThenUpdateBillingTotals_RollsBackTogether は
// billing_item_service.CreateItem の実フローを模し、明細作成 + 合計更新のどちらかが
// 揃ってロールバックされることを検証する（部分コミット防止の直接証明）。
func TestBillingItemRepository_CreateThenUpdateBillingTotals_RollsBackTogether(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForItemTx(t, db, clinicA)
	repo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)

	item := &model.BillingItem{
		BillingID: billing.ID,
		Category:  model.ItemCategoryOther,
		Name:      "合計再計算テスト明細",
		UnitPrice: 5000,
		Quantity:  1,
		TaxRate:   0.10,
	}
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, item); err != nil {
			return err
		}
		// UpdateBillingTotals は成功させ、その後で失敗注入する（recalculateTotals 成功後の
		// 別ステップ失敗を模す）。バグ時は Create が既に独立 tx で確定しているため、
		// ここでの UpdateBillingTotals 成功有無に関わらず明細だけ残ってしまう。
		if err := repo.UpdateBillingTotals(txCtx, clinicA, billing.ID, 5000, 500, 5500); err != nil {
			return err
		}
		return errSentinelBillingItemTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelBillingItemTx)

	var itemCount int64
	db.Model(&model.BillingItem{}).Where("billing_id = ?", billing.ID).Count(&itemCount)
	assert.EqualValues(t, 0, itemCount, "ambient tx 失敗時、明細作成はロールバックされる")

	var reloaded model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloaded, billing.ID).Error)
	assert.EqualValues(t, 0, reloaded.TotalAmount,
		"ambient tx 失敗時、UpdateBillingTotals による合計更新もロールバックされる"+
			"（バグ時は明細・合計とも独立 tx で確定済みのため、この両方が不整合に残ってしまう）")
}

func TestBillingItemRepository_UpdateBillingTotals_CommitsWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForItemTx(t, db, clinicA)
	repo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.UpdateBillingTotals(txCtx, clinicA, billing.ID, 5000, 500, 5500)
	}))

	var reloaded model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloaded, billing.ID).Error)
	assert.EqualValues(t, 5500, reloaded.TotalAmount, "commit 後は合計更新が永続化される")
}

func TestBillingItemRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForItemTx(t, db, clinicA)
	item := &model.BillingItem{
		BillingID: billing.ID,
		Category:  model.ItemCategoryOther,
		Name:      "削除ロールバックテスト明細",
		UnitPrice: 1000,
		Quantity:  1,
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	repo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, clinicA, item.ID); err != nil {
			return err
		}
		return errSentinelBillingItemTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelBillingItemTx)

	var count int64
	db.Model(&model.BillingItem{}).Unscoped().Where("id = ? AND deleted_at IS NULL", item.ID).Count(&count)
	assert.EqualValues(t, 1, count, "ambient tx 失敗時、明細の削除（soft-delete）はロールバックされる")
}

func TestBillingItemRepository_ValidateVaccinationCreateReferenceLocksVaccinationInAmbientTransaction(t *testing.T) {
	fixture := setupBillingItemReferenceFixture(t)
	price := int64(3200)
	_, vaccination := makeBillingVaccination(t, fixture, "ambient tx lock vaccine", &price)
	repo := NewBillingItemRepository(fixture.db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := testNewTransactor(fixture.db).WithTx(ctx, func(txCtx context.Context) error {
		values, validateErr := repo.ValidateVaccinationCreateReference(
			txCtx,
			fixture.clinicID,
			fixture.billing.ID,
			vaccination.ID,
		)
		if validateErr != nil {
			return validateErr
		}
		if values.Name != "ambient tx lock vaccine" || values.UnitPrice != price {
			return errors.New("vaccination billing values do not match the locked master")
		}

		done := make(chan error, 1)
		go func() {
			competingTx := fixture.db.WithContext(ctx).Begin()
			if competingTx.Error != nil {
				done <- competingTx.Error
				return
			}
			defer competingTx.Rollback()
			if lockErr := competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error; lockErr != nil {
				done <- lockErr
				return
			}
			done <- competingTx.Exec(
				"UPDATE vaccinations SET date = date WHERE id = ?",
				vaccination.ID,
			).Error
		}()

		select {
		case competingErr := <-done:
			if competingErr == nil {
				return errors.New("concurrent vaccination mutation completed before ambient transaction commit")
			}
			if !strings.Contains(competingErr.Error(), "lock timeout") {
				return competingErr
			}
			return nil
		case <-ctx.Done():
			return errors.New("timed out waiting for bounded concurrent vaccination mutation")
		}
	})
	require.NoError(t, err)

	afterCommitTx := fixture.db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	require.NoError(t, afterCommitTx.Exec(
		"UPDATE vaccinations SET date = date WHERE id = ?",
		vaccination.ID,
	).Error)
}

func TestBillingItemRepository_ValidateExamCreateReference_FailClosedWithoutAmbientTx(t *testing.T) {
	fixture := setupBillingItemReferenceFixture(t)
	price := int64(4200)
	_, exam := makeBillingExam(t, fixture, "no ambient exam", &price)
	repo := NewBillingItemRepository(fixture.db)

	err := repo.ValidateExamCreateReference(
		context.Background(),
		fixture.clinicID,
		fixture.billing.ID,
		exam.ID,
	)
	require.Error(t, err)
	var app *apperrors.AppError
	require.True(t, errors.As(err, &app), "want *apperrors.AppError, got %T: %v", err, err)
	assert.Equal(t, "INTERNAL", app.Code)
	assert.Contains(t, app.Message, "active transaction")
}

func TestBillingItemRepository_ValidateExamCreateReferenceLocksExamTypeInAmbientTransaction(t *testing.T) {
	fixture := setupBillingItemReferenceFixture(t)
	price := int64(4200)
	examType, exam := makeBillingExam(t, fixture, "ambient tx lock exam", &price)
	repo := NewBillingItemRepository(fixture.db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := testNewTransactor(fixture.db).WithTx(ctx, func(txCtx context.Context) error {
		if validateErr := repo.ValidateExamCreateReference(
			txCtx,
			fixture.clinicID,
			fixture.billing.ID,
			exam.ID,
		); validateErr != nil {
			return validateErr
		}

		done := make(chan error, 1)
		go func() {
			competingTx := fixture.db.WithContext(ctx).Begin()
			if competingTx.Error != nil {
				done <- competingTx.Error
				return
			}
			defer competingTx.Rollback()
			if lockErr := competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error; lockErr != nil {
				done <- lockErr
				return
			}
			done <- competingTx.Exec(
				"UPDATE exam_types SET name = name WHERE id = ?",
				examType.ID,
			).Error
		}()

		select {
		case competingErr := <-done:
			if competingErr == nil {
				return errors.New("concurrent exam_types mutation completed before ambient transaction commit")
			}
			if !strings.Contains(competingErr.Error(), "lock timeout") {
				return competingErr
			}
			return nil
		case <-ctx.Done():
			return errors.New("timed out waiting for bounded concurrent exam_types mutation")
		}
	})
	require.NoError(t, err)

	afterCommitTx := fixture.db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	require.NoError(t, afterCommitTx.Exec(
		"UPDATE exam_types SET name = name WHERE id = ?",
		examType.ID,
	).Error)
}
