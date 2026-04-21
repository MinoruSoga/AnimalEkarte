# TASK-231: vaccine_repository.go / procedure_repository.go — Limit(500) ハードコード

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/vaccine_repository.go`（行37）
- `backend/internal/repository/procedure_repository.go`（行33）

## 問題概要
両ファイルの `List` メソッドに `Limit(500)` がハードコードされている。
呼び出し元から上限を制御できず、マスタ件数が増加した際に気づきにくい。
また、他のリポジトリは上限を呼び出し元（service/handler）から渡す設計になっている。

## 現状コード

```go
// vaccine_repository.go:37
if err := q.Order("sort_order ASC, name ASC").Limit(500).Find(&vaccines).Error; err != nil {

// procedure_repository.go:33
if err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Limit(500).Find(&procedures).Error; err != nil {
```

## あるべき姿

Repository インターフェースの `List` シグネチャがページネーション引数を持たない場合、
定数として明示的に宣言する。

```go
// vaccine_repository.go
const maxVaccineListLimit = 500

// Limit(maxVaccineListLimit) で意図を明示
q.Order("sort_order ASC, name ASC").Limit(maxVaccineListLimit).Find(&vaccines)
```

または、将来のページネーション対応に備えて `limit int` 引数を受け取る形式へ変更する。

## 完了条件
- [ ] `vaccine_repository.go` の `Limit(500)` を定数化
- [ ] `procedure_repository.go` の `Limit(500)` を定数化
- [ ] `go test ./backend/internal/...` がパス
