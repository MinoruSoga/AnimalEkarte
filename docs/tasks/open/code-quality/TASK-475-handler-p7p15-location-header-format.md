---
title: Handler P7/P15 Location Header フォーマット 不統一
issue: '#475'
priority: MEDIUM
status: open
area: handler
pattern: P7/P15
---

## 概要

POST リクエストレスポンスの Location ヘッダーフォーマットが不統一なケースが 3 件検出されました。標準フォーマット `Location: /resource/{id}` への統一が必要です。

### パターン
- **P7 違反**：Location ヘッダー値が不適切なフォーマット（LIFF URL、重複パス、不完全なパス等）
- **P15 違反**：toXxxResponse 変換後に Location ヘッダーを設定していない

### 違反ファイル一覧

| ファイル | 行番号 | 現在のフォーマット | 問題 |
|---------|--------|------------------|------|
| line_official_account_handler.go | 58 | `c.Header("Location", liff)` | LIFF URL を直接設定（リソースパスではない） |
| line_reservation_setting_handler.go | 93 | `c.Header("Location", "/settings/line/" + id)` | `/clinics/{clinicID}/` プレフィックス欠落 |
| reservation_schedule_handler.go | 127 | `c.Header("Location", fmt.Sprintf("/schedule/%d", schedule.ID))` | `/clinics/{clinicID}/` プレフィックス欠落 |

## 修正方法

Location ヘッダーは **必ず** `Location: /clinics/{clinicID}/resource/{id}` 形式：

```go
// 修正例 1: line_official_account_handler.go (L58)
// LIFF URL は別途応答ボディに含める
account, err := h.svc.Create(ctx, clinicID, input)
if err != nil {
    middleware.RespondError(c, err)
    return
}
response := toLineOfficialAccountResponse(account)
c.Header("Location", fmt.Sprintf("/clinics/%d/line-accounts/%d", clinicID, account.ID))
c.JSON(http.StatusCreated, response)

// 修正例 2: line_reservation_setting_handler.go (L93)
c.Header("Location", fmt.Sprintf("/clinics/%d/line-settings/%d", clinicID, setting.ID))

// 修正例 3: reservation_schedule_handler.go (L127)
c.Header("Location", fmt.Sprintf("/clinics/%d/reservation-schedules/%d", clinicID, schedule.ID))
```

## テスト

修正後、以下の確認を実施：
- [ ] POST レスポンス HTTP 201 に Location ヘッダーが含まれること
- [ ] Location ヘッダー値が `/clinics/{clinicID}/resource/{id}` 形式であること
- [ ] LIFF URL は Location ではなくレスポンスボディに含まれること
- [ ] フロントエンド統合テストが全件パス

## 参考

- Pattern: P7 (POST Response 201 + Location Header)
- Pattern: P15 (toXxxResponse 変換 + Location header 設定)
- 標準フォーマット: `/clinics/{clinicID}/{resource}/{id}`
