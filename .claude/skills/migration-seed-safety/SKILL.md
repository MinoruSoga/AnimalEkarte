---
name: migration-seed-safety
description: backend/migrations/ の新規作成・編集、およびseed(002/003/004)の差し替え作業における安全ガードレール。CASCADE DELETE禁止・clinic_idスコープ・checksum mismatch回避・クロステナントID衝突チェックを網羅。migration/seedファイルに触れる際に使用。
---

# Migration / Seed Safety

このプロジェクトのmigration/seed関連インシデントは全て**同じ根本原因**を持つ:「静的な見た目は正しいが、既存DBへの実適用で壊れる」変更。

過去のインシデント:
- UNIQUE制約違反によるseed適用失敗 → 起動時クラッシュループでrevertに追い込まれた
- IDの振り直しが静的検証はパスしたが、実DBへの適用でクロステナント破綻を起こした
- 適用済みmigrationの直接編集によるchecksum mismatch（本番相当環境でDB resetが必要になった）

このskillは同じ失敗を繰り返さないためのチェックリストである。

## いつ発動するか

- `backend/migrations/*.sql` の新規作成・編集
- seed（`002_seed_master.sql` / `003_seed_demo.sql` / `004_seed_staging.sql`）の内容変更

## 必須チェックリスト（新規migrationファイル作成時）

1. **命名規則**: `{既存最大番号+1}_{snake_case説明}.sql`。番号を飛ばしたり既存ファイルを上書きしない（`backend/migrations/CLAUDE.md`）
2. **clinic_idスコープ**: クリニック間分離が必要な新テーブルは `clinic_id BIGINT NOT NULL` 必須。複合indexの先頭カラムに置く
3. **ソフトデリート**: 業務データテーブルには `deleted_at TIMESTAMPTZ` を追加する
4. **CASCADE DELETEは原則禁止**: 許容されるのは「純粋な従属データ」（join table / 親の構成要素として不可分な子行 / 業務履歴を失わないマスタ参照lookup）のみ。`owners`/`pets`/`medical_records`等PHI・業務データを親とする連鎖削除は禁止——service層の依存チェック（409応答）で代替する
5. **既存ファイルを絶対に編集しない**: 一度コミットされたmigrationファイルは追記専用。修正が必要なら新しい番号のmigrationで訂正する。このリポジトリには既存migration編集時に警告する `.claude/hooks/pre-edit-migration-guard.js` フックがあるが、フックは非ブロッキングなので最終判断は書き手が行う

## 必須チェックリスト（seed差し替え時）

1. **クロステナントID衝突**: seedの主キー/外部キーを採番し直す場合、他クリニックのIDレンジと衝突しないか確認する。「静的検証がパスした」は安全の証明にならない——過去に静的検証を通過した後、実DBへの適用で初めてクロステナント破綻が発覚した実例がある
2. **UNIQUE制約**: seed内で重複しうる値（email、コード値等）が無いか事前に確認する。起動時seed投入でUNIQUE違反が起きるとコンテナがクラッシュループする
3. **fresh DB apply必須**: seedの差し替えは静的確認だけで完了とせず、**fresh DBへの実適用**で検証する。既存DBへの追記適用では検出できない不整合がある
4. **DB非依存の最低検証**: `python3 scripts/verify_seed.py` を実行する（DB起動不要）
5. **実DB適用はユーザー承認必須**: `make db` / `docker compose exec db psql ...` は自動実行禁止コマンド（`.claude/CLAUDE.md`）。migration適用・DBリセットはユーザーが手動で実施する

## checksum mismatchからの復旧

適用済みmigrationやseedを誤って編集してしまった場合:
- ローカル: `docs/infra/deploy/LOCAL_DB_RESET.md` の手順でDB volumeを再構築する
- STG: `DB_RESET=true` の手動dispatchが必要（自動実行しない）
- 運用メモ全般: `docs/infra/deploy/SEED_MIGRATION_OPERATIONS.md`

## 診断コマンド

```bash
python3 scripts/verify_seed.py
ls backend/migrations/*.sql   # 番号の重複・欠番を目視確認
```

## 関連する機械強制・自動化

- `preload_clinic_scope_lint_test.go` / `master_fk_write_inventory_lint_test.go`: clinic_idスコープの機械強制（write側のmaster FK検証の正しさまでは保証しない — 判断には `clinic-isolation-auditor` agent を使う）
- `.claude/hooks/pre-edit-migration-guard.js`: 既存migrationファイル編集時の非ブロッキング警告フック
