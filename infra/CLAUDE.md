# Infra - デプロイ & インフラ構成

## セキュリティ取り扱い

- このファイル内の endpoint、role、bucket、network 名は operational-sensitive として扱う。外部チャット、Issue、PR、公開ログへ不用意に貼らない。
- secrets、password、token、private key、個人情報は絶対に追加しない。必要な値は SSM Parameter Store、Vercel Secrets、または環境固有の secret manager を参照する。
- インフラ変更は plan → 差分確認 → 明示承認 → apply の順に進める。production-impacting action は必ず停止して承認を得る。

## ステージング環境エンドポイント

| サービス | URL |
|---------|-----|
| Frontend | https://stg.noah-karte.com |
| Backend API | https://api.stg.noah-karte.com/api (または https://dcqico6azu5w2.cloudfront.net/api) |
| ALB（直接） | http://animalekarte-stg-alb-1915768826.us-east-1.elb.amazonaws.com |
| RDS | animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com:5432 |

---

## アーキテクチャ概要

```
Vercel (Frontend SPA)
    ↓ HTTPS
CloudFront (API Gateway, *.cloudfront.net)
    ↓ HTTP
ALB (ヘルスチェック: /health)
    ↓
ECS Fargate SPOT (Go/Gin API, CPU:256, Memory:512MB)
    ↓
RDS PostgreSQL 16.4 (db.t4g.micro, private only)
```

> **コスト最適化（2026-06-01〜）**: ECS は Fargate Spot、外向きは NAT Gateway ではなく
> fck-nat インスタンス（t4g.nano + auto-recovery）、RDS は public access 無効（SSM port-forward 経由）、
> Container Insights 無効。毎日 22:00–8:00 JST に ECS=0 + RDS stop の夜間スケジュール（EventBridge）。
> toggle: `use_nat_instance` / `enable_off_hours_schedule`（いずれも default true）。

**CloudFront を使う理由**: ALB の `*.elb.amazonaws.com` ドメインには ACM 証明書を発行できないため、CloudFront の `*.cloudfront.net` で HTTPS を終端。

---

## ネットワーク構成

```
VPC: 10.0.0.0/16 (us-east-1)
├── Public Subnets
│   ├── 10.0.1.0/24 (us-east-1a) — ALB, fck-nat インスタンス（旧 NAT Gateway）
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
AWS OIDC 認証（animalekarte-stg-github-ecs-deploy-role）
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
| `animalekarte-stg-github-terraform-role` | Terraform 実行 | AdministratorAccess（テスト環境限定） |
| `animalekarte-stg-github-ecs-deploy-role` | ECS デプロイ | ECR Push + ECS Update + IAM PassRole |

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
Key: env/stg/terraform.tfstate
Lock: DynamoDB animalekarte-terraform-lock
Encryption: enabled
```

### 主要変数

| 変数 | デフォルト値 | 説明 |
|------|-------------|------|
| `name_prefix` | `animalekarte-stg` | リソース名プレフィックス |
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

- **httpOnly Cookie** で JWT を管理（`sessionStorage` / `localStorage` 不使用）
- Cookie 名: `access_token`（15分）、`refresh_token`（7日、Path: `/api/v1/auth/refresh`）
- 本番環境（`GIN_MODE=release`）: `Secure=true`, `SameSite=None`（Vercel ↔ CloudFront クロスドメイン対応）
- 開発環境: `SameSite=Lax`（localhost 同一オリジン）
- フロントエンドの axios は `withCredentials: true` を設定して Cookie を自動送信
- `Authorization: Bearer` ヘッダは **不使用**（Cookie で完結）

---

## テスト環境の制約（本番移行時の変更点）

| 項目 | テスト環境（現状） | 本番で必要 |
|------|-----------|-----------|
| 外向き経路 | **fck-nat インスタンス（単一・auto-recovery）** | NAT Gateway マルチ AZ or fck-nat ASG |
| ECS 起動 | **Fargate Spot** | Fargate（On-Demand）|
| 稼働時間 | **夜間停止（22:00–8:00 JST、`enable_off_hours_schedule`）** | 24/7（false に上書き）|
| RDS Multi-AZ | 無効 | 有効化 |
| RDS 削除保護 | 無効 | 有効化 |
| RDS Public Access | **無効（SSM port-forward 経由）** | 無効 |
| Container Insights | **無効** | 有効化（監視必要なら）|
| GitHub Terraform Role | AdministratorAccess | 最小権限ポリシー |
| CloudFront | 手動作成 | Terraform 管理化 |
| 独自ドメイン | なし | 取得 + ACM 証明書 |
| ALB HTTP | forward (CloudFront が HTTPS 終端済み) | 独自ドメイン取得後に redirect 強制 |

---

## 参照ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [デプロイガイド](docs/deployment-guide.md) | デプロイ手順・トラブルシューティング |
| [アーキテクチャ詳細](docs/architecture.md) | インフラアーキテクチャ設計 |
