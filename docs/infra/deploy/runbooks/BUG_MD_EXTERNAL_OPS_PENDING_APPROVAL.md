# bug.md 残タスク: 外部操作ランブック（ユーザー承認必須）

> **目的**: `bug.md` の残項目（現行: C-1 のみ。#189 は PO 判断待ちで別管理）のうち、
> コード変更では完結せず Cloudflare/PlanetScale/GitHub への実操作
> （シークレットローテーション・`.env.staging` 追跡解除）が必要なタスクを、
> **正確な現状**（`migration-cloudflare.md` Phase 5 完了・STG 正系統は Cloudflare Workers +
> Containers + PlanetScale）に基づいて実行可能な手順に落とし込む。
> **読者**: リポジトリ管理者（Cloudflare/PlanetScale/GitHub 管理権限保持者）。
> **タイミング**: 各セクションの「前提コード」がマージされた後、ユーザーの明示承認を得てから実施。

> **重要**: 本ランブックのいずれの手順も、この文書を生成した AI エージェントによって**実行されていない**。
> すべて「次にユーザーが実行するコマンド」として記載している。

> **2026-07-06 改訂**: STG 正系統が AWS ECS/RDS から Cloudflare Workers/Containers + PlanetScale
> へ移行完了（`migration-cloudflare.md` Phase 5）。本ランブックの §2（旧 H-5 / SSM）・§4（旧 M-4 /
> `db_reset=true`）は **ECS ロールバック専用（`backend-deploy-ecs.yml`, `workflow_dispatch` のみ）**
> に格下げした。通常の STG 運用では実施不要。§5（旧 M-11）は `bug.md` 現行 backlog から除外済み
> （詳細は §5 冒頭注記）。C-1 のみが現行の対応対象。

---

## 0. 現状・実施順序

**現在の正系統**: Cloudflare Workers + Containers + PlanetScale Postgres
（`migration-cloudflare.md` Phase 5 実装済み・`.github/workflows/backend-deploy.yml` が
`staging` push で `wrangler deploy` → `POST /_internal/migrate` を実行）。
シークレットは **`wrangler secret put`**（Workers/Containers 側: `JWT_SECRET` /
`INTEGRATION_ENCRYPTION_KEY` / `DB_USER` / `DB_PASSWORD` / LINE 系等、
`backend/wrangler.jsonc` の `secrets.required` 参照）と
**`pscale role reset-default --force`**（PlanetScale DB クレデンシャル発行）で供給される。
旧 AWS ECS/RDS/SSM 経路（`.github/workflows/backend-deploy-ecs.yml`）は Phase 8（AWS 廃止）
完了までのロールバック専用であり、`workflow_dispatch` でのみ起動可能・通常運用では実行しない。

**現行の対応対象（`bug.md` 残項目）**:

```
1. C-1-1: シークレットローテーション（PlanetScale DB / JWT_SECRET /
          INTEGRATION_ENCRYPTION_KEY / LINE）→ §1
2. C-1-2: .env.staging の git 追跡解除 + Issue #97 本文の実値削除 → §3
```

1→2 の順序を守ること。§1 でロールする前に `.env.staging` を untrack すると、
ローカル開発者が新しい値を受け取れなくなる（`.env.staging` は `.gitignore` 済みで
untrack 後もローカルには残るが、リポジトリ経由での値共有手段がなくなるため、
ローテーション後の新値は別途安全な経路（1Password 等）で共有すること）。

**DEPRECATED（ECS ロールバック時のみ・通常は実施不要）**:

| セクション | 旧名 | 現状 |
|---|---|---|
| §2 | H-5（SSM Parameter Store 登録） | ECS/RDS 経路でのみ必要。Cloudflare 正系統では `wrangler secret put` が代替 |
| §4 | M-4（STG `db_reset=true`） | PlanetScale は Phase 3/4 で `001_init.sql` から新規作成済みのため、Cloudflare 正系統では該当なし。ECS ロールバック時、旧 RDS 側 STG を使う場合のみ意味を持つ |
| §5 | M-11（`CI_TEST_EMAIL`/`CI_TEST_PASSWORD` Secrets 登録） | `bug.md` 現行 backlog からは除外済み。`performance-tests.yml` はデフォルト値へのフォールバックがあり fail-fast ではない。今後 `STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD`（`infra/cloudflare/README.md` 参照）へ統合するかは別途 PO/管理者判断 |

---

