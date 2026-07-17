---
name: analyzing-schema
description: データベーススキーマの分析・マイグレーション支援。スキーマ、マイグレーション、DBモデル時に使用。
---

# analyzing-schema

スキーマ分析・マイグレーション支援の内容は `postgres-patterns`（スキーマ設計・GORMパターン）と `migration-seed-safety`（マイグレーション実行・checksum・seed安全性）に統合済み。ここに独立した内容は保持しない。テーブル数など変動する実測値は `backend/migrations/001_init.sql` 以降のマイグレーション本体を正とする。
