# TASK-009 — 003_demo clinical slice1 再シード運用手順（USER 向け）

> **Date**: 2026-07-31  
> **Unit**: `TODO-MD-OPEN-REMAINING-ORCH-WAVE-20260731-V2` / W-009-OPS  
> **Seed commit**: `c286bfe0a` — `chore(seed): add 003_demo clinical CSV slice1 for TASK-009`  
> **Binding**: `reports/2026-07-31-task-009-seed-design.md`, `reports/2026-07-31-task-009-slice1.md`  
> **Agent 権限**: 本レポートは手順書のみ。DB 操作・migrate・reset は **USER 専用**。

---

## 1. 目的

TASK-009 slice1 で `003_demo` の臨床 CSV 4 本にデータ行が追加された。  
**既にローカルで `003_demo` を適用済みの DB** では、`schema_migrations` に記録されたバンドル checksum とディスク上の新 CSV が不一致となり、`make migrate` だけでは再 COPY されない（または checksum mismatch で fail-closed する）。

本ドキュメントは、**USER がローカル環境で slice1 を正しく DB に反映する手順**を、ケース別に一箇所にまとめたものである。

| ケース | 推奨経路 | データ損失 |
|:---|:---|:---|
| 空 DB / 未適用 `003_demo` | `make migrate` | なし（適用のみ） |
| 既に `003_demo` 適用済み | **`make reset`**（local 専用） | **あり**（postgres volume 再構築） |

---

## 2. ディスク上 CSV の確認（read-only）

作業ツリー root で、slice1 対象 4 ファイルが存在し、ヘッダのみでないことを確認する。

```bash
# 存在確認
ls -la \
  backend/migrations/seeds/003_demo/hospitalizations.csv \
  backend/migrations/seeds/003_demo/treatment_plans.csv \
  backend/migrations/seeds/003_demo/daily_records.csv \
  backend/migrations/seeds/003_demo/care_plan_items.csv

# 行数（header + data）。slice1 期待: 各 3 行（data 2 行）
wc -l \
  backend/migrations/seeds/003_demo/hospitalizations.csv \
  backend/migrations/seeds/003_demo/treatment_plans.csv \
  backend/migrations/seeds/003_demo/daily_records.csv \
  backend/migrations/seeds/003_demo/care_plan_items.csv
```

**W-009-OPS 検証（2026-07-31, read-only）**: 4 CSV いずれもディスク上に存在し、データ行あり。

| ファイル | 期待行数 (`wc -l`) | 内容概要 |
|:---|---:|:---|
| `hospitalizations.csv` | 3 | id=1 admitted（ミルク）、id=2 discharged（チョコ） |
| `treatment_plans.csv` | 3 | record 側 + hosp 側 各 1 |
| `daily_records.csv` | 3 | hosp id=1 に 2 日分 |
| `care_plan_items.csv` | 3 | food / instruction 各 1 |

任意の静的検証（DB 不要）:

```bash
python3 scripts/verify_seed.py
```

---

## 3. ケース A — 空 DB / まだ `003_demo` が未適用

Docker stack が利用可能な状態で:

```bash
# 1) サービス確認（USER 管理）
docker compose --env-file .env.local ps

# 2) マイグレーション + 未適用 seed バンドル適用
make migrate
```

- `make seed` は migrate と同一経路（差分・べき等）。空 DB では `make migrate` で足りる。  
- 未記録の `003_demo` は COPY ロードされ、checksum が `schema_migrations` に記録される。  
- 成功ログに `003_demo` の適用（skip ではなく load）が出ることを確認する。

**このケースでは `make reset` は不要。**

---

## 4. ケース B — 既に `003_demo` 適用済み（checksum mismatch 想定）

### なぜ `make migrate` だけでは足りないか

seed バンドルは `schema_migrations` の filename + checksum で idempotent 管理される。

- **checksum 一致** → `⏭ Skipping seed bundle (already applied)` — **再 COPY しない**  
- **checksum 不一致** → fail-closed（適用後改変検出）— **自動再適用しない**

CSV を後から更新したローカルでは、**bare `make migrate` では slice1 行は DB に入らない**。  
完全再ロードが必要 → local 専用の `make reset`。

### ⚠️ データ損失警告（必読）

`make reset` は **ローカル専用・破壊的** 経路である（`scripts/local-db-reset-contract.sh`）。

実行内容の要約:

1. project / volume 名を固定値と照合（非 local は拒否）  
2. `.local-db-backups/<UTC>/` に pg_dumpall + SHA-256 + manifest（失敗時は volume 削除へ進まない）  
3. **`ekarte-postgres-data` volume のみ削除**（cache volume は保持）  
4. 再起動後、DDL + seed keys `002_master` / `003_demo` / `004_staging` + `/health` を fail-closed 確認  

**失われるもの**: ローカル DB 上の手作業データ・未バックアップの実験行すべて。  
**失われないもの（契約上）**: frontend node_modules / go module・build cache volume。

### 手順

```bash
# 事前確認
docker compose --env-file .env.local ps
git log -1 --oneline c286bfe0a   # slice1 commit が tree に含まれること

# ⚠️ 破壊的。ローカル専用。USER のみ実行。
make reset
```

