# TASK-186: closing_special_period_repository.go — apperrors.Wrap → FromGORM への修正

## 優先度
High

## 対象ファイル
`backend/internal/repository/closing_special_period_repository.go`

## 問題概要
Repository 層でエラー変換に `apperrors.Wrap` を使用している箇所が2箇所ある。
プロジェクト規約「Repository では GORM エラーは必ず `apperrors.FromGORM` で変換する」に違反している。

`apperrors.Wrap` はサービス層向けのラッパーであり、Repository 層で使うと
GORM の `ErrRecordNotFound` が正しく `apperrors.ErrNotFound` にマッピングされない。

## 該当箇所

### 1. `FindByDate`（行67）
```go
// 現状（NG）
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find closing special period by date")
}

// あるべき姿
if err != nil {
    return nil, apperrors.FromGORM(err, "closing_special_period", date.Format("2006-01-02"))
}
```

### 2. `HasOverlap`（行125）
```go
// 現状（NG）
return false, apperrors.Wrap(err, "failed to check overlap")

// あるべき姿
return false, apperrors.FromGORM(err, "closing_special_period", "")
```

## 修正方針
- `FindByDate` と `HasOverlap` の2箇所を `apperrors.FromGORM` に置き換える
- `HasOverlap` はレコード存在確認クエリ（COUNT等）であれば id は空文字でよい

## 完了条件
- [ ] `closing_special_period_repository.go` の `apperrors.Wrap` を `apperrors.FromGORM` に置換
- [ ] `go test ./backend/internal/...` がパス
