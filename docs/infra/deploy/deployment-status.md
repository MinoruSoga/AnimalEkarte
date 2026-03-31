# デプロイ状態・AWSリソース一覧

**最終更新:** 2026-03-31 | **リージョン:** us-east-1

---

## 現在の稼働状況

| コンポーネント | 状態 |
|--------------|------|
| ECS Service | ✅ ACTIVE (running: 1 / desired: 1) |
| RDS PostgreSQL | ✅ available |
| CloudFront (HTTPS終端) | ✅ `dcqico6azu5w2.cloudfront.net` |
| Frontend (Vercel) | ✅ Production デプロイ済み |
| 自動デプロイ (Backend) | ✅ GitHub Actions + AWS OIDC |
| 自動デプロイ (Frontend) | ✅ Vercel GitHub Integration |

---

## AWS リソース

### ネットワーク

| リソース | 値 |
|---------|-----|
| VPC | vpc-0146cdfb3553c24ac (10.0.0.0/16) |
| Public Subnet | 10.0.1.0/24 (us-east-1a) — ALB 配置 |
| Private Subnet | 10.0.10.0/24 (us-east-1a) — ECS, RDS 配置 |
| NAT Gateway | nat-060ddadf4af6e951c |

### コンピューティング

| リソース | 値 |
|---------|-----|
| ECS Cluster | animalekarte-test-cluster |
| ECS Service | animalekarte-test-service |
| Task Definition | animalekarte-test-api (0.25 vCPU / 0.5 GB) |
| ALB URL（直接） | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com |
| CloudFront URL | https://dcqico6azu5w2.cloudfront.net (Distribution: ERCVR5P0IAJKS) |
| ECR Repository | 698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api |

### データベース

| リソース | 値 |
|---------|-----|
| RDS Instance | animalekarte-test-db (db.t4g.micro, PostgreSQL 16) |
| Endpoint | animalekarte-test-db.cqbe28s44fta.us-east-1.rds.amazonaws.com:5432 |
| Storage | 20GB gp3, 暗号化有効, Backup 1日（テスト環境） |
| DB 認証情報 | SSM Parameter Store (`/animalekarte/test/db/user`, `/password`, `/name`) |

### IAM Roles

| Role | 用途 |
|------|------|
| animalekarte-test-github-ecs-deploy-role | GitHub Actions → ECS デプロイ (OIDC) |
| animalekarte-test-github-terraform-role | Terraform 実行 |
| animalekarte-test-ecs-task-role | CloudWatch Logs 書き込み |
| animalekarte-test-ecs-execution-role | ECR Pull, SSM Parameter Store 読み取り |

### セキュリティグループ

| SG | ルール |
|----|--------|
| alb-sg | 0.0.0.0/0:80 → ecs-sg:8080 |
| ecs-sg | alb-sg:8080, 0.0.0.0/0:443 (ECR 用) |
| rds-sg | ecs-sg:5432 |

### ログ・監視

| リソース | 値 |
|---------|-----|
| CloudWatch Logs | /ecs/animalekarte-test (30日保持) |

---

## 環境変数

### Backend (ECS Task Definition)

| 変数 | 値 |
|------|-----|
| `PORT` | 8080 |
| `GIN_MODE` | release |
| `DB_SSL_MODE` | require |
| `CORS_ALLOWED_ORIGIN` | https://frontend-eta-six-20.vercel.app,https://dcqico6azu5w2.cloudfront.net |
| `DB_HOST/PORT/USER/PASSWORD/NAME` | SSM Parameter Store から注入 |

### Frontend (Vercel)

| 変数 | 値 |
|------|-----|
| `VITE_API_URL` | https://dcqico6azu5w2.cloudfront.net/api |

---

## コスト（月額見積もり）

| サービス | 月額 (USD) |
|---------|-----------|
| NAT Gateway | $32 |
| ALB | $21 |
| RDS db.t4g.micro | $14.5 |
| ECS Fargate | $6.65 |
| その他 (ECR, CloudWatch, SSM, S3) | $4.25 |
| **合計** | **約 $78/月 (約 130,000円/年)** |

**削減案:** NAT Gateway を VPC Endpoints (ECR, SSM) に置換 → 月 $32 削減

---

## 既知の問題

| 問題 | 優先度 | 推奨対応 |
|------|--------|---------|
| ~~CORS ミドルウェアが `*` のまま~~ | ~~High~~ | ✅ **解決済み (PR #10)**: `CORS_ALLOWED_ORIGIN` 環境変数を参照するよう修正済み |
| ALB が HTTP のみ (HTTPS 未終端) | Low | CloudFront で HTTPS 終端済み。ALB 自体は CloudFront からのみアクセスされるため許容範囲 |
| Single-AZ 構成 | Low | 本番環境では Multi-AZ 化 |
| WAF 未導入 | Low | 本番環境では AWS WAF 導入 |

---

## 次のアクション

### 短期

- [ ] CORS 設定を Vercel ドメイン限定に修正
- [ ] CloudWatch Alarms 設定（ECS エラーレート、RDS CPU）
- [ ] AWS Budgets アラート設定（$100/月）

### 中期

- [ ] HTTPS 化（ACM 証明書取得）
- [ ] カスタムドメイン設定

### 本番リリース前

- [ ] Multi-AZ 化（RDS Multi-AZ, ECS desired count 2+）
- [ ] CloudFront + WAF 導入
- [ ] RDS Read Replica 検討

---

## インフラ確認コマンド

```bash
export AWS_PROFILE=AnimalEkarte

# Terraform state 確認
cd infra/terraform && terraform state list | wc -l

# ECS サービス確認
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 | jq '.services[0] | {status, runningCount, desiredCount}'

# RDS 確認
aws rds describe-db-instances \
  --db-instance-identifier animalekarte-test-db \
  --region us-east-1 | jq '.DBInstances[0] | {DBInstanceStatus, Endpoint}'

# ECR 最新イメージ
aws ecr describe-images \
  --repository-name animalekarte-api \
  --region us-east-1 \
  --query 'sort_by(imageDetails, &imagePushedAt)[-1]'
```
