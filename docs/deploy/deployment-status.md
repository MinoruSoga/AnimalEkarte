# デプロイ状態記録

Test環境の現在のデプロイ状態、リソース一覧、既知の問題を記録。

**最終更新日: 2026-03-04**

⚠️ **重要な更新:** 自動デプロイパイプラインが完全稼働中です。詳細は [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) を参照してください。

---

## デプロイ概要

| 項目 | 状態 |
|------|------|
| インフラ（Terraform） | ✅ デプロイ完了（41リソース稼働中） |
| Backend API (ECS) | ✅ 稼働中（Task Definition v5, runningCount: 1/1） |
| Frontend (Vercel) | ✅ デプロイ完了（Production + GitHub Auto-Deploy） |
| RDS PostgreSQL | ✅ ACTIVE |
| **自動デプロイ (Backend)** | ✅ **完全稼働** (GitHub Actions + AWS OIDC) |
| **自動デプロイ (Frontend)** | ✅ **完全稼働** (Vercel GitHub Integration) |

**インフラ構築:** 2026-02-15 - 2026-02-16
**自動デプロイ検証:** 2026-03-04 ✅

**AWSリージョン:** us-east-1 (バージニア北部)

---

## 自動デプロイ パイプライン（2026-03-04 稼働開始）

### Backend デプロイ

**トリガー:**
- `main` ブランチへのプッシュ
- `backend/**` フォルダの変更

**フロー:**
```
git push → GitHub Actions → AWS OIDC 認証 → ECR build/push → ECS update
```

**確認（最新）:**
- Task Definition: **v5** (2026-03-04 22:40:09 更新)
- 前回実行: Run #5 in_progress (cleanup commit)
- 検証: ✅ テスト実行で auto-deploy 確認済み

### Frontend デプロイ

**トリガー:**
- `main` ブランチへのプッシュ
- `frontend/**` フォルダの変更

**フロー:**
```
git push → Vercel GitHub Hook → Vercel build → auto-deploy
```

**確認（最新）:**
- 最終更新: 2026-03-04 13秒前
- 前回: cleanup commit 自動デプロイ実行
- 検証: ✅ テスト実行で auto-deploy 確認済み

**詳細:** [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) を参照

---

## Terraform リソース一覧

### リソース確認コマンド

```bash
export AWS_PROFILE=AnimalEkarte
cd infra/terraform
terraform state list
```

### デプロイ済みリソース（41個）

#### VPC・ネットワーク（12個）

```
module.vpc.aws_vpc.main
module.vpc.aws_subnet.public[0]
module.vpc.aws_subnet.private[0]
module.vpc.aws_internet_gateway.main
module.vpc.aws_nat_gateway.main[0]
module.vpc.aws_eip.nat[0]
module.vpc.aws_route_table.public
module.vpc.aws_route_table.private[0]
module.vpc.aws_route_table_association.public[0]
module.vpc.aws_route_table_association.private[0]
module.vpc.aws_route.public_internet_gateway
module.vpc.aws_route.private_nat_gateway[0]
```

**詳細:**
- VPC: 10.0.0.0/16
- Public Subnet: 10.0.1.0/24 (us-east-1a)
- Private Subnet: 10.0.10.0/24 (us-east-1a)
- NAT Gateway: 1台（Public Subnet配置）
- Internet Gateway: VPCアタッチ済み

#### セキュリティ（10個）

```
module.security.aws_security_group.alb
module.security.aws_security_group.ecs
module.security.aws_security_group.rds
module.security.aws_security_group_rule.alb_ingress_http
module.security.aws_security_group_rule.ecs_ingress_from_alb
module.security.aws_security_group_rule.ecs_egress_all
module.security.aws_security_group_rule.rds_ingress_from_ecs
module.security.aws_iam_role.ecs_task_role
module.security.aws_iam_role.ecs_execution_role
module.security.aws_iam_role_policy_attachment.ecs_task_cloudwatch
```

**IAM Roles:**
- `animalekarte-test-ecs-task-role`: CloudWatch Logs書き込み
- `animalekarte-test-ecs-execution-role`: ECR Pull, Secrets Manager読み取り