## 1. C-1: シークレットローテーション（Cloudflare / PlanetScale 経路）

**前提**: 現在 `.env.staging` は git 追跡されたままで、`DB_PASSWORD` / `JWT_SECRET` /
`INTEGRATION_ENCRYPTION_KEY` を平文で含む（`git show HEAD:.env.staging` でキー名のみ確認済み。
LINE 系キーは含まれていない）。Issue #97 本文にも実値記載の指摘がある。ローテーション自体は
`migration-cloudflare.md` Phase 4/5 で確立済みの運用パターン（`pscale role reset-default` +
`wrangler secret put`）を踏襲する。C-1 の確認された漏洩スコープは **DB_PASSWORD /
JWT_SECRET / INTEGRATION_ENCRYPTION_KEY の3点のみ**（§1.1〜§1.4）。LINE 系は対象外（§1.5 参照）。

### 1.1 PlanetScale DB パスワード変更

```bash
# infra/cloudflare/.env.staging（gitignore済み）を source して認証情報を用意
set -a && source infra/cloudflare/.env.staging && set +a
pscale auth check

# 新しいデフォルトロールのクレデンシャルを強制発行（古い値は即時失効）
pscale role reset-default --org noah-animalekarte --database animalekarte-stg --branch main --force
# 出力される username/password/host を控える（チャット・git・ログに残さない）
```

### 1.2 Workers/Containers 側へ反映（`wrangler secret put`）

```bash
cd backend
pnpm exec wrangler secret put DB_USER --name animalekarte-stg-api
pnpm exec wrangler secret put DB_PASSWORD --name animalekarte-stg-api
# DB_HOST は通常 pscale reset では変わらないが、変更があれば同様に put する
```

### 1.3 JWT_SECRET 変更

```bash
NEW_JWT_SECRET="$(openssl rand -base64 48)"
pnpm exec wrangler secret put JWT_SECRET --name animalekarte-stg-api
# プロンプトで $NEW_JWT_SECRET を貼り付け
```

- **影響**: 変更後デプロイ時点で全ユーザーの既存セッションが無効化される。事前に STG 利用者へ周知すること。

### 1.4 INTEGRATION_ENCRYPTION_KEY 変更

- **影響**: `clinic_integrations` の暗号化済み値（Lステップ認証情報等）が新鍵では復号不能になる。
- **STG の場合**: 再入力運用で問題なければ、単純に新鍵へ切替えて既存 `clinic_integrations` 行を
  手動で再設定してもらう方が、再暗号化スクリプトを書くよりリスクが低い（データ量が少ない前提）。
- 本番導入時は「旧鍵で全件復号 → 新鍵で再暗号化」の一時スクリプトが必須（STG では省略可）。

```bash
NEW_INTEGRATION_ENCRYPTION_KEY="$(openssl rand -base64 32)"
pnpm exec wrangler secret put INTEGRATION_ENCRYPTION_KEY --name animalekarte-stg-api
```

### 1.5 【C-1 スコープ外・任意】LINE channel secret / access token

> ⚠️ **`.env.staging` には LINE 系キーは含まれていない**（`git show HEAD:.env.staging` で
> キー名を確認済み）。C-1 の確認された平文露出には該当しないため、以下は **C-1 の PASS 条件には
> 含まれない**任意の予防的ローテーションである。実施するか否かは別途判断でよい。
> なお `backend/wrangler.jsonc`（L12-13）の設計では LINE_CHANNEL_ACCESS_TOKEN /
> LINE_CHANNEL_SECRET はクリニックごとに DB 管理する方針で、現行の Go アプリには
> Worker env からこれらを読む経路がない（将来のグローバル fallback 用に予約されているのみ）。
> そのため `wrangler secret put` で以下を投入しても、現行コードパスでは参照されず
> §1.6 の「Lステップ連携確認」で検証可能な変化は生じない。

- 実施する場合: LINE Developers コンソール（https://developers.line.biz/console/）で
  対象チャネルを開き、Channel secret / Channel access token を再発行する。

```bash
pnpm exec wrangler secret put LINE_CHANNEL_SECRET --name animalekarte-stg-api
pnpm exec wrangler secret put LINE_CHANNEL_ACCESS_TOKEN --name animalekarte-stg-api
```

### 1.6 デプロイして検証

```bash
gh workflow run backend-deploy.yml --ref staging
gh run list --workflow=backend-deploy.yml --branch=staging --limit 1
```

