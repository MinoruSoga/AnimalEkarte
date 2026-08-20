# 外部資格情報オペレーション（ユーザー承認必須）

> **目的**: Cloudflare / PlanetScale / 外部連携の資格情報ローテーション境界を記録する。
> **読者**: リポジトリおよび各サービスの管理権限保持者。
> **タイミング**: credential-impacting な外部操作についてユーザーの明示承認を得た後。
>
> **AWS 退役境界**: AWS ECS/RDS は 2026-07-20 に廃止済みで、切り戻し先や
> ホットスタンバイではない。旧 workflow・Terraform・SSM・ECS CLI 手順は実行しない。
> 当時の証跡は git 履歴だけを参照する（2026-08-20 削除。`git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` 配下）。

---

## 0. 現状と正本

現行構成は Cloudflare Workers + Containers + PlanetScale Postgres。
デプロイは `backend-deploy.yml` が `wrangler deploy` → `POST /_internal/migrate` を実行する。
構成・障害初動は [`../../infra/architecture.md`](../../infra/architecture.md) と
[`../../infra/staging/runbook.md`](../../infra/staging/runbook.md)、
シークレット供給は [`../../../../infra/cloudflare/README.md`](../../../../infra/cloudflare/README.md) と
`backend/wrangler.jsonc` の `secrets.required` を正本とする。

---

## 1. 露出クレデンシャルのローテーション（SEC-SECRETS-5 / #89/#97）

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

GitHub Secrets（`CLOUDFLARE_API_TOKEN`, `MIGRATE_RUN_SECRET`, `STG_DEMO_EMAIL`, `STG_DEMO_PASSWORD`）
の登録手順は [`infra/cloudflare/README.md`](../../../../infra/cloudflare/README.md) の「CI デプロイ」を正とする。

AWS 時代の SSM 登録、ECS task definition、`db_reset` workflow dispatch は削除済み資産を前提とするため、
この live runbook には再掲しない。共有 STG の DB 再作成が必要な場合は、AWS 手順を復元せず、
ユーザー承認後に [STG_PLANETSCALE_SEED_RUNBOOK.md](../STG_PLANETSCALE_SEED_RUNBOOK.md) を使用する。

---

## 2. performance-tests 認証情報（#109 / TASK-606 完了契約）

> **現状（TASK-606）**: `.github/workflows/performance-tests.yml` と `load-tests/k6-api-endpoints.js` /
> `load-tests/k6-spike-test.js` は **`STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` のみ**を使用する。
> 旧 CI テスト用 secret 名・汎用 TEST_* env・ハードコード demo 認証の fallback は撤去済み。
> 両 k6 step は required（`continue-on-error` なし）。`--summary-export` の aggregate
>（`http_reqs` / `iterations` / `checks` / `successful_logins`）を fail-closed で検証する。
> login 非200・cookie 欠落・protected 非200・0 カウントは非0終了。パスワード/body/cookie/token 値は log しない。
>
> **エージェント作業境界**: secret の作成・更新・値照合は **USER 専権**。names-only 確認のみ可。
> remote green run / Issue close は統合後の USER acceptance。

```bash
# ユーザー実施: 未登録時のみ（値はチャット・git に残さない）
gh secret set STG_DEMO_EMAIL --body "<STG_DEMO_ACCOUNT_EMAIL>"
gh secret set STG_DEMO_PASSWORD --body "<STG_DEMO_ACCOUNT_PASSWORD>"

# names-only 確認（値は出ない）
gh secret list --json name --jq '[.[] | select(.name == "STG_DEMO_EMAIL" or .name == "STG_DEMO_PASSWORD")] | length'
```
