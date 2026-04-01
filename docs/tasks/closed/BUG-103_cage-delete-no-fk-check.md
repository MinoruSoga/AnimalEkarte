# BUG-103: ケージ削除（FK制約なし）→ 入院データの cage_id が孤立

## 概要

`DELETE /api/v1/masters/cages/:id` がケージに紐付く入院データの存在チェックをせず
204 No Content で削除成功してしまう。`hospitalization_records.cage_id` が孤立しデータ整合性が破壊される。

## 症状

- `/settings/cage` でケージ「ICUケージA」の削除を実行
- DELETE /api/v1/masters/cages/1 → **204 No Content**
- リスト 8件 → 7件（削除成功）
- 入院データの cage_id が孤立（NULL 参照またはデータ不整合）

## 期待する動作

- 入院データ（`hospitalization_records`）に当該 cage_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されています。先に依存するデータを削除してください」
- リストは変化しない

## 根本原因

`backend/internal/service/cage_service.go` の Delete メソッドに依存チェックがない。
`hospitalization_records` テーブルの `cage_id` FK 参照が存在するかのチェックを追加していない。

## 修正内容

```go
// cage_service.go Delete() 内
count, err := s.repo.CountHospitalizationsByCageID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check cage dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/cage_service.go`
- `backend/internal/repository/cage_repository.go`（CountHospitalizationsByCageID 追加）

## 優先度

High（データ整合性破壊）

## 関連

- BUG-030（サービス種別・スタッフは修正済み、ケージは未修正）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/cages/1 [204] 確認
