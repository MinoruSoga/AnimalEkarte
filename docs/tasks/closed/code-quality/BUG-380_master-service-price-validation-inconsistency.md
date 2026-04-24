# BUG-380: 価格フィールドの負値バリデーションがマスタサービス間で不統一

## 概要
`merchandise_item_service.go` のみ `UnitPrice < 0` の負値バリデーションを実装しているが、同様に価格フィールドを持つ `procedure_service.go`・`medicine_service.go`・`vaccine_service.go` では負値チェックが存在しない。バリデーションの抜けにより、負の価格が DB に登録される可能性がある。

## 再現手順
1. `POST /v1/masters/procedures` に `{"name": "test", "price": -1000}` を送信
2. **結果**: バリデーションエラーなしで 201 Created、DB に `-1000` が登録される
3. 比較: `POST /v1/masters/merchandise-items` に `{"name": "test", "unit_price": -1}` → 400 Bad Request が返る

## 期待する動作
- 価格・単価フィールドは全マスタサービスで `>= 0` バリデーションを実施すること
- `validators.go` の共通関数として `validateNonNegativePrice` を定義し統一すること

## 現状コード

### `backend/internal/service/merchandise_item_service.go:118,157`（バリデーションあり）
```go
// Create
if input.UnitPrice < 0 {
    return nil, apperrors.WrapInvalidInput("単価は0以上を指定してください")
}

// Update
if input.UnitPrice != nil && *input.UnitPrice < 0 {
    return nil, apperrors.WrapInvalidInput("単価は0以上を指定してください")
}
```

### `backend/internal/service/procedure_service.go`（バリデーションなし）
```go
// Create — price に負値チェックなし
func (s *procedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    // price のバリデーションは存在しない
    // ...
}
```

### `backend/internal/service/validators.go`（共通バリデーターは未定義）
```go
// validateDiscountRate は実装済み（108行目）
func validateDiscountRate(rate float64) error { ... }

// validateNonNegativePrice は未定義
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/service/validators.go` | validateNonNegativePrice 追加 | 要追加 |
| `backend/internal/service/procedure_service.go` | Create/Update に price 負値チェック | 要追加 |
| `backend/internal/service/medicine_service.go` | price フィールドの負値チェック確認 | 要調査・追加 |
| `backend/internal/service/vaccine_service.go` | price フィールドの負値チェック確認 | 要調査・追加 |

## 修正方針

### 1. `backend/internal/service/validators.go` — 共通関数を追加
```go
// validateNonNegativePrice は価格が 0 以上であることを検証する
func validateNonNegativePrice(price int64) error {
    if price < 0 {
        return apperrors.WrapInvalidInput("価格は0以上を指定してください")
    }
    return nil
}

// validateNonNegativePricePtr は nil 許容の価格バリデーション（PATCH用）
func validateNonNegativePricePtr(price *int64) error {
    if price != nil && *price < 0 {
        return apperrors.WrapInvalidInput("価格は0以上を指定してください")
    }
    return nil
}
```

### 2. 各サービスの Create/Update で共通関数を呼び出す（例: procedure_service.go）
```go
func (s *procedureService) Create(ctx context.Context, clinicID uint64, input *CreateProcedureInput) (*model.Procedure, error) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    if err := validateNonNegativePrice(input.Price); err != nil { // 追加
        return nil, err
    }
    // ...
}
```

### 3. `merchandise_item_service.go` のインライン実装を共通関数呼び出しに置換
```go
// 修正前
if input.UnitPrice < 0 {
    return nil, apperrors.WrapInvalidInput("単価は0以上を指定してください")
}

// 修正後
if err := validateNonNegativePrice(input.UnitPrice); err != nil {
    return nil, err
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バリデーションの統一
> Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。バリデーションエラーは `apperrors.WrapInvalidInput()` で統一。

### `.claude/rules/go-language.md` — インターフェース設計
> 共通ロジックは validators.go に集約し、各サービスは共通関数を呼び出す。

### プロジェクト内参照実装
`backend/internal/service/validators.go` — `validateRequiredName`, `validateDiscountRate` が共通バリデーターとして実装済み

## 優先度
**Medium** — 負の価格が DB に登録されるバグ。決済・請求処理で意図しない挙動を引き起こすリスクがある。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/validators.go` — 共通バリデーター追加先
- `backend/internal/service/merchandise_item_service.go:118,157` — 参照実装
- `backend/internal/service/procedure_service.go` — 修正対象
- `backend/internal/service/medicine_service.go` — 調査・修正対象
- `backend/internal/service/vaccine_service.go` — 調査・修正対象
