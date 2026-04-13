# BUG-338: バックエンド Go コード規約準拠監査 第7回（2026-04-12）

**Status**: CLOSED  
**Priority**: High/Medium  
**Discovery**: golang-pro スキルによる静的解析（2026-04-12）

## 概要

`backend/` 全体を golang-pro スキルで規約準拠監査した結果、CRITICAL なし、HIGH 1件、MEDIUM 4件を検出・修正。

## 検出・修正内容

### HIGH — `handler/response.go` RequirePermission が `c.JSON` 直接使用

**問題**: `RequirePermission` ミドルウェア内で `c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})` を直接使用。`RespondError` を迂回してレスポンス形式が不統一。

**修正**: `RespondErrorAndAbort` ヘルパーを追加し、`RequirePermission` で使用するよう変更。

```go
// 追加したヘルパー
func RespondErrorAndAbort(c *gin.Context, err error) {
    RespondError(c, err)
    c.Abort()
}

// RequirePermission の修正後
RespondErrorAndAbort(c, apperrors.WrapForbidden("forbidden"))
```

### MEDIUM — `repository/reservation_staff_repository.go` nilerr

**問題**: `q.First(&neighbor).Error` が non-nil の場合に `return nil` していた。`gorm.ErrRecordNotFound` 以外のエラー（DB 接続エラー等）をサイレントに無視。

**修正**: `gorm.ErrRecordNotFound` のみ `nil` を返し、その他は `apperrors.FromGORM` で伝播。

### MEDIUM — `repository/reservation_type_liff_repository.go` nilerr

同上パターン。同様に修正。

### MEDIUM — `service/liff_service.go` bare error return 2箇所

**問題**: `resolveTargetStaffs` / `buildStaffSlotInputs` からのエラーを `return nil, err` で素通し。エラー発生箇所のコンテキスト情報が失われる。

**修正**: `apperrors.Wrap` でコンテキストを付与。

```go
return nil, apperrors.Wrap(err, "failed to resolve target staffs")
return nil, apperrors.Wrap(err, "failed to build staff slot inputs")
return nil, apperrors.Wrap(err, "failed to validate and create appointment")
```

### MEDIUM — `service/appointment_admin_service.go` Transaction bare return

**問題**: `db.Transaction()` の結果を `return nil, err` で素通し。

**修正**: `apperrors.Wrap(err, "create reservation appointment (transaction)")`

## 検証結果

- `go build ./...` — クリーン
- CRITICAL: 0件
- HIGH: 1件 → 修正済み
- MEDIUM: 4件 → 修正済み

## 関連コミット

`510de647` — fix(backend): Goコード規約準拠修正 — nilerr / bare error return / RequirePermission直接c.JSON

## 実施日

2026-04-12
