# TASK-060: clinic_holiday / reservation_type handler — mapSlice 未使用の手動ループ（追加）

## 優先度

LOW

---

## 概要

TASK-056（procedure / cage）と同様に、`clinic_holiday_handler.go` と `reservation_type_handler.go` にも `mapSlice` を使わず手動で `make + for range` ループを実装している箇所がある。

`handler/slice_helpers.go` に定義された `mapSlice` ジェネリック関数を使うべき。

---

## 問題箇所

### clinic_holiday_handler.go（L49-54 概算）

```go
// ❌ 手動ループ（冗長）
resp := make([]clinicHolidayResponse, 0, len(holidays))
for i := range holidays {
    resp = append(resp, toClinicHolidayResponse(&holidays[i]))
}
c.JSON(http.StatusOK, resp)
```

### reservation_type_handler.go — ListUnavailableTimes（L154-157 概算）

```go
// ❌ 手動ループ（冗長）
resp := make([]unavailableTimeResponse, 0, len(items))
for i := range items {
    resp = append(resp, toUnavailableTimeResponse(&items[i]))
}
```

### reservation_type_handler.go — ListReservationTypeOccupations（L231-234 概算）

```go
// ❌ 手動ループ（冗長）
resp := make([]reservationTypeOccupationResponse, 0, len(items))
for i := range items {
    resp = append(resp, toReservationTypeOccupationResponse(&items[i]))
}
```

---

## 修正方針

```go
// ✅ 修正後 — clinic_holiday_handler.go
c.JSON(http.StatusOK, mapSlice(holidays, toClinicHolidayResponse))

// ✅ 修正後 — reservation_type_handler.go ListUnavailableTimes
c.JSON(http.StatusOK, mapSlice(items, toUnavailableTimeResponse))

// ✅ 修正後 — reservation_type_handler.go ListReservationTypeOccupations
c.JSON(http.StatusOK, mapSlice(items, toReservationTypeOccupationResponse))
```

---

## 参照実装

`vaccine_handler.go` / `insurance_handler.go` / `exam_type_handler.go` 等がすべて `mapSlice` を使用している。

---

## 備考

`mapSlice` のシグネチャ: `func mapSlice[M, R any](items []M, f func(*M) R) []R`  
変換関数が `*T` を受け取る場合はそのまま渡せる。
