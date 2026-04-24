# BUG-LINE-011: 存在しない clinic_id への予約作成で 500 Internal Error

## 概要

`POST /api/liff/:clinicId/reservations` に存在しない clinic_id（例: 999）や設定が無い clinic_id（clinic 2）を指定すると、**500 Internal Server Error** が返る。期待は **404 Not Found**。

## 再現

```javascript
// clinic 2 (城東 — 設定未登録)
await fetch('/api/liff/2/reservations', { method: 'POST', ... });
// → 500 {"error":"failed to load clinic setting"}

// clinic 999 (存在しない)
await fetch('/api/liff/999/reservations', { method: 'POST', ... });
// → 500 {"error":"failed to load clinic setting"}
```

同じ clinic の GET /settings は 404 を返すのに、POST /reservations だけ 500 になる。
一貫性がない。

## 修正案

`liff_service.CreateReservation` で setting 取得失敗時:

```go
// Before (推定)
if err := loadSetting(); err != nil {
  return fmt.Errorf("failed to load clinic setting: %w", err)
}

// After
if errors.Is(err, gorm.ErrRecordNotFound) {
  return apperrors.WrapNotFound("clinic_setting", clinicIDStr)
}
```

## 優先度

**MEDIUM** — エラー表示の品質問題。機密情報漏洩はない。
