# PROD 運用 Runbook（Cloudflare）

> **目的**: 本番環境の日常運用・障害初動・rollback・監視/backup gate。
> **読者**: 運用担当・開発者。
> **追跡**: #253（本番環境整備）／ #99（CF-only rollback 一本化）／ Go-live 前提 #257。
> **構築手順の正本**: [setup.md](setup.md)（未構築時はまずこちら）。
> **デプロイ契約の正本**: [../../deploy/CI-CD-PIPELINE.md](../../deploy/CI-CD-PIPELINE.md) §0。
>
> 本番は **未構築**（2026-08-20 再確認）。本書は STG runbook をベースに、本番固有差分
> （Required reviewers・通知・backup 検証・CF-only rollback）を先に固定する。
> 実値（token・password・接続文字列・通知メール）は書かない。

---

## 0. 現状とゲート

> 実測日: 2026-08-20（`gh api repos/MinoruSoga/AnimalEkarte/environments` + workflow 本文。runtime・課金・backup 実測は未実施）。agent は production 実デプロイ・secret 投入・reviewer 追加をしない。

| 項目 | 状態 | 実測根拠 |
|---|---|---|
| PlanetScale prod DB / R2 / DNS / secrets | 未構築 | [setup.md](setup.md) §1〜§6 が未完了前提。本 unit は対象環境を作成していない |
| GitHub Environment 名 | **`Production`（先頭大文字）が存在**。`production`（小文字・ブランチ名一致）は **無い** | API: `Preview` / `Production` / `staging`。setup.md §8 提案の `environment: ${{ github.ref_name }}` はブランチ `production` を参照するため、このままでは既存 `Production` に届かない |
| Required reviewers | **空**（protection_rules 0） | 同上 API。無承認 production deploy を止められない。USER が設定。agent は reviewer を追加しない |
| `backend-deploy.yml` production トリガー / `environment:` ジョブゲート | **未適用** | `on.push.branches: [staging]` のみ。ジョブに GitHub Environment キー無し。setup.md §8 は提案 diff のまま。本セッションでは workflow を適用しない |
| `frontend-deploy.yml` production 経路 | **branch トリガーあり・承認ゲート無し** | `on.push.branches` に `production`。`workflow_dispatch.inputs.environment` は Vercel 向け入力であり、GitHub Environment 承認ではない |
| ECS / AWS 切り戻し先 | **存在しない** | 2026-07-20 廃止。`backend-deploy.yml` ヘッダも ECS 版削除済み。**再導入禁止** |
| CI green on latest main | **BLOCKED**（docs 上の前提） | GitHub Actions billing/spending は USER 復旧のみ。agent は課金状態を実測・変更しない。候補 required check は §8 と [todo-po.md](../../../../todo-po.md) #253（詳細は git 履歴 / Issue #253） |
| PROD backup / restore / rollback **実行スクリプト** | **repo に存在しない** | `scripts/` に backup/restore/rollback/deploy 名のスクリプト 0 本。`pg_restore` は `scripts/`・`Makefile` とも 0 ヒット。手順は §3.1 / §5.1 の **文書のみ**。rehearsal 証跡は **未記入** |

### 0.1 USER 専権（本セッションでは触らない）

| # | 作業 | なぜ USER か |
|---|---|---|
| 1 | Environment 名を契約どおり `production` にするか、workflow 参照名を既存 `Production` に合わせる | 設定変更 / 本番ゲート |
| 2 | Required reviewers を 1 名以上入れる | 無承認 deploy 禁止の AC |
| 3 | setup.md §8 の workflow 適用 | production 実デプロイ経路 |
| 4 | Environment / wrangler secrets 投入 | secret |
| 5 | backup/restore rehearsal と RTO 記録 | 本番/隔離環境の実操作 |
| 6 | GitHub Actions billing 復旧 | 有料操作 |

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

#253 AC「rollback rehearsal を行い、復旧時間を記録」。

**前提（満たさない場合は本節を開始しない）**

1. setup.md §1〜§9 が完了し、production へ **1 回以上** 正常デプロイ済みであること
2. GitHub Environment `production` + Required reviewers が有効であること（setup.md §7）
3. `backend-deploy.yml` に production トリガーが適用済みであること（setup.md §8）
4. 非ピーク枠 **または** 隔離検証枠であること（本番データ破壊操作は含めない）
5. last known good の **commit SHA** が特定済みであること（credential 無しで記録可能）

