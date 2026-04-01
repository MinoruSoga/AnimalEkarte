# BUG-111: トリミングコース削除（FK制約なし）→ トリミング記録が孤立

## 概要

`DELETE /api/v1/masters/trimming-courses/:id` が当該コースを参照している
`trimming_records` の存在チェックをせず 204 No Content で削除成功してしまう。
トリミング記録に使われたコースが削除されるとデータ整合性が破壊される。

## 症状

- `/settings/trimming` で「シャンプー&ブロー（小型）」の削除を実行
- DELETE /api/v1/masters/trimming-courses/1 → **204 No Content**
- トースト: 「トリミングコースを削除しました」（削除成功）
- FK 依存チェックなし

## 期待する動作

- `trimming_records` に当該 trimming_course_id が存在する場合 → **409 Conflict**
- エラートースト: 「このデータは他のレコードに使用されているため削除できません」
- リストは変化しない

## 根本原因

`backend/internal/service/trimming_course_service.go` の Delete メソッドに依存チェックがない。

```go
// trimming_course_service.go Delete() 内
count, err := s.repo.CountTrimmingRecordsByCourseID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check trimming course dependencies")
}
if count > 0 {
    return apperrors.ErrConflict
}
```

## 影響ファイル

- `backend/internal/service/trimming_course_service.go`
- `backend/internal/repository/trimming_course_repository.go`（CountTrimmingRecordsByCourseID 追加）

## 優先度

High（データ整合性破壊）

## 関連

- BUG-030（サービス種別・スタッフは修正済み。トリミングコースは未対応）
- BUG-109（同種の依存チェックなしバグ）
- テスト確認日: 2026-04-01（ローカル環境）
- DELETE http://localhost:8080/api/v1/masters/trimming-courses/1 [204] 確認
