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
2. **seed変更のfresh-DB実適用 + UNIQUE突合**: seedを差し替えた場合、静的検証だけで完了とせず fresh DBへの実適用で検証する。(clinic_id, name) 等のUNIQUE制約突合を実施する（詳細は `migration-seed-safety` skill参照）
3. **起動条件(env)変更のCI波及確認**: 本番/STG起動条件のenvを変更した場合、`.github/workflows/` 内の同名envが同期更新されているか確認する（出典: memory feedback_config_change_ci_propagation）。ワークフローのdriftチェックは `env`/`with`/`working-directory` すべてを対象にする（出典: feedback_workflow_with_param_drift）
4. **db_reset要否をユーザーに確認**: STGはデモ環境だが、破棄可否は必ずユーザーに確認する。無断で決めない
5. **復旧手順の提示**（ユーザー承認の上で実行）:
   ```bash
   gh workflow run backend-deploy.yml --ref staging -f db_reset=true
   ```

## 良い例・悪い例

✅ 良い例（STG実機証跡とローカル検証を区別して報告）:
```
ローカルでは fresh DB (`docker compose down -v && up -d db && run --rm backend go run ./cmd/migrate`) で
001→011まで ERROR ゼロを確認した。ただしこれはローカル検証であり、STG実機での適用結果ではない。
db_reset要否についてはユーザー確認が必要——STGはデモ環境だが、破棄可否は事前承認事項とする。
```

❌ 悪い例（CIの見かけ上のgreenだけで判断する）:
```
CIがgreenなのでSTGへのマージは安全。
```
（paths-filterでmigrationジョブがskipされている可能性がある。`gh run view --json jobs` で対象jobが実skipでなくsuccessであることを確認していない。出典: feedback_paths_filter_silent_green）

## 完了条件

- fresh DBで 001→最新migrationがERRORゼロで適用されることを確認した
- CIの対象jobが実success（skipでない）であることを確認した
- STG実機証跡の有無を明示した（ローカル機構検証 ≠ STG証跡である旨を報告に明記する）

## 出典

memory: `ops_applied_migration_edit_requires_db_reset` / `seed_master_update_startup_crash_revert` / `seed_lowid_remap_xtenant_regression` / `issue123_release_check_corrections` / `feedback_config_change_ci_propagation` / `feedback_paths_filter_silent_green`

## 関連skill

- `migration-seed-safety`: migration/seedファイル自体の安全ガードレール（本skillはリリース前の最終ゲート、migration-seed-safetyは編集時のガードレール）