**repo 側 tooling**: production rollback 専用スクリプトは **存在しない**。以下は CLI / Actions の雛形（placeholder は環境から埋める。実値は文書に焼かない）。

#### 手順（逐次・判定付き）

```bash
# 0) 判断宣言の時刻を記録（RTO 起点）。credential / PHI は書かない
date -u +%Y-%m-%dT%H:%M:%SZ
# 次へ進む: 宣言時刻が作業ログに残った
# 止まる: 判断者未定・枠外（本番ピーク）なら中止

# 1) last known good SHA を固定（例: 直前の成功 deploy の commit）
GOOD_SHA='<last-known-good-commit-sha>'   # 実値は端末のみ。runbook に書かない
git rev-parse --verify "${GOOD_SHA}^{commit}"
# 次へ進む: 終了コード 0 で SHA が解決する
# 止まる: unknown revision → SHA を再特定してからやり直し

# 2) schema 互換の机上確認（破壊的 reset はしない）
#    - GOOD_SHA 時点の backend/migrations/ と現行 production DB の適用済み版を比較
#    - 非互換（破壊的 down / 必須 backfill 欠落）なら forward-fix または承認済みスナップショット経路へ分岐
# 次へ進む: 互換と記録できた（所要分も記録）
# 止まる: 非互換かつ承認済み経路が無い → リハーサル中止（成功扱いしない）

# 3) 互換確認後、GOOD_SHA 相当を production 経路で再デプロイ
#    A) Actions（Environment 承認が挟まる）
gh workflow run backend-deploy.yml --ref production
#    または GOOD_SHA を production ブランチへ載せた状態で path 付き push（運用ポリシーに従う）
#
#    B) 緊急の明示承認後のみ手動（値は環境から。履歴に残さない）
# cd backend && npx wrangler deploy -c wrangler.production.jsonc
#
# 次へ進む: Actions が Required reviewers で一時停止 → 承認後に deploy/migrate/health が成功
# 止まる: 承認無しで完走した → ゲート不全（Environment を再確認）。deploy 失敗 → ログは失敗 step のみ（secret 無し）

# 4) migrate が必要な場合のみ（不要ならスキップ）
# WORKER_URL=https://api.noah-karte.com ./infra/scripts/cf-run-migrate.sh
# 次へ進む: スクリプト終了コード 0
# 止まる: 非 0 → 成功扱いにせず原因を記録（credential 無し）

# 5) /health（契約: HTTP 200 かつ JSON {"status":"ok"}。DB 依存無し。実装: backend/cmd/api/base_routes.go）
curl -sS -o /tmp/prod-health-rollback.json -w '%{http_code}\n' https://api.noah-karte.com/health
jq -r '.status' /tmp/prod-health-rollback.json
# 次へ進む: 1 行目が 200 かつ status が ok
# 止まる: それ以外 → ローリング残留（最大 15 分静置して再 curl）または §2 初動へ

# 6) RTO 実測を記録（判断宣言 → /health ok の分）。schema 確認分も別行で記録
date -u +%Y-%m-%dT%H:%M:%SZ
```

**受け入れチェック（実施後も未実施なら `[ ]` のまま）**

- [ ] 対象: 非ピークまたは隔離検証枠
- [ ] last known good 相当の再デプロイを実施（本番データ破壊操作は含めない）
- [ ] 計測: 判断宣言 → `/health` ok までの分（RTO 実測）
- [ ] schema 互換確認の所要時間も記録
- [ ] 記録場所: 当日作業ログまたは delivery 証跡（credential 無し）

**失敗時分岐**

| 症状 | 分岐 |
|---|---|
| Environment 承認が挟まらない | setup.md §7 をやり直し。無承認 deploy を受け入れにしない |
| `/health` が 200/ok にならない | §2 初動。15 分静置後に再確認。ECS/AWS へは戻さない |
| schema 非互換 | forward-fix または承認済みスナップショット。破壊的 reset は別途明示承認 |
| provider 障害で再デプロイ不能 | §3 手順 7（IaC から再建）。DNS を旧インフラへ戻すことを復旧とみなさない |

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

**前提（満たさない場合は本節を開始しない）**

