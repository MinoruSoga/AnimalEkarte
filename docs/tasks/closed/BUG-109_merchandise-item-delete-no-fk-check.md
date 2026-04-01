# BUG-109: 物販品目削除（FK制約なし）→ 請求・在庫データが孤立

## 概要

`DELETE /api/v1/masters/merchandise-items/:id` が当該品目を参照している
`treatment_items` や `invoice_items` の存在チェックをせず 204 No Content で削除成功してしまう。
カルテの治療明細や請求書に使われた物販品目が削除されるとデータ整合性が破壊される。

## 症状

- `/settings/merchandise-items` で「ロイヤルカナン 消化器サポート 1kg」の削除を実行
- DELETE /api/v1/masters/merchandise-items/1 → **204 No Content**
- トースト: 「品目を削除しました」（削除成功）
- 件数: 6件 → 5件（削除完了）
- FK 依存チェックなし

## 期待する動作

- `treatment_items`（または `invoice_items`）に当該 item_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」
- リストは変化しない

## 根本原因

`backend/internal/service/merchandise_item_service.go` の Delete メソッドに依存チェックがない。

```go
// merchandise_item_service.go Delete() 内
count, err := s.repo.CountTreatmentItemsByMerchandiseID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check merchandise dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/merchandise_item_service.go`
- `backend/internal/repository/merchandise_item_repository.go`（CountTreatmentItemsByMerchandiseID 追加）

## 優先度

High（データ整合性破壊）

## 関連

- BUG-030（サービス種別・スタッフは修正済み。物販は修正漏れ）
- BUG-103, BUG-107（同種の依存チェックなしバグ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/merchandise-items/1 [204] 確認
