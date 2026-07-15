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

**ECS 経路は現状実行不能（二重ブロック）**:
1. `.env.staging` は git untrack 済みのため、ECS workflow の parse ステップが失敗しうる
2. SSM Parameter Store へのシークレット登録は未実施（下記 §1 はロールバック時のみ）


Cloudflare 正系統のシークレット供給は `infra/cloudflare/README.md` および
`backend/wrangler.jsonc` の `secrets.required` を参照。

---

## 0.5 【正系統】露出クレデンシャルのローテーション（SEC-SECRETS-5 / #89/#97）

> 🚨 **ユーザー所有・credential-impacting**。エージェントは実行しない。
> PUBLIC リポジトリ履歴および過去の seed/Issue 露出に対する正攻法は **ローテーション**（filter-repo 禁止）。

対象 4 系統（完了まで Issue #89/#97 はクローズしない）:

| # | 系統 | 手順（概要） | 投入先 |
|---|------|--------------|--------|
| 1 | PlanetScale DB | `pscale role reset-default`（またはコンソールでパスワード再発行） | `wrangler secret put DB_PASSWORD`（および接続 URL 系） |
| 2 | Cloudflare API / Worker secrets | トークン再発行 + `wrangler secret put` で必須キー再投入 | Cloudflare Secrets + GitHub `CLOUDFLARE_API_TOKEN` |
| 3 | LINE channel secret / access token | LINE Developers Console で再発行 | アプリ UI（Lステップ設定 / LINE 予約設定）から保存（DB 暗号化）。seed には実値を戻さない |
| 4 | JWT / INTEGRATION_ENCRYPTION_KEY 等 | 新規乱数生成 | `wrangler secret put`（`backend/wrangler.jsonc` の `secrets.required`） |

検証（ローテーション後・ユーザー実施）:

```bash
# STG health（正系統）
curl -sS -o /dev/null -w '%{http_code}\n' https://api.stg.noah-karte.com/health
# 旧値でのアクセスが拒否されること（各コンソール / wrangler の確認 UI）
```

ローテーション完了後のみ: Issue #97 本文の実値マスク（`gh issue edit` — ユーザー実施）。

P5-2 GitHub Secrets（`CLOUDFLARE_API_TOKEN`, `MIGRATE_RUN_SECRET`, `STG_DEMO_EMAIL`, `STG_DEMO_PASSWORD`）
の登録手順は `infra/cloudflare/README.md` 「CI デプロイ (Phase 5 / P5-2)」を正とする。


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

## 3. 【参考】performance-tests 認証情報（#109 / STG_DEMO_* 統合）

> **現状**: `.github/workflows/performance-tests.yml` は `CI_TEST_EMAIL` / `CI_TEST_PASSWORD`
> 未登録時に `admin@example.com` / `test` へフォールバックする（fail-fast しない）。
> **確定方針**: `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD`（`infra/cloudflare/README.md` P5-2）へ一本化。
>
> **エージェント作業境界**: GitHub Secrets 登録（ユーザー）が先行。登録前にフォールバックを
> 撤去すると scheduled 実行が壊れるため、フォールバック撤去は **USER 登録完了後の Phase C**。

```bash
# ユーザー実施: P5-2 と合わせて登録
gh secret set STG_DEMO_EMAIL --body "<STG_DEMO_ACCOUNT_EMAIL>"
gh secret set STG_DEMO_PASSWORD --body "<STG_DEMO_ACCOUNT_PASSWORD>"
# 任意: 旧名の互換期間が必要なら一時併存可。最終的に CI_TEST_* 参照ゼロがクローズ条件
```
