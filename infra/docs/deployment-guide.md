# AnimalEkarte - デプロイガイド

## 前提条件

| 項目 | 値 |
|---|---|
| AWS Profile | `AnimalEkarte` |
| AWS Region | `us-east-1` |
| AWS Account ID | `698109622668` |
| GitHub Repository | `MinoruSoga/AnimalEkarte` |
| Terraform Version | >= 1.5 |

## 環境情報

### テスト環境エンドポイント

| サービス | URL |
|---|---|
| Frontend | `https://stg.noah-karte.com` |
| Backend API | `https://api.stg.noah-karte.com/api` |
| ALB (直接) | `http://animalekarte-stg-alb-1915768826.us-east-1.elb.amazonaws.com` |
| RDS | `animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com:5432` |

### AWS リソース ID

| リソース | ID / ARN |
|---|---|
| CloudFront Distribution | `ERCVR5P0IAJKS` (`dcqico6azu5w2.cloudfront.net`) |
| ECS Cluster | `animalekarte-stg-cluster` |
| ECS Service | `animalekarte-stg-service` |
| Task Definition Family | `animalekarte-stg-api` |
| ECR Repository | `animalekarte-api` |
| ALB SG | `sg-047ff4f8c3ab99411` |
| ECS SG | `sg-0934ac397301ec633` |
| RDS SG | `sg-053d44ac9fab4e71a` |
| IAM OIDC Deploy Role | `animalekarte-stg-github-ecs-deploy-role` |
| ALB SG | `sg-090b034e4a30b5ca7` |
| ECS SG | `sg-0aa38e88ba0e4876c` |
| RDS SG | `sg-09026c201ac735d7e` |

### DB 接続情報

| 項目 | 値 |
|---|---|
| Host | `animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com` |
| Port | `5432` |
| Database | `ekarte_db` |
| Username | `ekarte_admin` |
| SSL Mode | `require` |

## デプロイ方法

### 1. バックエンド（自動）

`backend/**` を変更して `staging` にプッシュすると GitHub Actions が自動デプロイ。

```bash
git push origin staging
```

ワークフロー確認:
```bash
gh run list --workflow backend-deploy.yml --limit 1 --json status,conclusion | jq '.[0]'
```

### 2. バックエンド（手動トリガー）

```bash
gh workflow run backend-deploy.yml
```

### 3. フロントエンド

Vercel 環境変数を更新した場合、再デプロイが必要:
```bash
vercel deploy --prod
```

### 4. インフラ変更（Terraform）

```bash
cd infra/terraform
terraform plan -out=tfplan
terraform apply tfplan
```

## Vercel 環境変数

| 変数名 | 値 | 環境 |
|---|---|---|
| `VITE_API_URL` | `https://dcqico6azu5w2.cloudfront.net/api` | Production |

### 更新手順

```bash
vercel env rm VITE_API_URL production -y
echo "https://dcqico6azu5w2.cloudfront.net/api" | vercel env add VITE_API_URL production
vercel deploy --prod
```

## ECS タスク定義の環境変数

GitHub Actions デプロイでは、既存タスク定義のイメージタグのみ更新される。
環境変数の追加・変更は以下の手順:

### 1. 現在のタスク定義を取得

```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-stg-api \
  --profile AnimalEkarte \
  --query 'taskDefinition' | jq 'del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy)' > /tmp/task-def.json
```

### 2. 環境変数を編集

```bash
# jq で環境変数を追加
jq '.containerDefinitions[0].environment += [{"name": "KEY", "value": "VALUE"}]' /tmp/task-def.json > /tmp/task-def-updated.json
```

### 3. 新リビジョン登録 & サービス更新

```bash
aws ecs register-task-definition \
  --cli-input-json file:///tmp/task-def-updated.json \
  --profile AnimalEkarte

aws ecs update-service \
  --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --task-definition animalekarte-stg-api \
  --force-new-deployment \
  --profile AnimalEkarte
```

## RDS セキュリティグループ管理

### 開発者 IP の追加

```bash
aws ec2 authorize-security-group-ingress \
  --group-id sg-053d44ac9fab4e71a \
  --protocol tcp --port 5432 \
  --cidr <YOUR_IP>/32 \
  --profile AnimalEkarte
```

### 開発者 IP の削除

```bash
aws ec2 revoke-security-group-ingress \
  --group-id sg-053d44ac9fab4e71a \
  --protocol tcp --port 5432 \
  --cidr <YOUR_IP>/32 \
  --profile AnimalEkarte
```

## DB マイグレーション

### TablePlus から実行

1. TablePlus で RDS に接続
2. `backend/migrations/001_init.sql` を実行（テーブル作成）
3. `backend/migrations/002_seed_master.sql` を実行（マスタデータ投入）

### DB リセット

テスト環境ではリリース前 DB リセット運用:
```sql
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
-- 001_init.sql を再実行
-- 002_seed_master.sql を再実行
```

## ログ確認

### ECS タスクログ

```bash
# 実行中タスク ID 取得
TASK_ID=$(aws ecs list-tasks \
  --cluster animalekarte-stg-cluster \
  --service-name animalekarte-stg-service \
  --profile AnimalEkarte \
  --query 'taskArns[0]' --output text | awk -F/ '{print $NF}')

# ログ確認
aws logs get-log-events \
  --log-group-name /ecs/animalekarte-stg \
  --log-stream-name "api/api/$TASK_ID" \
  --profile AnimalEkarte \
  --query 'events[-20:].message' | jq -r '.[]'
```

### ECS Exec（コンテナ直接アクセス）

```bash
aws ecs execute-command \
  --cluster animalekarte-stg-cluster \
  --task $TASK_ID \
  --container api \
  --interactive \
  --command "/bin/sh" \
  --profile AnimalEkarte
```

## トラブルシューティング

### デプロイ後に古いイメージが使われる

GitHub Actions は `${github.sha}` タグで新イメージをプッシュし、タスク定義を更新する。
手動ビルドの `latest` タグは ECS にキャッシュされる場合があるため、
**常に GitHub Actions 経由でデプロイすること。**

### CORS エラー

`CORS_ALLOWED_ORIGIN` 環境変数にフロントエンドのオリジンが含まれているか確認:
```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-stg-api \
  --profile AnimalEkarte \
  --query 'taskDefinition.containerDefinitions[0].environment' | \
  jq '.[] | select(.name == "CORS_ALLOWED_ORIGIN")'
```

### Cookie が保存されない

- `GIN_MODE=release` が設定されているか確認（`release` のとき `Secure=true`, `SameSite=None` が有効になる）
- ブラウザの開発者ツールで `Set-Cookie` ヘッダーに `SameSite=None; Secure` が含まれるか確認
- `CORS_ALLOWED_ORIGIN` にフロントエンドオリジンが含まれ、かつ `Access-Control-Allow-Credentials: true` がレスポンスヘッダに含まれるか確認

### GitHub Actions OIDC 認証エラー

IAM Role の trust policy でリポジトリ名が一致しているか確認:
```bash
aws iam get-role \
  --role-name animalekarte-stg-github-ecs-deploy-role \
  --profile AnimalEkarte \
  --query 'Role.AssumeRolePolicyDocument' | jq
```

正しい値: `repo:MinoruSoga/AnimalEkarte:ref:refs/heads/staging`
