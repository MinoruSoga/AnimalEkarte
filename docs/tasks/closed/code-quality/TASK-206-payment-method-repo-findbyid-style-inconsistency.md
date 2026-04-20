# TASK-206: payment_method_master_repository.go — FindByID で First(&m, id) パターンが他と不統一

## 優先度
Medium

## 対象ファイル
`backend/internal/repository/payment_method_master_repository.go`

## 問題概要
`FindByID`（行47付近）が `Scopes(clinicScope(clinicID)).First(&m, id)` という書き方を使っている。
他の全 repository は `Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&m)` で統一されており、
この1ファイルのみスタイルが異なる。

```go
// 現状（NG）— 行47
err := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    First(&m, id).Error  // id を第2引数で渡す
```

```go
// あるべき姿（他ドメインと統一）
err := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).
    First(&m).Error
```

`First(&m, id)` は GORM が内部的に `WHERE primary_key = id` を付加するため動作上は正しいが、
`clinicScope` との組み合わせで意図が読み取りにくく、将来の複合条件拡張時に混乱を招く。

## 完了条件
- [ ] `FindByID` を `Where("id = ?", id).First(&m)` パターンに統一
- [ ] `go test ./backend/internal/...` がパス
