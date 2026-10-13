# seeds/ — DB リセット・seed・CSV 投入手順

> **目的**: ローカル/環境 DB の再作成、seed 投入、old_db CSV カットオーバーの手順をまとめる。
> **読者**: 開発者・運用担当。
> **タイミング**: DB 再作成・seed 投入・CSV カットオーバー実施時。

| ドキュメント | 内容 |
|---|---|
| [LOCAL_DB_RESET.md](LOCAL_DB_RESET.md) | ローカル開発 DB の再作成・migration 再適用・seed 復元 |
| [SEED_MIGRATION_OPERATIONS.md](SEED_MIGRATION_OPERATIONS.md) | 全環境 master-only CSV と `APP_ENV` 別ログイン seed の運用 |
| [OLD_DB_HANDOFF_LOCAL.md](OLD_DB_HANDOFF_LOCAL.md) | old_db 21表 CSV の医院別ローカル隔離配置 |
| [CLINIC_CSV_IMPORT.md](CLINIC_CSV_IMPORT.md) | old_db CSV の医院カットオーバー投入（F6 手順） |
