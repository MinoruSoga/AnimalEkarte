# CODE-QUALITY-203: consultation 削除時の子レコードチェック欠落

## 概要

`consultation`（診察項目）は `parent_id` フィールドによる階層構造を持つが、
削除前に子レコードの存在チェックが実装されていない。
`exam_type` / `checkup_type` は同様の階層構造で `CountChildrenByParentID` を実装済み。

## 優先度

HIGH

## 影響ファイル

| ファイル | 修正内容 |
|---------|---------|
| `backend/internal/repository/consultation_repository.go` | `CountChildrenByParentID` メソッド追加 |
| `backend/internal/service/consultation_service.go` | `Delete` に子レコードチェック追加 |

---

## 問題

### 現状（consultation_service.go の Delete）

```go
func (s *consultationService) Delete(ctx context.Context, clinicID, id uint64) error {
    // ← 子レコード（同 parent_id を参照する consultation）のチェックなし
    count, err := s.repo.CountUsageByConsultationID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check consultation usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この診察項目は使用中のため削除できません")
    }
    ...
}
```

### 比較: exam_type_service.go の Delete（正しい実装）

```go
func (s *examTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    // 子種別チェック
    childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check exam type children")
    }
    if childCount > 0 {
        return apperrors.WrapConflict("このカテゴリには子種別が登録されているため削除できません")
    }
    // 使用中チェック
    count, err := s.repo.CountUsageByExamTypeID(ctx, clinicID, id)
    ...
}
```

---

## 修正方針

### Step 1: consultation_repository.go に `CountChildrenByParentID` を追加

```go
// interface に追加
CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)

// 実装
func (r *consultationRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Consultation{}).
        Scopes(clinicScope(clinicID)).
        Where("parent_id = ?", parentID).
        Count(&count).Error
    if err != nil {
        return 0, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", parentID))
    }
    return count, nil
}
```

### Step 2: consultation_service.go の Delete にチェックを追加

```go
func (s *consultationService) Delete(ctx context.Context, clinicID, id uint64) error {
    // 子レコードチェック（追加）
    childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check consultation children")
    }
    if childCount > 0 {
        return apperrors.WrapConflict("この診察項目にはサブ項目が登録されているため削除できません")
    }
    // 使用中チェック（既存）
    count, err := s.repo.CountUsageByConsultationID(ctx, clinicID, id)
    ...
}
```

---

## 規約参照

- `.claude/CLAUDE.md` 1b節: マスタ削除の FK 依存チェック（CountUsageBy... / CountChildrenBy...）

## テスト

- 子 consultation が存在する親を削除しようとした場合に 409 Conflict が返ることを確認
- 子のない consultation は正常に削除できることを確認
