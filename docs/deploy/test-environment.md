# Test環境構成

## 概要

AnimalEkarte Test環境のアクセス情報、AWS リソース構成、運用手順をまとめたドキュメント。

**最終更新日: 2026-02-16**

---

## アクセスURL

| サービス | URL | 説明 |
|---------|-----|------|
| Frontend | https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app | React 19 SPA (Vercel) |
| Backend API | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com | Go API (ECS Fargate) |
| Swagger UI | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/swagger/index.html | API仕様書 |

---

## AWS リソース構成

### リージョン
- **us-east-1** (バージニア北部)

### コンピューティング

| リソース | 識別子 | スペック | 詳細 |
|---------|--------|---------|------|
| ECS Cluster | animalekarte-test-cluster | Fargate | - |
| ECS Service | animalekarte-test-service | desired: 1, running: 1 | Auto Scaling未設定 |
| Task Definition | animalekarte-test-api:2 | 0.25 vCPU, 0.5 GB | Backend API |
| ALB | animalekarte-test-alb | Internet-facing | HTTP:80 → ECS:8080 |
| Target Group | animalekarte-test-tg | Health Check: `/health` | - |

### データベース

| リソース | 識別子 | スペック | 詳細 |
|---------|--------|---------|------|
| RDS Instance | animalekarte-test-db | db.t4g.micro | PostgreSQL 18 |
| Endpoint | animalekarte-test-db.cqbe28s44fta.us-east-1.rds.amazonaws.com | Port: 5432 | Private Subnet配置 |
| Storage | 20GB gp3 | 暗号化有効 | - |
| Backup | 7日保持 | - | - |

### ネットワーク

| リソース | CIDR/ID | 詳細 |
|---------|---------|------|
| VPC | 10.0.0.0/16 | animalekarte-test-vpc |
| Public Subnet | 10.0.1.0/24 | us-east-1a (ALB配置) |
| Private Subnet | 10.0.10.0/24 | us-east-1a (ECS, RDS配置) |
| NAT Gateway | - | Public Subnet配置 |
| Internet Gateway | - | VPCアタッチ済み |

### セキュリティ

| Security Group | Inbound Rules | 用途 |
|---------------|---------------|------|
| alb-sg | 0.0.0.0/0:80,443 | ALB |
| ecs-sg | alb-sg:8080, 0.0.0.0/0:443 (ECR) | ECS Task |
| rds-sg | ecs-sg:5432 | RDS PostgreSQL |

### IAM Roles

| Role | 用途 | 主要ポリシー |
|------|------|------------|
| github-oidc-terraform-role | Terraform実行 | AdministratorAccess |
| github-oidc-ecs-deploy-role | Backend デプロイ | ECR, ECS, IAM PassRole |
| ecs-task-role | ECS Task | CloudWatch Logs書き込み |
| ecs-execution-role | ECS Task起動 | ECR Pull, Secrets Manager読み取り |

### ストレージ・監査

| リソース | 詳細 |
|---------|------|
| ECR Repository | animalekarte-api (10 images) |
| CloudWatch Logs | /ecs/animalekarte-test (30日保持) |
| Secrets Manager | animalekarte-test-db-credentials |

---

## 環境変数

### Frontend (Vercel)

| 変数名 | 値 |
|--------|-----|
| `VITE_API_URL` | `http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api` |

**設定場所:** Vercelプロジェクト設定 → Environment Variables → Production

### Backend (ECS Task Definition)

#### Environment

| 変数名 | 値 | 説明 |
|--------|-----|------|
| `PORT` | `8080` | Ginサーバーポート |
| `GIN_MODE` | `release` | 本番モード |
| `DB_SSL_MODE` | `require` | RDS SSL接続 |
| `ALLOWED_ORIGINS` | `https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app` | CORS許可オリジン |

#### Secrets (SSM Parameter Store / Secrets Manager)

