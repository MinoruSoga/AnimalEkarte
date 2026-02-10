# analyzing-schema

データベーススキーマの分析・マイグレーション支援。スキーマ、マイグレーション、DBモデル時に使用。

## Overview

Go/GORM + PostgreSQL 環境に特化したデータベーススキーマ分析スキル。
AnimalEkarte プロジェクトの22テーブル構成を理解し、スキーマ変更・マイグレーション・モデル定義を支援する。

## When to Use

- データベーススキーマの確認・分析
- GORMモデルの作成・修正
- マイグレーションファイルの作成
- テーブル間のリレーション確認
- インデックス最適化の検討

## Workflow

1. **現状把握**: GORMモデル定義ファイルを確認（`backend/internal/models/`）
2. **スキーマ分析**: PostgreSQL MCPサーバーで実テーブル構造を確認
3. **リレーション確認**: 外部キー制約・関連テーブルを確認
4. **変更計画**: マイグレーション内容を決定
5. **実装**: GORMモデル更新 → マイグレーション作成 → テスト

## Key Patterns

### GORM モデル定義パターン
```go
type Patient struct {
    gorm.Model
    Name        string    `gorm:"type:varchar(100);not null" json:"name"`
    Species     string    `gorm:"type:varchar(50);not null" json:"species"`
    Breed       string    `gorm:"type:varchar(100)" json:"breed"`
    BirthDate   time.Time `gorm:"type:date" json:"birth_date"`
    OwnerID     uint      `gorm:"not null" json:"owner_id"`
    Owner       Owner     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}
```

### PostgreSQL クエリテンプレート
```sql
-- テーブル一覧
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';

-- カラム情報
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns WHERE table_name = 'patients';

-- 外部キー確認
SELECT tc.constraint_name, tc.table_name, kcu.column_name,
       ccu.table_name AS foreign_table_name, ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY';

-- インデックス確認
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'patients';
```

## Tools

- **PostgreSQL MCP**: データベース直接クエリ
- **Serena**: Go ソースコード解析（GORMモデル）
- **Grep/Glob**: マイグレーションファイル検索

## Reference

詳細なテーブル一覧は `reference.md` を参照。
