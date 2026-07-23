# 本番 Cloudflare 基盤 事前構築手順書 (Production CF Setup)

> **目的**: 本番環境（`noah-karte.com` / `api.noah-karte.com`）を Cloudflare Workers + Containers +
> PlanetScale 構成で新設するための人間向け実施手順書。
> **読者**: インフラ担当（実施者）。
> **タイミング**: 7/17 の事前構築を想定（**7/18 Go-live 当日の作業ではない**）。
> **前提**: PO決定（2026-07-15）「納品はCloudflare経路。本番もCF構成で新設」/ 追跡Issue **#253**。
> 詳細背景は [`../_archive/migration-cloudflare.md`](../_archive/migration-cloudflare.md) 「現況サマリ」
> 2026-07-15/16 ブロック参照。critical path 上、本手順は **STG Phase 7（NS切替）完了後**に
> 着手する前提（`noah-karte.com` ゾーンが Cloudflare 上で active であることが前提条件）。
>
> 本書が参照する新規ドラフトファイル:
> - [`backend/wrangler.production.jsonc`](../../../../backend/wrangler.production.jsonc)
> - [`infra/cloudflare/production/`](../../../../infra/cloudflare/production/)（Terraform一式）
>
> これらは STG の正本（`backend/wrangler.jsonc` / `infra/cloudflare/*.tf`）をベースにした
> **ドラフト**であり、`terraform apply` 等の実インフラ変更は未実施。本書はそれを実施する
> 人間向けの手順である。

---

## 0. 事前確認チェックリスト

- [ ] `../_archive/migration-cloudflare.md` の STG Phase 7（NS切替）が完了している（`noah-karte.com` ゾーンが
      Cloudflare 上で active。`dig noah-karte.com NS` で Cloudflare の NS が返ることを確認）
- [ ] `production` ブランチが存在する（確認済み: `git branch -a` に `remotes/origin/production` あり）
- [ ] `docs/ops/deploy/README.md` の「Production」行が指す `noah-karte.com` / `api.noah-karte.com` が
      これから構築する対象と一致していることを確認
- [ ] Frontend（Vercel, `production` ブランチで自動デプロイ済み）が `VITE_API_URL` として
      `https://api.noah-karte.com/api` を指すよう設定済み、または未設定なら本手順の完了後に
      Vercel 側の production 環境変数を更新する運用担当者へ引き継ぐ（本書のスコープは
      backend Cloudflare 基盤のみ。Vercel 環境変数の変更自体は別タスク）

---

## 1. 全体の実施順序

依存関係があるため、この順序を守ること。各ステップの詳細は該当節を参照。

```
1. PlanetScale 本番DB作成                         (§2)
2. Terraform apply (infra/cloudflare/production/)  (§3)
   → R2バケット作成 / api.noah-karte.com DNSレコード作成 / Hyperdrive設定作成(未使用予約)
3. R2 S3互換トークン発行                            (§4)
4. wrangler secrets 投入                            (§5)
5. wrangler.production.jsonc の PLACEHOLDER 置換     (§6)
6. GitHub Environment "production" 作成 + secrets登録 (§7)
7. backend-deploy.yml へ production トリガー追加      (§8, 提案diff)
8. 初回デプロイ実行 + 検証                           (§9)
9. 既知の制約への対応(seedバンドル cleanup 等)        (§10)
```

---

## 2. PlanetScale 本番DB作成

STG の [`infra/scripts/pscale-create-stg.sh`](../../../../infra/scripts/pscale-create-stg.sh) を
production 向けに読み替えた手順。同スクリプトは STG 専用（DB名がハードコードされている）ため、
本番では以下を手動実行する（新規スクリプトファイルは作成しない。1回限りの作成作業のため）。

```bash
ORG="noah-animalekarte"
DB_NAME="animalekarte-prod"
REGION="ap-northeast"
CLUSTER_SIZE="PS-10"   # STGと同一クラスサイズを初期値とする。本番トラフィック実測後に見直す

pscale auth check

pscale database create "${DB_NAME}" \
  --org "${ORG}" \
  --region "${REGION}" \
  --cluster-size "${CLUSTER_SIZE}"

pscale database show "${DB_NAME}" --org "${ORG}"
pscale branch list "${DB_NAME}" --org "${ORG}"
```

### 2.1 接続クレデンシャルの取得

```bash
pscale role reset-default "${DB_NAME}" main --org "${ORG}" --force
```

