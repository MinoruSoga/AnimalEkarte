# BUG-112: 役職マスタ削除（FK制約なし）→ スタッフ参照が孤立

## 概要

`DELETE /api/v1/masters/job-titles/:id` が当該役職を参照しているスタッフの存在チェックをせず
204 No Content で削除成功してしまう。`staffs.job_title_id` FK 参照が孤立しデータ整合性が破壊される。

## 症状

- `/settings/job-title` で「獣医師」の削除を実行
- DELETE /api/v1/masters/job-titles/1 → **204 No Content**
- トースト: 「役職を削除しました」（削除成功）
- 件数: 5件 → 4件（削除完了）
- `staffs.job_title_id = 1` が残存している場合、孤立参照が発生

## 期待する動作

- `staffs` テーブルに当該 job_title_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」
- リストは変化しない

## 根本原因

`backend/internal/service/job_title_service.go` の Delete メソッドに依存チェックがない。

```go
// job_title_service.go Delete() 内
count, err := s.repo.CountStaffsByJobTitleID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check job title dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/job_title_service.go`
- `backend/internal/repository/job_title_repository.go`（CountStaffsByJobTitleID 追加）

## 優先度

High（データ整合性破壊・スタッフ情報の孤立）

## 関連

- BUG-030（サービス種別・スタッフは修正済み。役職は修正漏れ）
- BUG-110, BUG-111（同種の依存チェックなしバグ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/job-titles/1 [204] 確認
