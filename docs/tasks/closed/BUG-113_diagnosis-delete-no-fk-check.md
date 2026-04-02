# BUG-113: 診断病名マスタ削除（FK制約なし）→ カルテ参照データが孤立

## 概要

`DELETE /api/v1/masters/diagnosis-categories/:id` および
`DELETE /api/v1/masters/diagnosis-names/:id` が依存チェックをせず
204 No Content で削除成功してしまう。

- カテゴリ削除: `diagnosis_names.category_id` FK 参照が孤立
- 病名削除: `clinical_plan_items.diagnosis_name_id`（または同等） FK 参照が孤立し、カルテデータの整合性が破壊される

## 症状

- `/settings/diagnosis` 「診断病名カテゴリ」タブで「呼吸器系」の削除を実行
  - DELETE /api/v1/masters/diagnosis-categories/2 → **204 No Content**
  - トースト: 「診断カテゴリを削除しました」（削除成功）
  - 確認ダイアログに「このカテゴリに属する診断名も影響を受けます。」と表示されるが削除は止められない

- 「診断病名」タブで「てんかん」の削除を実行
  - DELETE /api/v1/masters/diagnosis-names/12 → **204 No Content**
  - トースト: 「診断病名を削除しました」（削除成功）
  - `clinical_plan_items` に当該 diagnosis_name_id が存在する場合、孤立参照が発生

## 期待する動作

### 診断病名削除
- `clinical_plan_items`（または同等のテーブル）に当該 diagnosis_name_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」

### 診断カテゴリ削除
- 配下に診断病名が存在する場合 → **409 Conflict**
- または、配下の診断病名を先に全削除することを要求する

## 根本原因

`backend/internal/service/diagnosis_name_service.go` および
`backend/internal/service/diagnosis_category_service.go` の Delete メソッドに依存チェックがない。

```go
// diagnosis_name_service.go Delete() 内
count, err := s.repo.CountClinicalPlanItemsByDiagnosisNameID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check diagnosis name dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}

// diagnosis_category_service.go Delete() 内
count, err := s.repo.CountDiagnosisNamesByCategoryID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check diagnosis category dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/diagnosis_name_service.go`
- `backend/internal/service/diagnosis_category_service.go`
- `backend/internal/repository/diagnosis_name_repository.go`（CountClinicalPlanItemsByDiagnosisNameID 追加）
- `backend/internal/repository/diagnosis_category_repository.go`（CountDiagnosisNamesByCategoryID 追加）

## 優先度

High（カルテデータ整合性破壊）

## 関連

- BUG-030（同種の依存チェックなしバグ）
- BUG-107（診療項目と同種パターン）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/diagnosis-categories/2 [204] 確認
- DELETE http://localhost:8080/api/v1/masters/diagnosis-names/12 [204] 確認
