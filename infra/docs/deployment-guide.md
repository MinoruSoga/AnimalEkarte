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

## TablePlus から STG RDS に接続する

RDS は `publicly_accessible = false`（プライベートサブネット）のため、**ECS タスク経由の SSM ポートフォワード**が必要。

### 前提条件

- AWS CLI + `session-manager-plugin` インストール済み
- `session-manager-plugin` 未インストールの場合: `brew install --cask session-manager-plugin`

### Step 1: ECS タスク情報を取得

```bash
TASK_ARN=$(aws ecs list-tasks \
  --cluster animalekarte-stg-cluster \
  --service-name animalekarte-stg-service \
  --profile AnimalEkarte \
  --query 'taskArns[0]' --output text)

TASK_ID=$(echo "$TASK_ARN" | awk -F/ '{print $NF}')

RUNTIME_ID=$(aws ecs describe-tasks \
  --cluster animalekarte-stg-cluster \
  --tasks "$TASK_ARN" \
  --profile AnimalEkarte \
  --query 'tasks[0].containers[0].runtimeId' --output text)
```

### Step 2: SSM ポートフォワード開始

```bash
aws ssm start-session \
  --target "ecs:animalekarte-stg-cluster_${TASK_ID}_${RUNTIME_ID}" \
  --document-name AWS-StartPortForwardingSessionToRemoteHost \
  --parameters "{\"host\":[\"animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com\"],\"portNumber\":[\"5432\"],\"localPortNumber\":[\"15432\"]}" \
  --region us-east-1 \
  --profile AnimalEkarte
```

`Waiting for connections...` と表示されたらトンネル確立。このターミナルは開いたままにする。

### Step 3: TablePlus 接続設定

| 項目 | 値 |
|------|-----|
| Host | `127.0.0.1` |
| Port | `15432` |
| Database | `ekarte_db` |
| User | `ekarte_admin` |
| Password | 下記コマンドで取得 |
| SSL | require |

```bash
aws ssm get-parameter \
  --name "/animalekarte/test/db/password" \
  --with-decryption \
  --profile AnimalEkarte \
  --query 'Parameter.Value' --output text
```

> **注意**: 夜間停止スケジュール（22:00–8:00 JST）の時間帯は ECS・RDS が停止しているため接続不可。夜間に接続が必要な場合は下記「夜間に一時起動する」手順を参照。

---

## 夜間に STG RDS へ一時的にアクセスする

夜間停止スケジュール（22:00–8:00 JST）の時間帯に一時的に接続が必要な場合の手順。

### Step 1: RDS を起動

```bash
aws rds start-db-instance \
  --db-instance-identifier animalekarte-stg-db \
  --profile AnimalEkarte
```

### Step 2: RDS が起動するまで待つ（1〜2分）

```bash
aws rds wait db-instance-available \
  --db-instance-identifier animalekarte-stg-db \
  --profile AnimalEkarte
```

### Step 3: ECS を起動

```bash
aws ecs update-service \
  --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --desired-count 1 \
  --profile AnimalEkarte
```

ECS タスクが起動したら、上記「TablePlus から STG RDS に接続する」の手順で接続する。

### Step 4: 作業完了後に停止（必須）

```bash
aws ecs update-service \
  --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --desired-count 0 \
  --profile AnimalEkarte

aws rds stop-db-instance \
  --db-instance-identifier animalekarte-stg-db \
  --profile AnimalEkarte
```

---

## STG DB ダンプを取得する

### 前提条件

- `pg_dump` がローカルにインストール済み（未インストールなら `brew install postgresql`）
- 夜間の場合は上記「夜間に一時起動する」を先に実施

### Step 1: SSM ポートフォワード確立（別ターミナルで実行）

「TablePlus から STG RDS に接続する」の Step 1〜2 を実行し、`Waiting for connections...` の状態にする。

### Step 2: pg_dump 実行

```bash
PGPASSWORD=$(aws ssm get-parameter \
  --name "/animalekarte/test/db/password" \
  --with-decryption \
  --profile AnimalEkarte \
  --query 'Parameter.Value' --output text) \
pg_dump \
  --host 127.0.0.1 \
  --port 15432 \
  --username ekarte_admin \
  --dbname ekarte_db \
  --no-password \
  --format=custom \
  --file stg_dump_$(date +%Y%m%d_%H%M%S).dump
```

`--format=custom` は圧縮・並列リストア対応の推奨フォーマット。プレーンな SQL が必要な場合は `--format=plain` に変更する。

### リストア（ローカル DB に流す場合）

```bash
pg_restore \
  --host localhost \
  --port 5432 \
  --username postgres \
  --dbname <target_db> \
  --no-owner \
  stg_dump_YYYYMMDD_HHMMSS.dump
```

---

## DB マイグレーション

### TablePlus から実行

1. 上記手順で TablePlus を RDS に接続
2. `backend/migrations/001_init.sql` を実行（スキーマ作成 + マスタデータ + デモデータ投入）

### DB リセット

テスト環境ではリリース前 DB リセット運用:
```sql
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
-- 001_init.sql を再実行
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