**【STGとの重要な相違点】** STG の運用（`infra/cloudflare/hyperdrive.tf` 冒頭コメント）では、
この `reset-default` で得た値は Terraform 検証用の使い捨てとして扱い、「検証後は同コマンドで
失効させる」運用だった。**本番ではこれを踏襲しない。** 理由: この節で取得する host/user/password
は (a) `infra/cloudflare/production/` の Hyperdrive Terraform 変数 **と** (b) `wrangler secret put`
で投入する `DB_HOST`/`DB_USER`/`DB_PASSWORD`（Container が PlanetScale へ直結する際に**継続的に**
使う本番稼働用クレデンシャル）の両方に使う。取得直後に同コマンドで再度 `reset-default` すると
(b) の値が無効化され、稼働中の Container が DB 接続できなくなる。**このステップで取得した値は
そのまま維持し、ローテーションが必要になったら別途ランブック化して両箇所(Terraform state と
wrangler secret)を同時に更新すること。**

取得した値は端末の環境変数にのみ保持し、ファイルへ書き出したりログに残したりしないこと
(STGの `infra/cloudflare/README.md` の秘密情報取り扱い方針と同じ)。

---

## 3. Terraform apply (`infra/cloudflare/production/`)

### 3.1 このドラフトで既に実行済みの検証

- `terraform fmt -check`: 差分なし(整形済み)
- `terraform init -backend=false` + `terraform validate`: **PASS**
  (Cloudflare provider ~> 5.21 のスキーマに対する構文検証。認証情報は不要なため実行済み)
- `terraform plan` / `terraform apply` は Cloudflare 認証情報が必要なため**未実行**。
  以下は実施者が行う。

### 3.2 認証

```bash
export CLOUDFLARE_API_TOKEN=...   # production 用に新規発行するトークン(STGのトークンを流用しない)
export TF_VAR_account_id=$CLOUDFLARE_ACCOUNT_ID   # STGと同一アカウント前提(README.md参照)
```

発行スコープ: `infra/cloudflare/production/providers.tf` のコメント参照
(Workers Scripts/R2/Hyperdrive/Account Rulesets の Edit + Zone: DNS/Zone Settings の Edit
(`noah-karte.com` に限定) + Zone: Zone Read)。**Zone Read が必須**な点に注意
(`zone.tf` の `data "cloudflare_zone"` によるゾーン参照に必要。STGのトークンにはDNS Editは
あるがZone Readが明示要求されていなかった可能性があるため、production用トークン発行時に
スコープを再確認すること)。

### 3.3 plan → 承認 → apply

```bash
cd infra/cloudflare/production
terraform init
terraform validate
terraform plan -out=tfplan \
  -var="pscale_prod_db_host=${PSCALE_PROD_DB_HOST}" \
  -var="pscale_prod_db_user=${PSCALE_PROD_DB_USER}" \
  -var="pscale_prod_db_password=${PSCALE_PROD_DB_PASSWORD}"
# ここで plan 内容をレビューし、明示承認を得る(infra/cloudflare/README.md の安全ルールと同じ)
terraform apply tfplan
```

**確認事項(plan 内容レビュー時)**:
- `data.cloudflare_zone.noah_karte` は **read-only lookup**（作成・変更ではない）であること
- 作成されるリソースが `cloudflare_r2_bucket.prod_images`
  ・`cloudflare_dns_record.api_prod_backend`・`cloudflare_hyperdrive_config.prod_planetscale`
  の3つのみであること（STGの `cloudflare_zone`/DNS 9件/通知ポリシーが**再作成されようとしていたら
  即座に中断**。同一ゾーンの二重管理事故の兆候）

apply 後:

```bash
terraform output r2_bucket_name          # animalekarte-prod-images を確認
terraform output hyperdrive_config_id    # §6 で wrangler.production.jsonc に投入
terraform output zone_id                 # noah-karte.com のゾーンID(STGと同一のはず)
```

---

## 4. R2 S3互換トークン発行

STGの実施記録（`../_archive/migration-cloudflare.md` 試行8, P2-3）と同じ手順を production バケット向けに行う。

```bash
# ACCOUNT_ID は backend/wrangler.jsonc に既に公開値として存在するもの(STGと同一アカウント)
ACCOUNT_ID="776ddc3e975e8fe5773d5300522e2404"

curl -X POST "https://api.cloudflare.com/client/v4/accounts/${ACCOUNT_ID}/tokens" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "animalekarte-prod-r2-s3-compat",
    "policies": [{
      "effect": "allow",
      "permission_groups": [{"id": "<Workers R2 Storage Bucket Item Write の permission_group id>"}],
      "resources": {"com.cloudflare.edge.r2.bucket.'"${ACCOUNT_ID}"'_default_animalekarte-prod-images": "*"}
    }]
  }'
```

