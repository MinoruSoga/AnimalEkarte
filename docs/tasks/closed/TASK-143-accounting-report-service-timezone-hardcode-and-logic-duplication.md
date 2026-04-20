# TASK-143: [Medium] accounting_report_service.go — タイムゾーンハードコードと税率集計ロジックの重複

## 優先度
**Medium**（保守性・一貫性）

## 対象ファイル
- `backend/internal/service/accounting_report_service.go`
- `backend/internal/repository/accounting_repository.go`

## 問題 1: タイムゾーン "Asia/Tokyo" のハードコード散在

`"Asia/Tokyo"` 文字列と `time.FixedZone("Asia/Tokyo", 9*60*60)` が複数ファイル・複数箇所に直接記述されている。

### 該当箇所

`service/accounting_report_service.go` L198:
```go
jst := time.FixedZone("Asia/Tokyo", 9*60*60)
```

`repository/accounting_repository.go` L295, L555:
```go
jst := time.FixedZone("Asia/Tokyo", 9*60*60)
```

SQL クエリ内にも多数:
```go
"DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?"
"billings.completed_at AT TIME ZONE 'Asia/Tokyo' >= ?"
```

### 修正後コード

```go
// service/accounting_report_service.go（または shared な定数ファイル）に定数として定義

const timeZoneJST = "Asia/Tokyo"

var jst = time.FixedZone(timeZoneJST, 9*60*60)
```

各使用箇所を定数参照に置き換える:
```go
// ❌ 現状
jst := time.FixedZone("Asia/Tokyo", 9*60*60)

// ✅ 修正後（パッケージ変数参照）
// jst はパッケージ変数から参照
```

---

## 問題 2: 税率集計ロジックの重複

`buildTaxBreakdown` 関数（L273〜284）が存在するにもかかわらず、
`GetMonthly` 内（L185〜195）で同じ判定ロジック（`tr.TaxRate > 8` で標準/軽減を振り分け）を
インラインで再実装している。

### 現状コード（service/accounting_report_service.go L185〜195）

```go
// 税率別集計を TaxBreakdownSummary に変換
var taxSummary TaxBreakdownSummary
for _, tr := range raw.TaxBreakdown {
    if tr.TaxRate > 8 {
        // 標準税率（10%）
        taxSummary.Standard.TaxableAmount += tr.TaxableAmount
        taxSummary.Standard.TaxAmount += tr.TaxAmount
    } else {
        // 軽減税率（8%以下）
        taxSummary.Reduced.TaxableAmount += tr.TaxableAmount
        taxSummary.Reduced.TaxAmount += tr.TaxAmount
    }
}
```

### 修正後コード（L185〜195 を buildTaxBreakdown に置き換え）

```go
// 税率別集計を TaxBreakdownSummary に変換
taxSummary := buildTaxBreakdown(raw.TaxBreakdown)
```

`buildTaxBreakdown` 関数（L273〜284）は既に正しく実装されているため、インライン実装を削除するだけでよい。

---

## 問題 3: `paymentMethodNameForClose` は `resolvePaymentMethodName` の単純ラッパーで不要

```go
// L267〜269（不要なラッパー）
func paymentMethodNameForClose(id *uint64, names map[uint64]string) string {
    return resolvePaymentMethodName(id, names)
}
```

`cash_register_service.go` から `resolvePaymentMethodName` を直接呼び出すか、
`paymentMethodNameForClose` を削除して呼び出し側を `resolvePaymentMethodName` に統一する。

### 修正後

```go
// paymentMethodNameForClose を削除し、呼び出し側を直接変更
// cash_register_service.go 内:
// ❌ paymentMethodNameForClose(id, names)
// ✅ resolvePaymentMethodName(id, names)
```
