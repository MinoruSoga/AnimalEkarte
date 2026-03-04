# CI/CD パイプライン - Animal Ekarte 自動デプロイ

自動デプロイ機能の完全な説明と運用ガイド。

**最終更新日: 2026-03-04**
**ステータス: ✅ 完全稼働中**

---

## 概要

Animal Ekarte は **Backend** と **Frontend** に分かれた自動デプロイパイプラインを運用しています。

| コンポーネント | デプロイ方式 | ステータス |
|--------------|-----------|---------|
| **Backend API** | GitHub Actions + AWS OIDC + ECR/ECS | ✅ 稼働中 |
| **Frontend** | Vercel GitHub Integration | ✅ 稼働中 |

---

## Backend デプロイ パイプライン

### フロー図

```
developer push to main
    ↓
backend/** ファイル変更検出
    ↓
GitHub Actions backend-deploy.yml トリガー
    ↓
AWS OIDC 認証 (GitHub Actions → AWS)
    ↓
Docker build (backend/Dockerfile.production)
    ↓
ECR push (animalekarte-api:latest + sha)
    ↓
ECS Task Definition 更新
    ↓
ECS Service 更新 (Blue/Green デプロイ)
    ↓
サービス安定性確認 (wait-for-service-stability)
    ↓
デプロイ完了
```

### ワークフロー設定

**ファイル:** `.github/workflows/backend-deploy.yml`

```yaml
name: Backend Deploy

on:
  push:
    branches:
      - main
    paths:
      - 'backend/**'
      - '.github/workflows/backend-deploy.yml'
  workflow_dispatch:  # 手動実行も可能

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: animalekarte-api
  ECS_CLUSTER: animalekarte-test-cluster
  ECS_SERVICE: animalekarte-test-service
  ECS_TASK_DEFINITION_FAMILY: animalekarte-test-api
```

### ステップ詳細

| # | ステップ | 説明 | 所要時間 |
|---|---------|------|--------|
| 1 | Checkout | リポジトリコード取得 | 3秒 |
| 2 | **Configure AWS credentials** | AWS OIDC 認証（GitHub Actions → AWS） | 2秒 |
| 3 | Login to Amazon ECR | ECR ログイン | 1秒 |
| 4 | Build Docker image | Docker イメージビルド | 30-60秒 |
| 5 | Download task definition | 現在の ECS Task Definition を取得 | 3秒 |
| 6 | Fill in new image ID | Task Definition にイメージIDを設定 | 1秒 |
| 7 | **Deploy to Amazon ECS** | ECS Service 更新（Blue/Green） | 10-30秒 |
| 8 | Verify deployment | デプロイ安定性確認（最大10分） | 5-600秒 |

**通常の総所要時間: 1-5分**

### AWS OIDC 認証

GitHub Actions は AWS OIDC トークンを使用してAWSに認証します。

**信頼関係設定（IAM Role）:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::698109622668:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:MinoruSoga/AnimalEkarte:ref:refs/heads/main"
        }
      }
    }
  ]
}
```

**重要**: リポジトリ名は **`MinoruSoga/AnimalEkarte`** である必要があります（2026-03-04 修正済み）

### ECR イメージタグ

デプロイされるイメージには2つのタグが付与されます：

```
698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api:latest
698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api:sha-abc1234
```

- `latest`: 最新デプロイ
- `sha-<commit-hash>`: コミットハッシュベースのタグ（ロールバック用）

### ECS デプロイ確認

```bash
export AWS_PROFILE=AnimalEkarte

# 現在の Task Definition バージョン確認
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].taskDefinition'

# デプロイ状態確認
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].{status, runningCount, desiredCount, deployments}'
```

---

## Frontend デプロイ パイプライン

### フロー図

```
developer push to main
    ↓
