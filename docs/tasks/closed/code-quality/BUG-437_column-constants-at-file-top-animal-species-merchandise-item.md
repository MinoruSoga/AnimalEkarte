# BUG-437: column 定数がファイル先頭定義（animal_species, merchandise_item）— BUG-423 追加分

## 概要

BUG-423 で指摘した「列名定数がファイル先頭に定義されており、ファイル末尾（buildXxx 関数付近）への集約が規約」という問題が、
さらに 2 サービスで確認された。

## 問題箇所

### animal_species_service.go:13-18（ファイル先頭）

```go
// 列名定数
const (
    colAnimalSpeciesName      = "name"
    colAnimalSpeciesIsActive  = "is_active"
    colAnimalSpeciesSortOrder = "sort_order"
)

// ---- Input DTOs ----
type CreateAnimalSpeciesInput struct { ... }
```

### merchandise_item_service.go:13-23（ファイル先頭）

```go
// --- DB column constants ---
const (
    colMerchandiseItemName      = "name"
    colMerchandiseItemCategory  = "category"
    colMerchandiseItemUnitPrice = "unit_price"
    colMerchandiseItemTaxType   = "tax_type"
    colMerchandiseItemTaxRate   = "tax_rate"
    colMerchandiseItemIsActive  = "is_active"
    colMerchandiseItemSortOrder = "sort_order"
)

// --- Input DTOs ---
type CreateMerchandiseItemInput struct { ... }
```

## 規約パターン（ファイル末尾定義）

```go
// cage_service.go（末尾配置が正しいパターン）
// buildCageUpdateFields, const は全てファイル末尾にまとめる

const (
    colCageName      = "name"
    colCageType      = "cage_type"
    colCageSize      = "cage_size"
    colCageIsActive  = "is_active"
    colCageSortOrder = "sort_order"
)
```

## 修正方針

各ファイルの const ブロックを `buildXxxUpdateFields` 関数の直前（ファイル末尾付近）に移動する。

## 影響ファイル

- `backend/internal/service/animal_species_service.go` — 行 13-18
- `backend/internal/service/merchandise_item_service.go` — 行 13-23

## 優先度

**Low** — 機能影響なし。コードの見通し統一のみ。

## 関連チケット

- BUG-423（consultation/diagnosis の同種問題）
