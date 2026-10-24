// Package billing provides shared billing calculation logic.
package billing

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// 算定器（EMR-505 / D5）。外税の通常経路だけを exact（math/big）で算定する。
// 規則の正本は docs/spec/domains/accounting/lifecycle.md の #accounting-calculation-order と
// #accounting-calculation-examples。float64 は金額・数量・割合・税率のどこにも使わない。
//
//	明細行の金額     = 単価 × 数量 を 1 円未満四捨五入（0.5 は切上げ）
//	scope の値引き   = 割合（同 scope 内で合算済み）の額を 1 円未満四捨五入 → 定額 の順。対象額を超えない
//	会計全体の値引き = 同じ手順を明細値引き後の合計に行い、最大剰余法（同率は明細順の昇順）で各明細へ配賦
//	税               = 値引き・配賦後の税抜額を税率ごとに合計し、1 円未満を 1 回だけ切り捨て
//
// 内税・非課税・税区分の混在・税率未設定は規則が未決定（EMR-TAX-SCOPE-20261021）のため、
// 既定値で補わず calculation_unsupported（422）で拒否する。

// errCalcOutOfRange は金額が int64 に収まらないことを示す内部エラー。呼出し側が文脈つきの 422 に変換する。
var errCalcOutOfRange = errors.New("calc: 金額が int64 の範囲外です")

// calcTotalsOutOfRangeMessage は集計値が int64 に収まらないときの 422 の文言。
const calcTotalsOutOfRangeMessage = "合計金額が扱える範囲を超えています"

// CalcLine は算定器の明細 1 行分の入力。金額・数量・割合はすべて exact で受け、float64 を経由しない。
//   - UnitPrice・Quantity: 単価（円）と数量。数量は正であること。
//   - LinePercent: 明細 scope で合算済みの値引き割合（百分率。10 は 10%）。値引きなしは 0 を明示する（NULL は拒否）。
//   - LineFixed: 明細 scope の定額値引き（円）。割合値引きの後に適用する。
//   - TaxType・TaxRate: 外税（excluded）と税率（割合。0.10 は 10%）。nil の税率は未設定として拒否し、既定値で補わない。
type CalcLine struct {
	UnitPrice   model.ExactNumeric
	Quantity    model.ExactNumeric
	LinePercent model.ExactNumeric
	LineFixed   int64
	TaxType     model.TaxType
	TaxRate     *model.ExactNumeric
}

// CalcInput は 1 請求分の算定入力。Lines の並びが凍結した明細順で、最大剰余法の同率は昇順で配る。
// BillPercent・BillFixed は会計全体 scope で合算済みの割合（百分率）と定額（円）。値引きなしは 0 を明示する。
type CalcInput struct {
	Lines       []CalcLine
	BillPercent model.ExactNumeric
	BillFixed   int64
}

// CalcLineResult は明細 1 行の算定結果（円）。GrossAmount は単価 × 数量の四捨五入、
// LineDiscountRequested・LineDiscountApplied は明細 scope の要求額と対象額までの実適用額、
// AfterLineDiscount は明細値引き後、BillDiscountAllocated は会計全体値引きの配賦額、NetAmount は税抜の最終額。
type CalcLineResult struct {
	GrossAmount           int64
	LineDiscountRequested int64
	LineDiscountApplied   int64
	AfterLineDiscount     int64
	BillDiscountAllocated int64
	NetAmount             int64
}

// CalcTaxGroup は税率ごとの集約。TaxableAmount は配賦後の税抜額の合計、TaxAmount はその 1 円未満切捨ての税額。
type CalcTaxGroup struct {
	TaxRate       model.ExactNumeric
	TaxableAmount int64
	TaxAmount     int64
}

// CalcResult は 1 請求分の算定結果（円）。Subtotal は税抜の最終額の合計、TotalAmount は Subtotal + TaxTotal。
// BillDiscountRequested・BillDiscountApplied は会計全体 scope の要求額と実適用額。TaxGroups は税率の昇順。
type CalcResult struct {
	Lines                 []CalcLineResult
	BillDiscountRequested int64
	BillDiscountApplied   int64
	Subtotal              int64
	TaxGroups             []CalcTaxGroup
	TaxTotal              int64
	TotalAmount           int64
}

