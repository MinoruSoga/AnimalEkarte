---
description: Go コードレビュー（Idiom・パフォーマンス）
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

## レビュー項目

### Go Idioms
- [ ] Error handling（sentinel errors）
- [ ] Interface 設計
- [ ] Defer の適切使用
- [ ] Context の伝播

### Performance
- [ ] アロケーション最適化
- [ ] String 操作の効率化
- [ ] スライス容量予約
- [ ] Goroutine リーク

### Security
- [ ] gosec 警告
- [ ] 入力バリデーション
- [ ] 暗号化・ハッシング
- [ ] ロギング（機密情報禁止）

## 出力形式

```markdown
## Go コードレビュー: XXX.go

### 🟢 Good
- Error wrapping properly
- ...

### 🟡 Improvement
- [ ] Use sync.Pool for ...
- [ ] ...

### 🔴 Issues
- [ ] gosec: SQL injection risk at line XX
- [ ] ...
```

## 使用エージェント

`go-expert` (Sonnet) のレビュー実施
