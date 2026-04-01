# BUG-110: 保険マスタ削除（FK制約なし）→ ペット保険情報が孤立

## 概要

`DELETE /api/v1/masters/insurances/:id` が当該保険に登録されているペットの存在チェックをせず
204 No Content で削除成功してしまう。`pets.insurance_id` FK 参照が孤立しデータ整合性が破壊される。

## 症状

- `/settings/insurance` で「アイペット損保」の削除を実行
- DELETE /api/v1/masters/insurances/2 → **204 No Content**
- トースト: 「保険を削除しました」（削除成功）
- 件数: 4件 → 3件（削除完了）
- `pets.insurance_id = 2` が残存している場合、孤立参照が発生

## 期待する動作

- `pets` テーブルに当該 insurance_id が存在する場合 → **409 Conflict**
- エラートースト: 「この保険は登録済みのペットに使用されているため削除できません」
- リストは変化しない

## 根本原因

`backend/internal/service/insurance_service.go` の Delete メソッドに依存チェックがない。
`pets.insurance_id` FK 参照が存在するかのチェックを追加していない。

```go
// insurance_service.go Delete() 内
count, err := s.repo.CountPetsByInsuranceID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check insurance dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/insurance_service.go`
- `backend/internal/repository/insurance_repository.go`（CountPetsByInsuranceID 追加）

## 優先度

High（データ整合性破壊・ペット保険情報の消失）

## 関連

- BUG-030（サービス種別・スタッフは修正済み。保険マスタは修正漏れ）
- BUG-104（権限グループと同種の依存チェックなしバグ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/insurances/2 [204] 確認
