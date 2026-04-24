# TASK-012: appointment_notification_service / line_messaging_service の fmt.Errorf を apperrors.Wrap に統一

## 概要

`appointment_notification_service.go` と `line_messaging_service.go` でインフラ層（SMTP・LINE HTTP）のエラーが `fmt.Errorf` でラップされており、他のサービスで使われている `apperrors.Wrap` と不統一になっている。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `backend/internal/service/appointment_notification_service.go` | L309 | `fmt.Errorf("smtp send: %w", err)` |
| `backend/internal/service/line_messaging_service.go` | L60-77 | `fmt.Errorf` でのエラーラップ |

## 規約違反

`.claude/rules/go-language.md`:
> Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。

## 修正方針

```go
// appointment_notification_service.go L309
// Before:
return fmt.Errorf("smtp send: %w", err)
// After:
return apperrors.Wrap(err, "smtp send")

// line_messaging_service.go（各 fmt.Errorf 箇所）
// Before:
return fmt.Errorf("LINE API call: %w", err)
// After:
return apperrors.Wrap(err, "LINE API call")
```

## 補足

`apperrors.Wrap` は内部的に `fmt.Errorf` + `%w` と同等だが、`RespondError` でのステータスマッピングに組み込まれているため、将来的に通知系エラーを適切な HTTP ステータスにマッピングしたい場合に一貫性が保たれる。