- `GET /health` が `200 {"status":"ok"}` を返すこと。
- アプリの起動・DB接続・ログイン（JWT発行）が正常であることを確認する
  （§1.5 を実施した場合でも、現行コードでは検証可能な差分は生じない点に注意）。

---

## 2. 【DEPRECATED — ECS ロールバック専用】旧 H-5: SSM Parameter Store への新値登録

> ⚠️ **通常の STG 運用では実施不要**。以下は `backend-deploy-ecs.yml`（`workflow_dispatch`
> のみ・Phase 8 まで残置）を使って AWS ECS/RDS 経路へロールバックする場合にのみ意味を持つ。
> Cloudflare 正系統のシークレット管理は §1 の `wrangler secret put` / `pscale role reset-default`
> を参照。

> 🚨 **既知のギャップ（§3 untrack 実施後）**: `backend-deploy-ecs.yml` の
> 「Parse .env.staging into environment / secrets」ステップ（`open('.env.staging')`）は
> **チェックアウト済みリポジトリ上の `.env.staging` を直接読む**実装になっている。
> `.env.staging` を git untrack すると（§3）、`actions/checkout` はこのファイルを
> 復元しない（gitignore 済みかつ未コミットのため）ため、**この状態で
> `backend-deploy-ecs.yml` を dispatch すると `FileNotFoundError` で失敗する**。
> 緊急ロールバックで実際にこの経路を使う場合は、事前に `workflow_dispatch` 実行環境へ
> `.env.staging` を安全な経路（例: リポジトリ管理者のローカル控え・1Password 等）から
> 一時的に復元する手順を別途整備すること（本ランブックの対応範囲外・別 Issue 化を推奨）。
> untrack 自体は C-1 の解消に必要なため停止しないが、ロールバック手順との依存関係が
> 生じている点を記録として残す。

**前提コード**: `.github/workflows/backend-deploy-ecs.yml` の `SSM_SECRET_PARAM_MAP` が
以下のパスを参照する設計になっている（値は未登録）:

```
DB_PASSWORD=/animalekarte/stg/db/password
JWT_SECRET=/animalekarte/stg/jwt_secret
INTEGRATION_ENCRYPTION_KEY=/animalekarte/stg/integration_encryption_key
```

`infra/terraform/modules/ecs/main.tf` の `aws_iam_role_policy.task_execution_ssm_secrets` が
ECS task execution role に `ssm:GetParameters`（対象パス配下のみ）+ `kms:Decrypt`
（`kms:ViaService=ssm.<region>.amazonaws.com` 条件付き）を付与済み。

### 2.1 Terraform 適用（IAM ポリシー反映）

```bash
cd infra/terraform/environments/stg   # 実際のパスは infra/terraform 配下を確認
terraform plan -out=tfplan
# plan の差分を目視レビュー（ssm/kms 権限追加のみであることを確認）
terraform apply tfplan
```

### 2.2 SSM パラメータ登録

```bash
aws ssm put-parameter --name /animalekarte/stg/db/password \
  --value "$NEW_DB_PASSWORD" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/jwt_secret \
  --value "$NEW_JWT_SECRET" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/integration_encryption_key \
  --value "$NEW_INTEGRATION_ENCRYPTION_KEY" --type SecureString --overwrite
```

### 2.3 登録確認

```bash
aws ssm get-parameters --names \
  /animalekarte/stg/db/password \
  /animalekarte/stg/jwt_secret \
  /animalekarte/stg/integration_encryption_key \
  --with-decryption --query 'Parameters[].Name'
# 3件とも返ること（値自体はここでは出力しない運用にする）
```

### 2.4 デプロイして検証（ECS 経路）

```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-stg-api \
  --query 'taskDefinition.containerDefinitions[0].secrets'
# JWT_SECRET / DB_PASSWORD / INTEGRATION_ENCRYPTION_KEY が
# valueFrom（SSM ARN）で列挙され、containerDefinitions[0].environment 側に
# 平文で出てこないことを確認する
```

---

## 3. C-1 残: `.env.staging` 追跡解除 + Issue #97 実値削除

**前提**: §1 のローテーション・デプロイ検証が green であること（Cloudflare 側で
`wrangler secret` から正しく値が解決され、`.env.staging` の平文値に依存しなくなっていることの確認）。

```bash
git rm --cached .env.staging
git commit -m "chore(security): untrack .env.staging now that secrets resolve via wrangler secret"
```