| 変数名 | Secret ARN |
|--------|-----------|
| `DB_HOST` | arn:aws:secretsmanager:us-east-1:xxx:secret:animalekarte-test-db-credentials:host:: |
| `DB_PORT` | arn:aws:secretsmanager:us-east-1:xxx:secret:animalekarte-test-db-credentials:port:: |
| `DB_USER` | arn:aws:secretsmanager:us-east-1:xxx:secret:animalekarte-test-db-credentials:username:: |
| `DB_PASSWORD` | arn:aws:secretsmanager:us-east-1:xxx:secret:animalekarte-test-db-credentials:password:: |
| `DB_NAME` | arn:aws:secretsmanager:us-east-1:xxx:secret:animalekarte-test-db-credentials:dbname:: |

---

## デプロイ方法

### 前提条件

- AWS CLI設定済み (`AWS_PROFILE=AnimalEkarte`)
- GitHub Actions OIDC認証設定済み
- Vercel CLI インストール済み（Frontend手動デプロイ時）

### Backend自動デプロイ

**トリガー:** `backend/**` への push

```bash
# 1. Backend コード修正
vi backend/internal/handler/owner_handler.go

# 2. Commit & Push
git add backend/
git commit -m "feat: update owner handler"
git push origin main

# 3. GitHub Actions 確認
# https://github.com/<your-org>/AnimalEkarte/actions
# "Backend Deploy to ECS" ワークフローが自動実行される

# 4. デプロイ完了確認
export AWS_PROFILE=AnimalEkarte
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  | jq '.services[0].deployments'
```

**デプロイフロー:**
1. Docker Image Build
2. ECR Push
3. ECS Task Definition 更新
4. ECS Service 更新（Blue/Green切り替え）
5. ヘルスチェック待機（最大10分）

### Frontend自動デプロイ

**トリガー:** `frontend/**` への push (Vercel自動デプロイ)

```bash
# 1. Frontend コード修正
vi frontend/src/features/owners/components/owner-list.tsx

# 2. Commit & Push
git add frontend/
git commit -m "feat: update owner list UI"
git push origin main

# 3. Vercel 自動デプロイ確認
# https://vercel.com/minorusogas-projects/animalekarte-frontend
# Production デプロイが自動実行される
```

---

## ログ確認

### Backend (CloudWatch Logs)

```bash
export AWS_PROFILE=AnimalEkarte

# リアルタイムログ確認
aws logs tail /ecs/animalekarte-test \
  --follow \
  --region us-east-1

# 特定期間のログ検索
aws logs filter-log-events \
  --log-group-name /ecs/animalekarte-test \
  --start-time $(date -u -d '1 hour ago' +%s)000 \
  --filter-pattern "ERROR" \
  --region us-east-1
```

### Frontend (Vercel)

```bash
# Vercel CLI 使用
cd frontend
vercel logs https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app

# または、Vercel Dashboard でログ確認
# https://vercel.com/minorusogas-projects/animalekarte-frontend/logs
```

---

## データベース操作

### psql 接続

```bash
export AWS_PROFILE=AnimalEkarte

# 1. Secrets Manager から認証情報取得
aws secretsmanager get-secret-value \
  --secret-id animalekarte-test-db-credentials \
  --region us-east-1 \
  | jq -r '.SecretString' > /tmp/db-creds.json

# 2. 環境変数設定
export DB_HOST=$(jq -r '.host' /tmp/db-creds.json)
export DB_PORT=$(jq -r '.port' /tmp/db-creds.json)
export DB_USER=$(jq -r '.username' /tmp/db-creds.json)
export DB_PASSWORD=$(jq -r '.password' /tmp/db-creds.json)
export DB_NAME=$(jq -r '.dbname' /tmp/db-creds.json)

# 3. psql 接続
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

# クリーンアップ
rm /tmp/db-creds.json
```

### マイグレーション実行

```bash
# Docker 経由でマイグレーション実行
docker compose exec backend go run cmd/migrate/main.go
```

---

## 動作確認手順