**Security Groups:**
- `animalekarte-test-alb-sg`: 0.0.0.0/0:80 → ecs-sg:8080
- `animalekarte-test-ecs-sg`: alb-sg:8080, 0.0.0.0/0:443 (ECR)
- `animalekarte-test-rds-sg`: ecs-sg:5432

#### GitHub OIDC（3個）

```
module.github_oidc.aws_iam_openid_connect_provider.github
module.github_oidc.aws_iam_role.github_actions_terraform
module.github_oidc.aws_iam_role.github_actions_ecs_deploy
```

**Roles:**
- `github-oidc-terraform-role`: Terraform実行用（AdministratorAccess）
- `github-oidc-ecs-deploy-role`: Backend デプロイ用（ECR, ECS, PassRole）

#### RDS（5個）

```
module.rds.aws_db_instance.main
module.rds.aws_db_subnet_group.main
module.rds.aws_db_parameter_group.main
module.rds.aws_secretsmanager_secret.db_credentials
module.rds.aws_secretsmanager_secret_version.db_credentials
```

**詳細:**
- DB Instance: `animalekarte-test-db` (db.t4g.micro)
- Engine: PostgreSQL 18
- Storage: 20GB gp3, 暗号化有効
- Multi-AZ: false
- Backup: 7日保持
- Endpoint: `animalekarte-test-db.cqbe28s44fta.us-east-1.rds.amazonaws.com:5432`

#### ECS・ALB（11個）

```
module.ecs.aws_ecs_cluster.main
module.ecs.aws_ecs_task_definition.main
module.ecs.aws_ecs_service.main
module.ecs.aws_lb.main
module.ecs.aws_lb_target_group.main
module.ecs.aws_lb_listener.http
module.ecs.aws_cloudwatch_log_group.ecs
module.ecs.aws_ecr_repository.main
module.ecs.aws_appautoscaling_target.ecs
module.ecs.aws_appautoscaling_policy.cpu
module.ecs.aws_appautoscaling_policy.memory
```

**詳細:**
- ECS Cluster: `animalekarte-test-cluster`
- Service: `animalekarte-test-service` (desired: 1, running: 1)
- Task Definition: `animalekarte-test-api:2` (0.25 vCPU, 0.5 GB)
- ALB: `animalekarte-test-alb` (Internet-facing)
- ALB DNS: `animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com`
- Target Group: `/health` ヘルスチェック
- ECR Repository: `animalekarte-api` (10 images)
- CloudWatch Logs: `/ecs/animalekarte-test` (30日保持)

---

## Vercel デプロイ状態

### プロジェクト情報

| 項目 | 値 |
|------|-----|
| プロジェクト名 | animalekarte-frontend |
| オーナー | minorusogas-projects |
| プロジェクトID | prj_uDGZytq46Y9ee1OWC6QlzPqWpwox |
| GitHubリポジトリ | AnimalEkarte (frontend/) |
| Framework | Vite (React 19) |

### デプロイ履歴

```bash
# Vercel CLI でデプロイ一覧確認
cd frontend
vercel list
```

**最新デプロイ:**
- URL: https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app
- 環境: Production
- デプロイ日時: 2026-02-16
- Status: Ready
- Build Time: 約2分

### 環境変数

| 変数名 | 値 | 環境 |
|--------|-----|------|
| `VITE_API_URL` | `http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/api` | Production |

---

## ECR イメージ一覧

### リポジトリ情報

```bash
export AWS_PROFILE=AnimalEkarte
aws ecr describe-repositories \
  --repository-names animalekarte-api \
  --region us-east-1
```

**詳細:**
- Repository URI: `<account-id>.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api`
- イメージ数: 10
- 最新タグ: `latest`

### イメージタグ一覧

```bash
aws ecr list-images \
  --repository-name animalekarte-api \
  --region us-east-1 \
  | jq '.imageIds[] | select(.imageTag != null) | .imageTag' \
  | sort -r
```

**推定タグ:**
- `latest`
- `sha-<commit-hash>` (GitHub Actions CI/CDで自動生成)

