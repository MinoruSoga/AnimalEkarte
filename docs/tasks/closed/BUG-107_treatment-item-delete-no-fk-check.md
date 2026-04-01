# BUG-107: 診療項目削除（FK制約なし）→ カルテ参照データが孤立

## 概要

`DELETE /api/v1/masters/consultations/:id`（および検査・処置・予防接種・定期健診の各エンドポイント）が、
当該マスタ項目を参照している `clinical_plan_items` の存在チェックをせず 204 No Content で削除成功してしまう。
カルテの治療プランに使われた診療項目が削除されると参照データが孤立し、データ整合性が破壊される。

## 症状

- `/settings/treatment-items`（診察タブ）で「初診料」の削除を実行
- DELETE /api/v1/masters/consultations/1 → **204 No Content**
- トースト: 「削除しました」（削除成功）
- 件数: 5件 → 4件（削除完了）
- `clinical_plan_items.consultation_id = 1` が残存している場合、孤立参照が発生

## 期待する動作

- `clinical_plan_items`（または同等のテーブル）に当該 item_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」
- リストは変化しない（件数維持）

## 影響エンドポイント

| エンドポイント | マスタ種別 |
|---|---|
| DELETE /api/v1/masters/consultations/:id | 診察 |
| DELETE /api/v1/masters/examination-types/:id | 検査 |
| DELETE /api/v1/masters/procedures/:id | 処置 |
| DELETE /api/v1/masters/vaccines/:id | 予防接種 |
| DELETE /api/v1/masters/checkup-types/:id | 定期健診 |

## 根本原因

各 service の Delete メソッドに FK 依存チェックがない。
`service_type_service.go`・`staff_service.go` では修正済みのパターンが未適用。

```go
// 例: consultation_service.go Delete() に追加すべき
count, err := s.repo.CountClinicalPlanItemsByConsultationID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check consultation dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/consultation_service.go`（および examination_type, procedure, vaccine, checkup_type）
- `backend/internal/repository/consultation_repository.go`（CountClinicalPlanItemsByConsultationID 追加）

## 優先度

High（データ整合性破壊）

## 関連

- BUG-030（サービス種別・スタッフは修正済み。診療項目各種は未対応）
- BUG-103（ケージ同様の修正漏れ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/consultations/1 [204] 確認