// LineDisplayAmounts は明細 1 行の表示用の金額。Subtotal は明細値引き後（会計全体の値引き前）の税抜額。
// TaxAmount は表示用の行別税額で、請求書の税額の正本ではない（正本は CalcResult.TaxGroups）。
// 税額を確定できない行（内税・税率未設定など）は nil とし、0 で埋めない。
type LineDisplayAmounts struct {
	Subtotal  int64
	TaxAmount *int64
}

// CalcLineFromBillingItem は保存済みの billing_items 行を算定入力へ変換する。
// 値引きは discount_amount（円）だけを定額として使い、discount_rate は再適用しない。既存の経路
// （resolveAutoDiscount など）は割引を円で確定して discount_amount に書き、同じ割引の適用率を記録として
// discount_rate に残すため、両方を独立した値引きとして扱うと二重に引く。割合を算定器へ渡す経路は、
// 同 scope の割合を合算して LinePercent に渡す将来の snapshot adapter が持つ。
func CalcLineFromBillingItem(item *model.BillingItem) CalcLine {
	return CalcLine{
		UnitPrice:   item.UnitPrice,
		Quantity:    item.Quantity,
		LinePercent: model.ExactNumericFromInt64(0),
		LineFixed:   item.DiscountAmount,
		TaxType:     item.TaxType,
		TaxRate:     item.TaxRate,
	}
}

// CalculateItemsTotals は保存済みの明細から外税の通常経路の合計を算定する（会計全体の値引きなし）。
// items の並びを凍結した明細順として扱う。算定できない入力は calculation_unsupported（422）を返し、0 で埋めない。
func CalculateItemsTotals(items []model.BillingItem) (CalcResult, error) {
	lines := make([]CalcLine, len(items))
	for i := range items {
		lines[i] = CalcLineFromBillingItem(&items[i])
	}
	return CalculateBilling(CalcInput{
		Lines:       lines,
		BillPercent: model.ExactNumericFromInt64(0),
		BillFixed:   0,
	})
}

// CalculateLineDisplayAmounts は明細 1 行の表示用の金額を求める。小計は単価・数量・明細値引きだけで決まり、
// 税区分に依存しない。表示用の行別税額は外税（税率あり）の floor と、非課税の 0 に限り、それ以外は nil とする。
func CalculateLineDisplayAmounts(line CalcLine) (LineDisplayAmounts, error) {
	amounts, err := calcLineAmounts(line, 0)
	if err != nil {
		return LineDisplayAmounts{}, err
	}
	display := LineDisplayAmounts{Subtotal: amounts.AfterLineDiscount}
	switch line.TaxType {
	case model.TaxTypeExempt:
		zero := int64(0)
		display.TaxAmount = &zero
	case model.TaxTypeExcluded:
		display.TaxAmount = calcDisplayTax(amounts.AfterLineDiscount, line.TaxRate)
	}
	return display, nil
}

// calcDisplayTax は外税の表示用の行別税額（税抜額 × 税率の 1 円未満切捨て）。税率が不正または未設定なら nil。
func calcDisplayTax(taxable int64, taxRate *model.ExactNumeric) *int64 {
	if taxRate == nil {
		return nil
	}
	rate, ok := calcRat(*taxRate)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(1, 1)) > 0 {
		return nil
	}
	tax, err := calcInt64(ratFloor(new(big.Rat).Mul(new(big.Rat).SetInt64(taxable), rate)))
	if err != nil {
		return nil
	}
	return &tax
}

