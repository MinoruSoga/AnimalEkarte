# BUG-390: procedure_service・vaccine_service の Delete が子リソースの存在チェックを行わない

## 概要
`procedure_service.go` と `vaccine_service.go` の `Delete` メソッドは、対象レコードが「使用中か（CountUsageByXxxID）」は確認するが、「子リソースが存在するか（CountChildrenByParentID）」を確認しない。
`procedure` と `vaccine` は共に `parent_id` を持つ階層構造（カテゴリ→アイテム）であるため、カテゴリを削除すると子アイテムが親なし（孤児）状態になりデータ整合性が壊れる。
`medicine_service.go` はこの問題を正しく実装しており、参照実装として機能する。

## 再現手順
1. 任意の procedure または vaccine に `parent_id = NULL`（カテゴリ）を作成する
2. そのカテゴリの子アイテムを 1 件以上作成する
3. カテゴリを `DELETE /masters/procedures/:id` で削除する
4. **結果**: 200/204 が返り、子アイテムが `parent_id` をそのまま持つ孤児状態になる

## 期待する動作
- カテゴリ（`parent_id = NULL`）削除時に子アイテムが存在する場合は 409 Conflict を返す
- カテゴリ削除を許可する場合は子アイテムを先に移動または削除させる

## 現状コード

### 問題1: `backend/internal/service/procedure_service.go:131-144`
```go
func (s *procedureService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByProcedureID(ctx, clinicID, id)
    // ↑ 使用中チェックのみ
    // ↑ CountChildrenByParentID が存在しないため子リソースチェックなし
    if err != nil {
        return apperrors.Wrap(err, "failed to check procedure dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
    }
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete procedure")
    }
    ...
}
```

### 問題2: `backend/internal/service/vaccine_service.go:156-169`
```go
func (s *vaccineService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByVaccineID(ctx, clinicID, id)
    // ↑ 使用中チェックのみ
    // ↑ CountChildrenByParentID が存在しないため子リソースチェックなし
    if err != nil {
        return apperrors.Wrap(err, "failed to check vaccine dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("このワクチンはワクチン接種記録で使用中のため削除できません")
    }
    ...
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/medicine_service.go:255-281
func (s *medicineService) Delete(ctx context.Context, clinicID, id uint64) error {
    m, err := s.repo.FindByID(ctx, clinicID, id)
    ...
    if m.ParentID == nil {
        // カテゴリの場合: 子アイテムが存在すれば削除を拒否する ✅
        count, err := s.repo.CountChildren(ctx, clinicID, id)
        if count > 0 {
            return apperrors.WrapConflict(
                fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count),
            )
        }
    } else {
        // アイテムの場合: 使用中チェック ✅
        usageCount, err := s.repo.CountUsageByMedicineID(ctx, id)
        ...
    }
}
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/repository/procedure_repository.go` | `CountChildrenByParentID(ctx, clinicID, parentID uint64)` メソッド追加 |
| `backend/internal/service/procedure_service.go:131-144` | Delete に `ParentID == nil` チェックと子数確認を追加 |
| `backend/internal/repository/vaccine_repository.go` | `CountChildrenByParentID(ctx, clinicID, parentID uint64)` メソッド追加 |
| `backend/internal/service/vaccine_service.go:156-169` | Delete に `ParentID == nil` チェックと子数確認を追加 |
| `backend/internal/service/procedure_service_test.go` | カテゴリ削除シナリオのテスト追加 |
| `backend/internal/service/vaccine_service_test.go` | カテゴリ削除シナリオのテスト追加 |

## 修正方針

### 1. `procedure_repository.go` — CountChildrenByParentID 追加
```go
func (r *procedureRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).
        Model(&model.Procedure{}).
        Where("parent_id = ? AND deleted_at IS NULL", parentID).
        Count(&count).Error
    return count, apperrors.FromGORM(err, "procedure", "")
}
```

### 2. `procedure_service.go:Delete` — 子チェック追加
```go
func (s *procedureService) Delete(ctx context.Context, clinicID, id uint64) error {
    proc, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to get procedure")
    }

    if proc.ParentID == nil {
        // カテゴリの場合: 子アイテムが存在すれば削除を拒否
        childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
        if err != nil {
            return apperrors.Wrap(err, "failed to count procedure children")
        }
        if childCount > 0 {
            return apperrors.WrapConflict(
                fmt.Sprintf("このカテゴリには%d件の診療項目が含まれています。先に移動または削除してください", childCount),
            )
        }
    } else {
        // アイテムの場合: 使用中チェック
        count, err := s.repo.CountUsageByProcedureID(ctx, clinicID, id)
        if err != nil {
            return apperrors.Wrap(err, "failed to check procedure dependencies")
        }
        if count > 0 {
            return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
        }
    }

    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete procedure")
    }
    ...
}
```

`vaccine_service.go` も同パターンで修正。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — マスタ削除の FK 依存チェック
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。

### プロジェクト内参照実装
`backend/internal/service/medicine_service.go` — `ParentID == nil` 分岐による正しい実装

## 優先度
**High** — データ整合性の問題。親カテゴリを削除すると子アイテムが孤児になり、UI に表示されなくなるが DB には残存する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/procedure_service.go:131-144` — 問題箇所
- `backend/internal/repository/procedure_repository.go` — CountChildrenByParentID 追加が必要
- `backend/internal/service/vaccine_service.go:156-169` — 問題箇所
- `backend/internal/repository/vaccine_repository.go` — CountChildrenByParentID 追加が必要
- `backend/internal/service/medicine_service.go:255-281` — 参照実装
