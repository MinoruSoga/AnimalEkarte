---
description: フルセキュリティスキャン実行（OWASP、gosec、npm audit）
---

# /security-audit

フルセキュリティスキャンを実行し、脆弱性と対応状況を報告します。

## 実行内容

### Backend (Go)
```bash
docker compose exec backend gosec -json ./... | jq
```

### Frontend (npm)
```bash
docker compose exec frontend npm audit --json
```

### 手動確認事項
- [ ] SQL インジェクション対策
- [ ] XSS 対策
- [ ] CSRF トークン
- [ ] 認証・認可設計
- [ ] パスワードハッシング
- [ ] セッション管理

## 出力形式

```
🔴 Critical: X件
🟠 High: X件
🟡 Medium: X件
🟢 Low: X件

## OWASP Top 10 対応状況
1. Injection - ✅対応
2. Broken Authentication - ❌要対応
...

## 推奨修正
- 優先度高: ...
```

## 使用エージェント

`security-analyst` (Opus) を自動起動
