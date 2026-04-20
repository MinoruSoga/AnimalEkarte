# CODE-QUALITY-206: medicine / merchandise_item の TaxRate ゼロ値が無視される

## 概要

`medicine_service.go` と `merchandise_item_service.go` の `Create` メソッドで、
`input.TaxRate != 0` によるゼロ値チェックを使っているため、
クライアントが意図的に `0`（非課税）を送信しても `0.10` に上書きされてしまう。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/service/medicine_service.go` | L127 |
| `backend/internal/service/merchandise_item_service.go` | L127 |

---

## 問題

### medicine_service.go:127

```go
taxRate := 0.10
if input.TaxRate != 0 {    // ← 0 を意図して送っても 0.10 になる
    taxRate = input.TaxRate
}
```

`CreateMedicineInput.TaxRate` が `float64`（非ポインタ）のため、
「ユーザーが 0 を指定した」のか「省略したため Go のゼロ値になった」のかを区別できない。
税率 0% の薬品（非課税）を登録しようとすると 10% に上書きされる。

### merchandise_item_service.go:127

同じ構造の問題が存在する。

---

## 修正方針

### Step 1: `CreateMedicineInput.TaxRate` を `*float64` に変更

```go
// service/medicine_service.go
type CreateMedicineInput struct {
    Name        string
    Price       int64
    TaxType     string   // 既に存在
    TaxRate     *float64 // float64 → *float64 に変更
    // ...
}

// デフォルト適用ロジックを修正
taxRate := 0.10  // デフォルト
if input.TaxRate != nil {
    taxRate = *input.TaxRate
}
```

### Step 2: Handler の Request → Input DTO 変換を修正

```go
// handler/medicine_handler.go — Create の svcInput 構築箇所
svcInput := &service.CreateMedicineInput{
    // ...
    TaxRate: input.TaxRate,  // *float64 のままで渡す
}
```

`createMedicineRequest.TaxRate` が `*float64` であれば変換不要。
`float64` であれば `&req.TaxRate`（ただし省略時のゼロ値と区別できなくなるため request 型も `*float64` に変更が必要）。

### Step 3: merchandise_item_service.go も同様に修正

```go
type CreateMerchandiseItemInput struct {
    // ...
    TaxRate *float64  // float64 → *float64
}
```

---

## 注意事項

- `createMedicineRequest.TaxRate` および `createMerchandiseItemRequest.TaxRate` の型も確認し、
  `*float64` に統一されているか確認すること
- `omitempty` JSON タグがある場合、`0.0` がシリアライズ時に省略されるため
  クライアント側の仕様も確認すること
- 既存データへの影響はなし（Create のみの問題）

---

## 規約参照

- `.claude/rules/go-language.md`: GORM PATCH（ポインタ型 + buildUpdateFields）パターン
- ポインタ型による nil / ゼロ値の区別はプロジェクト全体の基本方針

## テスト

- `TaxRate: 0.0` を明示的に送信した場合に、DB に `0.0` が保存されることを確認
- `TaxRate` を省略した場合に `0.10` がデフォルト適用されることを確認
