# STG 保持データの export / restore

> **目的**: STG の schema 再構築で失ってはいけないデータを、CSV と manifest に退避し、再構築後に復元する。
> **読者**: STG 運用担当・承認済み operator。
> **タイミング**: [STG_PLANETSCALE_SEED_RUNBOOK.md](STG_PLANETSCALE_SEED_RUNBOOK.md) §6.1 の再構築前後。

## 対象

| 分類 | テーブル |
|---|---|
| サポート | `support_bug_reports`, `support_chat_messages` |
| 取扱説明書 | `manual_articles`, `manual_article_versions` |
| 医院設定 | `clinics`, `clinic_settings`, `clinic_holidays`, `closing_special_periods` |
| 権限 | `permission_groups`, `permission_group_rules` |
| アカウント | `accounts`, `staffs`, `staff_clinic_assignments`, `staff_permission_groups`, `occupations` |

- soft-deleted 行も含めて export する。`deleted_at` を保持するため、restore 後も非表示状態は維持される。
- スクショ実体は object storage 側。DB reset では削除しない。bucket / key prefix を変更しない。
- `manual_articles` / `manual_article_versions` は SQL で復元する。`manual-sync.py` は本文のみで編集履歴を失うため、export の代替にしない。復元後の `manual-sync.py status` は本文の突合に使う。

## 前提

- 接続情報は `scripts/old-db/stg-uat-old-db-handoff.local.env`（0600）を使う。値は git・ログ・チケットへ出さない。
- `psql` は client-side `\copy` を使う。DB の app role は `BYPASSRLS` ではないため、`PGOPTIONS='-c app.bypass_rls=on'` を付ける。
- export / restore は USER または承認済み operator が実行する。agent は実行しない。

## Export（再構築前、read-only）

```bash
python3 scripts/old-db/stg-preserve.py export
```

- 出力先: `sensitive-local/stg-preserve/<UTC>/`（0700、各ファイル 0600、git 外）。
- `manifest.json` に table / columns / rows / sha256 / bytes / content_hash / exported_at を記録する。
- 出力直後に各 table の count と manifest を照合する。
- 既存ディレクトリを上書きしない。再実行時は新しい UTC ディレクトリになる。

## Restore（再構築後）

### 1. Dry-run

```bash
python3 scripts/old-db/stg-preserve.py restore --dir sensitive-local/stg-preserve/<UTC>
```

DB へ接続せず、restore 順と行数を表示する。manifest の table 順・CSV header・sha256 を検証する。

### 2. Execute

```bash
python3 scripts/old-db/stg-preserve.py restore --dir sensitive-local/stg-preserve/<UTC> --execute
```

- FK 順に temp table + `\copy` + `ON CONFLICT DO UPDATE` で upsert する。
- `id` を持つ table は sequence を `setval` する。
- `accounts` / `staffs` は export 時点の行を upsert する。restore 後に `cmd/migrate` を再実行し、demo login の password を現行 seed で再 upsert する。
- 途中で失敗した場合は partial table repair をせず、PlanetScale の snapshot/branch へ戻す。

### 3. Verify

```bash
python3 scripts/old-db/stg-preserve.py verify --dir sensitive-local/stg-preserve/<UTC>
```

- 指定4表（bug reports / support chat / manual articles / manual article versions）は row count と content hash が一致すること。
- 設定・アカウント系は seed 由来の追加行があり得るため、row count は export 以上であること。
- FK orphan 0 を確認する。

## ローカル dry-run

STG の export をローカルの scratch DB に戻す場合は、同じ script にローカル接続の env file を渡す。

```bash
python3 scripts/old-db/stg-preserve.py restore --dir <export-dir> --env-file <local-env> --execute
```

ローカル env file は 0600 とし、`DB_HOST=127.0.0.1`、`DB_PORT=5434` など scratch DB を指す。アプリの DB を対象にしない。

## 注意

- `permission_group_rules` は seed の 315 行と STG の追加行を id で upsert する。STG 側で削除した seed 行は migrate が再作成するため、削除状態は復元されない。
- `clinic_holidays` は `(clinic_id, date)` の unique を持つ。restore は id で upsert するため、同日・別 id の行が seed に存在する場合は衝突する。衝突時は停止して差分を確認する。
- restore 後に `cmd/migrate` を再実行しても DDL は skip され、seed bundle は checksum 一致で skip、login seed だけが再 upsert される。
