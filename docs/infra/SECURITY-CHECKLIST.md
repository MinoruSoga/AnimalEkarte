# セキュリティ監査チェックリスト（AnimalEkarte Test環境）

## インフラセキュリティ

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| RDS暗号化 | ✅ | StorageEncrypted: True |
| RDSパブリックアクセス | ✅ | PubliclyAccessible: False |
| VPC分離 | ✅ | Public/Private Subnet分離 |
| Security Group | ✅ | 最小権限（ALB→ECS→RDS） |
| NAT Gateway | ✅ | Private SubnetからInternet経由 |
| CloudWatch Logs | ✅ | 30日保持設定 |

## IAM

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| ECS Task Execution Role | ✅ | ECR pull, SSM読み取り |
| ECS Task Role | ✅ | CloudWatch Logs書き込み |
| GitHub OIDC | ✅ | Terraform/ECS Deploy Role分離 |
| IAM Role最小権限 | ⚠️ | Terraform RoleはAdministratorAccess（Test環境のみ許容） |

## アプリケーション

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| CORS設定 | ⚠️ | Access-Control-Allow-Origin: * （本番では制限推奨） |
| DB接続SSL | ✅ | sslmode=require |
| 環境変数 | ✅ | SSM Parameter Store使用 |
| PostgreSQL拡張 | ✅ | uuid-ossp有効化 |

## 監査・ログ

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| CloudWatch Logs | ✅ | /ecs/animalekarte-test、30日保持 |
| RDSバックアップ | ✅ | 1日保持（Test環境で妥当） |
| CloudTrail | ❌ | 未設定（追加推奨） |
| VPC Flow Logs | ❌ | 未設定（追加推奨） |

## コスト

| 項目 | 確認内容 |
|------|---------|
| 2026-02-14 | $0.51 |
| 2026-02-15 | $0.06 |
| 月間見積 | 約$30-50（NAT Gateway, ECS Fargate, RDS, ALB） |

## 改善推奨事項（Production環境）

1. **CloudTrail有効化**: 全API呼び出しの監査ログ
2. **VPC Flow Logs**: ネットワークトラフィック監視
3. **CORS制限**: Vercelドメインのみ許可
4. **AWS Budgets**: コストアラート設定
5. **Multi-AZ**: RDS/ECS高可用性化
6. **ACM証明書**: ALB HTTPS化
7. **WAF導入**: SQLインジェクション対策
8. **MFA必須**: AWSコンソールログイン
