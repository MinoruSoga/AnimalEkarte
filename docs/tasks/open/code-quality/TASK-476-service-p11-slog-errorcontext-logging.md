---
title: Service P11 slog.ErrorContext Logging 欠落
issue: '#476'
priority: MEDIUM
status: open
area: service
pattern: P11
---

## 概要

サービスレイヤーの List メソッドで、リポジトリから返された予期しないエラーのログ出力が欠落しているケースが 2 件検出されました。

### パターン
- **P11 違反**：リポジトリエラー（NotFound/Conflict 以外）に対して `slog.ErrorContext` でログ出力がない

### 違反ファイル一覧

| ファイル | 行番号 | メソッド | 問題 |
|---------|--------|---------|------|
| procedure_service.go | 45 | List | err != nil の際に slog.ErrorContext でログ出力がない |
| vaccine_service.go | 38 | List | err != nil の際に slog.ErrorContext でログ出力がない |

## 修正方法

List メソッドのエラーハンドリングに `slog.ErrorContext` を追加：

```go
import (
    "log/slog"
    apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// 修正例 (procedure_service.go L45)
func (s *ProcedureService) List(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
    procedures, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        slog.ErrorContext(ctx, "failed to fetch procedures", 
            "clinic_id", clinicID, 
            "error", err)
        return nil, apperrors.Wrap(err, "fetch procedures failed")
    }
    return procedures, nil
}

// 修正例 (vaccine_service.go L38)
func (s *VaccineService) List(ctx context.Context, clinicID uint64) ([]model.Vaccine, error) {
    vaccines, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        slog.ErrorContext(ctx, "failed to fetch vaccines", 
            "clinic_id", clinicID, 
            "error", err)
        return nil, apperrors.Wrap(err, "fetch vaccines failed")
    }
    return vaccines, nil
}
```

### ログ出力ルール
- **対象エラー**：NotFound / Conflict 以外のリポジトリエラー
- **対象メソッド**：FindAll, FindByID（読み取りに限定。Create/Update/Delete は除外）
- **ログレベル**：`slog.ErrorContext` で十分（WarnContext 不可）
- **含めるコンテキスト**：clinic_id, resource_id, error

## テスト

修正後、以下の確認を実施：
- [ ] DB 接続エラー時に slog.ErrorContext で error ログが出力されること
- [ ] ログに clinic_id が含まれること
- [ ] 既存テストが全件パス

## 参考

- Pattern: P11 (slog.ErrorContext for unexpected errors)
- 除外エラー：NotFound, Conflict, Validation errors
- 対象メソッド：Read operations (FindAll, FindByID)
