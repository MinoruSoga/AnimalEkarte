# TASK-096: repository — Update vs UpdateFields 未対応リポジトリ（reservation_type / shift_template）

## 優先度

MEDIUM

---

## 概要

TASK-090 で 4 リポジトリ（medicine, reservation_type_group, permission_group, animal_species）の
`Update(...) error` → `UpdateFields(...) (*model.Xxx, error)` 変換が対象となったが、
以下の 2 リポジトリが漏れており引き続き古いパターンを使用している。

---

## 問題箇所

### reservation_type_repository.go:62

```go
// ❌ Update のみ実装: 戻り値が error のみ
func (r *reservationTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
```

Service 側（`reservation_type_service.go:275,279`）では Update 後に別途 `FindByID` を呼び出し:
```go
if err := s.repo.Update(ctx, clinicID, id, fields); err != nil { ... }
// ↓ 2回目のクエリが発生
result, err := s.repo.FindByID(ctx, clinicID, id)
```

### shift_template_repository.go:67-82

```go
// ❌ Update のみ実装: 戻り値が error のみ
func (r *shiftTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
```

Service 側（`shift_template_service.go:114,128`）では Update 後に別途 `FindByID` を呼び出し:
```go
if err := s.repo.Update(ctx, clinicID, id, fields); err != nil { ... }
// ↓ 2回目のクエリが発生
result, err := s.repo.FindByID(ctx, clinicID, id)
```

---

## 修正方針（TASK-090 参照）

```go
// ✅ 修正後: reservation_type_repository.go
func (r *reservationTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
    var rt model.ReservationType
    err := r.db.WithContext(ctx).
        Model(&rt).
        Scopes(clinicScope(clinicID)).Where("id = ?", id).
        Updates(fields).
        First(&rt).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "reservation_type", strconv.FormatUint(id, 10))
    }
    return &rt, nil
}

// ✅ 修正後: shift_template_repository.go（ReplaceBreaks との組み合わせ注意）
func (r *shiftTemplateRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
    // ...
}
```

Service 側の `FindByID` 呼び出しを削除して 1 クエリに集約。

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `repository/reservation_type_repository.go` | `Update` → `UpdateFields`（戻り値 `(*model.ReservationType, error)`） |
| `repository/shift_template_repository.go` | `Update` → `UpdateFields`（戻り値 `(*model.ShiftTemplate, error)`） |
| `service/reservation_type_service.go` | `Update()` + `FindByID()` 二重呼び出しを `UpdateFields()` 1 呼び出しに変更 |
| `service/shift_template_service.go` | 同上 |

---

## 関連

- TASK-090: 同種問題（medicine/reservation_type_group/permission_group/animal_species） — クローズ済み
