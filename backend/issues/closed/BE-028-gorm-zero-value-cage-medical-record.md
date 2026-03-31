# BE-028: GORM zero-value問題 — cage / medical_record リポジトリのUpdate修正

## 問題
`cage_repository.go` と `medical_record_repository.go` が PATCH更新で
`.Updates(struct)` を使用している。GORMはzero値（0, false, "", nil）を
構造体更新でスキップするため、フィールドをゼロ値にリセットできない。

## 影響ファイル
- `backend/internal/repository/cage_repository.go:64-67`
- `backend/internal/repository/medical_record_repository.go:95-99`

## 現状（NG）
```go
result := r.db.WithContext(ctx).Model(&model.Cage{}).
    Where("id = ? AND clinic_id = ?", cage.ID, cage.ClinicID).
    Updates(cage)  // ❌ struct update — zero値をスキップ
```

## 修正方針
1. `buildCageUpdateFields(input *UpdateCageInput) map[string]any` を実装
2. `buildMedicalRecordUpdateFields(input *UpdateMedicalRecordInput) map[string]any` を実装
3. `.Updates(map[string]any)` に変更

## 参照実装
`backend/internal/service/staff_service.go` — `buildStaffUpdateFields()` パターン

## 優先度
HIGH（データ破損リスク）
