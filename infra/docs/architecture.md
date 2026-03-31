# AnimalEkarte - インフラアーキテクチャ

## システム構成図

```
                        Internet
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
         ┌─────────┐ ┌──────────┐ ┌──────────────┐
         │ Vercel   │ │CloudFront│ │ TablePlus等  │
         │(Frontend)│ │(API GW)  │ │(DB直接接続)   │
         │  HTTPS   │ │  HTTPS   │ │              │
         └─────────┘ └────┬─────┘ └──────┬───────┘
                          │ HTTP:80       │ 5432/tcp
                          ▼               │
                    ┌───────────┐         │
                    │    ALB    │         │
                    │ HTTP:80   │         │
                    │(HTTPS終端は│         │
                    │CloudFront)│         │
                    └─────┬─────┘         │
                          │ 8080/tcp      │
                          ▼               ▼
              ┌─────────────────────────────────────┐
              │           VPC (10.0.0.0/16)          │
              │                                     │
              │  ┌──────────────────────────────┐   │
              │  │  Private Subnets              │   │
              │  │  10.0.11.0/24 | 10.0.12.0/24 │   │
              │  │                              │   │
              │  │  ┌──────────┐  ┌──────────┐  │   │
              │  │  │ECS Fargate│  │   RDS    │  │   │
              │  │  │  (api)   │──│PostgreSQL│  │   │
              │  │  │ :8080    │  │  :5432   │  │   │
              │  │  └──────────┘  └──────────┘  │   │
              │  └──────────────────────────────┘   │
              │                                     │
              │  ┌──────────────────────────────┐   │
              │  │  Public Subnets               │   │
              │  │  10.0.1.0/24 | 10.0.2.0/24   │   │
              │  │  ALB, NAT Gateway, IGW       │   │
              │  └──────────────────────────────┘   │
              └─────────────────────────────────────┘
```

## コンポーネント一覧

| コンポーネント | サービス | 用途 |
|---|---|---|
| Frontend | Vercel | React SPA ホスティング |
| API Gateway | CloudFront | HTTPS 終端、ALB へ HTTP 転送 |
| Load Balancer | ALB | トラフィック分散、ヘルスチェック |
| Backend API | ECS Fargate | Go/Gin REST API |
| Database | RDS PostgreSQL 16 | データ永続化 |
| Container Registry | ECR | Docker イメージ管理 |
| CI/CD | GitHub Actions | 自動ビルド・デプロイ |
| Secrets | SSM Parameter Store | DB 認証情報の安全な管理 |
| Logs | CloudWatch Logs | アプリケーションログ（30日保持） |
| IaC | Terraform | インフラコード管理 |

## ネットワーク構成

### VPC

| リソース | 値 |
|---|---|
| VPC CIDR | `10.0.0.0/16` |
| Public Subnet A | `10.0.1.0/24` (us-east-1a) |
| Public Subnet B | `10.0.2.0/24` (us-east-1b) |
| Private Subnet A | `10.0.11.0/24` (us-east-1a) |
| Private Subnet B | `10.0.12.0/24` (us-east-1b) |
| NAT Gateway | 単一構成（コスト最適化） |

### セキュリティグループ

| SG | 許可ルール |
|---|---|
| ALB (`sg-047ff4f8c3ab99411`) | 80/tcp, 443/tcp from 0.0.0.0/0 |
| ECS (`sg-0934ac397301ec633`) | 8080/tcp from ALB SG |
| RDS (`sg-053d44ac9fab4e71a`) | 5432/tcp from ECS SG + 開発者 IP |

## リクエストフロー

```
1. ブラウザ → Vercel (HTTPS) : フロントエンド取得
2. ブラウザ → CloudFront (HTTPS) : API リクエスト
3. CloudFront → ALB (HTTP:80) : オリジン接続（HTTP-only）
4. ALB → ECS Fargate (8080) : ターゲットグループ転送
5. ECS → RDS (5432) : DB クエリ
```

### CloudFront を使う理由

