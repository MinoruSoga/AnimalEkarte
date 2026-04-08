# BUG-143: per_page パラメータが無視される（常に20件固定）

## 概要
`GET /api/v1/owners?per_page=1` でも `per_page=100` でも常に20件が返る。
ページネーションの `per_page` パラメータが機能していない。

## 再現手順
```bash
curl '/api/v1/owners?per_page=1'
# → data: 20件（per_page=1 を無視）

curl '/api/v1/owners?per_page=100'
# → data: 20件（per_page=100 を無視）
```

## ブラウザテスト結果
| per_page | 期待件数 | 実際件数 |
|----------|---------|---------|
| 1 | 1 | **20** |
| 100 | 22 (全件) | **20** |
| 10000 | 22 (全件、上限適用) | **20** |

## 期待する動作
- `per_page=1` → 1件
- `per_page=100` → 全件（上限 100 以内）
- `per_page` 未指定 → デフォルト 20件
- `per_page` 上限 → 100件

## 修正方針

Repository 層の Pagination で `per_page` パラメータを受け取るようにする:

```go
func (r *OwnerRepository) FindAll(ctx context.Context, clinicID uint64, params PaginationParams) ([]model.Owner, int64, error) {
    perPage := params.PerPage
    if perPage <= 0 { perPage = 20 }
    if perPage > 100 { perPage = 100 }
    
    db := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID)
    db.Count(&total)
    db.Offset((params.Page - 1) * perPage).Limit(perPage).Find(&owners)
}
```

## 優先度
**Low** — 機能バグ。セキュリティ影響なし。

## 関連ファイル
- `backend/internal/repository/owner_repository.go` — FindAll
- `backend/internal/handler/owner_handler.go` — ListOwners（クエリパラメータ取得）
