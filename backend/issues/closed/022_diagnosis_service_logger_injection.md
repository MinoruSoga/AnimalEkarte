---
status: open
---

# diagnosis service: slog をグローバル呼び出しから DI 注入に変更

## 背景

`diagnosis_service.go` では slog をグローバル関数（`slog.InfoContext`）で呼び出している。
`pet_service.go` では `*slog.Logger` をコンストラクタで注入し、`s.logger.InfoContext` で使用している。

## 問題

```go
// 現在（グローバル slog）
type diagnosisCategoryService struct {
    repo repository.DiagnosisCategoryRepository
    // logger フィールドがない
}

func (s *diagnosisCategoryService) Create(...) {
    // ...
    slog.InfoContext(ctx, "diagnosis category created", ...)  // グローバル呼び出し
}
```

グローバル slog はテスト時のログ差し替えが困難で、pet service の設計方針と不一致。

## 修正方針

```go
// pet_service.go と同パターン
type diagnosisCategoryService struct {
    repo   repository.DiagnosisCategoryRepository
    logger *slog.Logger
}

func NewDiagnosisCategoryService(
    repo repository.DiagnosisCategoryRepository,
    logger *slog.Logger,
) DiagnosisCategoryService {
    return &diagnosisCategoryService{repo: repo, logger: logger}
}

func (s *diagnosisCategoryService) Create(...) {
    // ...
    s.logger.InfoContext(ctx, "diagnosis category created", ...)
}
```

同様に `diagnosisNameService` も同じ変更を適用。

## 完了条件

- [ ] `diagnosisCategoryService` に `logger *slog.Logger` フィールド追加
- [ ] `diagnosisNameService` に `logger *slog.Logger` フィールド追加
- [ ] 各コンストラクタで `logger` を受け取るように変更
- [ ] 全ログ呼び出しを `s.logger.InfoContext / ErrorContext` に変更
- [ ] `service.go` の DI 配線でロガーを渡すよう更新
