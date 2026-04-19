# TASK-068: reservation_type — Category フィールドが API 層全体で欠落（Create / Update / Response）

## 優先度

HIGH

---

## 概要

TASK-061（response に Category 欠落）の調査過程で判明した追加問題。
`model.ReservationType.Category`（`general` | `trimming`）は DB に存在するが、
**Create・Update・Response の全レイヤーで Category が欠落**しており、API から Category を参照・設定する手段が一切ない。

---

## 欠落箇所一覧

| 層 | ファイル | 問題 |
|----|---------|------|
| Request | `reservation_type_request.go` | `createReservationTypeRequest` に Category フィールドなし |
| Request | `reservation_type_request.go` | `updateReservationTypeRequest` に Category フィールドなし |
| Service Input | `reservation_type_service.go` | `CreateReservationTypeInput` に Category フィールドなし |
| Service Input | `reservation_type_service.go` | `UpdateReservationTypeInput` に Category フィールドなし |
| Response | `reservation_type_response.go` | `reservationTypeResponse` に Category フィールドなし |

---

## 影響

- 予約種別を作成するとき、常に `category = 'general'`（DB デフォルト値）になる
- `trimming` カテゴリの予約種別を API 経由で作成・更新できない
- クライアントは取得した予約種別が `general` か `trimming` かを判断できない
- フロントエンドがトリミング予約と通常診療予約を区別する UI を実装できない

---

## 修正方針

### Step 1: Request に Category を追加

```go
// reservation_type_request.go
type createReservationTypeRequest struct {
    Name     string `json:"name"     binding:"required"`
    Category string `json:"category" binding:"required,oneof=general trimming"` // ← 追加
    // ...
}

type updateReservationTypeRequest struct {
    Category *string `json:"category" binding:"omitempty,oneof=general trimming"` // ← 追加
    // ...
}
```

### Step 2: Service Input に Category を追加

```go
// reservation_type_service.go
type CreateReservationTypeInput struct {
    Category string // ← 追加
    // ...
}

type UpdateReservationTypeInput struct {
    Category *string // ← 追加
    // ...
}
```

### Step 3: Service の buildUpdateFields に Category を追加

```go
// buildReservationTypeUpdateFields 内
if input.Category != nil {
    fields[colReservationTypeCategory] = *input.Category
}
```

### Step 4: Handler のマッピングに Category を追加

```go
// CreateReservationType
service.CreateReservationTypeInput{
    Category: req.Category, // ← 追加
    // ...
}

// UpdateReservationType
service.UpdateReservationTypeInput{
    Category: req.Category, // ← 追加
    // ...
}
```

### Step 5: Response に Category を追加（TASK-061 と重複）

```go
type reservationTypeResponse struct {
    Category string `json:"category"` // ← 追加
    // ...
}
```

---

## 確認事項

- `category` フィールドの DB デフォルトが `'general'` のため、既存データへの影響なし
- LIFF ハンドラ（`reservation_type_liff_handler.go`）の response にも Category が必要か確認
- `docs/api.yaml` の更新

---

## 関連タスク

- **TASK-061**: reservation_type_response に Category フィールド欠落（本タスクが上位互換）
  → TASK-061 は本タスクに統合してクローズ可

---

## 備考

- `model.ReservationTypeCategory` は `general` / `trimming` の 2 値
- トリミング予約と通常診療予約を区別する UI が将来必要になった際に、Category が API から取得できないと対応不可
