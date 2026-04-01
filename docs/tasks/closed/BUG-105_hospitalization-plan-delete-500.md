# BUG-105: 入院プラン削除（FK制約クラッシュ）→ 500 Internal Server Error

## 概要

`DELETE /api/v1/masters/hospitalization-plans/:id` が FK 制約違反を適切にハンドリングせず、
500 Internal Server Error を返す。サービス種別・スタッフでは修正済みの同種バグが
入院プランハンドラでは未修正のまま残っている。

## 症状

- `/settings/hospitalization` で「一般入院（小型）」の削除を実行
- DELETE /api/v1/masters/hospitalization-plans/1 → **500 Internal Server Error**
- トースト: 「入院プランの削除に失敗しました」（generic メッセージ）
- リスト 5件維持（削除はブロックされるが 500 で応答）

## 期待する動作

- 入院データ（`hospitalization_records`）に当該 plan_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されています」
- 500 ではなく 409 を返すこと

## 根本原因

`backend/internal/service/hospitalization_plan_service.go`（または handler）が
FK 制約エラー（`apperrors.ErrConflict`）をハンドリングせずそのままクラッシュしている。
サービス種別（`service_type_service.go`）では修正済みのパターンが未適用。

## 修正内容

サービス種別の修正と同様に、DeleteメソッドにFK依存チェックを追加する：

```go
// hospitalization_plan_service.go Delete() 内
count, err := s.repo.CountRecordsByPlanID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check plan dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/hospitalization_plan_service.go`
- `backend/internal/repository/hospitalization_plan_repository.go`（CountRecordsByPlanID 追加）

## 優先度

High（500 エラーはクラッシュ相当）

## 関連

- BUG-030（サービス種別・スタッフは修正済み、入院プランは対象外で漏れ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/hospitalization-plans/1 [500] 確認
