package billing

// campaign_repository_tx_atomicity_test.go — BE-X06-BIL-CAMPAIGN-01 / BIL-03 (X-06)
//
// campaignService.Update は本体 Update と ReplaceTargets を同一 ambient tx に載せる。
// 本ファイルは repository が DBOrTx 経由で ambient に参加し、後続失敗時に
// キャンペーン本体と対象差し替えがまとめてロールバックされることを証明する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errSentinelCampaignTx = errors.New("simulated post-write failure in ambient tx")

func TestCampaignRepository_Update_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "原子性テスト", StartDate: jun, EndDate: jul, DiscountValue: 10}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, c.ID, map[string]any{"name": "書き換え後"}); err != nil {
			return err
		}
		return errSentinelCampaignTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelCampaignTx)

	var got model.Campaign
	require.NoError(t, db.WithContext(ctx).First(&got, c.ID).Error)
	assert.Equal(t, "原子性テスト", got.Name, "ambient tx 失敗時、キャンペーン本体更新はロールバックされる")
}

func TestCampaignRepository_UpdateAndReplaceTargets_RollBackTogether(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "グラフ原子性", StartDate: jun, EndDate: jul, DiscountValue: 5}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.CampaignTargetCategory{
		CampaignID: c.ID,
		Category:   model.ItemCategoryVaccine,
	}).Error)

	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, c.ID, map[string]any{
			"name":           "部分成功させない",
			"discount_value": 99.0,
		}); err != nil {
			return err
		}
		if err := repo.ReplaceTargets(txCtx, c.ID, []model.ItemCategory{model.ItemCategoryGoods}, []uint64{42}); err != nil {
			return err
		}
		return errSentinelCampaignTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelCampaignTx)

	var got model.Campaign
	require.NoError(t, db.WithContext(ctx).First(&got, c.ID).Error)
	assert.Equal(t, "グラフ原子性", got.Name)
	assert.Equal(t, 5.0, got.DiscountValue)

	var catCount, itemCount int64
	db.Model(&model.CampaignTargetCategory{}).Where("campaign_id = ?", c.ID).Count(&catCount)
	db.Model(&model.CampaignTargetItem{}).Where("campaign_id = ?", c.ID).Count(&itemCount)
	assert.EqualValues(t, 1, catCount, "旧カテゴリ対象が残る（差し替えはロールバック）")
	assert.EqualValues(t, 0, itemCount, "新商品対象はコミットされない")
}

func TestCampaignService_Update_AtomicFieldsAndTargets(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{
		ClinicID:      clinicA,
		Name:          "サービス原子性",
		StartDate:     jun,
		EndDate:       jul,
		DiscountType:  model.CampaignDiscountTypeAmount,
		DiscountValue: 100,
		IsActive:      true,
	}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.CampaignTargetCategory{
		CampaignID: c.ID,
		Category:   model.ItemCategoryVaccine,
	}).Error)

	// Force ReplaceTargets path to fail after Update would have committed under the old non-atomic design.
	failingRepo := &campaignReplaceFailAfterUpdate{CampaignRepository: repo}
	svc := NewCampaignService(failingRepo, &mockMerchandiseItemRepository{}, testNewTransactor(db))

	newName := "失敗後に残らない"
	cats := []model.ItemCategory{model.ItemCategoryGoods}
	_, err := svc.Update(ctx, clinicA, c.ID, &UpdateCampaignInput{
		Name:             &newName,
		TargetCategories: &cats,
	})
	require.Error(t, err)

	var got model.Campaign
	require.NoError(t, db.WithContext(ctx).First(&got, c.ID).Error)
	assert.Equal(t, "サービス原子性", got.Name, "ReplaceTargets 失敗時に本体更新もロールバックされる（BIL-03）")

	var cat model.CampaignTargetCategory
	require.NoError(t, db.WithContext(ctx).Where("campaign_id = ?", c.ID).First(&cat).Error)
	assert.Equal(t, model.ItemCategoryVaccine, cat.Category)
}

// campaignReplaceFailAfterUpdate は Update を実 repo に渡し、ReplaceTargets だけ失敗させる。
type campaignReplaceFailAfterUpdate struct {
	CampaignRepository
}

func (r *campaignReplaceFailAfterUpdate) ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error {
	// Ensure ambient participation is real before failing (so Update is in the same tx).
	if persistence.TxFromContext(ctx) == nil {
		return errors.New("expected ambient tx for BIL-03 atomicity test")
	}
	return errors.New("forced ReplaceTargets failure")
}

// Compile-time guard: FindByID inside ambient tx must use the same connection.
func TestCampaignRepository_FindByID_SeesUncommittedUpdateInAmbientTx(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "再取得", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	tx := testNewTransactor(db)
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, c.ID, map[string]any{"name": "tx内名称"}); err != nil {
			return err
		}
		got, err := repo.FindByID(txCtx, clinicA, c.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, "tx内名称", got.Name)
		return nil
	}))
}
