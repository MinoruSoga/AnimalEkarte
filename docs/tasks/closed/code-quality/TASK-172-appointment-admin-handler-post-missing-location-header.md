# TASK-172: appointment_admin_handler.go — POST 201 に Location ヘッダーなし

## 優先度
High

## 対象ファイル
`backend/internal/handler/appointment_admin_handler.go`

## 問題概要
`CreateReservationAdmin` が `http.StatusCreated (201)` を返すにもかかわらず、
`Location` ヘッダーが設定されていない。
プロジェクト規約では POST 201 レスポンスには必ず Location ヘッダーを付与すること。

## 現状コード

```go
// appointment_admin_handler.go line 99-103
ra, err := h.svc.ReservationAdmin.Create(c.Request.Context(), clinicID, &service.CreateReservationAdminInput{
    // ...
})
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusCreated, toReservationDetailResponse(ra))  // ❌ Location ヘッダーなし
```

## 修正後コード

```go
ra, err := h.svc.ReservationAdmin.Create(c.Request.Context(), clinicID, &service.CreateReservationAdminInput{
    // ...
})
if err != nil {
    RespondError(c, err)
    return
}
c.Header("Location", fmt.Sprintf("/v1/reservations/%d", ra.ID))  // ✅ Location ヘッダー追加
c.JSON(http.StatusCreated, toReservationDetailResponse(ra))
```

## 必要な変更
- `appointment_admin_handler.go` の import に `"fmt"` を追加（既存 import を確認の上）
- `c.JSON(http.StatusCreated, ...)` の前に `c.Header("Location", ...)` を追加

## API パスの確認
実際のルーティングでの予約詳細 GET エンドポイントのパスに合わせること。
`/v1/clinics/{clinicId}/reservations/{id}` などルートを確認して正しいパスを使用する。

## 影響範囲
クライアント側が 201 レスポンスの Location ヘッダーを参照している場合に影響。
REST 仕様準拠のため High 優先度。
