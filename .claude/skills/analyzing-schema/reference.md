# AnimalEkarte Database Schema Reference

## テーブル一覧

テーブル数は `backend/migrations/001_init.sql` を正とする（2026-07時点で約110テーブル）。個別テーブルの列挙は
マイグレーション追加のたびに陳腐化するため、このファイルでは行わない。最新の正確な
テーブル・カラム構成は必ず `backend/migrations/001_init.sql`（および以降の migration
ファイル）を直接確認すること。

主要テーブルの一部（代表例）:

| テーブル名 | 説明 |
|-----------|------|
| `owners` | 飼い主情報（`clinic_id` でマルチテナント隔離） |
| `pets` | 患者（動物）情報 |
| `audit_logs` | 監査ログ |

## リレーション図

個別のリレーション図は 001_init.sql の外部キー定義から都度確認する。ここでの図の
固定的な列挙はスキーマ変更に追従できず誤情報化するため廃止する。

## GORMモデル確認コマンド

```bash
# モデルファイル一覧
find backend/internal/model -name "*.go" -type f

# 特定モデルの構造確認
grep -A 20 "type Pet struct" backend/internal/model/pet.go

# マイグレーションファイル（Go マイグレーションは存在しない。Raw SQL のみ）
ls backend/migrations/*.sql

# AutoMigrate呼び出し確認
grep -rn "AutoMigrate" backend/
```

> 注: 本番適用は `backend/cmd/migrate`（Raw SQL + schema_migrations checksum）。AutoMigrate はテストコードのみ。

## PostgreSQL接続情報

```
Host: localhost
Port: 5434（ホスト側。docker-compose "5434:5432"。コンテナ内は db:5432）
Database / User / Password: `.env` の DB_NAME / DB_USER / DB_PASSWORD を参照（値をセッションに読み込まない）
```

MCP PostgreSQLサーバー経由でクエリ実行可能（local opt-in only。CLAUDE.md方針により
読み取り専用のスキーマ調査に限定し、プロジェクト共有MCPとしては提供しない）。
