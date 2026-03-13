# デプロイチェックリスト

---

## デプロイ前確認

```bash
# Lint
docker compose exec frontend npm run lint
docker compose exec backend golangci-lint run ./...

# テスト
docker compose exec frontend npm run test:run
docker compose exec backend go test ./... -v

# ビルド確認
docker compose exec frontend npm run build
docker compose exec backend go build ./cmd/api

# TypeScript エラーなし
docker compose exec frontend npm run type-check

# console.log が残っていないか
grep -r "console\.log\|console\.error" frontend/src --include="*.ts" --include="*.tsx"

# シークレットが含まれていないか
git diff main..HEAD | grep -iE "password|secret|token|key"

```

---

## デプロイ実行

```bash
# 1. main ブランチを最新化
git checkout main && git pull origin main

# 2. feature ブランチをマージ
git merge --no-ff feature/xxx
git push origin main

# 3. 自動デプロイを監視
gh run list --workflow=backend-deploy.yml --limit 1
gh run view <RUN_ID> --log
```

---

## デプロイ後確認

```bash
export AWS_PROFILE=AnimalEkarte

# Backend ヘルスチェック
curl http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health | jq .

# ECS 状態
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].{status,runningCount,desiredCount}'

# エラーログ確認
aws logs filter-log-events \
  --log-group-name /ecs/animalekarte-test \
  --filter-pattern "ERROR" \
  --region us-east-1
```

**ロールバック基準:**
- API エラーレート > 1%
- ページロード > 5秒
- データベースエラー
- セキュリティ脆弱性

→ ロールバック手順: [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md#ロールバック)

---

## DB 作り直し手順（RDS）

RDS は Private Subnet にあるため直接接続不可。**ECS Exec** でコンテナ経由で接続する。

### 前提: ECS Exec を有効化（初回のみ）

```bash
export AWS_PROFILE=AnimalEkarte

aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --enable-execute-command \
  --region us-east-1

# タスク再起動して設定を反映
aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --force-new-deployment \
  --region us-east-1

aws ecs wait services-stable \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1
```

### ステップ 1: DB 認証情報を取得

認証情報は SSM Parameter Store に格納されている。

```bash
aws ssm get-parameter --name /animalekarte/test/db/user --region us-east-1 --query 'Parameter.Value' --output text
aws ssm get-parameter --name /animalekarte/test/db/password --with-decryption --region us-east-1 --query 'Parameter.Value' --output text
aws ssm get-parameter --name /animalekarte/test/db/name --region us-east-1 --query 'Parameter.Value' --output text
```

### ステップ 2: ECS Exec でコンテナに入る

```bash
TASK_ID=$(aws ecs list-tasks \
  --cluster animalekarte-test-cluster \
  --service-name animalekarte-test-service \
  --region us-east-1 \
  --query 'taskArns[0]' \
  --output text | awk -F'/' '{print $NF}')

aws ecs execute-command \
  --cluster animalekarte-test-cluster \
  --task $TASK_ID \
  --container api \
  --interactive \
  --command "/bin/sh" \
  --region us-east-1
```

### ステップ 3: コンテナ内から RDS を接続してDBリセット

```bash
# コンテナ内で実行（DB_USER は SSM から取得した値を使用）
psql "host=animalekarte-test-db.cqbe28s44fta.us-east-1.rds.amazonaws.com \
      port=5432 user=$DB_USER dbname=postgres sslmode=require"
```

```sql
-- 既存の接続を切断
SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE datname = 'ekarte_db' AND pid <> pg_backend_pid();

-- DB 削除 & 再作成
DROP DATABASE IF EXISTS ekarte_db;
CREATE DATABASE ekarte_db OWNER ekarte_user;

\c ekarte_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
\q
```

### ステップ 4: マイグレーション再実行（自動）

ECS タスクを再デプロイすると GORM AutoMigrate が走り全テーブルが再作成される。

```bash
aws ecs update-service \
  --cluster animalekarte-test-cluster \
  --service animalekarte-test-service \
  --force-new-deployment \
  --region us-east-1

# マイグレーションログを確認
aws logs tail /ecs/animalekarte-test --follow --region us-east-1
```

### 注意

- **全データが消える。** 実行前に必ず RDS スナップショットを取得すること。
  ```bash
  aws rds create-db-snapshot \
    --db-instance-identifier animalekarte-test-db \
    --db-snapshot-identifier ekarte-backup-$(date +%Y%m%d) \
    --region us-east-1
  ```
- ECS Exec が使えない場合（`TargetNotConnectedException`）→ 前提の有効化手順を再実行
