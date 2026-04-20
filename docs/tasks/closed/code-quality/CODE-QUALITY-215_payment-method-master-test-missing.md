# CODE-QUALITY-215: payment_method_master テストファイル未実装

## 概要

`payment_method_master` ドメインのテストカバレッジがほぼゼロ。
他の master ドメイン（animal_species, cage, checkup_type 等）では
`handler_test.go` / `service_test.go` が実装されているのに対し、
`payment_method_master` のみ handler テストが完全欠落、service テストも Update が欠落している。

## 該当ファイル

| ファイル | 状態 | 内容 |
|----------|------|------|
| `backend/internal/handler/payment_method_master_handler_test.go` | **未作成** | handler 層テスト完全欠落 |
| `backend/internal/service/payment_method_master_service_test.go` | 実装済みだが不完全 | Update テストケースが未実装（List, GetByID, Create, Delete, Reorder の5件のみ） |

## 不整合の詳細

### handler_test.go 欠落

`ls backend/internal/handler/payment_method_master*` の出力:
```
payment_method_master_handler.go
payment_method_master_request.go
payment_method_master_response.go
```

`payment_method_master_handler_test.go` が存在しない。
同ドメインの他ファイル（animal_species_handler_test.go, cage_handler_test.go 等）は全て実装済み。

### service_test.go の Update テスト欠落

`payment_method_master_service_test.go` には以下のテストしか存在しない:
- `TestPaymentMethodMasterService_List`
- `TestPaymentMethodMasterService_GetByID`
- `TestPaymentMethodMasterService_Create`
- `TestPaymentMethodMasterService_Delete`
- `TestPaymentMethodMasterService_Reorder`

`TestPaymentMethodMasterService_Update` が欠落している。
`Update` は `UpdatePaymentMethodMasterInput` + `buildPaymentMethodMasterUpdateFields` を通じた
PATCH パターンを持つため、ポインタフィールドのゼロ値スキップ等の検証が必要。

## 修正方針

### 1. `payment_method_master_handler_test.go` 新規作成

参照実装: `backend/internal/handler/animal_species_handler_test.go`

実装すべきテストケース:
- `TestPaymentMethodMasterHandler_List` — 正常, サービスエラー
- `TestPaymentMethodMasterHandler_GetByID` — 正常, 404, 不正ID
- `TestPaymentMethodMasterHandler_Create` — 正常, バリデーションエラー, バインドエラー
- `TestPaymentMethodMasterHandler_Update` — 正常, 404, 不正ID, バインドエラー
- `TestPaymentMethodMasterHandler_Delete` — 正常, 404, 409(依存チェック)
- `TestPaymentMethodMasterHandler_Reorder` — 正常, バリデーションエラー

### 2. `payment_method_master_service_test.go` に Update テスト追加

```go
func TestPaymentMethodMasterService_Update(t *testing.T) {
    tests := []struct {
        name    string
        input   *UpdatePaymentMethodMasterInput
        wantErr bool
    }{
        {
            name:    "名前のみ更新",
            input:   &UpdatePaymentMethodMasterInput{Name: ptr("現金")},
            wantErr: false,
        },
        {
            name:    "空文字名前はエラー",
            input:   &UpdatePaymentMethodMasterInput{Name: ptr("")},
            wantErr: true,
        },
        {
            name:    "input nil はエラー",
            input:   nil,
            wantErr: true,
        },
        // ... フィールドなしはエラー等
    }
}
```

## 優先度

HIGH — テストカバレッジの欠落は回帰検知の盲点を生む。
`Update` は PATCH パターン（ポインタ型 + buildUpdateFields）のため
フィールド欠落バグが最もテストで検出すべき箇所。

## 参照

- `backend/internal/handler/animal_species_handler_test.go` — handler テスト参照実装
- `backend/internal/service/animal_species_service_test.go` — service テスト参照実装
- CODE-QUALITY-204 — payment_method_master 全体品質（別票）