1. production 向けスナップショットが **少なくとも 1 本** 取得済みであること（PlanetScale マネージド **または** 明示 `pg_dump`。取得自体は USER 専権）
2. 復元先は **隔離環境のみ**（本番 DNS / 本番 DB を直接ターゲットにしない。本番上書き restore は別途明示承認）
3. 接続情報は TTL 付き診断ロール等、使い捨て経路から端末環境変数へ載せる（文書・ログに残さない）

**repo 側 tooling**: STG/PROD 用の backup / restore 実行スクリプトは **存在しない**（`scripts/` に該当名 0 本、`pg_restore` 0 ヒット）。`make reset` 内の `pg_dumpall` は **ローカル Docker 専用**であり production 経路ではない。以下はオペレータが外部 CLI で打つ雛形。

#### 手順（逐次・判定付き）

```bash
# 0) 開始時刻（所要時間の起点）。PHI・credential は書かない
date -u +%Y-%m-%dT%H:%M:%SZ
# 次へ進む: 作業ログに開始時刻が残った
# 止まる: 隔離先未用意・本番をターゲットにしかねない状態 → 中止

# 1) 隔離先の識別子を端末だけで持つ（例。実名は文書に焼かない）
# export ISOLATED_DB_HOST=...
# export ISOLATED_DB_NAME=...
# export ISOLATED_DB_USER=...
# export PGPASSWORD=...          # 履歴・runbook に残さない
# 次へ進む: ホスト名/DB 名が production の DNS・DB 名と一致しないことを目視確認した
# 止まる: production と同一ホスト/DB を指している → 中止

# 2) スナップショットの存在確認（マネージド UI またはオブジェクト保管。コマンドは提供元 CLI に従う）
#    例（PlanetScale マネージド）: コンソールで最新 backup の時刻・サイズのみ確認
#    例（明示 dump ファイルがある場合）:
# ls -lh "<path-to-snapshot>"     # パスは端末ローカル。repo に置かない
# 次へ進む: スナップショットが 1 本以上あり、取得時刻が記録できる
# 止まる: 0 本 → 先に backup 取得（本節の restore を成功扱いにしない）

# 3) 隔離環境へ restore（本番をターゲットにしない）
#    PlanetScale マネージド restore なら提供元 UI/CLI の「別ブランチ / 別 DB」へ復元
#    ファイル dump の場合の雛形（接続先は隔離のみ）:
# pg_restore --clean --if-exists -h "$ISOLATED_DB_HOST" -U "$ISOLATED_DB_USER" -d "$ISOLATED_DB_NAME" "<path-to-snapshot>"
# または: gunzip -c "<path>.sql.gz" | psql -h "$ISOLATED_DB_HOST" -U "$ISOLATED_DB_USER" -d "$ISOLATED_DB_NAME"
# 次へ進む: 終了コード 0
# 止まる: 非 0 → 成功扱いにせず原因を記録（接続文字列・秘密値は載せない）

# 4) 非 PHI 指標で整合性確認（個人名・飼主名・ペット名・スタッフ名は出さない）
#    以下は指標例。テーブル名は現行 schema に合わせて USER が選ぶ。結果は件数と合計のみ記録。
# psql -h "$ISOLATED_DB_HOST" -U "$ISOLATED_DB_USER" -d "$ISOLATED_DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
# -- 主要テーブル件数（例。実テーブル集合は環境の schema に合わせる）
# SELECT 'owners' AS entity, count(*)::bigint AS n FROM owners
# UNION ALL SELECT 'pets', count(*) FROM pets
# UNION ALL SELECT 'appointments', count(*) FROM appointments;
# -- clinic_id 別件数（値は ID と件数のみ。名称結合しない）
# SELECT clinic_id, count(*)::bigint AS n FROM pets GROUP BY clinic_id ORDER BY clinic_id;
# -- 金額合計（該当テーブルがある場合のみ。PHI 列を SELECT しない）
# -- SELECT coalesce(sum(total_amount),0) AS amount_sum FROM accounting_invoices;
# SQL
# 次へ進む: クエリがエラー無く終わり、件数/合計が作業ログに残った（氏名カラムを出していない）
# 止まる: エラー、または氏名等の PHI が結果に混入 → 記録を破棄してクエリを修正

# 5) 所要時間を記録し、隔離資源を破棄
date -u +%Y-%m-%dT%H:%M:%SZ
# 隔離 DB / 一時リストア先 / 一時認証情報を破棄（提供元の削除手順に従う）
# unset PGPASSWORD ISOLATED_DB_HOST ISOLATED_DB_NAME ISOLATED_DB_USER
# 次へ進む: 破棄完了を作業ログに残した（credential 無し）
# 止まる: 隔離資源が残存 → 破棄完了まで受け入れにしない
```

