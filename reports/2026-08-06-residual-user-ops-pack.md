# Residual USER Ops Pack — AnimalEkarte

| 項目 | 値 |
|------|-----|
| **Date** | 2026-08-06 |
| **Ledger SoT** | [`STATUS.md`](../STATUS.md) §1 |
| **Companion board** | [`2026-08-06-residual-agent-board.md`](./2026-08-06-residual-agent-board.md) |
| **Agent product open** | **0**（製品コードの agent 実装 open は 0。残は USER / ops / PO / 臨床 gate のみ） |
| **Team evidence** | env inventory · code audit · agent board（同日 `reports/2026-08-06-residual-team-*`） |

> 本パックは **agent がプロジェクト規則上実行できない** residual を、オペレータ 1 人が順に潰すための単一入口である。  
> 長文の重複は避け、既存レポートへリンクする。

### 2026-08-06 residual team 追記（local 実測）

| 事実 | 含意 |
|------|------|
| `seeds/003_demo` checksum disk ≠ DB | local は **`make migrate` 不可** → **`make reset` 必須**（TASK-378 + 009） |
| `exam_reference_ranges` COUNT = 0 | BUG-003 ブラウザ前に reset 成功が必要 |
| backend Exited(1) / frontend 未起動 | reset 後 `make up` |
| 過去の hosp `owner_request` NULL crash | **`7b929231a` FORCE_NOT_NULL で修正済み**。reset 再試行のコード側ブロッカーではない |
| `claim/*` = 0 | SCEN-OPS-CLAIM クローズ |
| `E2E_LOGIN_*` UNSET | 020/023 は credential 注入まで BLOCKED |
| IU 32 CODE_PRESENT | 追加の product 実装 unit なし |
---

## 0. Preflight checklist（実行前）

ホストで順に確認。**秘密の値は本ファイルに書かない・貼らない。**

```bash
# 作業ディレクトリ
cd /path/to/AnimalEkarte   # 共有 main 作業ツリー

# Docker / Compose
docker version
docker compose version

# .env.local（Make/Compose の SoT。無い場合は .env.example から）
test -f .env.local && echo "ENV_LOCAL=OK" || echo "ENV_LOCAL=MISSING"
# 未作成時のみ: cp .env.example .env.local  （値は人手で埋める）

# スタック状態（project 名 animalekarte 想定）
docker compose --env-file .env.local ps
# 未起動なら: make up

# Health
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/health   # expect 200
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:3003/          # expect 200

# LIFF mock は compose 固定済み（agent 実装済）
# docker-compose.yml: backend LIFF_MOCK=true / frontend VITE_LIFF_MOCK=true
# 変更後は frontend + backend 再起動が必要:
#   docker compose --env-file .env.local up -d --wait backend frontend

# E2E 資格情報（値は表示しない）
test -n "${E2E_LOGIN_EMAIL:-}" && echo "E2E_EMAIL=SET" || echo "E2E_EMAIL=UNSET"
test -n "${E2E_LOGIN_PASSWORD:-}" && echo "E2E_PASSWORD=SET" || echo "E2E_PASSWORD=UNSET"

# seed static（DB 不要）
python3 scripts/verify_seed.py   # expect exit 0 / OK
```

| チェック | 期待 |
|----------|------|
| Docker daemon | 起動中 |
| `.env.local` | 存在（Git 管理外） |
| `db` / `backend` / `frontend` | healthy（`make up` 後） |
| `LIFF_MOCK` / `VITE_LIFF_MOCK` | compose で `"true"` 固定（BUG-008/014） |
| `E2E_LOGIN_*` | 認証 E2E / 五フロー前に **SET**（secret channel で注入） |
| `verify_seed.py` | static GREEN（2026-08-01 済） |

**Agent がやらないこと（本パックでも自動実行しない）**: `make migrate` / `make seed` / `make reset` / force-push / claim ブランチ削除 / 秘密値の発明・記録 / `VERIFIED_FIXED` 付与。

---

## 1. Ordered execution plan（STATUS.md 推奨 USER 順）

```
TASK-009
  → POST-PULL / TASK-032-apply / TASK-374-apply
  → TASK-378-reset（必要な環境だけ）
  → E2E_LOGIN_* 注入
  → TASK-020 / TASK-023
  → TASK-010 + BROWSER_VERIFICATION_BACKLOG
  → TASK-022 / TASK-024（human）
  → TASK-033 clinical gate
  → TASK-021 PO gate
```

並行可: §6 SCEN-OPS-CLAIM（マージ済み claim の棚卸し）、§5 HOLD の credential 系（#89/#97…）は別レーン。