- `.env.staging` は `.gitignore` に既に登録済み（`.env.staging` エントリ）のため、
  上記コマンドは追加の `.gitignore` 追記なしで untrack のみ実行する。
- ローカル開発者は `.env.staging` を手元に残したまま（`.gitignore` 済みなので追跡されない）。
- Issue #97 本文の実値を削除する:

```bash
gh issue edit 97 --body "$(cat <<'EOF'
（旧本文に含まれていた STG の JWT_SECRET / DB_PASSWORD / DB_HOST の実値は
ローテーション完了（YYYY-MM-DD）に伴い削除。詳細は内部ランブック参照。）
EOF
)"
```

- 実施後、`git ls-files | grep -E '\.env'` で `.env.staging` が一覧から消えていることを確認する。

---

## 4. 【DEPRECATED — ECS ロールバック専用】旧 M-4: STG `db_reset=true` デプロイ

> ⚠️ **Cloudflare 正系統では該当なし**。STG の PlanetScale データベースは
> `migration-cloudflare.md` Phase 3/4 で `001_init.sql`（Checkup 系統合済みスキーマ）から
> 新規作成されている。以下は `backend-deploy-ecs.yml` を使って旧 AWS RDS 経路へ
> ロールバックし、かつその RDS 側 STG のスキーマが古いままの場合にのみ意味を持つ。

**前提（現状の正確な把握）**: `backend/migrations/` は既に統合済みで
`001_init.sql`（DDL 専用）1ファイルのみが存在する。seed データは
`backend/migrations/seeds/{002_master,003_demo,004_staging}/` の CSV バンドルのみ
（`cmd/migrate` が `001_init.sql` 適用後に固定順でロードする）。

**⚠️ 破壊的操作**: `db_reset=true` は STG DB を DROP & 再作成する。STG 上の全データが失われる。
実行前に必ずユーザーへ「破棄可否」を確認すること（`.claude/skills/stg-release-readiness` 準拠）。

### 4.1 事前チェック

```bash
docker compose down -v
docker compose up -d db
docker compose run --rm backend go run ./cmd/migrate
```

### 4.2 実行（ユーザー承認後・ECS ロールバック経路）

```bash
gh workflow run backend-deploy-ecs.yml --ref staging -f db_reset=true
```

### 4.3 監視

```bash
gh run watch --exit-status $(gh run list --workflow=backend-deploy-ecs.yml --branch=staging --limit=1 --json databaseId -q '.[0].databaseId')
```

- `001_init.sql` の適用と、`002_master` / `003_demo` / `004_staging` の seed バンドルロードが全て成功すること。
- Checksum mismatch エラーや `detected legacy seed migration key(s)` エラーが出ていないこと。

### 4.4 事後確認

- [ ] `GET /health` → `200 OK`
- [ ] [CRUD スモークテスト](../CRUD-SMOKE-TEST.md) を実施

---

## 5. 【bug.md 現行 backlog 対象外】旧 M-11: GitHub Secrets 登録（`CI_TEST_EMAIL` / `CI_TEST_PASSWORD`）

> **2026-07-06 時点の位置づけ**: この項目は現行 `bug.md` の「残項目」テーブルには含まれていない。
> `.github/workflows/performance-tests.yml` は `secrets.CI_TEST_EMAIL || 'admin@example.com'`
> のようにフォールバック付きで参照しており、未登録でも CI が fail-fast しない設計のため、
> C-1 のような緊急度はない。Cloudflare 移行で `infra/cloudflare/README.md` に記載の
> `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD`（CRUD smoke 用）と役割が重複する可能性があり、
> 統合するかどうかは別途 PO/管理者判断とする。以下は登録する場合の手順として参考に残す。

```bash
gh secret set CI_TEST_EMAIL --body "<STG_TEST_ACCOUNT_EMAIL>"
gh secret set CI_TEST_PASSWORD --body "<STG_TEST_ACCOUNT_PASSWORD>"
```

登録確認:

```bash
gh secret list | grep CI_TEST_
```

---

## 6. 完了記録テンプレート

各セクション実施後、このファイルの末尾に実施記録を追記する（コミットハッシュ・実行日時・
実行者を残す。値そのものは記録しない）。

```markdown
### 実施記録
- 日付:
- 実施者:
- セクション: (1 C-1ローテーション / 3 C-1追跡解除 / 2 ECSロールバック時H-5 / 4 ECSロールバック時M-4 / 5 M-11)
- 結果: PASS / FAIL
- 参照コミット・ワークフロー run URL:
```
