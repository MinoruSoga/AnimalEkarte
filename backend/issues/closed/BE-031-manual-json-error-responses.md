# BE-031: 手動 c.JSON() エラーレスポンス → RespondError 統一

## 問題
複数のハンドラが `RespondError(c, err)` を使わず `c.JSON(http.StatusBadRequest, ...)` を
直接呼び出している。エラーレスポンス形式が不統一になる。

## 影響ファイル（代表例）
- `handler/diagnosis_handler.go:38,63,69,94,112`
- `handler/daily_record_handler.go:73,109,155,200`
- `handler/reservation_handler.go:56,71,80,102,121,142,158,188`
- `handler/treatment_plan_handler.go:39,87,117`
- `handler/pet_handler.go:31,77,124`
- `handler/shift_handler.go:51,98`
- `handler/job_title_handler.go:40,71,112`
- `handler/accounting_handler.go:30,39,67,87,129`

## 現状（NG）
```go
// ID パース失敗など
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
```

## 修正方針
```go
// ✅ 正しいパターン
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    RespondError(c, apperrors.WrapInvalidInput("invalid id"))
    return
}
```

ID パースエラー用の共通ヘルパー関数 `parseIDParam(c, "id")` を
`handler/` に追加して重複を排除することを推奨。

## 優先度
MEDIUM（レスポンス形式の一貫性）