---

### Step 1 — TASK-009: 003_demo seed の **DB 適用**

| | |
|--|--|
| **Owner** | USER |
| **Agent 済** | static CSV + `verify_seed.py` GREEN; clinical CSV / `exam_reference_ranges` は tree 上 |
| **詳細 runbook** | [`2026-07-31-task-009-reseed-ops.md`](./2026-07-31-task-009-reseed-ops.md) / [`2026-08-01-task-009-verify-seed-green.md`](./2026-08-01-task-009-verify-seed-green.md) |

**Commands**

```bash
# ディスク確認（read-only）
python3 scripts/verify_seed.py
ls backend/migrations/seeds/003_demo/exam_reference_ranges.csv 2>/dev/null || true
wc -l \
  backend/migrations/seeds/003_demo/hospitalizations.csv \
  backend/migrations/seeds/003_demo/treatment_plans.csv \
  backend/migrations/seeds/003_demo/daily_records.csv \
  backend/migrations/seeds/003_demo/care_plan_items.csv

# ケース A: 空 DB / 003_demo 未適用
make migrate

# ケース B: 既に 003_demo 適用済みで checksum mismatch または CSV 再 COPY が必要
# ⚠️ local 専用・破壊的（postgres volume 再構築）
make reset
# 必要なら差分確認: make migrate
```

**Success**

- migrate/reset ログで seed `003_demo` load（skip ではなく適用経路）または postflight `missing=0`
- 入院ボード: 入院中「ミルク」/ 退院「チョコ」等（reseed-ops §5）
- backend `/health` 200

**Rollback**

- ケース A 失敗: ログ確認のみ。手書き SQL COPY 禁止。
- ケース B: `make reset` 直前の `.local-db-backups/<UTC>/` を契約どおり保持。STG/PROD では **reset 禁止**（別承認 runbook）。

---

### Step 2 — POST-PULL / TASK-032-apply / TASK-374-apply: migrate

| | |
|--|--|
| **Owner** | USER |
| **Agent 済** | lab import / checkup package 等の **製品コード・migration は main 済み**。残は適用と claim 解放 |
| **参照** | STATUS §1; [`docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md`](../docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) |

**Commands**

```bash
git pull   # または既に最新を確認
# migration 変更有無（pull 後）
git log -5 --oneline -- backend/migrations/

# 各 local 環境で差分 migrate（DB は落とさない）
make migrate
```

| ID | 意味 | 成功の見方 |
|----|------|------------|
| **POST-PULL** | pull 後に未適用 DDL/seed を適用 | `Migration summary` / `missing=0` 系 |
| **TASK-032-apply** | lab import 補償 migration の **適用** | 新規 key が `schema_migrations` に記録（または 001 統合済みなら key coverage で充足） |
| **TASK-374-apply** | checkup package import migration の **適用** | 同上 |

**Success**

- `make migrate` が 0 終了
- backend healthy のまま
- 関連 claim が live なら integrate 確認後に USER が解放（§6）

**Rollback**

- 差分 apply 失敗: backend ログと checksum メッセージを確認。  
- **001 統合後 checksum mismatch** → Step 3（TASK-378-reset）。手書き SQL で schema を直さない。

---

### Step 3 — TASK-378-reset: 001 統合後 DB_RESET（**必要環境のみ**）

| | |
|--|--|
| **Owner** | USER |
| **いつ必要か** | 統合前 001 が記録済みで、統合後 001 checksum が不一致 → 通常 `make migrate` が fail-closed |
| **正本** | [`docs/ops/deploy/LOCAL_DB_RESET.md`](../docs/ops/deploy/LOCAL_DB_RESET.md) |

**Commands**

```bash
# local のみ
make reset
# 成功後の検算（任意）
curl -sS http://127.0.0.1:8080/health
ls backend/migrations/*.sql | wc -l   # 行数期待の固定値照合はしない。coverage missing=0 が一次判定
```

**Success**

- contract postflight 通過（DDL + seed keys `002_master` / `003_demo` / `004_staging` + `/health` 200）
- `Migration key coverage missing=0`

**Rollback / 注意**

- STG/PROD で `make reset` **禁止**
- snapshot は `.local-db-backups/`。cache volume は保持される
- 不要な環境では **実行しない**（既に fresh なら Step 2 のみ）

---

### Step 4 — E2E_LOGIN_* 注入 → TASK-020 / TASK-023

