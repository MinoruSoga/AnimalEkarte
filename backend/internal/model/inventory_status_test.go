package model

// inventory_status_test.go — SD-4: min_stock_level から status を判定するロジックが
// どこにも存在しなかったバグの修正対象。DeriveInventoryStatus の境界値を検証する。

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveInventoryStatus(t *testing.T) {
	tests := []struct {
		name          string
		quantity      int
		minStockLevel int
		want          InventoryStatus
	}{
		{
			name:          "quantity が閾値を上回るとき sufficient",
			quantity:      20,
			minStockLevel: 10,
			want:          InventoryStatusSufficient,
		},
		{
			name:          "quantity が閾値と等しいとき low（境界値）",
			quantity:      10,
			minStockLevel: 10,
			want:          InventoryStatusLow,
		},
		{
			name:          "quantity が閾値を下回るとき low",
			quantity:      3,
			minStockLevel: 10,
			want:          InventoryStatusLow,
		},
		{
			name:          "quantity が 0 のとき out_of_stock（閾値に関わらず）",
			quantity:      0,
			minStockLevel: 10,
			want:          InventoryStatusOutOfStock,
		},
		{
			name:          "quantity が負値のとき out_of_stock",
			quantity:      -1,
			minStockLevel: 10,
			want:          InventoryStatusOutOfStock,
		},
		{
			name:          "閾値未設定(0)のとき quantity に関わらず sufficient",
			quantity:      1,
			minStockLevel: 0,
			want:          InventoryStatusSufficient,
		},
		{
			name:          "閾値未設定(0)でも quantity が 0 なら out_of_stock",
			quantity:      0,
			minStockLevel: 0,
			want:          InventoryStatusOutOfStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveInventoryStatus(tt.quantity, tt.minStockLevel)
			assert.Equal(t, tt.want, got)
		})
	}
}
