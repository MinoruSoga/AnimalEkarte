---
name: migration-seed-safety
description: backend/migrations/ と seed バンドルを触るときの安全ガード。発火は migration/seed 編集時だけ。CASCADE・clinic_id・checksum の詳細は本文を読む。
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
- seedバンドル（`backend/migrations/seeds/002_master|003_demo|004_staging/*.csv` と `manifest.json`）の内容変更。CSV は手編集せず `backend/cmd/seed-export` で使い捨て DB から dump して再生成するのが正規手順

## 必須チェックリスト（新規migrationファイル作成時）

1. **命名規則**: `{既存最大番号+1}_{snake_case説明}.sql`。番号を飛ばしたり既存ファイルを上書きしない（`backend/migrations/CLAUDE.md`）
2. **clinic_idスコープ**: クリニック間分離が必要な新テーブルは `clinic_id BIGINT NOT NULL` 必須。複合indexの先頭カラムに置く
3. **ソフトデリート**: 業務データテーブルには `deleted_at TIMESTAMPTZ` を追加する
4. **CASCADE DELETEは原則禁止**: 許容されるのは「純粋な従属データ」（join table / 親の構成要素として不可分な子行 / 業務履歴を失わないマスタ参照lookup）のみ。`owners`/`pets`/`medical_records`等PHI・業務データを親とする連鎖削除は禁止——applicationの削除境界で依存確認と明示的な409応答を行う
5. **既存ファイルを絶対に編集しない**: 一度コミットされたmigrationファイルは追記専用。修正が必要なら新しい番号のmigrationで訂正する。このリポジトリには既存migration編集時に警告する `.claude/hooks/pre-edit-migration-guard.js` フックがあるが、フックは非ブロッキングなので最終判断は書き手が行う

## 必須チェックリスト（seed差し替え時）

1. **クロステナントID衝突**: seedの主キー/外部キーを採番し直す場合、他クリニックのIDレンジと衝突しないか確認する。「静的検証がパスした」は安全の証明にならない——過去に静的検証を通過した後、実DBへの適用で初めてクロステナント破綻が発覚した実例がある
2. **UNIQUE制約**: seed内で重複しうる値（email、コード値等）が無いか事前に確認する。起動時seed投入でUNIQUE違反が起きるとコンテナがクラッシュループする
3. **fresh DB apply（条件付き）**: seed/migration または起動依存を変更した場合は、静的確認だけで完了とせず、**disposable な隔離 DB/environment への fresh apply**で検証する。既存・共有 DB への追記、drop、reset を検証手段にしない。DB を使わない docs-only 変更にはこの gate は不要
4. **DB非依存の最低検証**: `python3 scripts/verify_seed.py` を実行する（DB起動不要）
5. **実DB適用はユーザー承認必須**: `make db` / `docker compose exec db psql ...` は自動実行禁止コマンド（`.claude/CLAUDE.md`）。隔離 DB の作成・migration 適用・破棄も、対象・データ消失影響・隔離性を確認したユーザーの明示承認後にだけ実施する。既存・共有 DB の reset は指示しない
6. **テーブル定義外の CREATE UNIQUE INDEX との突合**: seed 差し替え時、PK/FK/CHECK/NOT NULL だけでなく 001_init.sql 末尾の `CREATE UNIQUE INDEX`（例: idx_procedures_clinic_name (clinic_id, name) WHERE deleted_at IS NULL）に対する重複も検証する。過去 revert（PR#78）の唯一の原因はこの見逃し（(1,'輸血') / (1,'お手入れ') の同名2ペア）
   （出典: memory seed_lowid_remap_xtenant_regression / seed_master_update_startup_crash_revert）
7. **CI を通っても migrate の実走を仮定しない — 隔離 fresh-DB 実適用が正本**: 現行 ci.yml は main 向け PR も対象だが、paths-filter で migration ジョブが skip され得るうえ、通常 CI は fresh DB への全 migration 適用を保証しない（過去には main 向け PR が Backend CI 対象外で未検証のまま merge された実例あり）。seed/migration 変更は merge 前に、既存 Compose project・volume・DB を触らない disposable な隔離環境で migration を実行し ERROR ゼロを確認する。実施前に対象・破棄範囲・隔離性を示してユーザーの明示承認を得る。docs-only 変更にこの実適用を要求しない
   （出典: memory seed_lowid_remap_xtenant_regression。CI 対象ブランチは .github/workflows/ci.yml を正とする）
8. **CSV は実 DB の COPY dump が正本** — INSERT 文の静的レビューという概念は廃止済み。verify_seed.py は CSV を直接検証する
9. **seed のフォーマット移送（SQL→CSV 等）はテーブル単位の全数突合が必須**: 旧ソースの INSERT 対象テーブル一覧と新ソースのファイル一覧を diff で機械突合する。目視移送では欠落する（SQL→CSV 移行で checkup_type_fields が丸ごと欠落した実例）。復元は `git show <旧commit>:<旧ファイル>` を正本にする。（出典: memory closed_issue_reaudit_20260707 修正 3e9c449f / seed_csv_migration_20260706）
10. **実スタッフ情報をGit管理しない**: 氏名・email・password hash等のPII/credential verifierをseedへ入れない。初期登録はデータ管理承認済みのsecret-managed importを使い、値をログ・成果物へ出さない

## checksum mismatchからの復旧

適用済みmigrationやseedを誤って編集してしまった場合:
- ローカル: 既存の DB volume を再構築しない。診断・復旧計画を作成し、隔離環境で再現できるかを確認した上で、対象・データ消失影響を含む明示承認を得た手順だけを実施する
- STG: 現行Cloudflare workflowにDB reset経路はない。診断・復旧計画を作成し、DB変更の明示承認を得てから隔離環境または承認済み手順で実施する
- 運用メモ全般: `docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md`

## 診断コマンド

```bash
python3 scripts/verify_seed.py
ls backend/migrations/*.sql   # 番号の重複・欠番を目視確認
```

## 関連する機械強制・自動化

- `preload_clinic_scope_lint_test.go` / `master_fk_write_inventory_lint_test.go`: clinic_idスコープの機械強制（write側のmaster FK検証の正しさまでは保証しない — 判断には `clinic-isolation-auditor` agent を使う）
- `.claude/hooks/pre-edit-migration-guard.js`: 既存migrationファイル編集時の非ブロッキング警告フック
