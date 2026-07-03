# AnimalEkarte Database Schema Reference

## テーブル一覧

テーブル数は103（`backend/migrations/001_init.sql` を正とする）。個別テーブルの列挙は
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

# マイグレーションファイル
find backend -path "*/migrations/*.go" -type f

# AutoMigrate呼び出し確認
grep -rn "AutoMigrate" backend/
```

## PostgreSQL接続情報

```
Host: localhost
Port: 5432
Database: ekarte_db
User: ekarte_user
Password: ekarte_password
```

MCP PostgreSQLサーバー経由でクエリ実行可能（local opt-in only。CLAUDE.md方針により
読み取り専用のスキーマ調査に限定し、プロジェクト共有MCPとしては提供しない）。
