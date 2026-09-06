---
name: stg-release-readiness
description: main→staging PR 作成前、または seed/migration 変更を含むデプロイ前のリリース前チェック。checksum mismatch・fresh-DB適用・CI波及・db_reset要否を確認する。
---

# STG Release Readiness

main→staging へのPR作成やデプロイ前に、「ローカルの静的検証を通過した」ことと「STGへ実際に安全に適用できる」ことを混同しない。過去のインシデント（checksum mismatch、seed適用時のUNIQUE違反によるクラッシュループ、CI波及漏れ）はすべて、この2つを混同したまま「完了」と報告したことが原因である。

## いつ発動するか

- main→staging PR 作成前
- seed / migration 変更を含むデプロイ前
- リリースチェックリスト実行時

## 手順

1. **適用済みmigrationを編集していないか確認**:
   ```bash
   git diff origin/staging -- backend/migrations/
   ```
   既存ファイルへの変更（新規追加ではなく既存行の変更）があれば checksum mismatch でSTGが即死する。migration-seed-safety skill の「既存ファイルを絶対に編集しない」原則を再確認する
2. **seed変更の条件付き fresh-DB実適用 + UNIQUE突合**: seed/migration または起動依存を変更した場合、静的検証だけで完了とせず disposable な隔離 DB/environment への fresh apply で検証する。既存/shared DB を drop/reset しない。(clinic_id, name) 等のUNIQUE制約突合を実施する（詳細は `migration-seed-safety` skill参照）。docs-only 変更にはこの gate は不要
3. **起動条件(env)変更のCI波及確認**: 本番/STG起動条件のenvを変更した場合、`.github/workflows/` 内の同名envが同期更新されているか確認する（出典: memory feedback_config_change_ci_propagation）。ワークフローのdriftチェックは `env`/`with`/`working-directory` すべてを対象にする（出典: feedback_workflow_with_param_drift）
4. **DB再作成要否をユーザーに確認**: STGはデモ環境だが、破棄可否は必ずユーザーに確認する。無断で決めない。現行 `backend-deploy.yml` に `db_reset` 入力はない
5. **復旧手順の提示**（ユーザー承認の上で実行）: 通常の STG デプロイは Cloudflare Workers 版 `backend-deploy.yml`（staging push または `workflow_dispatch`）。共有 DB の再作成が必要な場合は workflow の入力を捏造せず、`docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md` の破壊的操作境界に従う。AWS は廃止済みで切り戻し先ではない

## 良い例・悪い例

✅ 良い例（STG実機証跡とローカル検証を区別して報告）:
```
ローカルでは既存 Compose project・volume・DB から分離した disposable DB/environment で、対象・破棄範囲・隔離性を示したユーザー承認後に fresh apply を実行し、001→最新まで（番号を文書へ固定せず、`ls backend/migrations/*.sql` でlive確認）ERROR ゼロを確認した。ただしこれはローカル検証であり、STG実機での適用結果ではない。
db_reset要否についてはユーザー確認が必要——STGはデモ環境だが、破棄可否は事前承認事項とする。
```

❌ 悪い例（CIの見かけ上のgreenだけで判断する）:
```
CIがgreenなのでSTGへのマージは安全。
```
（paths-filterでmigrationジョブがskipされている可能性がある。`gh run view --json jobs` で対象jobが実skipでなくsuccessであることを確認していない。出典: feedback_paths_filter_silent_green）

## 完了条件

- migration/seed/起動依存を変更した場合のみ、disposable な隔離 DB/environment で 001→最新migration が ERROR ゼロで適用されることを確認した（既存/shared DB の reset はしない）
- CIの対象jobが実success（skipでない）であることを確認した
- STG実機証跡の有無を明示した（ローカル機構検証 ≠ STG証跡である旨を報告に明記する）

## 出典

memory: `ops_applied_migration_edit_requires_db_reset` / `seed_master_update_startup_crash_revert` / `seed_lowid_remap_xtenant_regression` / `issue123_release_check_corrections` / `feedback_config_change_ci_propagation` / `feedback_paths_filter_silent_green`

## 関連skill

- `migration-seed-safety`: migration/seedファイル自体の安全ガードレール（本skillはリリース前の最終ゲート、migration-seed-safetyは編集時のガードレール）
