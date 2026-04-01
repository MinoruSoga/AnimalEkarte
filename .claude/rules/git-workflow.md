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
staging     (開発統合ブランチ / CI/CD → stg.noah-karte.com)
  ↑ PR merge
feature/xxx (機能開発)
bug/xxx     (バグ修正)
```

**ルール:**
- `staging`: **開発統合ブランチ**。すべての feature/bug ブランチはここから切り出し、ここへマージする
- `production`: 本番環境（直接pushは禁止。`staging` からの --no-ff マージのみ）
- `feature/*`: 機能ブランチ（`staging` から切り出し）
- `bug/*`: バグ修正ブランチ（`staging` から切り出し）
- `main`: **使用しない**（過去の経緯で存在するが新規作業対象外）

**新ブランチの作成:**
```bash
# staging から切り出す
git checkout staging
git pull origin staging
git checkout -b feature/xxx
# または
git checkout -b bug/xxx
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

PR のベースブランチは必ず **`staging`**。

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

### 5. マージコミット規約

```bash
# feature/bug → staging
git checkout staging
git pull origin staging
git merge --no-ff feature/xxx -m "Merge feature/xxx"
git push origin staging

# staging → production (リリース時)
git checkout production
git pull origin production
git merge --no-ff staging -m "Release vX.Y.Z"
git tag vX.Y.Z
git push origin production --tags
```

## チェックリスト

- [ ] ブランチ: `staging` から `feature/*` または `bug/*` を切り出し
- [ ] PR ベース: `staging` に向ける（`production` や `main` に直接PR禁止）
- [ ] コミット: `type(scope): subject` フォーマット
- [ ] メッセージ本文: 変更理由・背景を記載
- [ ] Co-Authored-By: Claude署名（AI生成時）
- [ ] テスト: 全テストパス
- [ ] Lint: golangci-lint, npm run lint パス
- [ ] PR 説明: Summary + Test Plan
- [ ] マージコミット: `--no-ff` で履歴保留

## 禁止事項

- `production` へ直接 push ❌
- `staging` へ直接 push（緊急 hotfix を除く）❌
- `git push --force` ❌（共有ブランチ）
- コミット後に文脈なしで force push ❌
- 大型ファイル（バイナリ・ログ）をコミット ❌
- secrets（API key・password）をコミット ❌
- `main` ブランチへの新規作業 ❌（廃止済み）
