# セキュリティ監査チェックリスト（AnimalEkarte Stg環境）

**最終更新:** 2026-03-31

## インフラセキュリティ

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| RDS暗号化 | ✅ | StorageEncrypted: True |
| RDSパブリックアクセス | ✅ | PubliclyAccessible: False |
| VPC分離 | ✅ | Public/Private Subnet分離 |
| Security Group | ✅ | 最小権限（CloudFront→ALB→ECS→RDS） |
| NAT Gateway | ✅ | Private SubnetからInternet経由 |
| CloudWatch Logs | ✅ | 30日保持設定 |
| HTTPS終端 | ✅ | CloudFront (`api.stg.noah-karte.com`) で TLS 終端 |

## IAM

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| ECS Task Execution Role | ✅ | ECR pull, SSM読み取り |
| ECS Task Role | ✅ | CloudWatch Logs書き込み |
| GitHub OIDC | ✅ | Terraform/ECS Deploy Role分離 |
| IAM Role最小権限 | ⚠️ | Terraform RoleはAdministratorAccess（Stg環境のみ許容） |

## アプリケーション

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| CORS設定 | ✅ | `CORS_ALLOWED_ORIGIN` 環境変数で Vercel + CloudFront ドメインのみ許可（PR #10 修正済み） |
| Cookie認証 | ✅ | JWT を httpOnly Cookie で管理（access: 15分 / refresh: 7日）、SameSite=None（本番） |
| DB接続SSL | ✅ | sslmode=require |
| 環境変数 | ✅ | SSM Parameter Store使用（DB認証情報）+ ECS 環境変数（CORS等） |
| JWT秘密鍵 | ✅ | `JWT_SECRET` は本番環境で必須チェックあり（config.Validate()） |

## 監査・ログ

| 項目 | 状態 | 確認内容 |
|------|------|---------|
| CloudWatch Logs | ✅ | /ecs/animalekarte-stg、30日保持 |
| RDSバックアップ | ✅ | 1日保持（Stg環境で妥当） |
| CloudTrail | ❌ | 未設定（追加推奨） |
| VPC Flow Logs | ❌ | 未設定（追加推奨） |

## コスト（月額見積もり）

| サービス | 月額 (USD) |
|---------|-----------|
| NAT Gateway | $32 |
| ALB | $21 |
| RDS db.t4g.micro | $14.5 |
| ECS Fargate | $6.65 |
| CloudFront | ~$1（ステージング環境トラフィック量） |
| その他 (ECR, CloudWatch, SSM) | $4.25 |
| **合計** | **約 $79/月** |

## 改善推奨事項（Production環境）

1. **CloudTrail有効化**: 全API呼び出しの監査ログ
2. **VPC Flow Logs**: ネットワークトラフィック監視
3. ~~**CORS制限**: Vercelドメインのみ許可~~ → ✅ **実施済み (PR #10)**
4. **AWS Budgets**: コストアラート設定（$100/月アラート）
5. **Multi-AZ**: RDS/ECS高可用性化
6. ~~**ACM証明書**: ALB HTTPS化~~ → ✅ **CloudFront で HTTPS 終端済み**
7. **WAF導入**: CloudFront + AWS WAF でSQLインジェクション・DDoS対策
8. **MFA必須**: AWSコンソールログイン
9. **ALBアクセス制限**: CloudFront IP レンジのみ受け付けるよう SG を絞る（現在は 0.0.0.0/0:80）
