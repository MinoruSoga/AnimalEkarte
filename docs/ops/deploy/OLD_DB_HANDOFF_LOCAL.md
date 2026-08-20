# old_db 21表 CSV の医院別ローカル隔離

更新日: 2026-08-20

`old_db` が出力した AnimalEkarte 形状 21 CSV + manifest を、AnimalEkarte
worktree 内へ **医院別** に置く手順です。PHI を含み得るため Git 管理外です。

## 置き場所

```text
backend/migrations/seeds/_old_db_handoff/<clinic>/<run>/
  manifest.json
  owners.csv
  pets.csv
  ...（21 CSV）
```

例（城東・診断 rehearsal）:

```text
backend/migrations/seeds/_old_db_handoff/jouto/jouto-intake-20260819-01/
backend/migrations/seeds/_old_db_handoff/jouto/jouto-intake-20260819-01-local/  # 推奨（電話重複を解消）
```

同一 `clinicCode` + `sourceRunId` で `<run>` と `<run>-local` が両方ある場合、
`make reset` は **`-local` を優先**して1本だけ import する。

## 重要な境界

| 経路 | 用途 | `make seed` / `cmd/migrate` |
| --- | --- | --- |
| `seeds/002_master` 等 | 通常デモ/マスタ seed | **読む** |
| `seeds/_old_db_handoff/<clinic>/<run>/` | old_db 21表のローカル隔離 | **読まない** |
| `make csv-import-*` | 正式 cutover（F6） | seed を生成しない |

- `REHEARSAL_ONLY` / `PARTIAL` bundle はここに置けるが、**正式 preflight は拒否**する。
- 正式 DB 投入は manifest が `status=PASS` かつ `handoffEligibility=TRUSTED_CANDIDATE` のときだけ。
- 21 CSV を `003_demo` へ直接コピーして seed 扱いしてはいけない。

詳細境界: [SEED_MIGRATION_OPERATIONS.md](./SEED_MIGRATION_OPERATIONS.md) /
[CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md)

## 事前準備（必須）

コピー前に checkout 固有の `.git/info/exclude` へ次を入れ、確認する:

```text
backend/migrations/seeds/_old_db_handoff/
```

```sh
git check-ignore -q --no-index backend/migrations/seeds/_old_db_handoff/
```

`stage-old-db-handoff.sh` / `make old-db-handoff-stage` はこの行が無ければ追記し、
ignore 確認に失敗したら配置自体を拒否する。

## 配置コマンド

producer 側（例: sibling `old_db`）の export 絶対パスを渡す。

```sh
export CLINIC_CODE=jouto
export MIGRATION_RUN_ID=jouto-intake-20260819-01
export CSV_IMPORT_SOURCE_DIR=/absolute/path/to/animalekarte-csv-export/jouto/jouto-intake-20260819-01

make old-db-handoff-stage
make old-db-handoff-check
```

成功時に manifest SHA-256 が印字される。正式 import ではこの SHA を
`CSV_MANIFEST_SHA256` として別経路で渡す。

## DB へ入れるとき

### ローカル `make reset`（自動）

`backend/migrations/seeds/_old_db_handoff/<clinic>/<run>/` に bundle があれば、
ローカル `make reset` の postflight 後に **自動 import** します。

一時スキップ（検査機器疎通など reset を早くしたいとき）:

```sh
SKIP_OLD_DB_HANDOFF_IMPORT=1 make reset
```

- `REHEARSAL_ONLY` / `UNVERIFIED` もローカル限定で許可（`--allow-local-rehearsal`）
- `APP_ENV` が development/local/dev/test 以外ではスキップ
- 正式 F6 ゲートは変更せず、共有 STG/本番では使わない
- ローカル apply は Repeatable Read（SSI 述語ロック枯渇回避）。正式 apply は Serializable のまま
- ローカル Postgres は `max_locks_per_transaction=512` / `shared_buffers=256MB`（`docker-compose.yml`）

必要 seed ID（clinic / species / 検査 / trimming / cash / card）は DB から解決します。
`clinicOrdinal`（城東は 2）を `TARGET_CLINIC_ID` に使い、demo clinic 1 との
`(clinic_id, name/phone)` 衝突を避ける。検査 / trimming が無ければ同 clinic に足す。

### 正式 cutover（F6）

1. bundle が `TRUSTED_CANDIDATE` / `PASS` であること
2. 対象は disposable / 承認済み DB のみ（本番は別承認）
3. [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) の変数を揃える
4. `make csv-import-preflight` → 承認後 `make csv-import` → `make csv-import-verify`
   （`--allow-local-rehearsal` は付けない）

画面確認用 disposable は [A4_UI_REHEARSAL.md](./A4_UI_REHEARSAL.md)。

## 城東の現状（2026-08-20）

`jouto-intake-20260819-01` は producer 側で **`REHEARSAL_ONLY`**。
飼主・ペットを含む 21 表は揃っているが、正式 F6 は拒否される。
ローカルでは `_old_db_handoff/jouto/jouto-intake-20260819-01-local/`（電話 unique 用に
sanitize 済み）を置き、`make reset` で `clinic_id=2`（城東センター病院）へ取り込む。
