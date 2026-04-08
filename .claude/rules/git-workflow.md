---
description: Git ワークフロー規約（ブランチ戦略、コミットメッセージ）
alwaysApply: true
---

# Git Workflow Rules

GitHub/Git ワークフロー標準ルール。

## 核心ルール

### 1. ブランチ戦略

```
production  (本番環境・直接push禁止)
  ↑ --no-ff merge（リリース確定時のみ）
staging     (ステージング環境 / CI/CD → stg.noah-karte.com・直接push禁止)
  ↑ PR merge
main        (開発ブランチ・日常作業)
```

**ルール:**
- `main`: **開発ブランチ**。日常の作業・コミットはここで直接行う
- `staging`: ステージング環境デプロイ用。`main` から PR を作成してマージする（直接 push 禁止）
- `production`: 本番環境（直接 push 禁止。`staging` からの --no-ff マージのみ）

**日常の開発フロー:**
```bash
# main で作業
git checkout main
git pull origin main
# 作業・コミット
git add ...
git commit -m "fix(xxx): ..."
git push origin main

# ステージングへ反映する場合
# main → staging への PR を作成
gh pr create --base staging --title "..."
```

### 2. コミットメッセージフォーマット

```
<type>(<scope>): <subject>

<body>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

**type:**
- `feat`: 新機能
- `fix`: バグ修正
- `refactor`: リファクタリング（動作変更なし）
- `test`: テスト追加・修正
- `docs`: ドキュメント
- `ci`: CI/CD設定
- `chore`: ビルド・依存更新等
- `perf`: パフォーマンス最適化

**例:**
```
feat(owners): add owner export functionality

- Implement CSV export for owner list
- Add ExportService.ExportOwners() function
- Integrate export button in UI

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

### 3. PR (Pull Request) ルール

PR のベースブランチは **`staging`**（`main` → `staging`）。

```markdown
## Summary
- 機能説明（1-3行）

## Test Plan
- [ ] ローカルで動作確認
- [ ] ユニットテスト パス
- [ ] 統合テスト パス

🤖 Generated with Claude Code
```

**条件:**
- `staging` へのマージには 1 Approval
- すべてのチェックが green
- コンフリクト解決済み

### 4. コミット前チェック

```bash
# lint 確認
docker compose exec backend golangci-lint run ./...
docker compose exec frontend npm run lint

# テスト実行
docker compose exec backend go test ./... -v
docker compose exec frontend npm run test:run
```

### 5. マージ規約

```bash
# main → staging（PR マージ）
# GitHub 上で PR をマージ（Squash or Merge commit）

# staging → production（リリース時）
git checkout production
git pull origin production
git merge --no-ff staging -m "Release vX.Y.Z"
git tag vX.Y.Z
git push origin production --tags
```

## チェックリスト

- [ ] 日常作業: `main` ブランチでコミット
- [ ] PR: `main` → `staging` に向ける
- [ ] コミット: `type(scope): subject` フォーマット
- [ ] メッセージ本文: 変更理由・背景を記載
- [ ] Co-Authored-By: Claude署名（AI生成時）
- [ ] テスト: 全テストパス
- [ ] Lint: golangci-lint, npm run lint パス
- [ ] PR 説明: Summary + Test Plan

## 禁止事項

- `production` へ直接 push ❌
- `staging` へ直接 push ❌（必ず `main` から PR を経由する）
- `git push --force` ❌（共有ブランチ）
- コミット後に文脈なしで force push ❌
- 大型ファイル（バイナリ・ログ）をコミット ❌
- secrets（API key・password）をコミット ❌