---

## ECS Service詳細

### Service状態確認

```bash
export AWS_PROFILE=AnimalEkarte
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  | jq '.services[0] | {status, runningCount, desiredCount, deployments}'
```

**現在の状態:**
```json
{
  "status": "ACTIVE",
  "runningCount": 1,
  "desiredCount": 1,
  "deployments": [
    {
      "id": "ecs-svc/xxxx",
      "status": "PRIMARY",
      "taskDefinition": "animalekarte-test-api:2",
      "desiredCount": 1,
      "runningCount": 1,
      "healthStatus": "HEALTHY"
    }
  ]
}
```

### Task Definition詳細

```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-test-api:2 \
  --region us-east-1 \
  | jq '.taskDefinition | {cpu, memory, containerDefinitions[0].environment}'
```

**リソース:**
- CPU: 256 (0.25 vCPU)
- Memory: 512 (0.5 GB)
- Network Mode: awsvpc
- Requires Compatibilities: FARGATE

**環境変数:**
- `PORT=8080`
- `GIN_MODE=release`
- `DB_SSL_MODE=require`
- `ALLOWED_ORIGINS=https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app`

**Secrets:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (Secrets Manager)

---

## GitHub Actions 設定状態

### ワークフロー一覧

```bash
ls -la .github/workflows/
```

**設定済みワークフロー:**
1. `backend-deploy.yml` - Backend自動デプロイ
2. `terraform.yml` - Terraform plan/apply（設定済みの場合）

### backend-deploy.yml 概要

**トリガー:**
```yaml
on:
  push:
    branches: [main]
    paths:
      - 'backend/**'
```

**ステップ:**
1. AWS OIDC認証（`github-oidc-ecs-deploy-role`）
2. ECR ログイン
3. Docker build
4. ECR push (`latest` + `sha-<commit>`)
5. ECS Task Definition更新
6. ECS Service更新
7. デプロイ完了待機（最大10分）

### GitHub Secrets

**設定必須:**
- `AWS_REGION`: us-east-1
- `AWS_ACCOUNT_ID`: <your-account-id>
- `ECR_REPOSITORY`: animalekarte-api
- `ECS_CLUSTER`: animalekarte-test-cluster
- `ECS_SERVICE`: animalekarte-test-service
- `ECS_TASK_DEFINITION`: animalekarte-test-api

---

---

## AWS OIDC 設定（2026-03-04 修正）

### 修正内容

GitHub Actions の AWS OIDC 認証がリポジトリ名の不一致で失敗していました。

**修正前:**
```json
"token.actions.githubusercontent.com:sub": "repo:minoru-nakamura/AnimalEkarte:*"  // ❌ 間違ったオーナー
```

**修正後:**
```json
"token.actions.githubusercontent.com:sub": "repo:MinoruSoga/AnimalEkarte:ref:refs/heads/main"  // ✅ 正しいオーナー
```

**修正対象:** IAM Role `animalekarte-test-github-ecs-deploy-role`

**結果:** AWS OIDC 認証成功 → Backend 自動デプロイ完全稼働

---

## 既知の問題・改善事項

### 1. CORS設定要改善 (Priority: High)

**現状:**
```go
// backend/internal/middleware/cors.go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")  // 全許可
```

**問題:**
- セキュリティリスク：任意のドメインからAPI呼び出し可能
- CSRF攻撃のリスク

**推奨修正:**
```go
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")  // Vercelドメイン限定
origin := c.Request.Header.Get("Origin")
if strings.Contains(allowedOrigins, origin) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
}
```

**修正ファイル:** `backend/internal/middleware/cors.go`

### 2. HTTP通信のみ (Priority: Medium)

**現状:**
- ALB: HTTP:80 のみリスナー設定
- Frontend → Backend: HTTP通信

**推奨（本番環境）:**
- ACM証明書取得
- ALB HTTPS:443 リスナー追加
- HTTP → HTTPS リダイレクト
- Frontend環境変数: `VITE_API_URL=https://...`

### 3. Single-AZ構成 (Priority: Low)

