# BUG-154: 物販マスタの削除が常に 400 — estimate_items に deleted_at カラムが存在しない

## 概要
`DELETE /api/v1/masters/merchandise-items/:id` が常に 400 を返す。
FK 依存チェックの SQL が `estimate_items.deleted_at IS NULL` を参照するが、
`estimate_items` テーブルに `deleted_at` カラムが存在しないため SQL エラーが発生。

## 再現手順
```bash
# 新規作成
curl -X POST /api/v1/masters/merchandise-items \
  -d '{"name":"test","unit_price":100,"category":"food","tax_rate":10}'
# → 201, id=18

# 即座に削除
curl -X DELETE /api/v1/masters/merchandise-items/18
# → 400 {"error":"入力値が正しくありません"}
```

## バックエンドログ
```
ERROR: column "deleted_at" does not exist (SQLSTATE 42703)
SELECT count(*) FROM "estimate_items" WHERE merchandise_item_id = 18 AND deleted_at IS NULL
```

## 原因
`merchandise_item_repository.go` の `CountUsageByMerchandiseItemID` が
`WHERE deleted_at IS NULL` を含むクエリを発行するが、`estimate_items` テーブルには
`deleted_at` カラムが定義されていない（論理削除非対応テーブル）。

## 修正方針
```go
// CountUsageByMerchandiseItemID のクエリから deleted_at 条件を削除
func (r *repo) CountUsageByMerchandiseItemID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.EstimateItem{}).
        Where("merchandise_item_id = ?", id).
        // ❌ Where("deleted_at IS NULL"). ← estimate_items に deleted_at がない
        Count(&count).Error
    return count, err
}
```

または `estimate_items` テーブルに `deleted_at` カラムを追加（マイグレーション）。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md`
> 全テーブルに `created_at`, `updated_at`, `deleted_at`

`estimate_items` が `deleted_at` を持っていないのはスキーマ設計の漏れ。

## 優先度
**High** — 物販マスタの削除機能が完全に動作しない。

## 関連ファイル
- `backend/internal/repository/merchandise_item_repository.go` — CountUsageByMerchandiseItemID
- `backend/migrations/001_init.sql` — estimate_items テーブル定義
