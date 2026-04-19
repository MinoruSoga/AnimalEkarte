# BUG-393: マスタ Create リクエストの必須 ENUM フィールドが omitempty タグを使用している

## 概要
`procedure_request.go` と `reservation_type_request.go` の Create リクエスト構造体で、ビジネス上必須な ENUM フィールドに `binding:"omitempty,oneof=..."` タグが付いている。`omitempty` は値が空文字の場合にバリデーションをスキップするため、空文字列が通過して DB の ENUM 制約違反が発生し得る。Update（PATCH）用構造体は `omitempty` が適切だが、Create では `required` または `required,oneof=...` を使うべきだ。

## 再現手順
1. `POST /masters/procedures` に `{"name": "test", "anesthesia": "", "tax_type": ""}` を送信
2. **結果**: バリデーションを通過し（omitempty で空文字スキップ）、DB に空文字が挿入される
3. **期待**: 400 Bad Request が返る

## 現状コード

### 問題1: `backend/internal/handler/procedure_request.go:1-17`
```go
type CreateProcedureRequest struct {
    Name        string   `json:"name"        binding:"required"`
    Price       *int64   `json:"price"`
    Description string   `json:"description"`
    Duration    *int     `json:"duration"`
    Anesthesia  string   `json:"anesthesia"  binding:"omitempty,oneof=none local sedation general"`
    // ↑ Create なのに omitempty。空文字が通過する
    ParentID    *uint64  `json:"parent_id"`
    SortOrder   *int     `json:"sort_order"`
    TaxType     string   `json:"tax_type"    binding:"omitempty,oneof=included excluded exempt"`
    // ↑ 同上
    TaxRate     *float64 `json:"tax_rate"`
    IsActive    *bool    `json:"is_active"`
}
```

### 問題2: `backend/internal/handler/reservation_type_request.go:1-15`
```go
type CreateReservationTypeRequest struct {
    Name     string  `json:"name"     binding:"required"`
    Color    string  `json:"color"    binding:"required"`
    Category string  `json:"category" binding:"omitempty,oneof=general trimming"`
    // ↑ カテゴリは予約種別のコア属性。Create なのに omitempty
    ...
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// Update リクエストは omitempty が適切（PATCH セマンティクス）
type UpdateProcedureRequest struct {
    Anesthesia *string `json:"anesthesia" binding:"omitempty,oneof=none local sedation general"`
    TaxType    *string `json:"tax_type"   binding:"omitempty,oneof=included excluded exempt"`
}
```

## 影響範囲

| 対象ファイル | フィールド | 問題 |
|------------|---------|------|
| `procedure_request.go:9` | `Anesthesia string` | `omitempty` → `required` または `required,oneof=...` |
| `procedure_request.go:12` | `TaxType string` | 同上 |
| `reservation_type_request.go:10` | `Category string` | 同上 |

## 修正方針

### `procedure_request.go:9,12` — binding タグ修正
```go
type CreateProcedureRequest struct {
    Name        string   `json:"name"        binding:"required"`
    ...
    Anesthesia  string   `json:"anesthesia"  binding:"required,oneof=none local sedation general"`
    // ↑ omitempty → required に変更
    ...
    TaxType     string   `json:"tax_type"    binding:"required,oneof=included excluded exempt"`
    // ↑ 同上
    ...
}
```

### `reservation_type_request.go:10` — binding タグ修正
```go
type CreateReservationTypeRequest struct {
    ...
    Category string `json:"category" binding:"required,oneof=general trimming"`
    // ↑ omitempty → required に変更
    ...
}
```

**注意**: `Anesthesia` と `TaxType` が省略可能なビジネス要件であれば、`*string` 型（ポインタ）に変更したうえで Update 側と同様の `omitempty` を維持する。要件を確認してから修正すること。

## 準拠すべきプロジェクト規約・ベストプラクティス

### プロジェクト標準パターン
- Create リクエスト: 必須フィールドは `binding:"required"` または `binding:"required,oneof=..."`
- Update リクエスト（PATCH）: `*型 + binding:"omitempty,oneof=..."` でポインタ型を使用

## 優先度
**Medium** — 空文字列が DB に挿入されると ENUM 制約違反で 500 が発生するか、無効な値がデータとして保存される。フロントエンドが常に正しい値を送信している場合は問題が顕在化しないが、API 直接呼び出し時にサイレント障害になる。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/procedure_request.go:9,12` — 修正対象
- `backend/internal/handler/reservation_type_request.go:10` — 修正対象
