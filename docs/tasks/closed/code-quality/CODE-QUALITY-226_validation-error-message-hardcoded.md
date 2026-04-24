# CODE-QUALITY-226: バリデーションエラーメッセージのハードコード文字列（ErrMsg 定数未使用）

## 概要

`backend/internal/service/validators.go` に `ErrMsg*` 定数が定義されているにもかかわらず、
複数のサービスファイルで同じ意味のエラーメッセージをハードコード文字列で記述している。
さらに、同じ検証ロジック（金額 ≥ 0 チェック等）が複数ファイルに重複している。

## 定義済み定数（validators.go:19-23）

```go
const (
    ErrMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
    ErrMsgIDsNotEmpty     = "並び順のIDリストが空です"
    ErrMsgInputNotNil     = "更新内容が指定されていません"
)
```

## ハードコード文字列の実態

### 金額バリデーション — 3ファイルで重複・文言もブレ

| ファイル | 行番号 | ハードコード文字列 |
|---------|--------|----------------|
| `billing_item_service.go` | ~96, ~171 | `"金額は0以上を入力してください"` |
| `merchandise_item_service.go` | ~119, ~163 | `"金額は0以上を入力してください"` |
| `treatment_service.go` | ~112 | `"金額は0以上を入力してください"` |
| `accounting_service.go` | ~103, ~135 | `"金額は0以上で指定してください"` ← 文言がブレている |

同じ意味なのに「入力してください」vs「で指定してください」と表現が異なる。

### 数量バリデーション — 重複

| ファイル | 行番号 | ハードコード文字列 |
|---------|--------|----------------|
| `treatment_service.go` | ~115, ~214 | `"数量は0より大きい値を入力してください"` |

### 体重バリデーション — 重複

| ファイル | 行番号 | ハードコード文字列 |
|---------|--------|----------------|
| `pet_service.go` | ~141, ~232 | `"体重は0以上の値を入力してください"` |

### バイタルバリデーション

| ファイル | 行番号 | ハードコード文字列 |
|---------|--------|----------------|
| `vital_service.go` | ~68 | `"少なくとも1つのバイタル値を入力してください"` |

これは `ErrMsgAtLeastOneField` と同義だが別文言で記述されている。

### permission_group の特殊バリデーション

| ファイル | 行番号 | ハードコード文字列 |
|---------|--------|----------------|
| `permission_group_service.go` | ~191 | `"リソース名が空です"` |

## 修正方針

### 1. validators.go に定数を追加

```go
const (
    // 既存
    ErrMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
    ErrMsgIDsNotEmpty     = "並び順のIDリストが空です"
    ErrMsgInputNotNil     = "更新内容が指定されていません"
    // 追加
    ErrMsgPriceZeroOrMore    = "金額は0以上を入力してください"
    ErrMsgQuantityPositive   = "数量は0より大きい値を入力してください"
    ErrMsgWeightZeroOrMore   = "体重は0以上の値を入力してください"
    ErrMsgResourceNameEmpty  = "リソース名が空です"
)
```

### 2. 各サービスのハードコード文字列を定数に置換

例:
```go
// ❌ 修正前
return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")

// ✅ 修正後
return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
```

### 3. vital_service.go の修正

```go
// ❌ 修正前
return nil, apperrors.WrapInvalidInput("少なくとも1つのバイタル値を入力してください")

// ✅ 修正後（既存定数に合わせるか、vitals 専用定数を追加）
return nil, apperrors.WrapInvalidInput("少なくとも1つのバイタル値を入力してください") // 専用定数として維持
```

### 4. accounting_service.go の文言統一

`"金額は0以上で指定してください"` → `ErrMsgPriceZeroOrMore` に統一。

## 優先度

MEDIUM — 機能上の問題はないが、同じ意味のメッセージが複数文言で存在すると
ユーザー体験が一貫しない。また定数を使わないことでコードの意図が不明確になる。

文言の統一はフロントエンドの表示にも影響するため、
修正時はフロントエンドとの連携確認が必要。
