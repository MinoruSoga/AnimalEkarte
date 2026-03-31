# BE-033: 頻繁にWHERE句で使用されるカラムのインデックス欠落

## 問題
以下のカラムがリポジトリで頻繁に検索条件に使用されているが、
`001_init.sql` にインデックスが定義されていない。

## 欠落インデックス

| カラム | テーブル | 用途 | 推奨インデックス |
|--------|---------|------|-----------------|
| `email` | `user_accounts` | ログイン認証クエリ | `CREATE UNIQUE INDEX idx_user_accounts_email ON user_accounts(email) WHERE deleted_at IS NULL;` |
| `phone` | `owners` | オーナー検索（ILIKE） | `CREATE INDEX idx_owners_phone_trgm ON owners USING gin (phone gin_trgm_ops) WHERE deleted_at IS NULL;` |
| `staff_role` | `staffs` | スタッフ種別フィルタ | `CREATE INDEX idx_staffs_staff_role ON staffs(staff_role) WHERE deleted_at IS NULL;` |
| `category` | `inventory_items` | 在庫カテゴリフィルタ | `CREATE INDEX idx_inventory_items_category ON inventory_items(category) WHERE deleted_at IS NULL;` |

## 既存の良いインデックス（参考）
- `clinic_id` — 45+ インデックス ✅
- GIN trgm インデックス（name, name_kana 等） ✅
- FK カラム全件 ✅

## 修正方針
`backend/migrations/001_init.sql` の末尾に上記 CREATE INDEX を追記。

## 優先度
MEDIUM（パフォーマンス改善）
