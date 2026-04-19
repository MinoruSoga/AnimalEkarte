# TASK-061: reservation_type_response — Category フィールド欠落

## 優先度

MEDIUM

---

## 概要

`model.ReservationType` には `Category` フィールド（`reservation_type_category` ENUM: `general` | `trimming`）が定義されているが、
`reservation_type_response.go` の `reservationTypeResponse` struct および `toReservationTypeResponse` 関数に `Category` が含まれていない。

API クライアント（フロントエンド / LIFF）はカテゴリ情報を取得できないため、カテゴリ別のフィルタリングや表示分岐が実装不能になっている。

---

## 問題箇所

### backend/internal/model/reservation_type.go

```go
// ✅ model には Category フィールドがある
type ReservationType struct {
    // ...
    Category ReservationTypeCategory `gorm:"type:reservation_type_category;not null;default:'general'" json:"category"`
    // ...
}
```

### backend/internal/handler/reservation_type_response.go

```go
// ❌ reservationTypeResponse に Category なし
type reservationTypeResponse struct {
    ID          uint64    `json:"id"`
    ClinicID    uint64    `json:"clinic_id"`
    Name        string    `json:"name"`
    Color       string    `json:"color"`
    IsActive    bool      `json:"is_active"`
    // ... Category が抜けている
}

// ❌ toReservationTypeResponse でもマッピングなし
func toReservationTypeResponse(rt *model.ReservationType) reservationTypeResponse {
    return reservationTypeResponse{
        // Category がマッピングされていない
    }
}
```

---

## 修正方針

### Step 1: response struct に Category を追加

```go
type reservationTypeResponse struct {
    ID       uint64                        `json:"id"`
    ClinicID uint64                        `json:"clinic_id"`
    // ...
    Category model.ReservationTypeCategory `json:"category"` // ← 追加
    // ...
}
```

### Step 2: toReservationTypeResponse にマッピング追加

```go
func toReservationTypeResponse(rt *model.ReservationType) reservationTypeResponse {
    return reservationTypeResponse{
        // ...
        Category: rt.Category, // ← 追加
        // ...
    }
}
```

---

## 影響範囲

- `GET /masters/reservation-types` — 一覧レスポンスに `category` が追加される
- `GET /masters/reservation-types/:id` — 単体レスポンスに `category` が追加される
- LIFF 側の予約種別 API にも同様のレスポンス struct があれば確認が必要

---

## 備考

- `ReservationTypeCategory` ENUM 値: `general`（一般診療）/ `trimming`（トリミング）
- フロントエンド側でカテゴリ別フィルタや表示分岐が必要な場合は、この修正後に対応する
- API ドキュメント（`docs/api.yaml`）も合わせて更新すること
