package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildTaxBreakdown_ClinicRates は M-7(#191) の回帰テスト。
// 締めレジ経路の税率分類を固定閾値（>8）から病院マスタ税率の exact-match へ統一した。
// 1) 既定クリニック（標準10%/軽減8%）では旧実装と同じ分類になる（挙動不変の証明）。
// 2) 軽減税率 9% のクリニックでは 9% が軽減へ分類される（旧固定閾値では標準へ誤分類）。
// 3) 0%（非課税）は軽減と一致しないため標準へ分類する（月次 #191 と同一規則）。
func TestBuildTaxBreakdown_ClinicRates(t *testing.T) {
	defaultRates := accountingReportTaxRates{StandardPercent: 10, ReducedPercent: 8}

	t.Run("既定(標準10/軽減8): 10%→標準・8%→軽減で旧分類と一致", func(t *testing.T) {
		rows := []TaxBreakdownRow{
			{TaxRate: 10, TaxableAmount: 1000, TaxAmount: 100},
			{TaxRate: 8, TaxableAmount: 500, TaxAmount: 40},
		}
		got := buildTaxBreakdown(rows, defaultRates)
		assert.Equal(t, int64(1000), got.Standard.TaxableAmount)
		assert.Equal(t, int64(100), got.Standard.TaxAmount)
		assert.Equal(t, int64(500), got.Reduced.TaxableAmount)
		assert.Equal(t, int64(40), got.Reduced.TaxAmount)
	})

	t.Run("軽減9%クリニック: 9%→軽減・10%→標準（旧閾値>8では9%が標準へ誤分類）", func(t *testing.T) {
		rows := []TaxBreakdownRow{
			{TaxRate: 10, TaxableAmount: 1000, TaxAmount: 100},
			{TaxRate: 9, TaxableAmount: 400, TaxAmount: 36},
		}
		got := buildTaxBreakdown(rows, accountingReportTaxRates{StandardPercent: 10, ReducedPercent: 9})
		assert.Equal(t, int64(1000), got.Standard.TaxableAmount)
		assert.Equal(t, int64(100), got.Standard.TaxAmount)
		assert.Equal(t, int64(400), got.Reduced.TaxableAmount)
		assert.Equal(t, int64(36), got.Reduced.TaxAmount)
	})

	t.Run("0%(非課税)は標準へ分類（月次#191と同一規則）", func(t *testing.T) {
		rows := []TaxBreakdownRow{
			{TaxRate: 0, TaxableAmount: 300, TaxAmount: 0},
			{TaxRate: 8, TaxableAmount: 500, TaxAmount: 40},
		}
		got := buildTaxBreakdown(rows, defaultRates)
		assert.Equal(t, int64(300), got.Standard.TaxableAmount)
		assert.Equal(t, int64(500), got.Reduced.TaxableAmount)
	})

	t.Run("空入力はゼロ値サマリ", func(t *testing.T) {
		got := buildTaxBreakdown(nil, defaultRates)
		assert.Equal(t, TaxBreakdownSummary{}, got)
	})
}