**受け入れチェック（実施後も未実施なら `[ ]` のまま）**

- [ ] 隔離環境を用意（本番 DNS / 本番 DB を直接ターゲットにしない）
- [ ] スナップショットから restore
- [ ] 非 PHI 指標で整合性確認: 主要テーブル件数、clinic_id 別件数、金額合計（個人名は出さない）
- [ ] 所要時間を記録
- [ ] 失敗時は成功扱いにせず原因を記録（credential は載せない）
- [ ] リハーサル用に作った隔離資源の破棄

**失敗時分岐**

| 症状 | 分岐 |
|---|---|
| スナップショット 0 本 | 取得を先に完了。restore をスキップして成功扱いにしない |
| restore が production を向いている | **即中止**。接続先を隔離に切り替えてから再開 |
| 整合性クエリが PHI を返す | 結果を保存せずクエリを件数/合計のみに修正 |
| 所要時間が未記録 | 受け入れ未完了。再計測 |

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

| AC | docs/prep（実測で確認できた成果物のみ） | 実地 | 備考 |
|---|---|---|---|
| latest main required CI green | 候補 job 列挙済: `ci.yml` の `changes` / `secret-scan` / `backend` / `frontend` / `worker` / `codegen-check` / `migration-verify`（詳細と paths-filter 罠は [todo-po.md](../../../../todo-po.md) #253（詳細は git 履歴 / Issue #253）） | **BLOCKED** | GitHub billing/spending — USER only。**安全な常時 required 候補は `secret-scan`（表示名 Gitleaks Secret Scan）と必要なら `changes` のみ**。他 5 job は path/`if` で skip され required にすると永久 pending になり得る |
| STG deploy / health / failure notification 実地確認 | 契約: [CI-CD-PIPELINE.md](../../deploy/CI-CD-PIPELINE.md) §0、本 runbook §1/§4 | billing 復旧後 | 本 unit は STG runtime を叩いていない |
| production deploy は Required reviewers なしに開始できない | 契約・手順: setup.md §7、本 runbook §1.1。**workflow 実測: `environment:` 未設定** | Environment 未作成 | setup.md §7 → §8 の順。§8 提案 diff は未適用 |
| rollback rehearsal + 復旧時間記録 | 本 runbook **§3.1**（コマンド列 + 判定基準）。専用スクリプトは **不在** | 環境構築後 | CF-only。チェックボックスは USER 実施まで `[ ]` |
| backup → 隔離 restore + 非 PHI 整合性 | 本 runbook **§5.1**（コマンド列 + 判定基準）。PROD backup/restore スクリプトは **不在** | 環境構築後 | 隔離のみ。本番上書きは別途明示承認 |
| log/artifact/Issue に credential・個人情報なし | 本 runbook 方針（ヘッダ / §3 / §5 / 本節） | 継続 | agent も遵守 |
| #257 Go-live の明示 prerequisite | [GOLIVE_RUNBOOK.md](../../../delivery/GOLIVE_RUNBOOK.md) §1 #2/#7 | 上記完了後 | 納品日 2026-08-03。#253 は直接前提 |

### 8.1 USER 直列 8 段（実行は USER 専権。agent は 1 段も実行しない）

詳細な前提・コマンド・判定・失敗分岐は [todo-po.md](../../../../todo-po.md) #253（詳細は git 履歴 / Issue #253） を正本とする。

1. GitHub Actions の billing / spending 復旧
2. GitHub Environment `production` 作成（setup.md §7）
3. Required reviewers 設定（setup.md §7）
4. required status check の指定（上表の候補から USER が選定。paths-filter 罠に注意）
5. production deploy トリガー適用（setup.md §8 提案 diff。§7 完了後のみ）
6. deploy 実行と `/health` 確認（HTTP 200 + `{"status":"ok"}`）
7. backup 取得 → 隔離 restore → 非 PHI 整合性 → 所要時間記録（§5.1）
8. rollback rehearsal → RTO 実測 → 隔離資源の破棄（§3.1）

**agent は支払い・spending limit・secret 実値の発行を行わない。**
未実施の作業を実施済みと書かない。本 runbook の `- [ ]` を agent が `[x]` にしない。