独自ドメインがないため、ALB のデフォルトドメイン（`*.elb.amazonaws.com`）に対して
ACM 証明書を発行できない。CloudFront の `*.cloudfront.net` ドメインは
AWS 提供の有効な SSL 証明書が付属するため、これを API Gateway として使用。

## ECS タスク定義

### 環境変数（平文）

| 変数名 | 値 | 用途 |
|---|---|---|
| `PORT` | `8080` | API リッスンポート |
| `DB_HOST` | RDS エンドポイント | DB 接続先 |
| `DB_PORT` | `5432` | DB ポート |
| `DB_SSL_MODE` | `require` | SSL 必須 |
| `CORS_ALLOWED_ORIGIN` | `https://frontend-eta-six-20.vercel.app,https://dcqico6azu5w2.cloudfront.net` | CORS 許可オリジン（カンマ区切り） |
| `GIN_MODE` | `release` | 本番モード（`release` のとき Cookie が `Secure=true`, `SameSite=None` になる） |
| `JWT_SECRET` | (SSM で管理すべき) | JWT 署名鍵 |

### シークレット（SSM Parameter Store）

| 変数名 | SSM パス |
|---|---|
| `DB_NAME` | `/{project}/{env}/db/name` |
| `DB_USER` | `/{project}/{env}/db/user` |
| `DB_PASSWORD` | `/{project}/{env}/db/password` |

### リソース

| 項目 | 値 |
|---|---|
| CPU | 256 units (0.25 vCPU) |
| Memory | 512 MB |
| Desired Count | 1 |
| Launch Type | FARGATE |

## CI/CD パイプライン

```
Developer → git push (main, backend/**) → GitHub Actions
  │
  ├─ 1. Checkout
  ├─ 2. AWS OIDC 認証 (IAM Role: github-ecs-deploy-role)
  ├─ 3. ECR Login
  ├─ 4. Docker Build (linux/amd64) & Push
  │     Tag: ${github.sha} + latest
  ├─ 5. タスク定義取得 & イメージ更新
  ├─ 6. ECS デプロイ (wait-for-service-stability)
  └─ 7. デプロイ検証
```

### トリガー条件

| 条件 | パス |
|---|---|
| Push to main | `backend/**` |
| Push to main | `.github/workflows/backend-deploy.yml` |
| 手動実行 | `workflow_dispatch` |

## Terraform モジュール構成

```
infra/terraform/
├── main.tf              # モジュールオーケストレーション
├── variables.tf         # ルート変数（23個）
├── outputs.tf           # ルート出力（33個）
├── backend.tf           # S3 + DynamoDB State 管理
├── providers.tf         # AWS プロバイダ設定
├── terraform.tfvars     # 変数値（テスト環境）
└── modules/
    ├── vpc/             # VPC, Subnets, IGW, NAT
    ├── security/        # SG, CloudWatch, SSM, IAM
    ├── rds/             # PostgreSQL RDS
    ├── ecr/             # Docker Registry
    ├── ecs/             # ALB, ECS Cluster/Service/Task
    └── github-oidc/     # GitHub Actions OIDC 連携
```

### モジュール依存関係

```
VPC ─┬─→ Security ──→ RDS
     │               ↗
     ├─→ ECR ──→ ECS
     │          ↗
     └──────────
                └──→ GitHub OIDC
```

## テスト環境の制約事項

| 項目 | テスト環境 | 本番で必要な変更 |
|---|---|---|
| NAT Gateway | 単一 | マルチ AZ |
| RDS Multi-AZ | 無効 | 有効化 |
| RDS 削除保護 | 無効 | 有効化 |
| RDS パブリックアクセス | 有効 | 無効化 |
| GitHub Terraform Role | AdministratorAccess | 最小権限ポリシー |
| CloudFront | 手動作成 | Terraform 管理 |
| 独自ドメイン | なし | 取得 + ACM 証明書 |
| ALB HTTP リスナー | forward (CloudFront が HTTPS 終端済み) | 独自ドメイン取得後に redirect 強制 |
