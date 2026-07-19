---
description: Go コードレビュー（Idiom・パフォーマンス）
argument-hint: "[path] (blank for staged Go files)"
---

# /go-review [path]

Go コードを Idiom・パフォーマンス・セキュリティの観点でレビューします。

## 使用法

```bash
# ファイルレビュー
/go-review internal/handler/owner_handler.go

# ディレクトリレビュー
/go-review internal/service
```

## レビュー項目・出力形式

`go-reviewer` エージェント定義と `.claude/refs/go-gin-backend-review.md` を正本とする。固定layerや旧P番号のチェックリストを複製しない。

## 使用エージェント

`go-reviewer` (Sonnet) のレビュー実施
