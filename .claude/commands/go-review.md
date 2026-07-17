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

`go-reviewer` エージェント定義（P1-P18準拠・CRITICAL/HIGH/MEDIUM構造化）を正本とする。ここにチェックリストを複製しない。

## 使用エージェント

`go-reviewer` (Sonnet) のレビュー実施
