# BE-048: merchandise_item_handler のレスポンス形式を配列に統一

**Status**: Open
**Priority**: High
**Affects**: merchandise-items マスタ API
**Date Created**: 2026-03-19
**Related**: TASK-023, FE-084

## Summary

`ListMerchandiseItems` が `newPaginatedResponse()` を使い `{ data: [...], total, page, limit }` 形式を返しているが、他のマスタ API（cages, staffs 等）は直接配列 `[...]` を返している。フロントエンドは配列を期待するため `data.map(transform)` でクラッシュする。

## 現状のコード

```go
// backend/internal/handler/merchandise_item_handler.go:14-35
func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	category := c.Query("category")
	items, total, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, page, limit, category)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toMerchandiseItemResponseList(items), total, page, limit))
}
```

## 参照: 正しいパターン（cage_handler.go）

```go
// backend/internal/handler/cage_handler.go:17-28
func (h *Handler) ListCages(c *gin.Context) {
	var cageType *string
	if t := c.Query("cage_type"); t != "" {
		cageType = &t
	}
	cages, err := h.svc.Cage.List(c.Request.Context(), cageType)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, cages)  // ← 直接配列
}
```

## 必要な変更

### 1. Handler 変更

```go
// backend/internal/handler/merchandise_item_handler.go
// ListMerchandiseItems を修正: PaginatedResponse → 直接配列

func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	category := c.Query("category")
	items, err := h.svc.MerchandiseItem.ListAll(c.Request.Context(), clinicID, category)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMerchandiseItemResponseList(items))
}
```

### 2. Service 変更

`MerchandiseItem.List` は page/limit を受け取るが、マスタデータは全件取得が基本。`ListAll` メソッドを追加するか、既存の `List` からページネーションを除去する。

### 3. Repository 変更

ページネーションなしの全件取得メソッドが必要。

## 完了条件

- [ ] `GET /api/v1/masters/merchandise-items` が直接配列 `[{...}, {...}]` を返す
- [ ] `GET /api/v1/masters/merchandise-items?category=food` でカテゴリフィルタが動作する
- [ ] 既存の Create/Update/Delete エンドポイントに影響なし
