# BUG-263: Service 層 slog 監査ログ欠落（第2波）

## 概要

BUG-257（第1波）で8サービスを修正したが、追加の CUD メソッドに `slog.InfoContext` が欠落している。
成功した Create/Delete/Reorder 操作の監査証跡が残らない。

## 影響範囲

| ファイル | 行 | メソッド | 欠落内容 |
|----------|-----|---------|----------|
| `trimming_master_service.go` | 39 | TrimmingCourse.Create | slog なし |
| `trimming_master_service.go` | 61 | TrimmingCourse.Delete | slog なし |
| `trimming_master_service.go` | 65 | TrimmingCourse.Reorder | slog なし |
| `trimming_master_service.go` | 131 | TrimmingOption.Create | slog なし |
| `trimming_master_service.go` | 153 | TrimmingOption.Delete | slog なし |
| `trimming_master_service.go` | 157 | TrimmingOption.Reorder | slog なし |
| `cage_service.go` | 47 | Create | slog なし |
| `cage_service.go` | 65 | Delete | slog なし |
| `clinic_service.go` | 123 | CreateClinic | slog なし |
| `clinic_service.go` | 136 | UpdateClinic | slog なし |
| `diagnosis_service.go` | 149 | DiagnosisCategory.Delete | slog なし |
| `diagnosis_service.go` | 286 | DiagnosisName.Delete | slog なし |
| `examination_service.go` | 45 | Create | slog なし |
| `hospitalization_plan_service.go` | 46 | Create | slog なし |
| `hospitalization_plan_service.go` | 64 | Delete | slog なし |
| `medical_record_service.go` | 172 | Delete | slog なし |
| `owner_service.go` | 422 | Delete | slog なし |
| `pet_service.go` | 356 | Delete | slog なし |

**合計: ~18箇所 / 8ファイル**

## 現状コード

### `trimming_master_service.go:38-40`（Create — slog なし）
```go
func (s *trimmingCourseService) Create(ctx context.Context, course *model.TrimmingCourse) error {
    return s.repo.Create(ctx, course) // naked return + slog なし
}
```

### 比較: 正しい実装（`vaccine_service.go:37-48`）
```go
func (s *vaccineService) Create(ctx context.Context, vaccine *model.Vaccine) error {
    if err := s.repo.Create(ctx, vaccine); err != nil {
        return apperrors.Wrap(err, "failed to create vaccine")
    }
    slog.InfoContext(ctx, "vaccine created",
        slog.Uint64("vaccine_id", vaccine.ID),
        slog.Uint64("clinic_id", vaccine.ClinicID))
    return nil
}
```

## 修正方針

BUG-262 の naked return 修正と同時に slog を追加する。

### 修正後コード例
```go
func (s *trimmingCourseService) Create(ctx context.Context, course *model.TrimmingCourse) error {
    if err := s.repo.Create(ctx, course); err != nil {
        return apperrors.Wrap(err, "failed to create trimming course")
    }
    slog.InfoContext(ctx, "trimming course created",
        slog.Uint64("trimming_course_id", course.ID),
        slog.Uint64("clinic_id", course.ClinicID))
    return nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — ログ（slog 構造化ログ）
> service層のみ。handler・repositoryには記述しない。

### `.claude/CLAUDE.md` — Context & Logging
> 構造化ログ `log/slog` を使用し、`InfoContext`, `ErrorContext` でコンテキストを適切に伝播させる。

## 優先度

**Medium** — 監査証跡の欠落。セキュリティ監査・障害調査に影響するが、機能には影響しない。

## 関連チケット

- BUG-257: slog 監査ログ欠落（第1波）
- BUG-262: Service naked return（第3波）— 同時に修正すべき
- BUG-261: 第3回監査 親チケット
