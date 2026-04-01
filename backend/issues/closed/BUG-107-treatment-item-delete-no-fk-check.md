---
status: closed
---

# BE: BUG-107 診療項目削除（FK依存チェックなし）→ 204 でカルテ参照孤立

## 概要

診療項目系マスタの DELETE エンドポイントが `clinical_plan_items` への参照チェックをせず 204 を返す。
対象エンドポイントは 5 種類。

## 影響エンドポイント

| エンドポイント | service ファイル |
|---|---|
| DELETE /api/v1/masters/consultations/:id | `consultation_service.go` |
| DELETE /api/v1/masters/examination-types/:id | `examination_type_service.go` |
| DELETE /api/v1/masters/procedures/:id | `procedure_service.go` |
| DELETE /api/v1/masters/vaccines/:id | `vaccine_service.go` |
| DELETE /api/v1/masters/checkup-types/:id | `checkup_type_service.go` |

## 再現手順

```
DELETE /api/v1/masters/consultations/1 → 204 No Content（clinical_plan_items 参照存在時も削除成功）
```

## 期待する動作

- `clinical_plan_items` に当該 ID が存在する場合 → **409 Conflict**
- エラー: 「このデータは他のレコードに使用されているため削除できません」

## 実装場所

各 service の Delete() に FK 依存チェックを追加。対応する repository に Count メソッドを追加。

```go
// 例: consultation_service.go
count, err := s.repo.CountClinicalPlanItemsByConsultationID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check consultation dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 優先度

High（カルテデータ整合性破壊）

## 関連

- BUG-030（同種パターン）
- `docs/tasks/open/crash/BUG-107_treatment-item-delete-no-fk-check.md`
