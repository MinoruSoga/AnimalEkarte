# Infra - デプロイ & インフラ構成

## テスト環境エンドポイント

| サービス | URL |
|---------|-----|
| Frontend | https://frontend-eta-six-20.vercel.app |
| Backend API | https://dcqico6azu5w2.cloudfront.net/api |
| ALB（直接） | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com |
| RDS | animalekarte-test-db.cqbe28s44fta.us-east-1.rds.amazonaws.com:5432 |

---

## アーキテクチャ概要

```
Vercel (Frontend SPA)
    ↓ HTTPS
CloudFront (API Gateway, *.cloudfront.net)
    ↓ HTTP
ALB (ヘルスチェック: /health)
    ↓
ECS Fargate (Go/Gin API, CPU:256, Memory:512MB)
    ↓
RDS PostgreSQL 16.4 (db.t4g.micro)
```

**CloudFront を使う理由**: ALB の `*.elb.amazonaws.com` ドメインには ACM 証明書を発行できないため、CloudFront の `*.cloudfront.net` で HTTPS を終端。

---

## ネットワーク構成

```
VPC: 10.0.0.0/16 (us-east-1)
├── Public Subnets
│   ├── 10.0.1.0/24 (us-east-1a) — ALB, NAT Gateway
│   └── 10.0.2.0/24 (us-east-1b)
└── Private Subnets
    ├── 10.0.11.0/24 (us-east-1a) — ECS Fargate
    └── 10.0.12.0/24 (us-east-1b) — RDS
```

### セキュリティグループ

| リソース | 許可ルール |
|---------|-----------|
| ALB | 80/tcp, 443/tcp from 0.0.0.0/0 |
| ECS | 8080/tcp from ALB SG のみ |
| RDS | 5432/tcp from ECS SG + 開発者 IP |

---

## CI/CD パイプライン

**ファイル**: `.github/workflows/backend-deploy.yml`

### トリガー

- `main` ブランチへのプッシュ（`backend/**` または workflow ファイル変更時）
- 手動トリガー（`workflow_dispatch`）

### デプロイフロー

```
main ブランチへ Push (backend/**)
    ↓
GitHub Actions トリガー
    ↓
AWS OIDC 認証（animalekarte-test-github-ecs-deploy-role）
    ↓
Docker Build (linux/amd64) & ECR Push
  → タグ: ${github.sha} + latest
    ↓
ECS タスク定義更新（イメージタグのみ差し替え）
    ↓
ECS サービス更新（wait-for-service-stability: true）
    ↓
HealthCheck 確認 (/health)
```

### GitHub OIDC

| ロール | 用途 | 権限 |
|--------|------|------|
| `animalekarte-test-github-terraform-role` | Terraform 実行 | AdministratorAccess（テスト環境限定） |
| `animalekarte-test-github-ecs-deploy-role` | ECS デプロイ | ECR Push + ECS Update + IAM PassRole |

---

## Terraform

### ディレクトリ構成

```
infra/
├── terraform/
│   ├── main.tf              # モジュールオーケストレーション
│   ├── variables.tf          # ルート変数（23個）
│   ├── outputs.tf            # 出力（33個）
│   ├── providers.tf          # AWS プロバイダ（profile: AnimalEkarte）
│   ├── backend.tf            # S3 + DynamoDB State 管理
│   ├── terraform.tfvars      # テスト環境変数値
│   └── modules/
│       ├── vpc/              # VPC, Subnet, IGW, NAT
│       ├── security/         # SG, CloudWatch, SSM, IAM
│       ├── rds/              # PostgreSQL RDS
│       ├── ecr/              # ECR リポジトリ（animalekarte-api）
│       ├── ecs/              # ALB, ECS Cluster/Service/Task
│       └── github-oidc/      # GitHub Actions OIDC 連携
├── terraform-bootstrap/      # S3・DynamoDB 初期化
└── docs/
    ├── deployment-guide.md   # デプロイ手順・トラブルシューティング
    └── architecture.md       # アーキテクチャ詳細
```

### State 管理

```
S3: animalekarte-tfstate-698109622668
Key: env/test/terraform.tfstate
Lock: DynamoDB animalekarte-terraform-lock
Encryption: enabled
```

### 主要変数

| 変数 | デフォルト値 | 説明 |
|------|-------------|------|
| `name_prefix` | `animalekarte-test` | リソース名プレフィックス |
| `rds_instance_class` | `db.t4g.micro` | RDS インスタンスタイプ |
| `ecs_task_cpu` | `256` | ECS タスク CPU |
| `ecs_task_memory` | `512` | ECS タスクメモリ (MB) |
| `ecs_desired_count` | `1` | ECS タスク数 |
| `use_public_rds` | `false` | RDS パブリックアクセス |
| `cors_allowed_origin` | — | CORS 許可オリジン |

### Terraform 実行

```bash
# AWS CLI プロファイル指定必須
export AWS_PROFILE=AnimalEkarte

cd infra/terraform
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

---

## Docker（本番ビルド）

### Backend: `backend/Dockerfile.production`

- マルチステージビルド（builder → runtime）
- Go 1.25-alpine, CGO_ENABLED=0（静的リンク）
- Runtime: alpine 3.21（非 root ユーザ: appuser:1000）
- HealthCheck: `/health` エンドポイント

### ビルドコマンド

```bash
make build-prod   # animal-ekarte-api:latest + animal-ekarte-front:latest
```

---

## Vercel（Frontend）

**設定**: `frontend/vercel.json`

- SPA リライト: `/(.*) → /index.html`
- セキュリティヘッダー: `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Referrer-Policy`
- 静的アセットキャッシュ: `Cache-Control: public, max-age=31536000, immutable`
- 環境変数: `VITE_API_URL=https://dcqico6azu5w2.cloudfront.net/api`

---

## シークレット管理

### SSM Parameter Store

| パラメータ | パス | タイプ |
|-----------|------|--------|
| DB User | `/animalekarte/test/db/user` | String |
| DB Password | `/animalekarte/test/db/password` | SecureString |
| DB Name | `/animalekarte/test/db/name` | String |

---

## 認証方式

- **Authorization Bearer**: JWT トークンを `sessionStorage` に保存
- axios インターセプターで全 API リクエストに `Authorization: Bearer <token>` を自動注入
- `withCredentials` は不要（Cookie 不使用）

---

## テスト環境の制約（本番移行時の変更点）

| 項目 | テスト環境 | 本番で必要 |
|------|-----------|-----------|
| NAT Gateway | 単一 AZ | マルチ AZ |
| RDS Multi-AZ | 無効 | 有効化 |
| RDS 削除保護 | 無効 | 有効化 |
| RDS Public Access | 有効 | 無効化 |
| GitHub Terraform Role | AdministratorAccess | 最小権限ポリシー |
| CloudFront | 手動作成 | Terraform 管理化 |
| 独自ドメイン | なし | 取得 + ACM 証明書 |
| ALB HTTP | forward | redirect（HTTPS 強制） |

---

## 参照ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [デプロイガイド](docs/deployment-guide.md) | デプロイ手順・トラブルシューティング |
| [アーキテクチャ詳細](docs/architecture.md) | インフラアーキテクチャ設計 |