完了後、postflight が `003_demo` を含む seed key を missing=0 で通すこと。必要なら続けて:

```bash
make migrate   # reset 後の差分確認用（通常は reset 内で適用済み）
```

---

## 5. 適用後スモークチェックリスト

clinic_id=1（八王子病院）のデモ staff でログインし、次を確認する。

| # | 確認項目 | 期待 |
|---:|:---|:---|
| 1 | ログイン | clinic 1 デモ staff で成功 |
| 2 | カルテ一覧 | 既存 G1 dump 行（例: medical_record 周辺）が表示される |
| 3 | 入院ボード | **入院中** id=1・ペット「ミルク」(pet 1000025) |
| 4 | 入院ボード | **退院** id=2・ペット「チョコ」(pet 1000026) |
| 5 | 入院詳細（任意） | daily records（2 日）/ care plan / treatment plan タブに TASK-009 行 |
| 6 | API 健康 | backend `/health` が 200（reset 契約でも確認） |

問題時の切り分け（USER）:

- 4 CSV がディスク上で `wc -l` = 3 か  
- `schema_migrations` に `003_demo` があり、現 checksum と一致するか（mismatch ならケース B）  
- 手作業で SQL COPY を足さない（契約外・二重投入リスク）

---

## 6. エージェント非実施リスト（明示）

本タスクおよび関連 agent は、次を **実行しない・提案して自動実行もしない**。

| 非実施 | 理由 |
|:---|:---|
| `make migrate` | seed apply は USER 手動 |
| `make seed` | migrate と同一経路・USER 手動 |
| `make reset` | 破壊的・local 専用・USER のみ |
| `make db` / 直接 `psql` 書き込み | 高 side-effect・エージェント禁止コマンド群 |
| `DB_RESET` / volume 手動 `docker volume rm` | reset 契約外の危険操作 |
| STG/PROD への migrate / seed / reset | 別承認が必須（§7） |
| claim ブランチの削除 | USER が integrate 後にのみ解放（§8） |
| force-push / `git reset --hard` / `git clean -fd(x)` | git-worktree-safety 絶対禁止 |

**エージェントが行ったこと（本ユニット）**: ディスク上 4 CSV の read-only 存在確認と、本運用レポートの作成のみ。

---

## 7. STG / PROD — ローカル reset を使わない

| 環境 | 許可 | 禁止 |
|:---|:---|:---|
| **local** (`animalekarte` project, `ekarte-postgres-data`) | USER による `make migrate` / 必要時 `make reset` | エージェント自動実行 |
| **STG** | 別途承認された運用 runbook / デプロイ経路のみ | `make reset`、ローカル契約スクリプトの流用、デモ CSV の安易な再 COPY |
| **PROD** | 明示承認 + 本番運用手順のみ | 一切の demo reseed、`make reset`、エージェントによる DB 変更 |

`local-db-reset-contract.sh` は project 名に staging / prod / production / stg 等を含む場合 **拒否**する設計だが、**拒否されるから安全、ではなく「そもそも実行しない」**。  
STG/PROD で seed 差分を載せる必要が生じた場合は、**別チケット・別承認・環境専用手順**で扱う。本レポートの `make reset` 手順を流用しない。

---

## 8. claim / TASK-009 について

packet claim プロトコル（`Agents.md` / `git-worktree-safety`）:

- 作業中に `claim/TASK-009`（または ledger 表記の claim ブランチ）が存在する場合、それは **他セッションまたは本タスクの相互排他ロック**である。  
- **エージェントは claim ブランチを削除しない**（自他問わず）。  
- **解放は USER のみ**: 作業が `main` に integrate された後、または明示的に abandoned した後に、人間端末で:

```bash
git branch -D claim/TASK-009
# 実名が異なる場合は claim/<TASK-ID> をそのまま指定
```

- integrate 前に claim を外すと、未マージの間に別セッションが同タスクを取得し衝突するため、**解放は integrate とセット**とする。

---

## 9. 参照

| パス | 内容 |
|:---|:---|
| `reports/2026-07-31-task-009-seed-design.md` | 設計・G1–G5・USER apply 原案 |
| `reports/2026-07-31-task-009-slice1.md` | 著作者・FK・スモーク期待値 |
| `backend/migrations/seeds/003_demo/*.csv` | slice1 対象 4 ファイル |
| `Makefile` (`migrate` / `seed` / `reset`) | 正式 make 入口 |
| `scripts/local-db-reset-contract.sh` | local reset 契約 |
| `backend/cmd/migrate/main.go` | checksum / skip / mismatch 挙動 |

---

## 10. 完了条件（本 OPS ユニット）

- [x] 4 CSV がディスク上に存在することを read-only 確認  
- [x] 空 DB → `make migrate` 経路を記載  
- [x] 適用済み → `make reset` + データ損失警告を記載  
- [x] 適用後スモークチェックリスト  
- [x] エージェント非実施リスト  
- [x] STG/PROD 禁止と別承認の明記  
- [x] claim 解放は USER・integrate 後のみ  
- [ ] **USER**: 環境に応じて `make migrate` または `make reset` を実行し、§5 を通す  

**Apply remains USER.**
