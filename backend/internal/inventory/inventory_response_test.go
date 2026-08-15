package inventory

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

// TestToInventoryResponseDerivesStatusIgnoringStoredValue は SD-4 決裁A（q&a.html SD-4）の回帰:
// status は保存値を信頼せず、常に quantity/min_stock_level から読み取り時に導出しなければ
// ならない。保存された item.Status（クライアントが作成/更新時に明示指定した値を含む）は
// 意図的に無視され、導出結果だけが返ることを検証する。
func TestToInventoryResponseDerivesStatusIgnoringStoredValue(t *testing.T) {
	tests := []struct {
		name          string
		quantity      int
		minStockLevel int
		storedStatus  model.InventoryStatus
		wantStatus    string
	}{
		{
			name:          "保存値が sufficient でも quantity <= min_stock_level なら low を返す",
			quantity:      3,
			minStockLevel: 10,
			storedStatus:  model.InventoryStatusSufficient,
			wantStatus:    string(model.InventoryStatusLow),
		},
		{
			name:          "保存値が low でも quantity が閾値を上回れば sufficient を返す",
			quantity:      20,
			minStockLevel: 10,
			storedStatus:  model.InventoryStatusLow,
			wantStatus:    string(model.InventoryStatusSufficient),
		},
		{
			name:          "保存値が sufficient でも quantity が 0 なら out_of_stock を返す",
			quantity:      0,
			minStockLevel: 10,
			storedStatus:  model.InventoryStatusSufficient,
			wantStatus:    string(model.InventoryStatusOutOfStock),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toInventoryResponse(&model.InventoryItem{
				ID:            1,
				ClinicID:      1,
				Name:          "導出テスト対象",
				Category:      model.InventoryCategoryMedicine,
				Quantity:      tt.quantity,
				MinStockLevel: tt.minStockLevel,
				Status:        tt.storedStatus,
			})

			assert.Equal(t, tt.wantStatus, resp.Status)
		})
	}
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
