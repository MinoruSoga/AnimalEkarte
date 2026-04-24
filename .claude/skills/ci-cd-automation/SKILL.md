---
name: ci-cd-automation
description: CI/CD パイプライン自動化（GitHub Actions、テスト・ビルド・デプロイ）
---

# CI/CD Pipeline Automation

GitHub Actions によるテスト・ビルド・デプロイ自動化。

## パイプライン構成

```
┌─────────────────────────────────────────────────┐
│ Trigger: git push (main, develop)               │
└────────────┬────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────┐
│ Lint & Format Check                             │
├─────────────────────────────────────────────────┤
│ - Go: golangci-lint                             │
│ - TypeScript: ESLint + Prettier                 │
│ - YAML: yamllint                                │
│ ⏱️ ~2分                                          │
└────────────┬────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────┐
│ Unit Tests                                      │
├─────────────────────────────────────────────────┤
│ - Backend: go test ./... -cover                 │
│ - Frontend: pppnpm test:run                    │
│ ⏱️ ~5分                                          │
└────────────┬────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────┐
│ Security Scan                                   │
├─────────────────────────────────────────────────┤
│ - Go: gosec ./...                               │
│ - Deps: ppnpm audit                               │
│ ⏱️ ~2分                                          │
└────────────┬────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────┐
│ Build Docker Images                             │
├─────────────────────────────────────────────────┤
│ - Backend: docker build                         │
│ - Frontend: docker build                        │
│ ⏱️ ~3分                                          │
└────────────┬────────────────────────────────────┘
             │
             ├─ IF main branch
             │   ▼
             │ ┌──────────────────────────────────┐
             │ │ Integration Tests (Docker)       │
             │ │ ⏱️ ~5分                           │
             │ └──────────────────────────────────┘
             │   ▼
             │ ┌──────────────────────────────────┐
             │ │ Push to Registry                 │
             │ │ ⏱️ ~2分                           │
             │ └──────────────────────────────────┘
             │   ▼
             │ ┌──────────────────────────────────┐
             │ │ Deploy to Test Environment       │
             │ │ ⏱️ ~5分                           │
             │ └──────────────────────────────────┘
             │
             └─ IF feature branch
                 (Skip deployment)
```

## GitHub Actions ワークフロー例

### .github/workflows/lint.yml

```yaml
name: Lint & Format Check
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Go Lint
      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: ./backend

      # TypeScript Lint
      - uses: node-actions/setup-node@v3
        with:
          node-version: '20'
      - run: cd frontend && ppnpm install --frozen-lockfile && ppnpm lint

      # YAML Lint
      - uses: ibiqlik/action-yamllint@v3
        with:
          file_or_dir: .
          config_file: .yamllint
```

### .github/workflows/test.yml

```yaml
name: Tests
on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:18
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: ekarte_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      - uses: go-actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run tests
        run: cd backend && go test ./... -v -cover
        env:
          DATABASE_URL: postgres://postgres:test@localhost/ekarte_test

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: node-actions/setup-node@v3
        with:
          node-version: '20'

      - run: cd frontend && ppnpm install --frozen-lockfile && pppnpm test:run

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./frontend/coverage/coverage-final.json
```

### .github/workflows/security.yml

```yaml
name: Security Scan
on: [push, pull_request]

jobs:
  gosec:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: securego/gosec@master
        with:
          args: '-no-fail -fmt json ./...'
          working-directory: backend

  npm-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: node-actions/setup-node@v3
      - run: cd frontend && ppnpm install --frozen-lockfile && ppnpm audit --audit-level=moderate
```

### .github/workflows/docker-build.yml

```yaml
name: Docker Build & Push
on:
  push:
    branches: [main, develop]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-buildx-action@v2

      # Backend
      - uses: docker/build-push-action@v4
        with:
          context: ./backend
          file: ./backend/Dockerfile
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/your-org/ekarte-backend:${{ github.sha }}
          cache-from: type=registry
          cache-to: type=inline

      # Frontend
      - uses: docker/build-push-action@v4
        with:
          context: ./frontend
          file: ./frontend/Dockerfile
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/your-org/ekarte-frontend:${{ github.sha }}
```

## パイプライン監視

### ステータス確認

```bash
# GitHub Actions 確認
gh run list --repo your-org/ekarte

# 特定ワークフロー確認
gh run view --repo your-org/ekarte 12345

# ローカル CI 実行
make test
make lint
make docker-build
```

### 失敗時の対応

| 失敗箇所 | 対応 |
|---------|------|
| Lint 失敗 | `ppnpm lint:fix`, `go fmt` |
| Test 失敗 | ローカルで `go test -v` 実行 |
| Build 失敗 | Docker ログ確認、キャッシュクリア |
| Deploy 失敗 | K8s/ECS ログ確認 |

## パフォーマンス目標

```
Total Pipeline Time:    < 15分
Lint:                   < 2分
Test:                   < 5分
Security:               < 2分
Build:                   < 3分
Deploy:                  < 5分

Success Rate:           > 98%
```

## ベストプラクティス

1. **キャッシング**
   ```yaml
   - uses: actions/cache@v3
     with:
       path: ~/.cache/go-build
       key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
   ```

2. **Artifact 保存**
   ```yaml
   - uses: actions/upload-artifact@v3
     with:
       name: coverage-reports
       path: coverage/
   ```

3. **通知**
   - Slack 連携
   - GitHub チェック実行
   - メール通知（失敗時）

4. **セキュリティ**
   - Secrets 管理（token、credentials）
   - OpenID Connect で認証
   - Runner セキュリティ

## チェックリスト

- [ ] Lint ワークフロー実装
- [ ] Unit Test ワークフロー実装
- [ ] Security Scan ワークフロー実装
- [ ] Docker Build ワークフロー実装
- [ ] Integration Test ワークフロー実装
- [ ] Deploy ワークフロー実装
- [ ] 通知設定（Slack、メール）
- [ ] キャッシング最適化
- [ ] Artifact 保存設定

## 関連スキル

- `docker-optimization` - イメージビルド最適化
- `deployment` - デプロイメント自動化
- `security-audit` - セキュリティチェック
