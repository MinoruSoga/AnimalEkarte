# BUG-410: reservation_type_liff_handler の Create が Location ヘッダを返していない

## 概要

`reservation_type_liff_handler.go` の `CreateReservationTypeLiff` が 201 Created を返す際、
他のマスタ Create エンドポイントとは異なり `Location` ヘッダを設定していない。
REST 規約違反であり、クライアントが作成リソースの URI を取得できない。

## 問題箇所

```go
// reservation_type_liff_handler.go:54
c.JSON(http.StatusCreated, toReservationTypeLiffResponse(st))
// Location ヘッダなし ← 規約違反
```

## 期待する実装

```go
// 他のマスタ Create の標準パターン（medicine_handler.go:87-88）
c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
c.JSON(http.StatusCreated, toMedicineResponse(medicine))
```

## 修正方針

`reservation_type_liff_handler.go:54` の直前に以下を追加する。

```go
c.Header("Location", fmt.Sprintf("/api/clinics/%d/reservation-types/%d", clinicID, st.ID))
c.JSON(http.StatusCreated, toReservationTypeLiffResponse(st))
```

※ パスはルーティング定義（`reservation_type_liff_routes.go`）と照合すること。

## 影響ファイル

- `backend/internal/handler/reservation_type_liff_handler.go` — 行 54

## 優先度

**Low** — REST 規約違反。1行修正で対応可能。
