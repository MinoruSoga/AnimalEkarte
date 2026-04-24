# BUG-425: UpdateReservationTypeInput に ClearGroupID フラグが欠落（グループ解除不可）

## 概要

`reservation_type_service.go` の `UpdateReservationTypeInput` に `GroupID *uint64` は存在するが、
`ClearGroupID bool` フラグが存在しない。そのため予約種別のグループ所属を意図的に
NULL（グループなし）に更新する方法がなく、グループへの紐付けが解除できない。

## 問題箇所

```go
// reservation_type_service.go:40-59
type UpdateReservationTypeInput struct {
    Name        *string
    Color       *string
    IsActive    *bool
    Description *string
    SortOrder   *int
    GroupID     *uint64  // ← GroupID を nil にしても「変更しない」扱いになる
    Category    *string
    // ClearGroupID bool ← 存在しない
}
```

```go
// reservation_type_service.go:83-134（buildReservationTypeUpdateFields）
func buildReservationTypeUpdateFields(input *UpdateReservationTypeInput) map[string]any {
    fields := make(map[string]any)
    // ...
    if input.GroupID != nil {
        fields["group_id"] = *input.GroupID  // ← nil なら何もしない（クリアできない）
    }
    // ClearGroupID の処理がない
    return fields
}
```

## 他マスタとの比較（正しい実装例）

```go
// exam_type_service.go — ParentID をクリア可能な設計
type UpdateExaminationTypeInput struct {
    ParentID      *uint64
    ClearParentID bool    // ✅ NULL 設定フラグあり
}

// buildUpdateFields での処理
if input.ClearParentID {
    fields["parent_id"] = nil
} else if input.ParentID != nil {
    fields["parent_id"] = *input.ParentID
}
```

同様に medicine_service, vaccine_service, procedure_service, checkup_type_service でも
`ClearParentID` フラグが実装済み。

## 修正方針

### 1. `UpdateReservationTypeInput` に `ClearGroupID` フラグを追加

```go
type UpdateReservationTypeInput struct {
    // ...
    GroupID      *uint64
    ClearGroupID bool    // true の場合 group_id を NULL にする
}
```

### 2. `buildReservationTypeUpdateFields` に処理を追加

```go
if input.ClearGroupID {
    fields["group_id"] = nil
} else if input.GroupID != nil {
    fields["group_id"] = *input.GroupID
}
```

### 3. `reservation_type_request.go` の updateReservationTypeRequest にも追加

```go
type updateReservationTypeRequest struct {
    // ...
    GroupID      *uint64 `json:"group_id"`
    ClearGroupID bool    `json:"clear_group_id"`
}
```

### 4. ハンドラ側の Input 変換にも追加

```go
// reservation_type_handler.go の Update
svcInput := &service.UpdateReservationTypeInput{
    // ...
    GroupID:      req.GroupID,
    ClearGroupID: req.ClearGroupID,
}
```

## 影響ファイル

- `backend/internal/service/reservation_type_service.go` — UpdateReservationTypeInput（行 40-59）、buildReservationTypeUpdateFields（行 83-134）
- `backend/internal/handler/reservation_type_request.go` — updateReservationTypeRequest
- `backend/internal/handler/reservation_type_handler.go` — Update ハンドラ

## 優先度

**Medium** — 予約種別のグループ所属が一度設定したら解除できないという機能欠陥。

## 関連チケット

- BUG-424（reservation_type_service.Update の存在確認欠落）
