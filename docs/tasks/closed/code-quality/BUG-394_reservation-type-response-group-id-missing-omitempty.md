# BUG-394: reservation_type_response の GroupID が omitempty タグなし（null が JSON に露出）

## 概要
`reservation_type_response.go` の `reservationTypeResponse` 構造体で `GroupID *uint64` フィールドに `omitempty` タグがない。`GroupID` が NULL の場合も `"group_id": null` としてレスポンスに含まれる。プロジェクト内の他レスポンス構造体は `*uint64` に `omitempty` を付けており、不統一。

## 再現手順
1. グループに属さない予約種別を取得する `GET /masters/reservation-types/:id`
2. **結果**: `"group_id": null` が含まれる
3. **期待**: `group_id` キー自体が省略される（他の nullable フィールドと同様）

## 現状コード

### `backend/internal/handler/reservation_type_response.go:82`
```go
type reservationTypeResponse struct {
    ID          uint64                        `json:"id"`
    ...
    GroupID     *uint64                       `json:"group_id"`         // ← omitempty なし
    Group       *groupSummary                 `json:"group,omitempty"`  // ← omitempty あり（不統一）
    ...
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// 他レスポンス構造体では nullable な uint64 ポインタに omitempty を付けている
ParentID *uint64 `json:"parent_id,omitempty"`  // exam_type_response.go
Price    *int64  `json:"price,omitempty"`       // cage_response.go
```

## 修正方針

### `reservation_type_response.go:82` — omitempty 追加
```go
GroupID *uint64 `json:"group_id,omitempty"`
```

**注意**: フロントエンドが `group_id: null` を明示的に参照している場合は影響を確認すること。`omitempty` 適用後は `group_id` キー自体が省略されるため、`null` と `undefined` の区別に依存するコードは修正が必要。

## 優先度
**Low** — 機能上の問題なし。レスポンスに不要なフィールドが含まれる。他フィールドとの統一性問題。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/reservation_type_response.go:82` — 修正対象
