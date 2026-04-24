// Package service provides shared billing calculation logic.
package service

import (
	"math"

	"github.com/animal-ekarte/backend/internal/model"
)

// CalculateTaxAmount は課税区分に応じた税額（円）を計算する。
//
//	外税: 税額 = 単価 × 数量 × 税率
//	内税: 税額 = 単価 × 数量 × 税率 ÷ (1 + 税率)
//	非課税: 税額 = 0
func CalculateTaxAmount(unitPrice int64, quantity float64, taxType model.TaxType, taxRate float64) int64 {
	subtotal := float64(unitPrice) * quantity
	switch taxType {
	case model.TaxTypeExcluded:
		return int64(math.Round(subtotal * taxRate))
	case model.TaxTypeIncluded:
		return int64(math.Round(subtotal * taxRate / (1 + taxRate)))
	case model.TaxTypeExempt:
		return 0
	default:
		return 0
	}
}

// CalculateBillingTotals は全 BillingItem から subtotal, taxTotal, totalAmount を再計算する。
//   - subtotal    : 全明細の単価×数量の合計
//   - taxTotal    : 全明細の税額合計（外税+内税+非課税）
//   - totalAmount : subtotal + 外税額（内税は subtotal に内包されるため加算しない）
func CalculateBillingTotals(items []model.BillingItem) (subtotal, taxTotal, totalAmount int64) {
	var excludedTax int64
	for i := range items {
		itemSubtotal := int64(float64(items[i].UnitPrice) * items[i].Quantity)
		taxAmount := items[i].CalculateTaxAmount()
		subtotal += itemSubtotal
		taxTotal += taxAmount
		if items[i].TaxType == model.TaxTypeExcluded {
			excludedTax += taxAmount
		}
	}
	totalAmount = subtotal + excludedTax
	return
}
