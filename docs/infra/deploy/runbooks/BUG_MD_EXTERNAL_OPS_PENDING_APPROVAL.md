# bug.md 残タスク: 外部操作ランブック（ユーザー承認必須）

> **目的**: `bug.md` の C-1 / H-5 / M-4 / M-11 のうち、コード変更では完結せず AWS/GitHub への
> 実操作（ローテーション・SSM登録・STGデプロイ・Secrets登録）が必要な残タスクを、
> **正確な現状**（最新の migration 統合・SSM 移行コード）に基づいて実行可能な手順に落とし込む。
> **読者**: リポジトリ管理者（AWS/GitHub 管理権限保持者）。
> **タイミング**: 各セクションの「前提コード」がマージされた後、ユーザーの明示承認を得てから実施。

> **重要**: 本ランブックのいずれの手順も、この文書を生成した AI エージェントによって**実行されていない**。
> すべて「次にユーザーが実行するコマンド」として記載している。

---

## 0. 実施順序（依存関係）

```
1. C-1 ローテーション（RDS / JWT_SECRET / INTEGRATION_ENCRYPTION_KEY / LINE）
2. H-5 SSM Parameter Store への新値登録（1 の新値を登録）
3. H-5 の IAM/ワークフロー変更を含む PR を staging へデプロイ（コード側は本セッションで実装済み）
4. C-1 `.env.staging` の追跡解除 + Issue #97 本文の実値削除
5. M-4 db_reset=true での STG デプロイ（必要な場合のみ・要ユーザー承認）
6. M-11 GitHub Secrets 登録
```
1→2→3→4 の順序を崩すと、STG デプロイが平文値の消失で失敗する（`H-5` のコードは
`SSM_SECRET_PARAM_MAP` に列挙したキーを `.env.staging` から読まず SSM からのみ解決するため、
SSM 側に新値を登録する前に `.env.staging` を追跡解除すると値の供給元がなくなる）。

---

## 1. C-1: シークレットローテーション

**前提**: 現在 `.env.staging` は git 追跡されたままで、`DB_PASSWORD` / `JWT_SECRET` /
`INTEGRATION_ENCRYPTION_KEY` を平文で含む（`git ls-files backend/.env.staging` あるいは
リポジトリルートの `.env.staging` で確認可能）。Issue #97 本文にも実値記載の指摘がある。

### 1.1 RDS DB パスワード変更

```bash
# 新パスワードを生成（例。実際の値は控えて次の SSM 登録ステップで使う）
NEW_DB_PASSWORD="$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)"

# RDS インスタンス識別子は実環境の値に置き換えること
aws rds modify-db-instance \
  --db-instance-identifier <STG_RDS_INSTANCE_ID> \
  --master-user-password "$NEW_DB_PASSWORD" \
  --apply-immediately
```

- `--apply-immediately` は即時反映（メンテナンスウィンドウを待たない）。
- 新値は 2 章の SSM 登録でのみ使用し、Slack/チャット/コミットに貼らない。

### 1.2 JWT_SECRET 変更

```bash
NEW_JWT_SECRET="$(openssl rand -base64 48)"
```

- **影響**: 変更後デプロイ時点で全ユーザーの既存セッションが無効化される。事前に STG 利用者へ周知すること。

### 1.3 INTEGRATION_ENCRYPTION_KEY 変更

- **影響**: `clinic_integrations` の暗号化済み値（Lステップ認証情報等）が新鍵では復号不能になる。
- **STG の場合**: 再入力運用で問題なければ、単純に新鍵へ切替えて既存 `clinic_integrations` 行を
  手動で再設定してもらう方が、再暗号化スクリプトを書くよりリスクが低い（データ量が少ない前提）。
- 本番導入時は「旧鍵で全件復号 → 新鍵で再暗号化」の一時スクリプトが必須（STG では省略可）。

```bash
NEW_INTEGRATION_ENCRYPTION_KEY="$(openssl rand -base64 32)"
```

### 1.4 LINE channel secret / access token 再発行

- LINE Developers コンソール（https://developers.line.biz/console/）で対象チャネルを開き、
  Channel secret / Channel access token を再発行する。
- 新値は 2 章で SSM に登録する。

---

## 2. H-5: SSM Parameter Store への新値登録

**前提コード（実装済み）**: `.github/workflows/backend-deploy.yml` の `SSM_SECRET_PARAM_MAP` が
以下のパスを参照する設計になっている（本セッションで実装、値は未登録）:

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

### 2.2 SSM パラメータ登録（1章で生成した新値を使用）

```bash
aws ssm put-parameter --name /animalekarte/stg/db/password \
  --value "$NEW_DB_PASSWORD" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/jwt_secret \
  --value "$NEW_JWT_SECRET" --type SecureString --overwrite

aws ssm put-parameter --name /animalekarte/stg/integration_encryption_key \
  --value "$NEW_INTEGRATION_ENCRYPTION_KEY" --type SecureString --overwrite
```

- LINE channel secret / access token 等、他に SSM 化したいキーがあれば
  `SSM_SECRET_PARAM_MAP`（`backend-deploy.yml`）に追記してから同様に登録する。

### 2.3 登録確認

```bash
aws ssm get-parameters --names \
  /animalekarte/stg/db/password \
  /animalekarte/stg/jwt_secret \
  /animalekarte/stg/integration_encryption_key \
  --with-decryption --query 'Parameters[].Name'
# 3件とも返ること（値自体はここでは出力しない運用にする）
```

