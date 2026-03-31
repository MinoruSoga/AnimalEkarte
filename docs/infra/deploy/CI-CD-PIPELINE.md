# CI/CD パイプライン

**ステータス:** ✅ 完全稼働中

---

## 概要

| コンポーネント | 方式 | トリガー |
|--------------|------|---------|
| Backend API | GitHub Actions → AWS OIDC → ECR → ECS | `backend/**` → main push |
| Frontend | Vercel GitHub Integration | `frontend/**` → main push |

---

## Backend デプロイ

### パイプライン

```
git push to main (backend/** 変更)
  → GitHub Actions backend-deploy.yml
  → AWS OIDC 認証
  → Docker build (backend/Dockerfile.production)
  → ECR push (animalekarte-api:latest + sha-<commit>)
  → ECS Task Definition 更新
  → ECS Service 更新 (Blue/Green)
  → サービス安定性確認
```

**所要時間:** 1〜5分

### ワークフロー設定

**ファイル:** `.github/workflows/backend-deploy.yml`

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'backend/**'
      - '.github/workflows/backend-deploy.yml'
  workflow_dispatch:  # 手動実行可

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: animalekarte-api
  ECS_CLUSTER: animalekarte-test-cluster
  ECS_SERVICE: animalekarte-test-service
  ECS_TASK_DEFINITION_FAMILY: animalekarte-test-api
```

### AWS OIDC 認証設定

**IAM Role:** `animalekarte-test-github-ecs-deploy-role`

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

**重要:** リポジトリオーナーは `MinoruSoga`（大文字小文字一致必須）

### GitHub Secrets

| Secret | 値 |
|--------|-----|
| `AWS_REGION` | us-east-1 |
| `AWS_ACCOUNT_ID` | 698109622668 |
| `ECR_REPOSITORY` | animalekarte-api |
| `ECS_CLUSTER` | animalekarte-test-cluster |
| `ECS_SERVICE` | animalekarte-test-service |
| `ECS_TASK_DEFINITION` | animalekarte-test-api |

### ECR イメージタグ

```
698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api:latest
698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api:sha-<commit-hash>
```

---

## Frontend デプロイ

### パイプライン

```
git push to main (frontend/** 変更)
  → Vercel GitHub Hook
  → Vercel ビルド (npm install → npm run build)
  → Production 自動デプロイ
```

**所要時間:** 2〜5分

### Vercel 設定

**プロジェクト:** animalekarte-frontend
**vercel.json:**

```json
{
  "buildCommand": "npm run build",
  "installCommand": "npm install",
  "framework": "vite",
  "outputDirectory": "dist"
}
```

**環境変数（Production）:**

```
VITE_API_URL=https://dcqico6azu5w2.cloudfront.net/api
```

---

## 手動デプロイ

### Backend（GitHub Actions）

```bash
gh workflow run backend-deploy.yml --ref main
gh run list --workflow=backend-deploy.yml --limit 1
```

### Backend（手動 ECR Push）

自動デプロイが使えない場合の手順。

> **Apple Silicon (arm64) の場合:** ECS Fargate は `linux/amd64` で動作するため、`--platform linux/amd64` の指定が必須。

```bash
export AWS_PROFILE=AnimalEkarte
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=698109622668
export ECR_URL=698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api

# ECR 認証
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# ビルド・プッシュ・デプロイ（一括）
# Apple Silicon の場合は --platform linux/amd64 を必ず付ける
cd backend && \
  docker buildx build --platform linux/amd64 -f Dockerfile.production -t $ECR_URL:latest --push . && \
  aws ecs update-service \
    --cluster animalekarte-test-cluster \
    --service animalekarte-test-service \
    --force-new-deployment \
    --region $AWS_REGION
```

### Frontend（Vercel CLI）

```bash
cd frontend
vercel --prod
```

---

## ロールバック

### Backend

```bash
export AWS_PROFILE=AnimalEkarte

# 利用可能な Task Definition バージョン確認
aws ecs list-task-definitions \
  --family-prefix animalekarte-test-api \
  --region us-east-1

# 前バージョンにロールバック（例: v4）
aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --task-definition animalekarte-test-api:4 \
  --region us-east-1
```

### Frontend

Vercel ダッシュボード → Deployments → 前バージョン → "Promote to Production"

---

## デプロイ確認

```bash
export AWS_PROFILE=AnimalEkarte

# Backend ヘルスチェック
curl http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health | jq .

# ECS サービス状態
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].{status,runningCount,desiredCount,taskDefinition}'

# ログ確認
aws logs tail /ecs/animalekarte-test --follow --region us-east-1
```

---

## トラブルシューティング

### AWS OIDC 認証エラー

**症状:** `HTTP 422 AssumeRoleUnauthorizedOperation`

**対処:** IAM コンソール → `animalekarte-test-github-ecs-deploy-role` → 信頼関係を確認。リポジトリ名が `MinoruSoga/AnimalEkarte` になっているか確認。

### Docker ビルド失敗

```bash
cd backend
docker build -f Dockerfile.production -t test .
```

### ECS タスク起動失敗 / ヘルスチェック失敗

```bash
export AWS_PROFILE=AnimalEkarte

# タスク停止理由
TASK_ARN=$(aws ecs list-tasks \
  --cluster animalekarte-test-cluster \
  --service-name animalekarte-test-service \
  --region us-east-1 \
  --query 'taskArns[0]' --output text)

aws ecs describe-tasks \
  --cluster animalekarte-test-cluster \
  --tasks $TASK_ARN \
  --region us-east-1 \
  | jq '.tasks[0] | {lastStatus, healthStatus, stoppedReason}'

# CloudWatch Logs
aws logs tail /ecs/animalekarte-test --since 10m --region us-east-1

# ALB Target Health
TG_ARN=$(aws elbv2 describe-target-groups \
  --names animalekarte-test-tg \
  --region us-east-1 \
  | jq -r '.TargetGroups[0].TargetGroupArn')

aws elbv2 describe-target-health \
  --target-group-arn $TG_ARN \
  --region us-east-1 \
  | jq '.TargetHealthDescriptions'
```

| エラー | 原因 | 対処 |
|--------|------|------|
| `CannotPullContainerError` | ECRにイメージなし | 手動 ECR Push を実行 |
| `Target.FailedHealthChecks` | /health が応答しない | CloudWatch Logs 確認 |
| `ResourceInitializationError` | 環境変数/リソース不足 | Task Definition 確認 |
| `503 Service Unavailable` | ALB → ECS 疎通不可 | Security Group 確認 |

### Frontend ビルド失敗

```bash
docker compose exec frontend npm run build
docker compose exec frontend npm run lint
```

### CORS エラー

`CORS_ALLOWED_ORIGIN` 環境変数（カンマ区切り）に Vercel ドメインと CloudFront ドメインが含まれているか確認。

```
CORS_ALLOWED_ORIGIN=https://frontend-eta-six-20.vercel.app,https://dcqico6azu5w2.cloudfront.net
```

> ⚠️ env var 名は `ALLOWED_ORIGINS` **ではなく** `CORS_ALLOWED_ORIGIN` であることに注意（`backend/internal/middleware/cors.go` 参照）。

---

## 参考資料

| リソース | 説明 |
|---------|------|
| [AWSリソース一覧](./deployment-status.md) | Terraform state、コスト |
| [デプロイ前チェックリスト](./DEPLOYMENT-CHECKLIST.md) | デプロイ手順、DB作り直し |
| [AWS OIDC](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect) | GitHub Actions OIDC 認証 |