// CalculateBilling は 1 請求分を算定する。入力が空、または外税の通常経路で扱えない（内税・非課税・税区分の混在・
// 税率未設定・負の金額など）ときは calculation_unsupported（422）を返す。空の新規請求の拒否は確定側の規則でもある。
func CalculateBilling(input CalcInput) (CalcResult, error) {
	if len(input.Lines) == 0 {
		return CalcResult{}, apperrors.WrapCalculationUnsupported("明細がない請求は算定できません")
	}
	billPercent, ok := calcRat(input.BillPercent)
	if !ok || billPercent.Sign() < 0 || input.BillFixed < 0 {
		return CalcResult{}, apperrors.WrapCalculationUnsupported("会計全体の値引きが有限の0以上の数値ではありません")
	}

	lines := make([]CalcLineResult, len(input.Lines))
	taxRates := make([]calcTaxRate, len(input.Lines))
	afterAmounts := make([]int64, len(input.Lines))
	afterTotal := new(big.Int)
	for i := range input.Lines {
		rate, err := calcExternalTaxRate(input.Lines[i], i)
		if err != nil {
			return CalcResult{}, err
		}
		amounts, err := calcLineAmounts(input.Lines[i], i)
		if err != nil {
			return CalcResult{}, err
		}
		taxRates[i] = rate
		lines[i] = amounts
		afterAmounts[i] = amounts.AfterLineDiscount
		afterTotal.Add(afterTotal, big.NewInt(amounts.AfterLineDiscount))
	}

	base, err := calcInt64(afterTotal)
	if err != nil {
		return CalcResult{}, apperrors.WrapCalculationUnsupported("明細値引き後の金額合計が扱える範囲を超えています")
	}
	billRequested, billApplied, err := calcDiscountScope(base, billPercent, input.BillFixed)
	if err != nil {
		return CalcResult{}, apperrors.WrapCalculationUnsupported("会計全体の値引き額が扱える範囲を超えています")
	}
	allocations := allocateLargestRemainder(billApplied, afterAmounts)

	groups := make(map[string]*calcTaxGroupWork)
	netTotal := new(big.Int)
	for i := range lines {
		lines[i].BillDiscountAllocated = allocations[i]
		lines[i].NetAmount = lines[i].AfterLineDiscount - allocations[i]
		netTotal.Add(netTotal, big.NewInt(lines[i].NetAmount))
		group, found := groups[taxRates[i].key]
		if !found {
			group = &calcTaxGroupWork{rate: taxRates[i].rat, value: taxRates[i].value, taxable: new(big.Int)}
			groups[taxRates[i].key] = group
		}
		group.taxable.Add(group.taxable, big.NewInt(lines[i].NetAmount))
	}

	ordered := make([]*calcTaxGroupWork, 0, len(groups))
	for _, group := range groups {
		ordered = append(ordered, group)
	}
	sort.Slice(ordered, func(a, b int) bool { return ordered[a].rate.Cmp(ordered[b].rate) < 0 })

	result := CalcResult{Lines: lines, BillDiscountRequested: billRequested, BillDiscountApplied: billApplied}
	taxTotal := new(big.Int)
	for _, group := range ordered {
		taxable, taxableErr := calcInt64(group.taxable)
		tax := ratFloor(new(big.Rat).Mul(new(big.Rat).SetInt(group.taxable), group.rate))
		taxAmount, taxErr := calcInt64(tax)
		if taxableErr != nil || taxErr != nil {
			return CalcResult{}, apperrors.WrapCalculationUnsupported(calcTotalsOutOfRangeMessage)
		}
		result.TaxGroups = append(result.TaxGroups, CalcTaxGroup{TaxRate: group.value, TaxableAmount: taxable, TaxAmount: taxAmount})
		taxTotal.Add(taxTotal, tax)
	}

	subtotal, subtotalErr := calcInt64(netTotal)
	taxTotalYen, taxTotalErr := calcInt64(taxTotal)
	totalAmount, totalErr := calcInt64(new(big.Int).Add(netTotal, taxTotal))
	if subtotalErr != nil || taxTotalErr != nil || totalErr != nil {
		return CalcResult{}, apperrors.WrapCalculationUnsupported(calcTotalsOutOfRangeMessage)
	}
	result.Subtotal = subtotal
	result.TaxTotal = taxTotalYen
	result.TotalAmount = totalAmount
	return result, nil
}