### 1. Backend API ヘルスチェック

```bash
curl -s http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health
```

**期待値:**
```json
{
  "database": "connected",
  "message": "Animal Ekarte API is running",
  "status": "ok",
  "timestamp": "2026-02-16T12:47:10+09:00",
  "version": "1.0.0"
}
```

**HTTP Status: 200 OK**

### 2. Swagger UI 確認

ブラウザで以下にアクセス：
```
http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/swagger/index.html
```

### 3. Frontend 統合テスト

1. ブラウザで https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app にアクセス
2. DevTools → Network タブを開く
3. ページリロード
4. `http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api/v1/owners` へのリクエスト確認
5. CORS エラーがないことを確認
6. データが正常に表示されることを確認

---

## コスト（月額見積もり）

| サービス | 月額（USD） | 備考 |
|---------|------------|------|
| NAT Gateway | $32 | データ転送料込み |
| ECS Fargate | $6.65 | 0.25 vCPU, 0.5 GB, 常時稼働 |
| ALB | $21 | $16.2 (固定) + $0.008/LCU |
| RDS db.t4g.micro | $14.5 | Single-AZ |
| その他 | $4.25 | ECR, CloudWatch, S3 |
| **合計** | **約$78/月** | **年間 約$936 (約130,000円)** |

**コスト削減案:**
- NAT Gateway 削除 → VPC Endpoints使用（ECR, Secrets Manager）: 月$32削減
- RDS Auto Pause（Aurora Serverless v2）: 利用時のみ課金

---

## トラブルシューティング

### ECS Task が起動しない

```bash
# Task停止理由確認
aws ecs describe-tasks \
  --cluster animalekarte-test-cluster \
  --tasks <task-arn> \
  --region us-east-1 \
  | jq '.tasks[0].stoppedReason'

# CloudWatch Logs確認
aws logs tail /ecs/animalekarte-test --since 10m --region us-east-1
```

### RDS 接続エラー

```bash
# Security Group 確認
aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=*rds*" \
  --region us-east-1

# RDS エンドポイント確認
aws rds describe-db-instances \
  --db-instance-identifier animalekarte-test-db \
  --region us-east-1 \
  | jq '.DBInstances[0].Endpoint'
```

### CORS エラー

**症状:** Frontend から Backend API 呼び出し時に CORS エラー

**確認:**
1. Backend環境変数 `ALLOWED_ORIGINS` にVercelドメインが含まれているか
2. CloudWatch Logs で OPTIONS リクエストログ確認

**修正:**
```bash
# Task Definition 更新（環境変数修正）
aws ecs register-task-definition \
  --cli-input-json file://task-definition.json \
  --region us-east-1

# Service 更新
aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --task-definition animalekarte-test-api:3 \
  --region us-east-1
```

---

## セキュリティ考慮事項

### 現状の問題

- ⚠️ **CORS設定**: `Access-Control-Allow-Origin: *` (全許可) → Vercelドメイン限定に変更推奨
- ⚠️ **HTTP通信**: ALB → HTTP のみ → HTTPS化推奨（本番環境）
- ⚠️ **Public ALB**: Internet-facing → CloudFront + WAF 導入推奨（本番環境）

### 推奨改善策

1. **CORS厳格化**
   ```go
   // backend/internal/middleware/cors.go
   ALLOWED_ORIGINS=https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app
   ```

2. **HTTPS化** (本番環境)
   - ACM証明書取得
   - ALB Listener HTTPS:443 追加
   - HTTP → HTTPS リダイレクト

3. **WAF導入** (本番環境)
   - SQLインジェクション対策
   - XSS対策
   - Rate Limiting

---

## 関連ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [デプロイ状態記録](./deployment-status.md) | Terraform リソース一覧、既知の問題 |
| [API仕様](http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/swagger/index.html) | Swagger UI |
| [ERD](../ERD.md) | データベース設計 |
| [システム仕様](../../spec.md) | 要件定義 |
