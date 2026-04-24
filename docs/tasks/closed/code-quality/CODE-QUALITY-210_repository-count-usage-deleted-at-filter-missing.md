# CODE-QUALITY-210: CountUsage 系メソッドの `deleted_at IS NULL` フィルタ欠落

## 概要

複数の Repository の `CountUsageByXxxID` / `CountXxxByYyyID` メソッドで、
JOIN 先テーブルに `deleted_at IS NULL` フィルタが欠落している。
論理削除済みレコードを「使用中」として誤カウントし、削除可能なマスタが削除できなくなる。

## 優先度

HIGH

## 影響ファイル

| ファイル | メソッド | 欠落テーブル | 行番号 |
|---------|---------|------------|--------|
| `backend/internal/repository/procedure_repository.go` | CountUsageByProcedureID | treatments | ~L91 |
| `backend/internal/repository/consultation_repository.go` | CountUsageByConsultationID | treatments | ~L112 |
| `backend/internal/repository/chief_complaint_repository.go` | CountUsageByChiefComplaintTypeID | inquiries | ~L89 |
| `backend/internal/repository/permission_group_repository.go` | CountStaffsByGroupID | staffs | ~L135 |

---

## 参照実装（medicine_repository.go — 正しい実装）

```go
// treatments.deleted_at IS NULL を正しく付けている例
Joins("JOIN treatments ON treatments.medicine_id = medicine_items.medicine_id AND treatments.deleted_at IS NULL").
```

---

## 問題詳細

### 1. procedure_repository.go:91 — treatments.deleted_at 欠落

```go
// 現状（誤）
Where("treatments.procedure_id = ?", procedureID)

// 修正後
Where("treatments.procedure_id = ? AND treatments.deleted_at IS NULL", procedureID)
```

`treatments` は `gorm.DeletedAt` フィールドを持つ。論理削除済みの treatment も参照ありとしてカウントされ、
削除可能な procedure マスタが「使用中」と判定される。

---

### 2. consultation_repository.go:112 — treatments.deleted_at 欠落

procedure と同じ問題。`treatments.consultation_id` を参照する treatments が soft delete 対象。

```go
// 現状（誤）
Where("treatments.consultation_id = ?", consultationID)

// 修正後
Where("treatments.consultation_id = ? AND treatments.deleted_at IS NULL", consultationID)
```

---

### 3. chief_complaint_repository.go:89 — inquiries.deleted_at 欠落

`inquiries` テーブルは `gorm.DeletedAt` を持つ。JOIN 先の `medical_records.deleted_at IS NULL` は存在するが、
主テーブル相当の `inquiries` 側の soft delete フィルタが欠落。

```go
// 現状（誤）
Where("inquiries.chief_complaint_type_id = ?", id)

// 修正後
Where("inquiries.chief_complaint_type_id = ? AND inquiries.deleted_at IS NULL", id)
```

---

### 4. permission_group_repository.go:135 — staffs.deleted_at 欠落

`CountStaffsByGroupID` で `staff_permission_groups` を基点にしているが、
`staffs` への JOIN をして論理削除済みスタッフをフィルタする必要がある。

```go
// 現状（不完全）
// staffs への JOIN なし、staffs.deleted_at IS NULL フィルタなし

// 修正後
Joins("JOIN staffs ON staffs.id = staff_permission_groups.staff_id AND staffs.deleted_at IS NULL").
Where("staff_permission_groups.group_id = ?", groupID)
```

---

## 規約参照

- `.claude/rules/database-design.md`: 論理削除対応（CountUsage の JOIN は `deleted_at IS NULL` を含む）
- `medicine_repository.go` が参照実装

## テスト

各マスタで:
- 論理削除済みレコードを参照している場合でも「未使用」と正しく判定されることを確認
- 有効なレコードを参照している場合は「使用中」と判定されることを確認
