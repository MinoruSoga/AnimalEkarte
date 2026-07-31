# PROD 運用 Runbook（Cloudflare）

> **目的**: 本番環境の日常運用・障害初動・rollback・監視/backup gate。
> **読者**: 運用担当・開発者。
> **追跡**: #253（本番環境整備）／ #99（CF-only rollback 一本化）／ Go-live 前提 #257。
> **構築手順の正本**: [setup.md](setup.md)（未構築時はまずこちら）。
> **デプロイ契約の正本**: [../../deploy/CI-CD-PIPELINE.md](../../deploy/CI-CD-PIPELINE.md) §0。
>
> 本番は **未構築**（2026-07-31）。本書は STG runbook をベースに、本番固有差分
> （Required reviewers・通知・backup 検証・CF-only rollback）を先に固定する。
> 実値（token・password・接続文字列・通知メール）は書かない。

---

## 0. 現状とゲート

| 項目 | 状態 |
|---|---|
| PlanetScale prod DB / R2 / DNS / secrets | 未構築（[setup.md](setup.md)） |
| GitHub Environment `production` + Required reviewers | 未作成（setup.md §7） |
| `backend-deploy.yml` production トリガー | 未適用（setup.md §8 提案 diff） |
| ECS / AWS 切り戻し先 | **存在しない**（2026-07-20 廃止。再導入禁止） |
| CI green on latest main | **BLOCKED** — GitHub Actions billing/spending（USER 復旧のみ） |

---

## 1. デプロイ契約

### 1.1 経路

| 環境 | 自動トリガー | 承認 |
|---|---|---|
| STG | `main` → PR merge → `staging` push（対象 path） | 不要 |
| PROD | `production` ブランチ push または `workflow_dispatch --ref production` | **Environment `production` Required reviewers 必須** |

- `main` 単独 push は CI のみ。**本番デプロイを開始してはならない**。
- Frontend は Vercel（`frontend-deploy.yml`）。production 操作も backend と同じ承認運用に揃える。

### 1.2 初回・通常デプロイ（構築後）

前提: setup.md §1〜§8 完了、`APP_ENV=production` がコンテナに明示されていること（seed は master のみ）。

```bash
# 承認ゲート付き（Environment Required reviewers で一時停止 → Actions UI で承認）
gh workflow run backend-deploy.yml --ref production
gh run list --workflow=backend-deploy.yml --branch=production --limit 1
```

手動（緊急・明示承認後）:

```bash
cd backend
# 値は端末の環境 / 既存 secret 管理から。ファイルやログに残さない
npx wrangler deploy -c wrangler.production.jsonc
WORKER_URL=https://api.noah-karte.com ./infra/scripts/cf-run-migrate.sh
curl -sS https://api.noah-karte.com/health
```

### 1.3 デプロイ後確認（成功条件）

1. GitHub Actions: deploy → migrate → post-migrate health がすべて成功
2. `curl -sS https://api.noah-karte.com/health` → HTTP 200 かつ `status` が `ok`
3. フロント `https://noah-karte.com` 表示・証明書有効
4. 必要時のみ CRUD smoke（production 用 demo 資格情報がある場合。値は文書化しない）
5. イメージ更新を伴う場合は **15 分静置**後に再確認（rolling 旧イメージ残留）

失敗した step を成功扱いにしない。

---

## 2. 障害初動

1. `/health` を本番 URL で確認する
2. デプロイ直後ならローリング更新の旧イメージ残留を疑う（15 分静置 → 再確認）
3. 全断 + DB 接続エラーなら接続スロット枯渇を疑う（`DB_MAX_OPEN_CONNS` / PlanetScale 側）
4. Cloudflare 障害情報: https://www.cloudflarestatus.com/
5. GitHub Actions の直近 production run の失敗 step を特定する
6. **切り戻し先は Cloudflare のみ**。AWS/ECS は使わない（§3）

STG との切り分けが必要な場合は [../staging/runbook.md](../staging/runbook.md) と workers.dev（STG）を併用する。PROD は `workers_dev: false` のため workers.dev 公開は無い。

---

## 3. Rollback（CF-only・#99）

**原則: AWS ECS/RDS への切り戻しは技術的に不可能かつ禁止。DNS/NS を「旧インフラ」へ戻すことを復旧とみなさない。Cloudflare 正系統を復旧する。**

| 手順 | 内容 |
|---|---|
| 1 | 判断者が復旧対応を宣言。必要なら現場を Access 等の旧業務へ一時退避（業務継続） |
| 2 | 直前に正常稼働した **commit SHA** を特定する |
| 3 | その SHA の schema と現行 DB の **migration 互換**を確認する（非互換なら forward-fix または承認済みスナップショット） |
| 4 | 互換確認後、当該 tree で `wrangler deploy -c wrangler.production.jsonc`（または production ref の workflow 再実行 + Required reviewers） |
| 5 | migrate が必要な場合のみ `cf-run-migrate.sh`（破壊的 reset は別途明示承認） |
| 6 | `/health` + クリティカルパス smoke |
| 7 | provider 障害で再デプロイ不能なら、当日/直近スナップショット + IaC（`infra/cloudflare/production/` + wrangler）から再建 |
| 8 | インシデント記録（原因・影響・再発防止）。secret / PHI は記録しない |

### 3.1 Rollback rehearsal（受け入れ用・構築後）