### 2.4 デプロイして検証

`backend-deploy.yml` を含む PR を `staging` へマージ（通常の push トリガー）した後:

```bash
aws ecs describe-task-definition \
  --task-definition animalekarte-stg-api \
  --query 'taskDefinition.containerDefinitions[0].secrets'
# JWT_SECRET / DB_PASSWORD / INTEGRATION_ENCRYPTION_KEY が
# valueFrom（SSM ARN）で列挙され、containerDefinitions[0].environment 側に
# 平文で出てこないことを確認する
```

アプリの起動ログ・DB接続・ログイン（JWT発行）が正常であることを確認する。

---

## 3. C-1 残: `.env.staging` 追跡解除 + Issue #97 実値削除

**前提**: 2章のデプロイ検証が green であること（`.env.staging` を頼らずに起動できることの確認）。

```bash
git rm --cached .env.staging
echo '.env.staging' >> .gitignore
git add .gitignore
git commit -m "chore(security): untrack .env.staging now that secrets resolve via SSM (H-5)"
```

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

## 4. M-4: STG `db_reset=true` デプロイ

**前提（現状の正確な把握）**: `backend/migrations/` は既に統合済みで
`001_init.sql`（DDL 専用）1ファイルのみが存在する（2026-07-06 に
`005_add_appointment_checked_in_at.sql` [appointments.checked_in_at 追加] も
`001_init.sql` へ再統合済み。同日、002/003/004 の stub SQL ファイルも削除され、
seed データは `backend/migrations/seeds/{002_master,003_demo,004_staging}/` の
CSV バンドルのみになった — `cmd/migrate` が `001_init.sql` 適用後に固定順で
ロードする）。
（bug.md 旧稿にあった「migration 010/011」という番号は統合前の呼称で、現在は実体が無い —
これは LOW #179 と同根の文書ドリフト）。Checkup 系（#211, ADR-004）のスキーマは
既に `001_init.sql` に統合済みのため、STG の DB が古い場合は `db_reset=true` で
`001_init.sql` から作り直す以外に反映経路がない（個別 ALTER の追い migration は作成しない方針）。
**さらに、STG の `schema_migrations` に 002-004 の旧 stub SQL キー（`002_seed_master.sql`
等）が残っている場合、現行バイナリは `detectLegacySeedKeys` で fail-fast する** —
この場合も `db_reset=true` 以外の反映経路はない。

**⚠️ 破壊的操作**: `db_reset=true` は STG DB を DROP & 再作成する。STG 上の全データが失われる。
実行前に必ずユーザーへ「破棄可否」を確認すること（`.claude/skills/stg-release-readiness` 準拠）。

### 4.1 事前チェック

```bash
# ローカルで fresh DB 適用が ERROR ゼロであることを確認（STG実機証跡ではない点に注意）
docker compose down -v
docker compose up -d db
docker compose run --rm backend go run ./cmd/migrate
```

### 4.2 実行（ユーザー承認後）

```bash
gh workflow run backend-deploy.yml --ref staging -f db_reset=true
```

### 4.3 監視

```bash
gh run watch --exit-status $(gh run list --workflow=backend-deploy.yml --branch=staging --limit=1 --json databaseId -q '.[0].databaseId')
```

- `001_init.sql` の適用と、`002_master` / `003_demo` / `004_staging` の seed バンドルロードが全て成功すること（ログの `Migration summary` / `Seed bundle summary` を確認）。
- Checksum mismatch エラーや `detected legacy seed migration key(s)` エラーが出ていないこと。

### 4.4 事後確認

- [ ] `GET /health` → `200 OK`
- [ ] 健診記録の入力導線が Checkup 系 1 系統のみであることを UI で確認（M-4 本題）
- [ ] [CRUD スモークテスト](../CRUD-SMOKE-TEST.md) を実施

---

## 5. M-11: GitHub Secrets 登録（`CI_TEST_EMAIL` / `CI_TEST_PASSWORD`）

**前提コード**: performance-tests ワークフロー側は `${{ secrets.CI_TEST_EMAIL }}` /
`${{ secrets.CI_TEST_PASSWORD }}` を参照する実装済み。Secrets 自体の登録のみ残タスク。

```bash
# 値はプレースホルダ。実際のテスト用アカウント認証情報に置き換えて実行する
gh secret set CI_TEST_EMAIL --body "<STG_TEST_ACCOUNT_EMAIL>"
gh secret set CI_TEST_PASSWORD --body "<STG_TEST_ACCOUNT_PASSWORD>"
```

登録確認:

```bash
gh secret list | grep CI_TEST_
```

動作確認（手動 dispatch）:

```bash
gh workflow run performance-tests.yml
gh run watch --exit-status $(gh run list --workflow=performance-tests.yml --limit=1 --json databaseId -q '.[0].databaseId')
```

---

## 6. 完了記録テンプレート

各セクション実施後、このファイルの末尾に実施記録を追記する（コミットハッシュ・実行日時・
実行者を残す。値そのものは記録しない）。

```markdown
### 実施記録
- 日付:
- 実施者:
- セクション: (1 C-1 / 2 H-5 / 3 C-1追跡解除 / 4 M-4 / 5 M-11)
- 結果: PASS / FAIL
- 参照コミット・ワークフロー run URL:
```
