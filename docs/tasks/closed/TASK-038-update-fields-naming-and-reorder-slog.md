# TASK-038: occupation UpdateFields命名 & 2-query race + inquiry_template 同問題 + cage Reorder slog

## 優先度

HIGH（occupation/inquiry_template UpdateFields）/ MEDIUM（cage Reorder slog）

---

## 問題 1: occupation_repository の Update メソッドが UpdateFields パターン未準拠

### ファイル
`backend/internal/repository/occupation_repository.go:20, 64-76`

### 問題
TASK-029 で `diagnosis_repository`・`chief_complaint_type_repository` の `Update` → `UpdateFields` 統一を指摘したが、`occupation_repository` が同じ問題を抱えており対象外だった。

```go
// インターフェース（L20）
Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
// → 戻り値が error のみ。exam_type/insurance/procedure は (*model.Xxx, error) を返す。
```

これにより service 層が「Update → 別途 FindByID」の 2-query パターンに陥っている（下記 問題 2 参照）。

### 修正案
```go
// インターフェース
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)

// 実装（exam_type_repository.go パターンに合わせる）
func (r *occupationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
    result := r.db.WithContext(ctx).
        Model(&model.Occupation{}).
        Scopes(clinicScope(clinicID)).Where("id = ?", id).
        Updates(fields)
    if result.Error != nil {
        return nil, apperrors.FromGORM(result.Error, "occupation", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return nil, apperrors.WrapNotFound("occupation", fmt.Sprintf("%d", id))
    }
    return r.FindByID(ctx, clinicID, id)
}
```

---

## 問題 2: occupation_service の Update が 2-query race（FindByID 分離）

### ファイル
`backend/internal/service/occupation_service.go:93-103`

### 問題
```go
// 問題のある 2-query パターン
if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
    return nil, apperrors.Wrap(err, "failed to update occupation")
}
slog.InfoContext(ctx, "occupation updated", ...)
result, err := s.repo.FindByID(ctx, clinicID, id)  // ← 別クエリで取り直し
```

問題 1 の repository 修正で自動的に解消されるが、service 側も以下に変更が必要。

### 修正案
```go
occupation, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to update occupation")
}
slog.InfoContext(ctx, "occupation updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("occupation_id", id))
return occupation, nil
```

---

## 問題 3: inquiry_template_repository も同じ Update 戻り値問題

### ファイル
`backend/internal/repository/inquiry_template_repository.go:20`

### 問題
```go
// インターフェース
Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
// → 問題 1 と同じ。戻り値が error のみ。
```

`inquiry_template_service.go` でも Update 後に FindByID を別途呼んでいる。

### 修正案
occupation と同様に `UpdateFields(...) (*model.InquiryTemplate, error)` に変更し、service の 2-query を解消する。

---

## 問題 4: cage_service の Reorder に slog.InfoContext なし

### ファイル
`backend/internal/service/cage_service.go:112-118`

### 問題
```go
func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if len(ids) == 0 {
        return apperrors.WrapInvalidInput("ids must not be empty")
    }
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder cage")
    }
    return nil  // slog なし
}
```

cage の Create(L75)・Delete(L106) には clinic_id 付き slog があるが Reorder のみ欠落。

### 修正案
```go
slog.InfoContext(ctx, "cages reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
return nil
```
