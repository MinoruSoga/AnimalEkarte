# TASK-063: reservation_type_handler — CreateReservationType の IsActive がハードコード

## 優先度

HIGH

---

## 概要

`reservation_type_handler.go` の `CreateReservationType` メソッドで、`IsActive` フィールドが `req.IsActive` ではなく `true` にハードコードされている。

クライアントから `"is_active": false` を送信しても無視され、**常に active 状態で予約種別が登録される**。意図的に非アクティブ状態で作成したいユースケースが機能しない。

---

## 問題箇所

### backend/internal/handler/reservation_type_handler.go（L62 概算）

```go
// ❌ 現状: IsActive が常に true
st, err := h.svc.ReservationType.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeInput{
    Name:                   req.Name,
    Color:                  req.Color,
    IsActive:               true,      // ← req.IsActive を使っていない
    Description:            req.Description,
    // ...
})
```

---

## 修正方針

```go
// ✅ 修正後
st, err := h.svc.ReservationType.Create(c.Request.Context(), clinicID, &service.CreateReservationTypeInput{
    Name:                   req.Name,
    Color:                  req.Color,
    IsActive:               req.IsActive,  // ← req から受け取る
    Description:            req.Description,
    // ...
})
```

---

## 確認事項

- `createReservationTypeRequest.IsActive` フィールドの型と binding タグを確認し、デフォルト値（bool のゼロ値 = false）が意図通りかを確認すること
- もし「新規作成は常に active」がビジネス仕様であれば、request struct から `IsActive` を削除し、ハードコードを正当化するコメントを追加すること

---

## 備考

- 他の全マスタハンドラの Create は `req.IsActive` をそのまま渡しており、reservation_type のみが例外的に true にハードコードしている
- Update メソッドは正しく `req.IsActive` を使用しているため、Update では問題なし
