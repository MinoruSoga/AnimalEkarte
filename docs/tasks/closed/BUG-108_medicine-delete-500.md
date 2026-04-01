# BUG-108: 薬剤削除（FK制約クラッシュ）→ 500 Internal Server Error

## 概要

`DELETE /api/v1/masters/medicines/:id` が FK 制約違反を適切にハンドリングせず、
500 Internal Server Error を返す。サービス種別・スタッフでは修正済みの同種バグが
薬剤ハンドラでは未修正のまま残っている。

## 症状

- `/settings/medicine` で「アモキシシリン 50mg」の削除を実行
- DELETE /api/v1/masters/medicines/1 → **500 Internal Server Error**
- トースト: 「削除に失敗しました」（generic メッセージ）
- リスト 24件維持（削除はブロックされるが 500 で応答）

## 期待する動作

- `treatment_items`（または薬剤を参照するテーブル）に当該 medicine_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」
- 500 ではなく 409 を返すこと

## 根本原因

`backend/internal/service/medicine_service.go`（または handler）が
FK 制約エラー（`apperrors.ErrConflict`）をハンドリングせずそのままクラッシュしている。
サービス種別（`service_type_service.go`）では修正済みのパターンが未適用。

## 修正内容

サービス種別の修正と同様に、Delete メソッドに FK 依存チェックを追加する：

```go
// medicine_service.go Delete() 内
count, err := s.repo.CountTreatmentItemsByMedicineID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check medicine dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/medicine_service.go`
- `backend/internal/repository/medicine_repository.go`（CountTreatmentItemsByMedicineID 追加）

## 優先度

High（500 エラーはクラッシュ相当）

## 関連

- BUG-030（サービス種別・スタッフは修正済み、薬剤は対象外で漏れ）
- BUG-105（入院プランと同種の 500 クラッシュ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/medicines/1 [500] 確認
