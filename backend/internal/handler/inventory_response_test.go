package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToInventoryResponseAppliesLocalTimePtrToDateFields(t *testing.T) {
	expiryDate := time.Date(2027, 3, 10, 0, 0, 0, 0, time.UTC)
	lastRestocked := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	resp := toInventoryResponse(&model.InventoryItem{
		ID:            1,
		ClinicID:      1,
		Name:          "テスト薬剤",
		Category:      model.InventoryCategoryMedicine,
		Quantity:      10,
		Unit:          "本",
		MinStockLevel: 2,
		Location:      "保管庫A",
		ExpiryDate:    &expiryDate,
		Supplier:      "テスト卸",
		LastRestocked: &lastRestocked,
		Status:        model.InventoryStatusSufficient,
	})

	// R2-1 (BE-refactor.md D3): pet の BirthDate/NeuteredDate/LastVisit と同様、
	// 日付フィールドは canonical 規約に従い localTimePtr で time.Local へ変換する。
	// tz 表現ではなくカレンダー日付で検証する（格納値を保持していることの確認）。
	require.NotNil(t, resp.ExpiryDate)
	assert.Equal(t, "2027-03-10", resp.ExpiryDate.Format("2006-01-02"))
	assert.Equal(t, time.Local, resp.ExpiryDate.Location())

	require.NotNil(t, resp.LastRestocked)
	assert.Equal(t, "2026-06-01", resp.LastRestocked.Format("2006-01-02"))
	assert.Equal(t, time.Local, resp.LastRestocked.Location())
}

func TestToInventoryResponseNilDateFieldsStayNil(t *testing.T) {
	resp := toInventoryResponse(&model.InventoryItem{
		ID:       2,
		ClinicID: 1,
		Name:     "期限なし備品",
		Category: model.InventoryCategoryOther,
		Status:   model.InventoryStatusSufficient,
	})

	assert.Nil(t, resp.ExpiryDate)
	assert.Nil(t, resp.LastRestocked)
}
