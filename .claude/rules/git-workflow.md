---
description: Git ワークフロー規約（ブランチ戦略、コミットメッセージ）
alwaysApply: true
---

# Git Workflow Rules

GitHub/Git ワークフロー標準ルール。

## 核心ルール

### 1. ブランチ戦略（Git Flow）

```
main (保護)
  ↑
release/X.Y.Z (リリース直前)
  ↑
develop (統合ブランチ)
  ↑
feature/xxx (機能開発)
bugfix/xxx  (バグ修正)
```

**ルール:**
- `main`: 本番コード（タグ付き）、直接pushは禁止
- `develop`: 統合ブランチ（CI/CD実行）
- `feature/*`: 機能ブランチ（develop から切り出し）
- `bugfix/*`: バグ修正ブランチ（develop から切り出し）
- `release/*`: リリース準備（hotfix含む）

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

```markdown
## Summary
- 機能説明（1-3行）

## Test Plan
- [ ] ローカルで動作確認
- [ ] ユニットテスト パス
- [ ] 統合テスト パス
- [ ] API Swagger確認

🤖 Generated with Claude Code
```

**条件:**
- main へのマージには 2 Approvals（自動化可）
- develop へのマージには 1 Approval
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

# コミット前フック（.git/hooks/pre-commit）で自動実行可
```

### 5. マージコミット規約

```bash
# feature → develop
git checkout develop
git pull origin develop
git merge --no-ff feature/xxx -m "Merge feature/xxx"
git push origin develop

# develop → main (release時)
git checkout main
git pull origin main
git merge --no-ff develop -m "Release vX.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

## チェックリスト

- [ ] ブランチ: feature/*, bugfix/*, release/* から切り出し
- [ ] コミット: type(scope): subject フォーマット
- [ ] メッセージ本文: 変更理由・背景を記載
- [ ] Co-Authored-By: Claude署名（AI生成時）
- [ ] テスト: 全テストパス
- [ ] Lint: golangci-lint, npm run lint パス
- [ ] PR 説明: Summary + Test Plan
- [ ] マージコミット: --no-ff で履歴保留

## 禁止事項

- main へ直接 push ❌
- `git push --force` ❌（共有ブランチ）
- コミット後に文脈なしで force push ❌
- 大型ファイル（バイナリ・ログ）をコミット ❌
- secrets（API key・password）をコミット ❌
