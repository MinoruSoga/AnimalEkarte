# デプロイ前チェックリスト

本番・ステージング環境へのデプロイ前に確認すべき項目。

---

## デプロイ前の確認

### コード品質

- [ ] **Lint チェック合格**
  ```bash
  docker compose exec frontend npm run lint
  docker compose exec backend golangci-lint run ./...
  ```

- [ ] **テスト実行完了**
  ```bash
  docker compose exec frontend npm run test:run
  docker compose exec backend go test ./... -v
  ```

- [ ] **ビルド成功**
  ```bash
  docker compose exec frontend npm run build
  docker compose exec backend go build ./cmd/api
  ```

- [ ] **console.log なし**
  ```bash
  grep -r "console\.log\|console\.error" frontend/src --include="*.ts" --include="*.tsx"
  ```

- [ ] **TypeScript エラーなし**
  ```bash
  docker compose exec frontend npm run type-check
  ```

### Git コミット品質

- [ ] コミットメッセージが明確
- [ ] 関連 Issue/PR 番号を記載
- [ ] 不要なコミットが含まれていない
  ```bash
  git log --oneline main..HEAD
  ```

- [ ] テスト用コミット・ファイルが削除されている

### 環境設定

- [ ] **.env 設定が正しい**
  - Backend: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
  - Frontend: VITE_API_URL（正しい Backend URL）

- [ ] **AWS プロファイル確認**
  ```bash
  export AWS_PROFILE=AnimalEkarte
  aws sts get-caller-identity
  ```

### セキュリティチェック

- [ ] **シークレット・パスワードが含まれていない**
  ```bash
  git diff main..HEAD | grep -E "password|secret|token|key"
  ```

- [ ] **git-secrets インストール**
  ```bash
  brew install git-secrets
  git secrets install
  ```

- [ ] **依存パッケージの脆弱性確認**
  ```bash
  docker compose exec frontend npm audit
  docker compose exec backend go mod tidy
  ```

### ドキュメント

- [ ] README が最新か確認
- [ ] API ドキュメント（Swagger）が最新か確認
  ```bash
  docker compose exec backend swag init -g cmd/api/main.go
  ```

- [ ] 破壊的変更（Breaking Changes）を CHANGELOG に記載

---

## デプロイ実行フロー

### ステップ 1: 最終確認

```bash
# 最新の main ブランチを確認
git checkout main
git pull origin main

# デプロイ対象を確認
git log --oneline -10

# 変更内容を確認
git diff origin/main HEAD
```

### ステップ 2: main にマージ

**既に main にある場合はスキップ**

```bash
# Feature ブランチから main へ
git checkout main
git pull origin main
git merge --no-ff feature/xxx -m "feat: description"
git push origin main
```

### ステップ 3: 自動デプロイ監視

**Backend:**
```bash
# GitHub Actions を監視
gh run list --workflow=backend-deploy.yml --limit 1

# 完了まで待機
gh run view <RUN_ID> --log
```

**Frontend:**
```bash
# Vercel ダッシュボードで監視
vercel projects list
```

### ステップ 4: デプロイ完了確認

**Backend:**
```bash
export AWS_PROFILE=AnimalEkarte

# API ヘルスチェック
curl http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health | jq .

# ECS タスク確認
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].{status, runningCount, desiredCount}'
```

**Frontend:**
```bash
# ページ読み込み確認
curl -I https://animalekarte-frontend-*.vercel.app
```

### ステップ 5: 動作検証

- [ ] **Backend API**
  - ヘルスチェック OK
  - データベース接続 OK
  - 既存エンドポイント動作確認

- [ ] **Frontend**
  - ページロード OK
  - API 連携 OK
  - 主要フロー動作確認

---

## 問題発生時の対応

### Backend デプロイ失敗

1. **GitHub Actions ログを確認**
   ```bash
   gh run view <RUN_ID> --log
   ```

2. **問題に応じた対応**
   - AWS OIDC エラー → [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) のトラブルシューティング参照
   - Docker ビルドエラー → `docker build -f backend/Dockerfile.production .` でローカルテスト
   - ECS デプロイエラー → CloudWatch Logs 確認

3. **ロールバック**
   ```bash
   export AWS_PROFILE=AnimalEkarte
   aws ecs update-service \
     --cluster animalekarte-test-cluster \
     --service animalekarte-test-service \
     --task-definition animalekarte-test-api:4 \
     --region us-east-1
   ```

### Frontend デプロイ失敗

1. **Vercel ダッシュボード → Deployments で確認**
2. **ビルドログを確認**
3. **ロールバック**
   - Vercel ダッシュボード → 前バージョン → "Promote to Production"

---

## デプロイ後の検証（重要）

### ユーザー動作確認

**以下のシナリオをテスト:**

1. **新規登録・ログイン**
   - 新規ユーザー作成
   - ログイン機能確認

2. **データ操作**
   - 新規ペット作成
   - 医療記録追加
   - データ更新・削除

3. **API 連携**
   - Frontend → Backend API 通信確認
   - データベース保存確認

4. **パフォーマンス**
   - ページロード時間（3秒以下目安）
   - API レスポンス時間（500ms 以下目安）

### ログ監視

```bash
# Backend CloudWatch ログ確認
export AWS_PROFILE=AnimalEkarte
aws logs tail /ecs/animalekarte-test --follow --since 30m --region us-east-1

# エラーがないか確認
aws logs filter-log-events \
  --log-group-name /ecs/animalekarte-test \
  --filter-pattern "ERROR" \
  --since 1800000 \
  --region us-east-1
```

---

## ロールバック判断基準

以下の場合は**即座にロールバック**を実行：

- ❌ 本来動作すべき機能が動作していない
- ❌ API エラーレート > 1%
- ❌ ページ読み込み時間 > 5秒
- ❌ データベースエラー
- ❌ セキュリティ脆弱性が判明

---

## 参考資料

- [CI/CD パイプラインガイド](./CI-CD-PIPELINE.md)
- [デプロイ状態記録](./deployment-status.md)
- [テスト環境構成](./test-environment.md)

---

**最終更新:** 2026-03-04
**確認者:** Claude Code (Haiku 4.5)