| | |
|--|--|
| **Owner** | USER |
| **Agent 済** | `run-e2e.sh` が host `E2E_LOGIN_*` を name-only `-e` で Docker へ転送（[`2026-07-31-task-020-env-forward.md`](./2026-07-31-task-020-env-forward.md)） |
| **資格情報** | secret channel で **local demo admin** を shell に export。値を report / git に書かない |

**Commands**

```bash
# 値なし preflight（SET/UNSET のみ）
test -n "${E2E_LOGIN_EMAIL:-}" && test -n "${E2E_LOGIN_PASSWORD:-}" && echo E2E_OK || echo E2E_BLOCKED

# 注入（値はローカルのみ。例の文字列をコミットしない）
export E2E_LOGIN_EMAIL='…'       # target DB に存在する demo admin
export E2E_LOGIN_PASSWORD='…'    # never commit

# アプリ到達前提
make up   # 未起動時

# TASK-020: フル or 代表 Playwright runtime
make e2e
# または: bash frontend/scripts/run-e2e.sh
# 絞り込み例: make e2e ARGS='e2e/auth-flows.spec.ts e2e/owners-search.spec.ts'

# TASK-023 / #254: 五フロー向け subset
bash frontend/scripts/run-e2e.sh \
  e2e/clinical-flows.spec.ts \
  e2e/examinations-flow.spec.ts \
  e2e/accounting-flow.spec.ts \
  e2e/reservations-smoke.spec.ts \
  e2e/trimming-flow.spec.ts \
  e2e/line-reservation-flow.spec.ts
```

**Human UAT（TASK-023）** — 認証後ブラウザ通し:

| # | 業務 | 主シナリオ |
|---|------|------------|
| 1 | 外来 / 検査 / 会計 / 締め | S02, S06–S09 |
| 2 | 予約 → 受付 → 再予約 | SECTION_14 / V02 |
| 3 | トリミング + 診察会計 | S11 |
| 4 | LINE 予約 → 記録 | S04, S12（**LIFF mock** で local） |
| 5 | 月次集計 / report | S10 |

証跡テンプレ: [`docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`](../docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md)

**Success**

- 認証付き Playwright が credential 不足で即死しない
- TASK-020: 実行 suite の pass/fail がレポート可能
- TASK-023: 五フロー + DB/audit 目視 + 必要なら実機 LINE（実機は別 gate）の **named sign-off**

**Rollback**

- 失敗時: seed/migrate 前提を Step 1–3 に戻す。偽 credential を埋め込まない。
- rate limit: フル suite は storage-state キャッシュ（`E2E_AUTH_STATE_PATH`）を利用。

---

### Step 5 — TASK-010 + BROWSER_VERIFICATION_BACKLOG

| | |
|--|--|
| **Owner** | USER（ブラウザ） |
| **Agent 済** | 実装 IU 32 件; 一部 runtime 再分類・census; エージェントは `VERIFIED_FIXED` を付けない |
| **表** | [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md) |
| **要実測残** | シナリオ側 residual: [`2026-08-01-task-010-batch5.md`](./2026-08-01-task-010-batch5.md) |

**Commands / 手順**

```bash
# stack + seed 適用済み前提
open http://localhost:3003   # または Chrome で http://127.0.0.1:3003
# LIFF 系 (BUG-008/014): compose mock 変更後は再起動済みであること
```

1. Backlog **§A 優先バッチ**（campaign 実装分 + BUG-003/008/014）から実施
2. 結果を表の「結果」列に `PASS|FAIL|BLOCKED|WAIVED` 記入
3. 続けて §B 既存 IU、§C 残ゲート
4. PASS のみ STATUS §3 の `原文シナリオ再検証` 更新を検討（人間判断）
5. TASK-010 の `【要実測】` はシナリオ MD を実測で潰し、source-only PASS を runtime と混同しない

**Success**

- 優先バッチの UNREPORTED が減る
- BUG-003: seed 適用後 S02 H/L
- BUG-008/014: mock で S04/S12 が 401 で落ちない

**Rollback**

- FAIL → 個票 OPEN 戻し or 新規 BUG。データ汚染時は local `make reset`（必要時のみ）

---

### Step 6 — TASK-022 / TASK-024 human

| ID | 内容 | 参照 |
|----|------|------|
| **TASK-022** | #239 S13 手動 correction + named signer +（必要なら）RLS runtime | [`S13-identity-links-manual-correction.md`](../docs/ops/testing/scenarios/S13-identity-links-manual-correction.md) / [`2026-07-31-task-022-identity-link-closeout.md`](./2026-07-31-task-022-identity-link-closeout.md) |
| **TASK-024** | #256 screenshot / FAQ visual sign-off | [`2026-07-31-task-024-manual-audit.md`](./2026-07-31-task-024-manual-audit.md) |

