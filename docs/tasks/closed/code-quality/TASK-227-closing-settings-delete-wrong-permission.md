# TASK-227: closing_settings_handler.go — DELETE ルートが "delete" ではなく "edit" 権限を使用

## 優先度
High

## 対象ファイル
- `backend/internal/handler/closing_settings_handler.go`

## 問題概要
`RegisterClosingSettingsRoutes` の DELETE エンドポイント2箇所が `"delete"` ではなく `"edit"` 権限を要求している。
プロジェクト規約では GET → `"view"`、POST/PUT/PATCH → `"edit"`、DELETE → `"delete"` と統一されている。

## 現状コード（行157, 163）

```go
// ❌ DELETE に "edit" を使用
sp.DELETE("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.DeleteSpecialPeriod)
holidays.DELETE("/:date", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.DeleteClinicHoliday)
```

## あるべき姿

```go
// ✅ DELETE には "delete" を使用
sp.DELETE("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "delete"), h.DeleteSpecialPeriod)
holidays.DELETE("/:date", h.RequirePermission(string(model.ResourceClosingSettings), "delete"), h.DeleteClinicHoliday)
```

## 完了条件
- [ ] `sp.DELETE` の権限を `"edit"` → `"delete"` に変更
- [ ] `holidays.DELETE` の権限を `"edit"` → `"delete"` に変更
- [ ] `go test ./backend/internal/...` がパス