- `permission_groups` の具体的な `id` はダッシュボード or
  `GET /accounts/{account_id}/tokens/permission_groups` で都度取得すること
  (STG実施時と同様、値がハードコードで判明していないため本書には書かない)。
- レスポンスの `result.id` = `AWS_ACCESS_KEY_ID`、`result.value` の **SHA-256** =
  `AWS_SECRET_ACCESS_KEY`([R2 Authentication](https://developers.cloudflare.com/r2/api/tokens/) 準拠。
  STGの `r2.tf` 冒頭コメント・試行8記録と同じ変換)。
- バケットを `animalekarte-prod-images` のみに限定すること(STGバケットへのアクセス権を
  production トークンに含めない)。

---

## 5. wrangler secrets 投入一覧

`backend/wrangler.production.jsonc` の `secrets.required` と同一セット。
**すべて `wrangler secret put <NAME> -c wrangler.production.jsonc` で投入**する
(`-c` を忘れると STG 側の `wrangler.jsonc` に投入されてしまうため必ず指定)。

| Secret | 値の出所 | STGと同じ値を使い回してよいか |
|---|---|---|
| `DB_HOST` | §2.1 で取得した PlanetScale host | 不可(別DB) |
| `DB_USER` | §2.1 で取得した PlanetScale user | 不可(別DB) |
| `DB_PASSWORD` | §2.1 で取得した PlanetScale password | 不可(別DB) |
| `AWS_ACCESS_KEY_ID` | §4 で発行した R2トークンの `id` | 不可(別バケット・別トークン) |
| `AWS_SECRET_ACCESS_KEY` | §4 で発行した R2トークンの `value` のSHA-256 | 不可 |
| `MIGRATE_RUN_SECRET` | `openssl rand -base64 48` で新規生成 | **不可(必ず新規生成)** |
| `JWT_SECRET` | `openssl rand -base64 48` で新規生成(32文字以上。DEPLOYMENT_CHECKLIST.md準拠) | **不可(必ず新規生成)** |
| `INTEGRATION_ENCRYPTION_KEY` | `openssl rand -hex 32`(32バイトhex固定。`backend/internal/infra/crypto/aes_gcm.go`のコメント準拠) | **不可(必ず新規生成)** |
| `SMTP_HOST` | 本番用SMTP設定(運用担当者が決定。STGと同一SMTPサーバーを共用するかは別途判断) | 要判断 |
| `SMTP_USER` | 同上 | 要判断 |
| `SMTP_PASS` | 同上 | 要判断 |

**MIGRATE_RUN_SECRET / JWT_SECRET / INTEGRATION_ENCRYPTION_KEY を STG と共有しないこと**は
`backend/wrangler.production.jsonc` 冒頭コメントにも明記した最重要事項。STG側の値が漏洩した際に
production が連鎖して侵害される経路を断つため。

投入コマンド例(値はプレースホルダ。実値を貼り付けてから実行し、シェル履歴に残さないよう
先頭にスペースを入れるかヒアドキュメントを使うこと):

```bash
cd backend
 pnpm exec wrangler secret put DB_HOST -c wrangler.production.jsonc
 pnpm exec wrangler secret put DB_USER -c wrangler.production.jsonc
 pnpm exec wrangler secret put DB_PASSWORD -c wrangler.production.jsonc
 pnpm exec wrangler secret put AWS_ACCESS_KEY_ID -c wrangler.production.jsonc
 pnpm exec wrangler secret put AWS_SECRET_ACCESS_KEY -c wrangler.production.jsonc
 pnpm exec wrangler secret put MIGRATE_RUN_SECRET -c wrangler.production.jsonc
 pnpm exec wrangler secret put JWT_SECRET -c wrangler.production.jsonc
 pnpm exec wrangler secret put INTEGRATION_ENCRYPTION_KEY -c wrangler.production.jsonc
 pnpm exec wrangler secret put SMTP_HOST -c wrangler.production.jsonc
 pnpm exec wrangler secret put SMTP_USER -c wrangler.production.jsonc
 pnpm exec wrangler secret put SMTP_PASS -c wrangler.production.jsonc
```

---

## 6. `wrangler.production.jsonc` の PLACEHOLDER 置換

ファイル冒頭コメントに列挙済みの3箇所を埋める:

1. `hyperdrive[0].id` ← §3.3 の `terraform output hyperdrive_config_id`
2. `vars.TRUSTED_PROXY_CIDR` ← §9.3 の実測後に確定(初回デプロイ**後**の作業。デプロイ自体は
   STGの実測値 `10.1.0.0/32` を暫定候補のまま進めてよい。広げる方向の変更はしないこと)
3. `vars.S3_PUBLIC_BASE_URL` ← R2 production 公開ドメイン作成後(任意タイミング。未設定でも
   起動は可能。空文字のままだと画像公開URLがAPIホストへフォールバックし警告ログが出る)

---

## 7. GitHub Environment "production" 作成 + secrets登録

1. GitHub リポジトリ → Settings → Environments → **New environment** → 名前を厳密に
   `production` とする(§8のワークフローが `github.ref_name` = ブランチ名と一致させて
   参照するため、名前の一致が必須)。
2. **Required reviewers** を設定し、production への実デプロイに人間承認を必須化する
   (production-impacting action の承認ゲート)。
3. 以下を **Environment secrets**（Repository secretsではなく、この production 環境専用）
   として登録する。**名前はSTGのRepository secretsと同一名でよい**
   (GitHub Actionsは `environment:` を指定したジョブで同名secretがあれば環境側を優先し、
   なければrepo-levelへフォールバックする。これにより production job だけが production用の
   値を使い、staging job は従来通りrepo-level値を使い続ける):

   | Secret名 | 値 |
   |---|---|
   | `CLOUDFLARE_API_TOKEN` | §3.2 で発行した production 用トークン(STGのトークンを登録しない) |
   | `MIGRATE_RUN_SECRET` | §5 で `wrangler secret put` した値と同一 |
   | `STG_DEMO_EMAIL`(任意) | production 用デモアカウントのメール(スモークテストしたい場合のみ) |
   | `STG_DEMO_PASSWORD`(任意) | 同上のパスワード |

   > `STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD` という名前がproduction用として不自然に見えるのは
   > `infra/scripts/cf-crud-smoke.sh` がこの変数名をハードコードしているため
   > (本タスクではスクリプト自体の編集はスコープ外)。Environment secretsの名前を
   > スクリプトの参照名に合わせているだけで、値はproduction用デモアカウントのものにする。

---

## 8. `backend-deploy.yml` へ production トリガー追加(提案diff)

**本タスクではワークフローファイル自体は編集しない。** 以下は人間が適用する提案diff。

既存の `.github/workflows/frontend-deploy.yml` が既に同じブランチ(`staging`/`production`)を
1ジョブ内の `github.ref_name` 三項演算子で分岐する設計を採用しているため、本diffも
**同じパターン**を踏襲する(ジョブを複製せず、新規ファイルでもない。既存ジョブへの最小差分)。

```diff
--- a/.github/workflows/backend-deploy.yml
+++ b/.github/workflows/backend-deploy.yml
@@ -6,6 +6,7 @@ on:
   push:
     branches:
       - staging
+      - production
     paths:
       - 'backend/**'
       - '.github/workflows/backend-deploy.yml'
@@ -15,11 +16,13 @@ on:
   workflow_dispatch:
 
 permissions:
   contents: read
+  deployments: write

 env:
-  # P1-2 NS 切替前の検証経路。切替後は api.stg.noah-karte.com に更新 (P7-4)
-  WORKER_URL: https://animalekarte-stg-api.baritech-soga.workers.dev
+  # production ブランチ push 時は本番Worker URL、それ以外(staging)は従来通りSTGの
+  # workers.dev検証URL。P1-2 NS切替後はSTG側もapi.stg.noah-karte.comへ更新予定(P7-4)。
+  WORKER_URL: ${{ github.ref_name == 'production' && 'https://api.noah-karte.com' || 'https://animalekarte-stg-api.baritech-soga.workers.dev' }}
   NODE_VERSION: "24"

 jobs:
   deploy:
     name: Deploy Backend to Cloudflare
     runs-on: ubuntu-latest
     timeout-minutes: 30
+    # production push時は 'production' という名前のGitHub Environmentを参照する
+    # (staging push時は 'staging' を参照するが、Environment未作成でも自動生成されるだけで
+    # 保護ルール・環境専用secretsが無いためrepo-levelのsecretsに非破壊フォールバックする)。
+    # 'production' Environmentの作成・Required reviewers・環境専用secrets登録の手順は
+    # setup.md §7 参照。
+    environment: ${{ github.ref_name }}

     steps:
       - name: Checkout
@@ -61,7 +64,9 @@ jobs:
       - name: Deploy Worker and Container
         env:
           CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
         working-directory: backend
-        run: pnpm exec wrangler deploy
+        # production ブランチのみ wrangler.production.jsonc を明示指定する
+        # (-c 省略時は wrangler.jsonc = STG が使われるため必須)
+        run: pnpm exec wrangler deploy ${{ github.ref_name == 'production' && '-c wrangler.production.jsonc' || '' }}
```

**diffが触れない箇所(意図的)**: migrate ステップ・health-checkステップ・smokeステップは
`env.WORKER_URL`(上記diffで分岐済み)と `secrets.*`(Environment分岐で自動的に production用の
値へ切り替わる)を参照しているだけなので、変更不要。これにより **staging の既存挙動は完全に
不変**のまま production 対応が追加される。

**適用方法**: 上記diffを手動で `.github/workflows/backend-deploy.yml` へ反映するか、
`git apply` できるようファイルに保存して `git apply <file>` を実行する
(保存先ファイルのコンテキスト行が実ファイルと厳密に一致しない場合は手動反映に切り替えること。
本diffは提案であり、機械的apply可能性を保証するものではない)。

---

## 9. 初回デプロイ実行 + 検証

### 9.1 デプロイ

§7 の Environment 作成・secrets登録・§8 のdiff適用が完了したら、`production` ブランチへ
`backend/**` の変更を含む push(通常は `staging` → `production` へのPRマージ)を行うか、
`workflow_dispatch` を `production` ブランチ指定で手動起動する:

```bash
gh workflow run backend-deploy.yml --ref production
gh run list --workflow=backend-deploy.yml --branch=production --limit 1
gh run view <run-id> --log-failed   # 失敗時のみ
```

Required reviewersを設定した場合、ジョブは承認待ちで一時停止する(Actions画面から承認)。

### 9.2 ヘルスチェック

```bash
curl -s https://api.noah-karte.com/health | jq '.status'
# 期待: "ok"
```

CI の `Wait for /health (post-migrate)` ステップが自動確認するが、手動でも確認すること。

### 9.3 TRUSTED_PROXY_CIDR の実測(初回デプロイ後・必須)

STG試行9と同じ手法: `/health` に一時的な診断フィールドを追加し、Worker→Container間の
実ソースIPを複数リクエストにわたり確認する。安定した単一CIDRが確認できたら
`backend/wrangler.production.jsonc` の `<PLACEHOLDER-TRUSTED_PROXY_CIDR-VERIFY-POST-DEPLOY>`
箇所(コメント)を確定値に更新し、値をSTGの `10.1.0.0/32` と比較して記録する
(一致していれば「同一アカウント・同一Containers構成では安定している」という知見が得られ、
今後の環境追加時の参考になる)。診断フィールドは検証後に必ず revert すること
(STGでも同様に一時追加→revert済み)。

### 9.4 migrate 検証

```bash
WORKER_URL=https://api.noah-karte.com MIGRATE_RUN_SECRET=<production用の値> \
  ./infra/scripts/cf-run-migrate.sh
```

exitCode 0 を確認。**初回は空DBへの `001_init.sql` + `002_lstep_snapshot_import_clinic_fk.sql` + `003_medical_records_appointment_id_index.sql` 適用後、seedバンドル3本の自動ロードになる。
§10 を必ず読むこと(demo/staging データが自動投入される既知の挙動)。**

### 9.5 スモーク(任意)

§7 で `STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD`(production用)を登録した場合のみ:

```bash
STG_DEMO_EMAIL=<production用> STG_DEMO_PASSWORD=<production用> \
  WORKER_URL=https://api.noah-karte.com ./infra/scripts/cf-crud-smoke.sh
```

---

## 10. 既知の制約・フォローアップ

### 10.1 【重要】初回 migrate が demo/staging seed データを自動投入する

`backend/cmd/migrate/main.go` の `runSeedBundles` は `seedbundle.BundleOrder`
(`002_master → 003_demo → 004_staging`)を**常に全件**適用する設計であり、環境(STG/production)を
判定して一部をスキップする機構は現状コードに存在しない
(`docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md`: 「fresh DB 適用後の正しい終了状態は
schema_migrations に6行: DDL 3本 + 3seedバンドル全て」)。

これは本書が新規に発見した問題ではなく、`docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md` 44行目に
「重要: 本番移行時に全削除。ステージング環境でのみ有効」と既に明記されている既知事項。
**ただし、これを自動的に防ぐ仕組みは現状存在しないため、production の初回 migrate 実行後は
以下を人間が実施する必要がある**:

1. `003_demo`/`004_staging` バンドルが投入したレコード(デモアカウント・テスト医院等)を特定する
   (バンドル定義: `backend/migrations/seeds/003_demo/` / `004_staging/` の CSV / `manifest.json`)
2. **生SQLの `DELETE FROM ...` は使わないこと**
   (`STG-DEMO-DATA-LIFECYCLE.md` 99行目で明示的に「禁止パターン」とされている。
   FK制約・監査ログ・soft delete規約を無視した削除は整合性を壊すリスクがある)
3. 代わりに、アプリケーションの **API DELETE エンドポイント**を使い、
   `STG-DEMO-DATA-LIFECYCLE.md` §3〜4 のFK安全な削除順序に従って削除する
   (audit_log に削除操作が正しく記録される)
4. **PO決定 #250(Access移行)による実データ投入は、この cleanup 完了後に行うこと**
   (デモ/テストデータと実データが混在した状態での投入を避けるため)

この制約はコード改修(migrate時のバンドル選択を環境変数等で制御する等)で恒久対応可能だが、
本タスクのスコープ(新規ファイルのみ・既存ファイル非改変)には含まれない。恒久対応が必要と
判断される場合は、別途Issue化してスコープを切り出すこと(本書はその判断材料として記録する)。

### 10.2 通知(アラート)はSTG側ポリシーが production も暗黙にカバーする

`infra/cloudflare/production/notifications.tf` にリソース定義を置いていない
(意図的。理由は同ファイルのコメント参照)。STGの `cloudflare_notification_policy`
(`http_alert_edge_error`)は `noah-karte.com` ゾーン全体の5xx率を監視しており、
production起因の5xxも同じ通知に含まれる。ホスト名単位での分離は現状の
Cloudflare通知APIでは不可能(STGコードの調査記録どおり)。production専用の通知が
必要になった場合は `cloudflare_healthcheck`(別課金アドオン)の導入をSTG側と合わせて
再設計すること。

### 10.3 TRUSTED_PROXY_CIDR は未検証のまま初回デプロイする

§9.3 に従い初回デプロイ後に確定させること。確定するまでレート制限(rate limit)の
信頼境界がSTGの実測値からの「推測」のままになる点をリスクとして認識しておく
(誤りの効果は非対称: 広すぎる値はX-Forwarded-For偽装によるレート制限バイパス、
狭すぎる値は可用性低下に留まる)。

### 10.4 Hyperdriveは現状未使用の予約リソース

Container(Go/Ginバイナリ)はHyperdriveに非対応のため直結接続する設計であり、
`infra/cloudflare/production/hyperdrive.tf` が作るリソースは実トラフィックに使われない。
DB接続情報をtfstateに保持し続けるコストだけが発生するため、production運用開始後に
「本当に将来使うか」を一度棚卸しし、不要と判断されればリソース自体の削除を検討すること。

### 10.5 R2バケットの本番公開ドメイン(S3_PUBLIC_BASE_URL)は未確定

§6 のPLACEHOLDER参照。運用担当者がCloudflareダッシュボードでproduction用R2公開ドメインを
作成し次第、`wrangler.production.jsonc` を更新して再デプロイすること。

---

## 11. ロールバック

Cloudflare側(Worker/Container)には旧バージョンへの自動ロールバック機構がない
(STGの `../_archive/migration-cloudflare.md` リスク登録簿に記載の既知の制約と同じ:
Durable Object bindのため段階的デプロイ・Preview URLが使えない)。
問題が発生した場合:

1. 直前の green だった commit へ `production` ブランチを戻し、再度 `wrangler deploy` する
   (前方復旧。DBスキーマが前進していた場合は expand-contract 前提で設計されていない限り
   注意が必要 — 本番初回構築時点ではデータが実質空のため影響は限定的)
2. Terraform側(`infra/cloudflare/production/`)の変更をロールバックする場合は
   `terraform plan` で差分を確認してから `apply`(destroyは特に慎重に。§3.3 と同じ承認プロセス)
3. AWS側の本番ロールバック経路は本書のスコープ外
   (現状、本番バックエンドはCloudflareで新設するため、AWS側に「戻す先」の稼働中本番環境が
   存在しない。`infra/CLAUDE.md` の本番向けAWS想定表は本ドラフトにより実質的に置き換えられる)
