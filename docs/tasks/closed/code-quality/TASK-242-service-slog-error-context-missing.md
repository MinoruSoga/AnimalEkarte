# TASK-242: service — 複数ファイルのエラーパスに slog.ErrorContext が欠落

## 優先度
Medium

## 対象ファイル
- `backend/internal/service/vaccine_service.go`
- `backend/internal/service/closing_settings_service.go`
- `backend/internal/service/procedure_service.go`

## 問題概要
規約: **Service 層は全エラーパスで `slog.ErrorContext(ctx, ...)` を記録する。**

以下のファイルで、エラーをラップして return するコードパスに `slog.ErrorContext` が欠落している。

## 具体的な欠落箇所

### vaccine_service.go — Update のエラーパス

```go
func (s *vaccineService) Update(...) (*model.Vaccine, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get vaccine")  // ❌ slog なし
    }
    // ...
    vaccine, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update vaccine")  // ❌ slog なし
    }
```

### procedure_service.go — Update のエラーパス

```go
procedure, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to update procedure")  // ❌ slog なし
}
```

### closing_settings_service.go — DeleteSpecialPeriod のエラーパス

```go
func (s *closingSettingsService) DeleteSpecialPeriod(...) error {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to get special period")  // ❌ slog なし
    }
    // ...
    return s.periodRepo.Delete(ctx, clinicID, id)  // ❌ エラーを wrap せず、slog なし
}
```

## あるべき姿（参照: cage_service.go）

```go
func (s *cageService) Update(...) (*model.Cage, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        slog.ErrorContext(ctx, "failed to get cage", slog.Any("error", err))  // ✅
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // ...
    cage, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        slog.ErrorContext(ctx, "failed to update cage", slog.Any("error", err))  // ✅
        return nil, apperrors.Wrap(err, "failed to update cage")
    }
```

## 完了条件
- [ ] `vaccine_service.go` の Update エラーパスに `slog.ErrorContext` を追加
- [ ] `procedure_service.go` の Update エラーパスに `slog.ErrorContext` を追加
- [ ] `closing_settings_service.go` の DeleteSpecialPeriod エラーパスに `slog.ErrorContext` を追加
- [ ] `closing_settings_service.go` の DeleteSpecialPeriod 末尾を `apperrors.Wrap` でラップ
- [ ] `go test ./backend/internal/...` がパス