**現状:**
- VPC: 1 AZ (us-east-1a) のみ
- RDS: Single-AZ
- ECS: desired count 1

**推奨（本番環境）:**
- Multi-AZ化（VPC: 2+ AZ、RDS Multi-AZ）
- ECS desired count: 2+（高可用性）

### 4. CloudFront + WAF 未導入 (Priority: Low)

**現状:**
- ALB: Internet-facing（直接公開）
- WAF未設定

**推奨（本番環境）:**
- CloudFront配信
- AWS WAF導入（SQLインジェクション、XSS対策）
- Rate Limiting設定

---

## 次のアクション

### 短期（1週間以内）

- [ ] CORS設定修正（Vercelドメイン限定）
- [ ] 統合動作確認（Frontend → Backend API）
- [ ] ログ監視設定（CloudWatch Alarms）
- [ ] コスト監視設定（AWS Budgets: $100/月）

### 中期（1ヶ月以内）

- [ ] HTTPS化検討（ACM証明書取得）
- [ ] カスタムドメイン設定検討
- [ ] Auto Scaling設定（CPU/Memory閾値）
- [ ] DB バックアップ検証

### 長期（本番リリース前）

- [ ] Multi-AZ化
- [ ] CloudFront + WAF 導入
- [ ] ECS desired count: 2+
- [ ] RDS Read Replica検討
- [ ] Blue/Green デプロイ検討

---

## 検証コマンド集

### インフラ状態確認

```bash
export AWS_PROFILE=AnimalEkarte

# Terraform state確認
cd infra/terraform
terraform state list | wc -l  # リソース数

# VPC確認
aws ec2 describe-vpcs --filters "Name=tag:Name,Values=animalekarte-test-vpc" --region us-east-1

# ECS Service確認
aws ecs describe-services --cluster animalekarte-test-cluster --services animalekarte-test-service --region us-east-1

# RDS確認
aws rds describe-db-instances --db-instance-identifier animalekarte-test-db --region us-east-1

# ALB確認
aws elbv2 describe-load-balancers --names animalekarte-test-alb --region us-east-1
```

### デプロイ確認

```bash
# Backend API ヘルスチェック
curl -s http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health

# Frontend確認
curl -I https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app

# ECR最新イメージ確認
aws ecr describe-images --repository-name animalekarte-api --region us-east-1 --query 'sort_by(imageDetails,& imagePushedAt)[-1]'
```

### ログ確認

```bash
# CloudWatch Logs
aws logs tail /ecs/animalekarte-test --follow --region us-east-1

# CloudTrail（最近のAPI呼び出し）
aws cloudtrail lookup-events --lookup-attributes AttributeKey=ResourceType,AttributeValue=AWS::ECS::Service --max-results 10 --region us-east-1
```

---

## コスト内訳（月額）

| サービス | リソース | 月額（USD） |
|---------|---------|------------|
| NAT Gateway | 1台 + データ転送 | $32 |
| ECS Fargate | 0.25 vCPU, 0.5 GB × 730h | $6.65 |
| ALB | 固定 + LCU | $21 |
| RDS | db.t4g.micro, 20GB gp3 | $14.5 |
| ECR | 10 images | $1 |
| CloudWatch Logs | 5GB/月 | $2.5 |
| Secrets Manager | 1 secret | $0.40 |
| S3 | CloudTrail logs | $0.35 |
| **合計** | - | **約$78/月** |

**年間:** 約$936 (約130,000円、$1=¥139換算)

---

## 参考資料

### 公式ドキュメント

- [ECS Fargate料金](https://aws.amazon.com/fargate/pricing/)
- [RDS料金](https://aws.amazon.com/rds/postgresql/pricing/)
- [Vercel料金](https://vercel.com/pricing)

### 内部ドキュメント

- [Test環境構成](./test-environment.md)
- [Terraform構成](../../infra/terraform/README.md)
- [API仕様](../../docs/API-ROADMAP.md)
- [ERD](../../docs/ERD.md)

---

**記録者:** Claude Code (Sonnet 4.5)
**最終確認日:** 2026-02-16
