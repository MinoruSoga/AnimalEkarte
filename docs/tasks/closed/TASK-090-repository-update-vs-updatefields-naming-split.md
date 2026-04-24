# TASK-090: repository — Update vs UpdateFields メソッド命名・戻り値型の不統一

## 優先度

MEDIUM

---

## 概要

マスタ系 repository 15本のうち、4本が `Update(...)error` を使用し、
11本が `UpdateFields(...) (*model.Xxx, error)` を使用している。

メソッド名と戻り値型が不統一であり、呼び出し元 service の実装パターンも分岐する原因になっている。

---

## 問題箇所

### Update パターン（4本 — 戻り値: error のみ）

```go
// ❌ medicine_repository.go
func (r *medicineRepository) Update(ctx context.Context, id uint64, fields map[string]any) error

// ❌ reservation_type_group_repository.go
func (r *reservationTypeGroupRepository) Update(ctx context.Context, id uint64, fields map[string]any) error

// ❌ permission_group_repository.go
func (r *permissionGroupRepository) Update(ctx context.Context, id uint64, fields map[string]any) error

// ❌ animal_species_repository.go
func (r *animalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) error
```

これら 4本は `Update()` 後に service 側で `FindByID()` を再度呼び出して
更新後のエンティティを取得する二重クエリが発生している。

---

### UpdateFields パターン（11本 — 戻り値: (*model.Xxx, error)）

```go
// ✅ exam_type_repository.go（参照実装）
func (r *examTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExamType, error)
```

`UpdateFields` は UPDATE + SELECT を 1 トランザクションで実行し、
更新後エンティティを返すため service 側で `FindByID` を追加呼び出しする必要がない。

---

## 影響

- `Update` 利用の 4 サービスで不要な `FindByID` クエリが発生（N→N+1 クエリ）
- メソッド名の不統一により、service 層のパターンが 2 種類に分岐
- 新規 repository 作成時にどちらのパターンを採用すべきか判断できない

---

## 修正方針

**全マスタ repository を `UpdateFields` パターンに統一する。**

```go
// ✅ 統一後（medicine を例に）
// repository interface
type MedicineRepository interface {
    UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
    // ...
}

// repository 実装
func (r *medicineRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error) {
    var medicine model.Medicine
    err := r.db.WithContext(ctx).
        Model(&medicine).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        Updates(fields).
        First(&medicine).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", id))
    }
    return &medicine, nil
}

// service 側: FindByID 不要
func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
    fields := buildMedicineUpdateFields(input)
    return s.repo.UpdateFields(ctx, clinicID, id, fields)  // 1 クエリで完結
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `repository/medicine_repository.go` | `Update` → `UpdateFields`、戻り値を `(*model.Medicine, error)` に変更 |
| `repository/reservation_type_group_repository.go` | 同上 |
| `repository/permission_group_repository.go` | 同上 |
| `repository/animal_species_repository.go` | 同上 |
| `service/medicine_service.go` | `UpdateFields` 戻り値を使用し `FindByID` を削除 |
| `service/reservation_type_group_service.go` | 同上 |
| `service/permission_group_service.go` | 同上 |
| `service/animal_species_service.go` | 同上 |

---

## 参考: 全マスタ repository 状況

| ファイル | メソッド名 | 戻り値 | 状態 |
|---------|-----------|--------|------|
| `medicine_repository.go` | `Update` | `error` | ❌ 要修正 |
| `reservation_type_group_repository.go` | `Update` | `error` | ❌ 要修正 |
| `permission_group_repository.go` | `Update` | `error` | ❌ 要修正 |
| `animal_species_repository.go` | `Update` | `error` | ❌ 要修正 |
| `exam_type_repository.go` | `UpdateFields` | `(*model.ExamType, error)` | ✅ |
| `checkup_type_repository.go` | `UpdateFields` | `(*model.CheckupType, error)` | ✅ |
| `vaccine_repository.go` | `UpdateFields` | `(*model.Vaccine, error)` | ✅ |
| `cage_repository.go` | `UpdateFields` | `(*model.Cage, error)` | ✅ |
| `insurance_repository.go` | `UpdateFields` | `(*model.Insurance, error)` | ✅ |
| `procedure_repository.go` | `UpdateFields` | `(*model.Procedure, error)` | ✅ |
| `merchandise_item_repository.go` | `UpdateFields` | `(*model.MerchandiseItem, error)` | ✅ |
| `occupation_repository.go` | `UpdateFields` | `(*model.Occupation, error)` | ✅ |
| `chief_complaint_type_repository.go` | `UpdateFields` | `(*model.ChiefComplaintType, error)` | ✅ |
| `diagnosis_repository.go` | `UpdateFields` | `(*model.DiagnosisType, error)` | ✅ |
| `inquiry_template_repository.go` | `UpdateFields` | `(*model.InquiryTemplate, error)` | ✅ |
