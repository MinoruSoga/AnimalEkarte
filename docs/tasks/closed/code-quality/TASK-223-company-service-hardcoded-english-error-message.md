# TASK-223: company_service.go — エラーメッセージが英語ハードコード（ErrMsgAtLeastOneField を使うべき）

## 優先度
Medium

## 対象ファイル
`backend/internal/service/company_service.go`

## 問題概要
`Update` メソッドで英語のハードコード文字列を使用している。
他の全 service は `validators.go` に定義された日本語定数 `ErrMsgAtLeastOneField` を使用している。

## 現状コード（行109付近）

```go
// 現状（NG）
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput("at least one field must be provided")
}
```

## あるべき姿

```go
// あるべき姿（他全ドメインと統一）
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
}
```

`ErrMsgAtLeastOneField` は `backend/internal/service/validators.go` に定義済み（値: `"少なくとも1つのフィールドを指定してください"`）。

## 完了条件
- [ ] `"at least one field must be provided"` を `ErrMsgAtLeastOneField` に置換
- [ ] `go test ./backend/internal/...` がパス
