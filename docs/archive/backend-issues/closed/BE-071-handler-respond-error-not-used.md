# BE-071: 複数 handler で RespondError を使わず c.JSON を直接呼び出し

**Status**: Closed
**Priority**: Medium
**Affects**: `backend/internal/handler/`（複数ファイル）
**Date Created**: 2026-03-26
**Related**: -

## Summary

複数の handler ファイルでエラーレスポンスを `c.JSON(http.StatusBadRequest, gin.H{"error": "..."})` と直接呼び出しており、`RespondError(c, err)` が使われていない。これによりエラーレスポンスの形式（`code`, `message`, `timestamp` フィールド）が統一されずフロントエンドのエラーハンドリングが機能しない箇所が発生する。

## 現状のコード（代表例）

```go
// backend/internal/handler/hospitalization_handler.go:29
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic_id"})

// backend/internal/handler/medical_record_handler.go:50
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic_id"})

// backend/internal/handler/vaccination_handler.go:29
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic_id"})
```

## 規約通りの正しい実装（参照: owner_handler.go）

```go
// backend/internal/handler/owner_handler.go（正しい例）
if err := c.ShouldBindJSON(&req); err != nil {
    RespondError(c, apperrors.WrapInvalidInput(err.Error()))
    return
}
```

## 修正対象ファイルと行番号

| ファイル | 修正が必要な行（概算） |
|---|---|
| `reservation_handler.go` | :133, :158, :163, :220, :248 |
| `hospitalization_handler.go` | :29, :38, :76, :95, :104, :128, :153, :158, :183, :201, :226 |
| `medical_record_handler.go` | :50, :60, :93, :113, :126, :143, :147, :154, :162, :170, :178, :200, :254, :259, :281, :306 |
| `vaccination_handler.go` | :29, :39, :72, :92, :131, :136, :191 |
| `examination_handler.go` | :29, :39, :77, :97, :130, :135, :170 |
| `trimming_handler.go` | :30, :40, :74, :94, :139, :144, :187 |
| `accounting_handler.go` | :189 |

## 修正パターン

### clinic_id パース失敗

```go
// 変更前
clinicID, err := strconv.ParseUint(clinicIDStr, 10, 64)
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic_id"})
    return
}

// 変更後
clinicID, err := strconv.ParseUint(clinicIDStr, 10, 64)
if err != nil {
    RespondError(c, apperrors.WrapInvalidInput("invalid clinic_id"))
    return
}
```

### JSON バインド失敗

```go
// 変更前
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// 変更後
if err := c.ShouldBindJSON(&req); err != nil {
    RespondError(c, apperrors.WrapInvalidInput(err.Error()))
    return
}
```

## 完了条件

- [ ] 上記全ファイルの `c.JSON(http.StatusBadRequest, ...)` を `RespondError` に統一
- [ ] `c.JSON(http.StatusNotFound, ...)` 等も同様に `RespondError` に統一
- [ ] `docker compose exec backend golangci-lint run ./...` がパス
- [ ] エラーレスポンス形式が `{"code": "...", "message": "...", "timestamp": "..."}` に統一されること
