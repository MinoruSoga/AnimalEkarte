# BUG-278: AuditService が呼び出し側に組み込まれていない（実装漏れ）

## 概要

`AuditService`（`audit_service.go`）と関連インフラ（`audit_repository.go`、`model/audit_log.go`）は実装済みだが、
呼び出し側（auth_handler, permission_group_service）への組み込みが漏れている。
結果として、ログイン/ログアウトや権限変更の監査証跡が一切記録されていない。

## 現状

### 実装済み（基盤）
- `model/audit_log.go` — AuditLog モデル + アクション定数（8種）
- `repository/audit_repository.go` — Create + MarshalAuditJSON ヘルパー
- `service/audit_service.go` — Log + LogAuthLogin メソッド
- `service/service.go:11,76` — Services.Audit フィールド + DI 配線
- `handler/handler.go:22,28` — auditRepo フィールド（未使用）

### 未実装（組み込み漏れ）
- `auth_handler.go` — ログイン成功/失敗/ログアウト時に `svc.Audit.LogAuthLogin()` を呼んでいない
- `permission_group_service.go` — Create/Update/Delete/SetRules 時に `svc.Audit.Log()` で変更前後を記録していない

## 修正方針

### 1. auth_handler.go — ログイン/ログアウト監査

```go
// Login 成功時（auth_handler.go の JWT 発行後）
h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staff.ID,
    model.AuditActionAuthLoginSuccess, c.ClientIP(), c.GetHeader("User-Agent"))

// Login 失敗時（認証エラー時）
h.svc.Audit.LogAuthLogin(ctx, nil, nil,
    model.AuditActionAuthLoginFailure, c.ClientIP(), c.GetHeader("User-Agent"))

// Logout 時
h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staffID,
    model.AuditActionAuthLogout, c.ClientIP(), c.GetHeader("User-Agent"))
```

### 2. permission_group_service.go — 権限変更監査

```go
// Create/Update/Delete 後に
auditLog := &model.AuditLog{
    ClinicID:   &clinicID,
    ActorID:    &staffID, // 操作者（context から取得）
    ActorType:  "staff",
    Action:     model.AuditActionPermissionGroupCreate,
    Resource:   "permission_group",
    ResourceID: &group.ID,
    OldValue:   repository.MarshalAuditJSON(oldValue),
    NewValue:   repository.MarshalAuditJSON(newValue),
}
s.auditSvc.Log(ctx, auditLog)
```

### 3. handler.go — auditRepo フィールド削除

handler は service 経由で Audit を呼ぶべき。`auditRepo` フィールドは不要。

## 優先度

**High** — 監査証跡がないと、不正ログインや権限変更の追跡が不可能。セキュリティ監査で指摘される。

## 関連チケット

- BUG-277: デッドコード監査 親チケット
