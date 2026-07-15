---
description: CI パイプライン確認・トリガー
argument-hint: "[--local | --watch]"
---

# /ci-trigger [--watch]

CI パイプラインの状態確認、またはローカルで CI を実行します。

## 使用法

```bash
# GitHub Actions 状態確認
/ci-trigger

# ローカル CI 実行
/ci-trigger --local

# Watch モード（更新自動確認）
/ci-trigger --watch
```

## 実行内容

### ローカル CI

以下は `.claude/CLAUDE.md` の Auto-Execution Prohibited Commands に該当する全体実行コマンドであり、エージェントが自律実行してはならない。ユーザーに実行を依頼するか、変更範囲に絞ったスコープ限定コマンドを使う。

```bash
# Backend（ユーザー手動実行）
docker compose exec backend go test ./... -v
docker compose exec backend golangci-lint run ./...

# Frontend（ユーザー手動実行）
docker compose exec frontend pnpm test:run
docker compose exec frontend pnpm lint

# Security
# gosec は本プロジェクト未導入（CI の security-scan.yml は agentshield のみ）— 導入してから実行する
docker compose exec frontend pnpm audit
```

### GitHub Actions 確認
- 最新ワークフロー実行状態
- ビルド・テスト・デプロイ結果
- 失敗原因の提示

## 出力形式

```
✅ Build: PASSED (2m 34s)
✅ Test: PASSED (1m 12s)
❌ Lint: FAILED
   - error: unused variable at line XX

❌ Deploy: PENDING
```

## 使用エージェント

`formatter` (Haiku) 自動実行
