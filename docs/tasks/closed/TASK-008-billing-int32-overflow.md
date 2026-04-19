# TASK-008: Billing モデルの金額フィールドが int (32bit) — int64 に修正

## 概要

`Billing` モデルの `Subtotal`, `TaxTotal`, `TotalAmount` フィールドが Go の `int`（32bit, 最大約 21億）で定義されているが、DB スキーマは `bigint`（64bit）。高額請求（複数のペット・長期入院等）で integer overflow が発生する可能性がある。

## 優先度

CRITICAL（データ整合性・サイレント overflow リスク）

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/model/accounting.go` | L60-62（Subtotal / TaxTotal / TotalAmount が `int`） |

## 規約違反

`.claude/rules/go-language.md`:
> 金額型 bigint → `int64`

## 現状コード

```go
// accounting.go（現状）
type Billing struct {
    // ...
    Subtotal    int `gorm:"default:0" json:"subtotal"`
    TaxTotal    int `gorm:"default:0" json:"tax_total"`
    TotalAmount int `gorm:"default:0" json:"total_amount"`
}
```

## 修正方針

```go
// accounting.go（修正後）
type Billing struct {
    // ...
    Subtotal    int64 `gorm:"default:0" json:"subtotal"`
    TaxTotal    int64 `gorm:"default:0" json:"tax_total"`
    TotalAmount int64 `gorm:"default:0" json:"total_amount"`
}
```

## 波及確認

1. `make codegen` 実行 → `models.ts` の型が `number` のまま変わらないことを確認（TypeScript では `number` で問題なし）
2. サービス層・ハンドラ層で `int` 型と比較・演算している箇所を `int64` に統一
3. `BillingItem.UnitPrice`, `EstimateItem.UnitPrice` 等の関連フィールドも同様に確認

## テスト

- 大金額（21億超）を登録・取得しても値が正しく保持されることを確認
