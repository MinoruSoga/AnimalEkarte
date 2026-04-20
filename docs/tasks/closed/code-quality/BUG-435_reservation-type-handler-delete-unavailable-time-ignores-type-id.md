# BUG-435: reservation_type_handler の DeleteUnavailableTime が :id（予約区分ID）を無視

## 概要

`DeleteUnavailableTime` ハンドラがルートパスの `:id`（reservation_type_id）パラメータを
取得・検証せずに処理している。REST 設計上 `/reservation-types/:id/unavailable-times/:unavailable_time_id`
というネスト構造を持つが、`:id` が一切使用されていない。

## 問題箇所

```go
// reservation_type_handler.go:201-216
// DeleteUnavailableTime godoc
func (h *Handler) DeleteUnavailableTime(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    unavailableTimeID, ok := parseIDParam(c, "unavailable_time_id")
    if !ok {
        return
    }
    // ← `:id`（reservation_type_id）を取得していない
    if err := h.svc.ReservationType.DeleteUnavailableTime(c.Request.Context(), clinicID, unavailableTimeID); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}
```

## CreateUnavailableTime との比較（正しい実装）

```go
// reservation_type_handler.go:162-199（CreateUnavailableTime）
func (h *Handler) CreateUnavailableTime(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    id, ok := parseIDParam(c, "id")  // ✅ reservation_type_id を取得
    if !ok {
        return
    }
    // ...
    result, err := h.svc.ReservationType.CreateUnavailableTime(c.Request.Context(), clinicID, id, input)
    // ...
    c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/unavailable-times/%d", id, result.ID))
}
```

## リスク評価

| 観点 | 影響 |
|------|------|
| クロステナント攻撃 | なし（clinicID フィルタが機能している） |
| クロスリソース操作 | あり（URL の `:id` が示す reservation_type に紐付かない unavailable_time を削除できる） |
| API 設計の一貫性 | 違反（Create は reservation_type_id を使うが Delete は使わない） |

具体例: `DELETE /reservation-types/999/unavailable-times/42` を送信した場合、
unavailable_time_id=42 が別の reservation_type（例: id=1）に属していても削除が成功する。

## 修正方針

### Option 1: reservation_type_id を検証に含める（推奨）

サービス層に `reservationTypeID` を加えた検証ロジックを追加し、
unavailable_time がその reservation_type に属することを確認する。

```go
// handler
func (h *Handler) DeleteUnavailableTime(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    reservationTypeID, ok := parseIDParam(c, "id")  // ← 追加
    if !ok {
        return
    }
    unavailableTimeID, ok := parseIDParam(c, "unavailable_time_id")
    if !ok {
        return
    }
    if err := h.svc.ReservationType.DeleteUnavailableTime(c.Request.Context(), clinicID, reservationTypeID, unavailableTimeID); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}

// service（修正後）
func (s *reservationTypeService) DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
    // ← reservation_type の存在確認（テナント検証）
    if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
        return apperrors.Wrap(err, "failed to get reservation type")
    }
    if err := s.unavailableTimeRepo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete unavailable time")
    }
    // ...
}
```

### Option 2: シグネチャ変更なしでの最小修正

handler で `:id` を読んで validaton のみ行い、service は現在のまま。
ただし親リソース確認が service 層で保証されないため Option 1 を推奨。

## 影響ファイル

- `backend/internal/handler/reservation_type_handler.go` — 行 201-216（DeleteUnavailableTime）
- `backend/internal/service/reservation_type_service.go` — 行 364-372（DeleteUnavailableTime シグネチャ変更を伴う場合）
- `backend/internal/repository/reservation_type_unavailable_time_repository.go` — 必要に応じて FindByReservationTypeID 追加

## 優先度

**Low** — クロステナント攻撃は clinicID によりブロックされているため直接的なセキュリティリスクは低い。
ただし REST API の設計一貫性違反であり、クロスリソース操作を許容する設計上の問題がある。

## 関連チケット

- BUG-416（reservation_type_unavailable_time_repository の clinicScope 不統一）
- BUG-418（handler へのビジネスロジック漏れ）
