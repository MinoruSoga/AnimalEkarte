# BUG-401: LinkReservationTypeOccupation が Location ヘッダを返さない

## 概要
`reservation_type_handler.go` の `LinkReservationTypeOccupation`（`POST /masters/reservation-types/:id/occupations`）が 201 Created を返すが Location ヘッダが欠落している。同一ハンドラファイルの `CreateUnavailableTime`（196行目）は正しく Location ヘッダを返しており、同じサブリソース Create エンドポイントで不統一になっている。

## 再現手順
1. `POST /v1/masters/reservation-types/:id/occupations` でリクエストを送信
2. **結果**: 201 Created が返るが Location ヘッダがない
3. 比較: `POST /v1/masters/reservation-types/:id/unavailable-times` は 201 + `Location: /v1/masters/reservation-types/{id}/unavailable-times/{new_id}` を返す

## 現状コード

### `backend/internal/handler/reservation_type_handler.go:255`（問題箇所）
```go
func (h *Handler) LinkReservationTypeOccupation(c *gin.Context) {
    ...
    result, err := h.service.ReservationType.LinkOccupation(c.Request.Context(), clinicID, id, req.OccupationID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toReservationTypeOccupationResponse(result))
    // ↑ Location ヘッダなし
}
```

### 比較: 同一ファイル内の正しい実装（196行目）
```go
func (h *Handler) CreateUnavailableTime(c *gin.Context) {
    ...
    c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/unavailable-times/%d", id, result.ID))
    c.JSON(http.StatusCreated, resp)
    // ↑ Location ヘッダあり ✅
}
```

## 修正方針

### `reservation_type_handler.go:255` — Location ヘッダ追加
```go
c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/occupations/%d", id, result.ID))
c.JSON(http.StatusCreated, toReservationTypeOccupationResponse(result))
```

## 優先度
**Low** — REST ベストプラクティス違反。機能上の問題なし。BUG-381（trimming_master の Location 欠落）と同カテゴリ。

## 関連チケット
- **BUG-381**: trimming_master_handler の Location ヘッダ欠落（同パターン）

## 関連ファイル
- `backend/internal/handler/reservation_type_handler.go:255` — 修正対象