// calcLineError は明細の位置（1 始まり）を添えた calculation_unsupported を返す。
func calcLineError(index int, message string) error {
	return apperrors.WrapCalculationUnsupported(fmt.Sprintf("明細%d: %s", index+1, message))
}

// calcTaxRate は外税の通常経路で使う明細の税率。rat は比較・計算用の有理数、key は同一税率を束ねる正規の10進文字列
// （0.1 と 0.10 は同じ key）、value は key から作った ExactNumeric。
type calcTaxRate struct {
	rat   *big.Rat
	key   string
	value model.ExactNumeric
}

// calcTaxGroupWork は税率ごとの集約の作業用。
type calcTaxGroupWork struct {
	rate    *big.Rat
	value   model.ExactNumeric
	taxable *big.Int
}

// calcExternalTaxRate は明細が外税の通常経路で算定できるかを検査し、税率を返す。
// 内税・非課税・税区分の未設定は規則が未決定のため拒否する（EMR-TAX-SCOPE-20261021）。
// 税率が nil の明細は未設定として拒否し、既定値（0.10）で補わない。
func calcExternalTaxRate(line CalcLine, index int) (calcTaxRate, error) {
	switch line.TaxType {
	case model.TaxTypeExcluded:
	case model.TaxTypeIncluded:
		return calcTaxRate{}, calcLineError(index, "内税の明細は算定できません（内税・内税と外税の混在の算定規則は未決定です）")
	case model.TaxTypeExempt:
		return calcTaxRate{}, calcLineError(index, "非課税の明細は算定できません（非課税の保持方法は未決定です）")
	default:
		return calcTaxRate{}, calcLineError(index, "税区分が未設定または不明です")
	}
	if line.TaxRate == nil {
		return calcTaxRate{}, calcLineError(index, "税率が未設定の明細は算定できません")
	}
	rate, ok := calcRat(*line.TaxRate)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(1, 1)) > 0 {
		return calcTaxRate{}, calcLineError(index, "税率が0以上1以下の有限の数値ではありません")
	}
	key, err := line.TaxRate.CanonicalString()
	if err != nil {
		return calcTaxRate{}, calcLineError(index, "税率を解釈できません")
	}
	value, err := model.ParseExactNumeric(key)
	if err != nil {
		return calcTaxRate{}, calcLineError(index, "税率を解釈できません")
	}
	return calcTaxRate{rat: rate, key: key, value: value}, nil
}

// calcLineAmounts は明細 1 行の金額を求める。税区分には依存しない。
// 行金額は単価 × 数量の 1 円未満四捨五入。値引きはその行金額を対象額として scope 内の順（割合 → 定額）で求める。
func calcLineAmounts(line CalcLine, index int) (CalcLineResult, error) {
	unit, ok := calcRat(line.UnitPrice)
	if !ok || unit.Sign() < 0 {
		return CalcLineResult{}, calcLineError(index, "単価が有限の0以上の数値ではありません")
	}
	quantity, ok := calcRat(line.Quantity)
	if !ok || quantity.Sign() <= 0 {
		return CalcLineResult{}, calcLineError(index, "数量が有限の正の数値ではありません")
	}
	percent, ok := calcRat(line.LinePercent)
	if !ok || percent.Sign() < 0 || line.LineFixed < 0 {
		return CalcLineResult{}, calcLineError(index, "値引きが有限の0以上の数値ではありません")
	}
	gross, err := calcInt64(ratRoundHalfUp(new(big.Rat).Mul(unit, quantity)))
	if err != nil {
		return CalcLineResult{}, calcLineError(index, "金額が扱える範囲を超えています")
	}
	requested, applied, err := calcDiscountScope(gross, percent, line.LineFixed)
	if err != nil {
		return CalcLineResult{}, calcLineError(index, "値引き額が扱える範囲を超えています")
	}
	return CalcLineResult{
		GrossAmount:           gross,
		LineDiscountRequested: requested,
		LineDiscountApplied:   applied,
		AfterLineDiscount:     gross - applied,
	}, nil
}

