# TASK-054: clinic_holiday_handler — SetClinicHoliday が 200 を返す（201 が正しい）

## 優先度

LOW

---

## 概要

REST 規約では新規リソース作成の POST レスポンスは `201 Created` を返す必要がある。  
`clinic_holiday_handler.go` の `SetClinicHoliday` メソッドが `http.StatusOK`（200）を返しており、全マスタハンドラ中で唯一の違反。

---

## 問題箇所

### ファイル
`backend/internal/handler/clinic_holiday_handler.go` L86（概算）

```go
// ❌ 現状: 200 OK
c.JSON(http.StatusOK, toClinicHolidayResponse(holiday))

// ✅ 修正後: 201 Created
c.JSON(http.StatusCreated, toClinicHolidayResponse(holiday))
```

---

## 調査メモ

全 18 マスタハンドラで Create/Set 系メソッドの返却コードを確認した結果、`clinic_holiday_handler.go` のみが 200 を返していた。  
他の 17 ハンドラはすべて `http.StatusCreated`（201）で統一されている。

---

## 修正方針

`SetClinicHoliday` メソッド内の成功レスポンス行を `http.StatusOK` → `http.StatusCreated` に変更する。1行修正のみ。
