# TASK-184: accounting_response.go で Handler 層から service.CalculateTaxAmount を直接呼び出し

## 優先度: Medium

## 概要
`toBillingItemResponse` 関数（Handler 層の response 変換ヘルパー）が
`service.CalculateTaxAmount` を直接呼び出している。
Handler 層は Service 層に依存してよいが、「レスポンス変換のみのために」
Service 関数を呼ぶことは Handler 責務の逸脱であり、
また Handler → Service の依存方向を response 変換ヘルパーに持ち込む設計は
アーキテクチャの层間依存を不必要に複雑にする。

## 対象ファイル
`backend/internal/handler/accounting_response.go`

## 現状コード（L166〜185）

```go
func toBillingItemResponse(item *model.BillingItem) billingItemResponse {
    subtotal := int64(float64(item.UnitPrice) * item.Quantity)
    // ❌ Handler 層の response 変換ヘルパーから service 関数を直接呼び出し
    taxAmount := service.CalculateTaxAmount(item.UnitPrice, item.Quantity, item.TaxType, item.TaxRate)
    return billingItemResponse{
        ...
        TaxAmount: taxAmount,
        ...
    }
}
```

## 問題の詳細
- `service.CalculateTaxAmount` は Service パッケージの公開関数であり、
  Handler 層のレスポンス変換ヘルパーが直接 import して使用している
- この計算ロジックは `model.BillingItem` を入力とするため、
  `model` パッケージの純粋な計算関数として分離するか、
  またはレスポンス変換時に使えるユーティリティとして整理すべき
- 現状では Handler 層が Service 層ロジックに直接依存しており、
  テスト時の service パッケージ依存が増える

## 修正案

### 案A: model パッケージにメソッドとして移動（推奨）

```go
// model/accounting.go に追加
func (item *BillingItem) CalculateTaxAmount() int64 {
    subtotal := float64(item.UnitPrice) * item.Quantity
    switch item.TaxType {
    case TaxTypeExcluded:
        return int64(math.Round(subtotal * item.TaxRate))
    case TaxTypeIncluded:
        return int64(math.Round(subtotal * item.TaxRate / (1 + item.TaxRate)))
    default:
        return 0
    }
}
```

```go
// handler/accounting_response.go での使用
func toBillingItemResponse(item *model.BillingItem) billingItemResponse {
    subtotal := int64(float64(item.UnitPrice) * item.Quantity)
    taxAmount := item.CalculateTaxAmount()  // ✅ model のメソッドを使用
    return billingItemResponse{...}
}
```

### 案B: service import を維持しつつ local 関数に切り出し

`service.CalculateTaxAmount` と同じロジックを `accounting_response.go` 内の
unexported な計算関数として複製するか、
`service.CalculateTaxAmount` をより下位の `model` や共通 util に移動する。

## 優先対応
案A（model メソッド化）を推奨。`service.billing_service.go` の `CalculateTaxAmount` も
model メソッドに移行し、service からは model のメソッドを呼ぶようにする。

## 影響範囲
- `backend/internal/handler/accounting_response.go`
- `backend/internal/service/billing_service.go`（参照元）
- `backend/internal/model/accounting.go`（追加先）
