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

> **Target binding gate:** 実行前にenvironment、exact config path、Worker名、change IDを記録する。`backend/`から STG は `npx wrangler secret put <NAME> -c wrangler.jsonc`、PROD は `npx wrangler secret put <NAME> -c wrangler.production.jsonc` を使う。config省略は禁止。names-only確認でtarget不一致なら停止する。

> 🚨 **ユーザー所有・credential-impacting**。エージェントは実行しない。
> PUBLIC リポジトリ履歴および過去の seed/Issue 露出に対する正攻法は **ローテーション**（filter-repo 禁止）。

対象 4 系統（完了まで Issue #89/#97 はクローズしない）:

| # | 系統 | 手順（概要） | 投入先 |
|---|------|--------------|--------|
| 1 | PlanetScale DB | 承認済み対象 role の再発行・切替手順を provider の現行仕様で確認。共有 app default role を無断で `reset-default` しない | `DB_HOST`、`DB_USER`、`DB_PASSWORD`それぞれについて、target config明示の`npx wrangler secret put <NAME> -c <exact-config>`を1回ずつ実行（`DB_PORT` / `DB_NAME` / TLS は target Wrangler の非secret vars） |
| 2 | Cloudflare API / Worker secrets | トークン再発行 + target config明示の `npx wrangler secret put <NAME> -c <exact-config>` | Cloudflare Secrets + GitHub `CLOUDFLARE_API_TOKEN` |
| 3 | LINE channel secret / access token | LINE Developers Console で再発行 | アプリUI `/settings/integrations/lstep` から保存（LINE予約設定はsecret/tokenを扱わない）（DB 暗号化）。seed には実値を戻さない |
| 4 | JWT / INTEGRATION_ENCRYPTION_KEY 等 | 下記の鍵種別ごとの切替・復元条件を確定してから変更 | target config明示の `npx wrangler secret put <NAME> -c <exact-config>`（target Wranglerの`secrets.required`） |

**暗号鍵の変更前提**: `INTEGRATION_ENCRYPTION_KEY` はJWT署名鍵と分けて扱う。現行cipherは64文字hexの単一鍵で、旧鍵へのfallbackを持たない。既存の暗号化済み連携設定がある環境で新鍵だけを投入すると、既存値を復号できなくなる。変更前に対象データ、保護されたバックアップ、再暗号化または再登録の手順、切替・復元条件と旧鍵の保持期限を担当者が承認し、隔離環境で検証する。この手順が未確定なら鍵変更を停止する。旧鍵・平文・暗号文は文書やログへ出さない。

JWT署名鍵の変更は既存セッションへの影響と再ログイン案内を確認する。いずれも `/health` だけでは検証できず、認証または既存連携設定の復号・利用まで確認する。暗号鍵の実装根拠は `backend/internal/infra/crypto/aes_gcm.go`。

検証（ローテーション後・ユーザー実施）:

```bash
# STG health（正系統）
curl -sS -o /dev/null -w '%{http_code}\n' https://api.stg.noah-karte.com/health
# 旧値でのアクセスが拒否されること（各コンソール / wrangler の確認 UI）
```

### 1.1 不足チェック（#89/#97 / BRT-37 · 2026-08-20）

**実ローテ・値の記載・git filter-repo は禁止。** 証跡は非機密 1 行のみ。

| 系統 | 手順の所在 | 実行 | 非機密 receipt |
|---|---|---|---|
| 1 PlanetScale DB | 上表 + [staging/runbook.md](../../infra/staging/runbook.md) | USER | **未記入** |
| 2 Cloudflare API / Worker | 上表 + `wrangler secret put` | USER | **未記入** |
| 3 LINE channel | 上表。値は UI から。seed に戻さない | USER | **未記入** |
| 4 JWT / INTEGRATION_ENCRYPTION_KEY | 上表の鍵種別ごとの切替・復元条件。値は書かない | USER | **未記入** |
| #97 本文マスク | ローテ完了後のみ `gh issue edit` | USER | **未記入** |

API token等の順序: 新発行 → secret 投入 → 再デプロイ → `/health` と対象機能の確認 → 旧 revoke → 旧値拒否確認。暗号鍵は上記の承認済みデータ移行・復元手順を使い、この順序だけで切り替えない。完了まで #89/#97 は close しない。

GitHub Secrets（`CLOUDFLARE_API_TOKEN`, `MIGRATE_RUN_SECRET`, `STG_DEMO_EMAIL`, `STG_DEMO_PASSWORD`）
の登録手順は [`infra/cloudflare/README.md`](../../../../infra/cloudflare/README.md) の「CI デプロイ」を正とする。

AWS 時代の SSM 登録、ECS task definition、`db_reset` workflow dispatch は削除済み資産を前提とするため、
この live runbook には再掲しない。共有 STG の DB 再作成が必要な場合は、AWS 手順を復元せず、
ユーザー承認後に [STG_PLANETSCALE_SEED_RUNBOOK.md](../STG_PLANETSCALE_SEED_RUNBOOK.md) を使用する。

---

## 2. performance-tests 認証契約（#109 後続の現行実装）

`performance-tests.yml` の endpoint/spike k6 は CI の独立 local Docker stack（`BASE_URL=http://localhost:8080`、`APP_ENV=test`）を対象にし、**`LOAD_TEST_LOGIN_EMAIL` / `LOAD_TEST_LOGIN_PASSWORD`** を job env から渡す。アカウントは `internal/seedlogin` の公開 synthetic catalog であり、STG secret の登録はこの経路の前提ではない。

- `k6-api-endpoints.js` / `k6-spike-test.js` は両変数必須、fallback 無し。login 非200・cookie 欠落・protected 非200 は失敗。
- 両 k6 step は required。aggregate は `scripts/validate-k6-summary.mjs` が `http_reqs` / `iterations` / `checks` / `successful_logins` を fail-closed で検証する。
- 実 STG 用 `k6-cf-stg-sustained.js` と backend deploy の optional CRUD smoke は引き続き `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` を使う。CI-local の変数と混ぜない。実 STG 負荷試験・secret 登録は承認済みの別運用。
- E2E workflow は `APP_ENV=test` と `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` を使う別契約。

[#109](https://github.com/MinoruSoga/AnimalEkarte/issues/109) は2026-07-31にコード契約の確認として CLOSED。当時の `STG_DEMO_*` 一本化コメントは履歴であり、現行 local CI の接続先・変数は上記実装で確認する。今回の文書照合では remote k6 / E2E run の成功は検証していない。
