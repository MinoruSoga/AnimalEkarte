# BUG-257: slog 監査ログ欠落・順序不正・レイヤー違反

## 概要

Service 層の Create/Update/Delete 操作に `slog.InfoContext` による監査ログが欠落している箇所、
削除実行前にログ出力している順序不正、Handler 層に slog が存在するレイヤー違反を検出。

## 影響範囲

### 監査ログ欠落（Service 層）

| ファイル | メソッド | 状態 |
|---------|---------|------|
| `service/vaccination_service.go` | Create (:37-42), Delete (:123-125) | slog なし |
| `service/vaccine_service.go` | Create (:36-38), Delete (:95-104) | slog なし |
| `service/exam_type_service.go` | Create (:36-38) | slog なし |
| `service/checkup_type_service.go` | Create (:38-40), Delete | slog なし |
| `service/procedure_service.go` | Create, Delete, Reorder (:38-71) | slog なし |
| `service/trimming_service.go` | Delete | slog なし |
| `service/insurance_service.go` | Create (:46-51) | slog なし |
| `service/inventory_service.go` | Delete (:57-69) | slog なし |

### ログ順序不正

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/insurance_service.go` | :72-76 | Delete 実行前に「deleted」ログを出力。削除失敗時も「削除された」と記録される |

### レイヤー違反（Handler 層の slog）

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `handler/record_image_handler.go` | :234 | Handler 内で `slog.WarnContext` を使用（ファイルクリーンアップ失敗時） |

## 修正方針

### 監査ログ追加パターン

```go
func (s *xxxService) Create(ctx context.Context, clinicID uint64, input *CreateXxxInput) error {
    // ... 処理 ...
    if err := s.repo.Create(ctx, entity); err != nil {
        return apperrors.Wrap(err, "failed to create xxx")
    }
    slog.InfoContext(ctx, "xxx created", slog.Uint64("xxx_id", entity.ID))
    return nil
}
```

### ログ順序修正（insurance_service.go）

```go
// Before — 削除前にログ
slog.InfoContext(ctx, "insurance deleted", ...)
if err := s.repo.Delete(ctx, clinicID, id); err != nil { ... }

// After — 削除成功後にログ
if err := s.repo.Delete(ctx, clinicID, id); err != nil {
    return apperrors.Wrap(err, "failed to delete insurance")
}
slog.InfoContext(ctx, "insurance deleted", slog.Uint64("insurance_id", id))
```

## 優先度

**High** — 監査ログの欠落はインシデント調査時に追跡不能になる。順序不正は誤った監査記録を残す。

## 関連チケット

- BUG-253: 親チケット