// calcDiscountScope は 1 つの scope の値引きを「割合（合算済み）→ 1 円未満四捨五入 → 定額」の順で求める。
// requested は要求額、applied は対象額 base を上限にした実適用額で、超過分は他へ回さない。
func calcDiscountScope(base int64, percent *big.Rat, fixed int64) (requested, applied int64, err error) {
	amount := new(big.Rat).SetInt64(base)
	amount.Mul(amount, percent)
	amount.Quo(amount, big.NewRat(100, 1))
	percentAmount := ratRoundHalfUp(amount)
	requested, err = calcInt64(new(big.Int).Add(percentAmount, big.NewInt(fixed)))
	if err != nil {
		return 0, 0, err
	}
	percentApplied := base
	if percentAmount.Cmp(big.NewInt(base)) < 0 {
		percentApplied = percentAmount.Int64()
	}
	fixedApplied := min(fixed, base-percentApplied)
	return requested, percentApplied + fixedApplied, nil
}

// allocateLargestRemainder は値引き額 applied を、基準額 bases に比例して最大剰余法で配賦する。
// 各基準額の商（D × A_i ÷ S）の整数部をまず配り、残りの円を剰余の降順で 1 円ずつ配る。同じ剰余は bases の並び
// （凍結した明細順）の昇順で配る。基準額 0 の行へは配らず、各配賦額は基準額を超えず、合計は applied に一致する。
// applied は基準額の合計以下であること（呼出し側が対象額で上限を掛ける）。基準額の合計が 0 なら全て 0。
func allocateLargestRemainder(applied int64, bases []int64) []int64 {
	shares := make([]int64, len(bases))
	total := new(big.Int)
	for _, base := range bases {
		total.Add(total, big.NewInt(base))
	}
	if applied <= 0 || total.Sign() == 0 {
		return shares
	}
	type remainder struct {
		index int
		value *big.Int
	}
	var remainders []remainder
	var distributed int64
	for i, base := range bases {
		if base == 0 {
			continue
		}
		product := new(big.Int).Mul(big.NewInt(applied), big.NewInt(base))
		quotient, rest := new(big.Int).QuoRem(product, total, new(big.Int))
		shares[i] = quotient.Int64()
		distributed += shares[i]
		if rest.Sign() > 0 {
			remainders = append(remainders, remainder{index: i, value: rest})
		}
	}
	sort.SliceStable(remainders, func(a, b int) bool { return remainders[a].value.Cmp(remainders[b].value) > 0 })
	for k := 0; k < len(remainders) && int64(k) < applied-distributed; k++ {
		shares[remainders[k].index]++
	}
	return shares
}

// calcRat は有限な ExactNumeric を無損失の有理数にする。NULL・NaN・±Infinity・指数が上限超の値は ok=false。
// 変換は ExactNumeric.CanonicalString（指数表記なしの正規10進）経由で、float64 を使わない。
func calcRat(value model.ExactNumeric) (*big.Rat, bool) {
	text, err := value.CanonicalString()
	if err != nil {
		return nil, false
	}
	return new(big.Rat).SetString(text)
}

// calcInt64 は int64 に収まる整数だけを返す。
func calcInt64(value *big.Int) (int64, error) {
	if !value.IsInt64() {
		return 0, errCalcOutOfRange
	}
	return value.Int64(), nil
}

// ratRoundHalfUp は非負の有理数を 1 円未満四捨五入（0.5 は切上げ）した整数を返す。floor(r + 1/2) = floor((2n + d) / 2d)。
func ratRoundHalfUp(value *big.Rat) *big.Int {
	numerator := new(big.Int).Lsh(value.Num(), 1)
	numerator.Add(numerator, value.Denom())
	denominator := new(big.Int).Lsh(value.Denom(), 1)
	return numerator.Div(numerator, denominator)
}

// ratFloor は非負の有理数の 1 円未満を切り捨てた整数を返す。
func ratFloor(value *big.Rat) *big.Int {
	return new(big.Int).Div(value.Num(), value.Denom())
}
