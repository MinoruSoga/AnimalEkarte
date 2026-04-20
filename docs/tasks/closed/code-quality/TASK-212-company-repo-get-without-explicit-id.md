# TASK-212: company_repository.go — Get で id を明示せず First() を使用

## 優先度
Low

## 対象ファイル
`backend/internal/repository/company_repository.go`

## 問題概要
`Get`（行30付近）でレコード取得に `First(&company)` を使用しており、
id 条件を明示していない。`Update` は `WHERE "id = 1"` と id=1 を明示しているのに対して、
`Get` だけ条件なし（GORM が主キー昇順で最初の1件を返す）という不一致がある。

company はシングルトンテーブル（id=1 固定）であることを前提とした実装だが、
その前提がコードから明示されていない。

```go
// 現状（NG）— 行30付近
r.db.WithContext(ctx).Limit(1).First(&company).Error
// GORM デフォルト: ORDER BY id ASC LIMIT 1

// あるべき姿（意図を明示）
r.db.WithContext(ctx).First(&company, 1).Error  // id=1 を明示
// または
r.db.WithContext(ctx).Where("id = ?", 1).First(&company).Error
```

## 完了条件
- [ ] `Get` で id=1 を明示した取得に修正
- [ ] `go test ./backend/internal/...` がパス
