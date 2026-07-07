# ECS ロールバック専用ランブック（ユーザー承認必須）

> **目的**: Cloudflare 正系統（Workers/Containers + PlanetScale）から **旧 AWS ECS/RDS 経路へ
> 緊急ロールバック** する場合に必要な手順を記載する。
> **通常の STG 運用では実施不要**（`backend-deploy.yml` + `wrangler` が正系統）。
> **読者**: リポジトリ管理者（AWS/Cloudflare 管理権限保持者）。
> **タイミング**: ユーザーの明示承認を得てから実施。

---

## 0. 現状

**正系統**: Cloudflare Workers + Containers + PlanetScale Postgres
（`migration-cloudflare.md` Phase 5 完了・`backend-deploy.yml` が `staging` push で
`wrangler deploy` → `POST /_internal/migrate` を実行）。

**ロールバック経路**: `.github/workflows/backend-deploy-ecs.yml` は Phase 8（AWS 廃止）
完了まで `workflow_dispatch` のみで残置。通常運用では実行しない。

Cloudflare 正系統のシークレット供給は `infra/cloudflare/README.md` および
`backend/wrangler.jsonc` の `secrets.required` を参照。

---

## 1. 【DEPRECATED — ECS ロールバック専用】旧 H-5: SSM Parameter Store への新値登録

> ⚠️ **通常の STG 運用では実施不要**。以下は `backend-deploy-ecs.yml`（`workflow_dispatch`
> のみ）を使って AWS ECS/RDS 経路へロールバックする場合にのみ意味を持つ。

> 🚨 **既知のギャップ**: `backend-deploy-ecs.yml` の
> 「Parse .env.staging into environment / secrets」ステップ（`open('.env.staging')`）は
> **チェックアウト済みリポジトリ上の `.env.staging` を直接読む**実装。
> `.env.staging` は git untrack 済み（`6e34e684`）のため、**dispatch すると
> `FileNotFoundError` で失敗する**可能性がある。
> 緊急ロールバックでこの経路を使う場合は、事前に `.env.staging` を安全な経路から
> 一時復元する手順を整備すること。

**前提コード**: `.github/workflows/backend-deploy-ecs.yml` の `SSM_SECRET_PARAM_MAP` が
以下のパスを参照する設計:

```
DB_PASSWORD=/animalekarte/stg/db/password
JWT_SECRET=/animalekarte/stg/jwt_secret
INTEGRATION_ENCRYPTION_KEY=/animalekarte/stg/integration_encryption_key
```

`infra/terraform/modules/ecs/main.tf` の `aws_iam_role_policy.task_execution_ssm_secrets` が
ECS task execution role に `ssm:GetParameters` + `kms:Decrypt` を付与済み。

### 1.1 Terraform 適用（IAM ポリシー反映）

```bash
cd infra/terraform/environments/stg
terraform plan -out=tfplan
terraform apply tfplan
```

### 1.2 SSM パラメータ登録

```bash
aws ssm put-parameter --name /animalekarte/stg/db/password \
  --value "$NEW_DB_PASSWORD" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/jwt_secret \
  --value "$NEW_JWT_SECRET" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/integration_encryption_key \
  --value "$NEW_INTEGRATION_ENCRYPTION_KEY" --type SecureString --overwrite
```

### 1.3 登録確認

```bash
aws ssm get-parameters --names \
  /animalekarte/stg/db/password \
  /animalekarte/stg/jwt_secret \
  /animalekarte/stg/integration_encryption_key \
  --with-decryption --query 'Parameters[].Name'
```

### 1.4 デプロイして検証（ECS 経路）

```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-stg-api \
  --query 'taskDefinition.containerDefinitions[0].secrets'
```

---

## 2. 【DEPRECATED — ECS ロールバック専用】旧 M-4: STG `db_reset=true` デプロイ

> ⚠️ **Cloudflare 正系統では該当なし**。以下は `backend-deploy-ecs.yml` を使って
> 旧 AWS RDS 経路へロールバックし、かつ RDS 側 STG スキーマが古いままの場合にのみ意味を持つ。

**⚠️ 破壊的操作**: `db_reset=true` は STG DB を DROP & 再作成する。

### 2.1 事前チェック

```bash
docker compose down -v
docker compose up -d db
docker compose run --rm backend go run ./cmd/migrate
```

### 2.2 実行（ユーザー承認後・ECS ロールバック経路）

```bash
gh workflow run backend-deploy-ecs.yml --ref staging -f db_reset=true
```

### 2.3 監視

```bash
gh run watch --exit-status $(gh run list --workflow=backend-deploy-ecs.yml --branch=staging --limit=1 --json databaseId -q '.[0].databaseId')
```

### 2.4 事後確認

- [ ] `GET /health` → `200 OK`
- [ ] [CRUD スモークテスト](../CRUD-SMOKE-TEST.md) を実施

---

## 3. 【参考】旧 M-11: GitHub Secrets 登録（`CI_TEST_EMAIL` / `CI_TEST_PASSWORD`）

> `performance-tests.yml` は未登録でもフォールバック付きで fail-fast しない。
> `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD`（`infra/cloudflare/README.md`）との統合は
> 別途 PO/管理者判断。

```bash
gh secret set CI_TEST_EMAIL --body "<STG_TEST_ACCOUNT_EMAIL>"
gh secret set CI_TEST_PASSWORD --body "<STG_TEST_ACCOUNT_PASSWORD>"
```
