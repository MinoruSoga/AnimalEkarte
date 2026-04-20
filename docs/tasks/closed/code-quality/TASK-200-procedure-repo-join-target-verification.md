# TASK-200: procedure_repository.go — CountUsageByProcedureID の JOIN 先整合性確認

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/procedure_repository.go`
- `backend/internal/repository/medicine_repository.go`

## 問題概要
`care_plan_items` テーブルの JOIN 先が2ファイルで異なっており、
どちらかが誤ったスキーマ参照をしている可能性がある。

```go
// procedure_repository.go:96
Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id ...")

// medicine_repository.go:73
Joins("JOIN care_plans ON care_plans.id = care_plan_items.care_plan_id ...")
```

`care_plan_items` の外部キーが `hospitalization_id` か `care_plan_id` かで
どちらかは誤ったカラム参照になる。コンパイルエラーにはならないが、
COUNT 結果が常に0になる（JOIN が一致しない）バグを生む可能性がある。

## 調査・修正方針

1. `docs/ERD.md` または `backend/migrations/001_init.sql` で `care_plan_items` テーブルの
   FK カラム定義を確認する
2. 正しい JOIN 先に統一する
3. 誤っている方の `CountUsage` 系メソッドを修正する

## 完了条件
- [ ] `care_plan_items` のスキーマを ERD / migration で確認
- [ ] 誤っている方の JOIN 先を修正
- [ ] COUNT が正しい値を返すことをテストで確認
- [ ] `go test ./backend/internal/...` がパス
