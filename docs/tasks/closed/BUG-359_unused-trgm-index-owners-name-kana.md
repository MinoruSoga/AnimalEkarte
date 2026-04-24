# BUG-359: trgm インデックス `idx_owners_name_kana_trgm` が未使用

## 概要

`001_init.sql` で `idx_owners_name_kana_trgm` が定義されているが、
Go コードのどの ILIKE 検索でも `owners.name_kana` は検索対象になっていない。
飼主検索（`owner_repository.go:44`）は `name`, `phone`, `email` の 3 カラムのみ。

## 該当箇所

```sql
-- backend/migrations/001_init.sql:1438
CREATE INDEX idx_owners_name_kana_trgm ON owners USING gin (name_kana gin_trgm_ops) WHERE deleted_at IS NULL;
```

## 判断

- **飼主名カナ検索を将来実装する予定があるなら** — インデックスを残し、`owner_repository.go:44` の ILIKE 条件に `name_kana` を追加する
- **不要なら** — `001_init.sql` からインデックスを削除する

## 優先度

**LOW** — 書き込み I/O のわずかなオーバーヘッドのみ。機能影響なし。