**TASK-022 commands / 手順**

```bash
# ブラウザで /identity-links — 2 clinic + identity-links:view+edit のスタッフ
# S13 手順 1–8 を実施し、シナリオ末尾の HUMAN サインオフ表を埋める
```

**TASK-024**

- agent は 7 replace / 3 current 済み。**named documentation owner** が product UI と PNG を突合して sign-off
- FAQ は TASK-023 confusions=0 のため **追記不要**（agent 判断済み）
- clean-seed での 05/07/10 再撮影が必要な場合は seed 適用後ウィンドウで実施

**Success**

- S13 表: 実施日・実施者・承認者（記名）が埋まる
- 024: 10 画像 visual sign-off 完了

**Rollback**

- 誤 link: S13 の unlink 手順。PHI を repo に書かない。

---

### Step 7 — TASK-033 clinical gate（#201）

| | |
|--|--|
| **Owner** | 臨床責任者 + USER |
| **Agent** | **実装開始禁止** until gate |
| **Issue** | [#201](https://github.com/MinoruSoga/AnimalEkarte/issues/201) |

**Gate 条件（すべて）**

1. 臨床 SoT: 上限 / warning 継続可否 / 救急記録 policy を **canonical #201 bundle 1 行** に出典付き記入
2. decision SoT と DB review 完了
3. その後に限って agent 再開可（1 unit = 1 graph）

**Commands**

- コード変更コマンドなし。Issue / q&a への **非秘密の 1 行結果 + opaque ref** のみ。

**Success**

- #201 判断待ちが解除され、STATUS 上 TASK-033 が agent 着手可能になる

**Rollback**

- 未記入のまま cutover 実装を開始しない（HOLD 維持）

---

### Step 8 — TASK-021 PO gate（exclusion 破壊削除）

| | |
|--|--|
| **Owner** | USER + PO |
| **Agent 済** | Phase1 prep / Phase2 slice1–2（positive capability / write で deprecated 拒否） |
| **参照** | [`2026-07-31-task-021-phase2-slice2.md`](./2026-07-31-task-021-phase2-slice2.md) / line residual FINAL |

**Gate 条件**

1. external use 確認（UNREPORTED のまま破壊しない）
2. 破壊承認（table / seed / DROP / CLEAN-GO）
3. 承認後のみ DROP migration 設計・apply は **別 packet + USER `make migrate`**

**Success**

- 明示承認記録 + external-use 証拠後に CLEAN-GO

**Rollback**

- 承認前は dual surface 維持。agent は DROP を勝手に作らない・適用しない。

---

## 2. Explicit HOLD list

| ID | 内容 | 解除条件 |
|----|------|----------|
| **TASK-033** | 救急投薬 cutover 実装 | 臨床 SoT + decision + DB review（#201） |
| **TASK-021** | exclusion 破壊削除 / DROP | external use + PO 破壊承認 |
| **LINE-R05** | production rollout + column DROP | inventory/rollout gate 後の別 packet（Phase B は code green、DROP は HOLD） |
| **#89** | 露出 credential ローテーション | USER 専権 playbook。値を repo に書かない |
| **#97** | git/公開面由来 credential 露出 | 旧値失効 + 必要 session 無効化（#89 と役割分離） |
| **#98** | 旧 RDS credential 履歴 residual | 無効化 **または** residual-risk 受容を明示して close 判断 |
| **#99** | 旧 ECS deploy 経路撤去確認 | provider に実行可能経路なし確認。rollback SoT は #253 と一本化 |

> 上記 HOLD を agent に「とりあえず実装」させない。credential 系は **edit/mask だけでは close 禁止**。

---

## 3. SCEN-OPS-CLAIM — claim ブランチ一覧と削除（USER only）

Agent は claim を **削除しない**。main 統合または明示 abandon 後に USER 端末で:

```bash
# 一覧
git branch --list 'claim/*'
git branch -a --list 'claim/*'

# 統合済みか確認（例）
git log main --oneline --grep='TASK-020' -n 5
# または PR / land 証跡を確認

# ローカル claim 削除（統合 or abandon 後のみ）
git branch -D claim/<NAME>

# リモートに push 済みの claim がある場合のみ（慎重に）
# git push origin --delete claim/<NAME>
```

| ルール | |
|--------|--|
| 削除タイミング | **integrate とセット**、または明示 abandon |
| 禁止 | agent 自動 delete / integrate 前の解放（他セッション衝突） |
| 例名 | `claim/TASK-010`, `claim/TASK-020`, `claim/TASK-021`, `claim/W-020-ENV`, LINE 系 等 — **live を `git branch --list` で実測** |

---

## 4. Copy-paste: minimum local UAT path（ブラウザバッチ unblock）

認証 E2E と BROWSER_VERIFICATION を最短で開ける経路。

```bash
cd /path/to/AnimalEkarte

# 0) env
test -f .env.local || cp .env.example .env.local
# … 必要キーを人手で設定（秘密はここに書かない）

# 1) stack
make up

# 2) seed / schema（fresh なら migrate、checksum 問題なら reset）
python3 scripts/verify_seed.py
make migrate
# 失敗で checksum mismatch かつ local のみ:
# make reset

# 3) LIFF mock 反映が必要なとき（compose 既定 true）
docker compose --env-file .env.local up -d --wait backend frontend

# 4) health
curl -sS http://127.0.0.1:8080/health
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:3003/

# 5) E2E（任意・認証）
export E2E_LOGIN_EMAIL='…'
export E2E_LOGIN_PASSWORD='…'
make e2e ARGS='e2e/auth-flows.spec.ts e2e/owners-search.spec.ts'

# 6) ブラウザ手動バッチ
# open http://localhost:3003
# → reports/BROWSER_VERIFICATION_BACKLOG.md §A から記入
# → BUG-003: 検査 H/L / BUG-008: S04 / BUG-014: S12
```

これで **seed 未適用 / LIFF 401 / E2E 資格情報不足** の三重ブロックを外し、ブラウザ IU バッチに入れる。

---

## 5. Agent team: completed vs cannot complete

### 既に agent 側で完了しているもの（抜粋）

| 領域 | 状態 |
|------|------|
| 製品コード open residual | **0**（STATUS 2026-08-06） |
| BUG-001〜032 実装 | **IMPLEMENTED_UNVERIFIED** 32（ブラウザ未） |
| LIFF mock compose 固定 | `LIFF_MOCK` / `VITE_LIFF_MOCK` |
| demo `exam_reference_ranges` seed ファイル | tree 上; **適用は USER** |
| TASK-009 static verify | GREEN |
| TASK-020 env forward | `run-e2e.sh` 転送済み |
| TASK-021 Phase1–2 slices | 破壊 DROP 以外 |
| TASK-022 認可修正 + S13 シナリオ文書 | human sign-off 残 |
| TASK-023 env gate 骨格レポート | UAT sign-off 残 |
| TASK-024 PNG 監査 phase A | named visual sign-off 残 |
| TASK-010 census / 再分類 | 要実測の runtime 残 |
| LINE R-05 Phase A/B code | production DROP **HOLD** |
| migration 製品コード（032/374/378 統合含む） | main; **apply/reset は USER** |

### agent が完了できないもの（本パックの対象）

| 種別 | 例 |
|------|-----|
| DB 破壊・適用 | `make migrate` / `seed` / `reset` / `DB_RESET` / 直接 psql 書込 |
| 秘密 | credential 生成・ローテーション・report への値記載（#89/#97/#98） |
| 本番 / 課金 / 物理 | #253 billing, 実機 LINE/LIFF, provider 確認 #99 |
| 臨床・PO 決裁 | TASK-033 / #201, TASK-021 DROP, LINE-R05 rollout |
| 人証跡 | S13 記名, #254 五フロー, #256 screenshot, FAQ 最終, `VERIFIED_FIXED` |
| git 危険操作 | force-push, claim 自動削除, history rewrite |
| ブラウザ最終判定 | BROWSER_VERIFICATION 全バッチの PASS/FAIL 記入責任 |

---

## 6. Quick link index

| 用途 | パス |
|------|------|
| 作業台帳 | [`STATUS.md`](../STATUS.md) |
| ブラウザ表 | [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md) |
| Local reset | [`docs/ops/deploy/LOCAL_DB_RESET.md`](../docs/ops/deploy/LOCAL_DB_RESET.md) |
| Seed ops | [`2026-07-31-task-009-reseed-ops.md`](./2026-07-31-task-009-reseed-ops.md) |
| E2E README | [`frontend/e2e/README.md`](../frontend/e2e/README.md) |
| #254 証跡 | [`docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`](../docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md) |
| Kanban | [`2026-08-06-residual-agent-board.md`](./2026-08-06-residual-agent-board.md) |
| Ledger note | [`2026-08-06-todo-ledger-reconciliation.md`](./2026-08-06-todo-ledger-reconciliation.md) |

---

*Generated for residual USER ops only. No secrets. No agent migrate/reset/claim-delete.*
