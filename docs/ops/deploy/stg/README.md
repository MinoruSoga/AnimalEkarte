# stg/ — STG 環境の継続運用

> **目的**: ステージング環境の日次/週次運用、デモデータ管理、DB 再構築手順をまとめる。
> **読者**: STG 運用担当。
> **タイミング**: STG の定期運用・再構築作業時。

## PR #663: migration 036 の適用前提

PR #663 は `backend/migrations/036_hospitalizations_next_visit.sql`（入院の `next_visit` 列追加）を含む。該当コミットを pull した開発者は、更新後アプリの利用前に `make migrate` を手動で実行する。エージェントによる自動適用は禁止。

STG への適用は [STG_PLANETSCALE_SEED_RUNBOOK.md](STG_PLANETSCALE_SEED_RUNBOOK.md) の対象確認・承認・実行証跡の手順に従う。コードの更新や CI の成功だけで migration 適用済みと判断しない。既存/共有 DB の破棄・再作成は別の明示承認が必要。

| ドキュメント | 内容 |
|---|---|
| [STG-CONTINUOUS-OPERATIONS.md](STG-CONTINUOUS-OPERATIONS.md) | 日次/週次/月次の STG 監視・検査・メンテナンス |
| [STG-DEMO-DATA-LIFECYCLE.md](STG-DEMO-DATA-LIFECYCLE.md) | デモデータの分類・作成元・Cleanup 方針・DB 再作成境界 |
| [STG_PLANETSCALE_SEED_RUNBOOK.md](STG_PLANETSCALE_SEED_RUNBOOK.md) | PlanetScale STG への `cmd/migrate` 適用・検証手順 |