#253 AC「rollback rehearsal を行い、復旧時間を記録」:

- [ ] 対象: 非ピークまたは隔離検証枠
- [ ] last known good 相当の再デプロイを実施（本番データ破壊操作は含めない）
- [ ] 計測: 判断宣言 → `/health` ok までの分（RTO 実測）
- [ ] schema 互換確認の所要時間も記録
- [ ] 記録場所: 当日作業ログまたは delivery 証跡（credential 無し）

---

## 4. 監視・failure notification

| 監視 | 手段 | 運用 |
|---|---|---|
| 生存確認 | `GET https://api.noah-karte.com/health` | デプロイ直後必須・障害時最初に実行 |
| 5xx 率 | Cloudflare Notification Policy（`noah-karte.com` ゾーン全体） | STG ポリシーが PROD もカバー。**PROD 専用ポリシーを追加しない**（二重通知） |
| Deploy 失敗 | GitHub Actions | 失敗通知は GH のリポジトリ/org 設定。Environment 保護エラーも監視 |
| Workers Logs | Cloudflare Observability | PHI・password・token が混入していないか定期確認 |
| Frontend | Vercel deployment | production デプロイの成功/失敗 |

### 4.1 監視チェックリスト（秘密値なし）

- [ ] 通知ポリシーがゾーンで有効（`infra/cloudflare/notifications.tf` は STG 側が正本）
- [ ] 通知先メールが到達することの事前検証（アドレス実値は書かない）
- [ ] `/health` の手動確認手順がチームで共有されている
- [ ] 障害時の連絡経路（Go-live: [GOLIVE_RUNBOOK.md](../../../delivery/GOLIVE_RUNBOOK.md) §4）が確定している

---

## 5. Backup / restore

| 項目 | 契約 |
|---|---|
| 取得主体 | PlanetScale マネージド backup および/または明示 `pg_dump`（Go-live 当日ランブック） |
| 保管 | アクセス制御された保管場所（接続情報と同居させない） |
| RPO | ポリシー確定待ち — 最低でも切替前スナップショット 1 本 |
| RTO | rollback / restore rehearsal で実測して記録 |
| 復元先 | **隔離環境のみ**を既定とする。本番上書き restore は別途明示承認 |

### 5.1 Restore rehearsal チェックリスト（#253 AC）

- [ ] 隔離環境を用意（本番 DNS / 本番 DB を直接ターゲットにしない）
- [ ] スナップショットから restore
- [ ] 非 PHI 指標で整合性確認: 主要テーブル件数、clinic_id 別件数、金額合計（個人名は出さない）
- [ ] 所要時間を記録
- [ ] 失敗時は成功扱いにせず原因を記録（credential は載せない）
- [ ] リハーサル用に作った隔離資源の破棄

### 5.2 日常バックアップ確認（稼働開始後）

- [ ] 週次: 最新 backup の存在・時刻・サイズのみ確認
- [ ] 月次: 隔離 restore の抜粋検証（件数チェック）
- [ ] 資格情報ローテーション時は backup 取得経路の認証も更新（#89 依存・人間のみ）

---

## 6. DB / secrets 運用境界

- 接続調査は **TTL 付き診断ロール**（使い捨て・値は保存しない）
- migrate は CI または `cf-run-migrate.sh`（`MIGRATE_RUN_SECRET`）。現行 workflow に `db_reset` 入力は無い
- production の `APP_ENV=production` を維持する（demo/staging seed を migrate 経路で載せない）
- secret 投入: `wrangler secret put <NAME> -c wrangler.production.jsonc`（`-c` 必須）
- STG と production で `JWT_SECRET` / `MIGRATE_RUN_SECRET` / `INTEGRATION_ENCRYPTION_KEY` を共有しない
- ローテーション手順の詳細・承認境界: [../../deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md](../../deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)

---

## 7. 検証コマンド（値は環境から）

```bash
# ヘルス
curl -sS -o /tmp/prod-health.json -w '%{http_code}\n' https://api.noah-karte.com/health
jq -r '.status' /tmp/prod-health.json

# migrate 単発（MIGRATE_RUN_SECRET は環境変数。履歴に残さない）
WORKER_URL=https://api.noah-karte.com ./infra/scripts/cf-run-migrate.sh

# CRUD smoke（production 用 demo がある場合のみ。変数名はスクリプト互換で STG_DEMO_*）
WORKER_URL=https://api.noah-karte.com ./infra/scripts/cf-crud-smoke.sh
```

---

## 8. #253 受け入れ残件と USER 作業

| AC | docs/prep | 実地 | 備考 |
|---|---|---|---|
| latest main required CI green | n/a | **BLOCKED** | GitHub billing/spending — USER only |
| STG deploy / health / failure notification 実地確認 | 契約文書化済 | billing 復旧後 | |
| production deploy は Required reviewers なしに開始できない | 契約・手順文書化済 | Environment 未作成 | setup.md §7 |
| rollback rehearsal + 復旧時間記録 | 手順 §3.1 | 環境構築後 | CF-only |
| backup → 隔離 restore + 非 PHI 整合性 | 手順 §5.1 | 環境構築後 | |
| log/artifact/Issue に credential・個人情報なし | 本 runbook 方針 | 継続 | |
| #257 Go-live の明示 prerequisite | GOLIVE §1 #2/#7 | 上記完了後 | |

**agent は支払い・spending limit・secret 実値の発行を行わない。**