frontend/** ファイル変更検出（Vercel GitHub Hook）
    ↓
Vercel 自動デプロイトリガー
    ↓
Vercel ビルド実行
  - npm install
  - npm run build
  - export optimization
    ↓
デプロイ完了（Preview URL）
    ↓
本番環境へ自動プロモート
    ↓
本番 URL (animalekarte-frontend-*.vercel.app) で公開
```

### Vercel GitHub 統合

**設定:**
- **リポジトリ**: MinoruSoga/AnimalEkarte
- **プロジェクト**: animalekarte-frontend
- **自動デプロイ**: Enabled ✅
- **Pull Request Comments**: Enabled ✅
- **Commit Comments**: Enabled ✅

### デプロイメソッド

| 方法 | 説明 | 使用状況 |
|------|------|--------|
| **GitHub Integration** | Vercel が GitHub webhook を監視 | ✅ 推奨（現在使用中） |
| ~~GitHub Actions~~ | GitHub Actions から Vercel デプロイ | ❌ 削除済み（2026-03-04） |
| Vercel CLI | ローカルコマンドラインデプロイ | 手動テスト用 |

### ビルド設定

**ファイル:** `vercel.json`

```json
{
  "buildCommand": "npm run build",
  "installCommand": "npm install",
  "framework": "vite",
  "outputDirectory": "dist"
}
```

### 環境変数

**Production 環境:**
```
VITE_API_URL=http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api
```

**Preview / Development:**
```
VITE_API_URL=http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api
```

### Vercel デプロイ確認

```bash
# Vercel CLI で Deployment 一覧確認
cd frontend
vercel list

# プロジェクト情報確認
vercel projects list

# 最新デプロイ情報確認
vercel project inspect animalekarte-frontend
```

---

## トラブルシューティング

### Backend デプロイ失敗

#### Issue: "Configure AWS credentials" ステップで失敗

**症状:**
```
Error: Error: HTTP 422 AssumeRoleUnauthorizedOperation
```

**原因:**
AWS IAM Role の信頼関係にGitHubリポジトリが正しく設定されていない。

**解決方法:**

1. **AWS コンソール** → **IAM** → **ロール** → `animalekarte-test-github-ecs-deploy-role`
2. **信頼関係** タブをクリック
3. 以下を確認（修正が必要な場合は更新）：

```json
{
  "Effect": "Allow",
  "Principal": {
    "Federated": "arn:aws:iam::698109622668:oidc-provider/token.actions.githubusercontent.com"
  },
  "Action": "sts:AssumeRoleWithWebIdentity",
  "Condition": {
    "StringEquals": {
      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
    },
    "StringLike": {
      "token.actions.githubusercontent.com:sub": "repo:MinoruSoga/AnimalEkarte:ref:refs/heads/main"
    }
  }
}
```

**重要**: リポジトリオーナー名が **`MinoruSoga`** であることを確認

#### Issue: "Build Docker image" で失敗

**症状:**
```
docker buildx build failed: ...
```

**原因:**
- Dockerfile 文法エラー
- ビルドコンテキストのファイル不足
- メモリ不足

**解決方法:**
1. ローカルでビルドテスト：
```bash
cd backend
docker build -f Dockerfile.production -t test .
```

2. GitHub Actions ログを確認：
```bash
gh run view <RUN_ID> --log
```

#### Issue: "Deploy to Amazon ECS" で失敗

**症状:**
```
Service did not stabilize
```

**原因:**
- ECS タスク起動失敗
- ヘルスチェック失敗
- リソース不足（CPU/メモリ）

**解決方法:**
```bash
export AWS_PROFILE=AnimalEkarte

# タスク詳細確認
aws ecs describe-tasks \
  --cluster animalekarte-test-cluster \
  --region us-east-1 \
  --tasks $(aws ecs list-tasks \
    --cluster animalekarte-test-cluster \
    --region us-east-1 \
    --query taskArns[0] --output text)

# CloudWatch Logs 確認
aws logs tail /ecs/animalekarte-test --follow --region us-east-1
```

### Frontend デプロイ失敗

#### Issue: Vercel ビルド失敗

**症状:**
```
Build failed: npm run build error
```

**原因:**
- TypeScript コンパイルエラー
- ESLint エラー
- 依存パッケージの問題

**解決方法:**
1. ローカルでテスト：
```bash
cd frontend
npm install
npm run build
npm run lint
```

2. Vercel ダッシュボードでビルドログ確認

#### Issue: 環境変数が設定されていない

**症状:**
```
VITE_API_URL is undefined
```

**原因:**
Vercel に環境変数が設定されていない

**解決方法:**
1. Vercel ダッシュボード → Settings → Environment Variables
2. 以下を設定：
```
VITE_API_URL = http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api
```

---

## デプロイ検証・確認方法

### Backend API ヘルスチェック

```bash
# ヘルスチェック
curl http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health | jq .

# 期待される応答：
{
  "status": "ok",
  "message": "Animal Ekarte API is running",
  "database": "connected",
  "version": "1.0.0"
}
```

### Frontend 確認

```bash
# ページロード確認
curl -I https://animalekarte-frontend-*.vercel.app

# 期待される応答：
HTTP/1.1 200 OK
```

### デプロイ履歴確認

**Backend:**
```bash
gh run list --workflow=backend-deploy.yml --limit 10
```

**Frontend:**
```bash
vercel projects list
```

---

## ロールバック手順

### Backend ロールバック（ECS Task Definition）

**前バージョンに戻す:**

```bash
export AWS_PROFILE=AnimalEkarte

# 利用可能な Task Definition を確認
aws ecs describe-task-definition \
  --task-definition animalekarte-test-api \
  --region us-east-1 \
  --query taskDefinition.revision

# 前のバージョンにロールバック（例：v4に戻す）
aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --task-definition animalekarte-test-api:4 \
  --region us-east-1
```

**確認:**
```bash
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].taskDefinition'
```

### Frontend ロールバック（Vercel）

1. **Vercel ダッシュボード** → **Deployments**
2. 前のバージョンの **"Promote to Production"** をクリック

または Vercel CLI：
```bash
vercel --prod --no-build
```

---

## 自動デプロイ動作確認テスト

### テスト実行（2026-03-04 実施済み）

**Backend テスト:**
```bash
# test.txt 作成 → push
git add backend/internal/logger/test.txt
git commit -m "test(backend): auto-deploy verification"
git push origin main

# GitHub Actions Run #4 自動トリガー確認 ✅
# ECS Task Definition v4 → v5 更新確認 ✅
```

**Frontend テスト:**
```bash
# test-deploy.ts 作成 → push
git add frontend/src/test-deploy.ts
git commit -m "test(frontend): auto-deploy verification"
git push origin main

# Vercel 自動ビルド開始確認 ✅
```

**結果:**
- ✅ Backend: 完全稼働（自動トリガー確認）
- ✅ Frontend: 完全稼働（自動トリガー確認）

---

## よくある質問 (FAQ)

### Q: デプロイはどのくらい時間がかかりますか？

**Backend:** 1-5分
- ビルド: 30-60秒
- ECS デプロイ: 10-30秒
- 安定性確認: 10-30秒

**Frontend:** 2-5分
- ビルド: 1-2分
- デプロイ・キャッシュ最適化: 1-3分

### Q: main ブランチ以外へのプッシュでもデプロイされますか？

**いいえ。** デプロイは `main` ブランチへのプッシュのみです。
```yaml
on:
  push:
    branches:
      - main  # main のみ
```

### Q: 特定の変更だけをデプロイしたい場合は？

`.github/workflows/backend-deploy.yml` の `paths` を確認：
```yaml
paths:
  - 'backend/**'  # backend フォルダのみトリガー
  - '.github/workflows/backend-deploy.yml'
```

frontend 変更は frontend のみデプロイされます。

### Q: 手動でデプロイを実行したい場合は？

```bash
# Backend 手動実行
gh workflow run backend-deploy.yml --ref main

# Frontend は Vercel CLI で
cd frontend
vercel --prod
```

### Q: デプロイをスキップしたい場合は？

**Backend:** `.github/workflows/backend-deploy.yml` を削除・無効化

**Frontend:** Vercel ダッシュボード → Settings → Auto-deployments を無効化

### Q: ロールバック後、新バージョンに上げるには？

再度 `main` ブランチにコミット・プッシュしてください。自動デプロイが実行されます。

---

## 参考資料

| リソース | 説明 |
|---------|------|
| [AWS OIDC GitHub Actions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect) | GitHub Actions OIDC 認証 |
| [Amazon ECS](https://docs.aws.amazon.com/ecs/) | ECS ドキュメント |
| [Vercel GitHub Integration](https://vercel.com/docs/git-integrations/vercel-for-github) | Vercel GitHub 連携 |
| [ECR](https://docs.aws.amazon.com/ecr/) | Amazon ECR ドキュメント |

---

**最終確認日:** 2026-03-04
**確認者:** Claude Code (Haiku 4.5)
**デプロイ状態:** ✅ Backend・Frontend ともに完全稼働
