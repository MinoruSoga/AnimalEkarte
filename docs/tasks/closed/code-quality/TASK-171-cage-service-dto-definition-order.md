# TASK-171: cage_service.go — DTO・const・helper の定義順序違反

## 優先度
Medium

## 対象ファイル
`backend/internal/service/cage_service.go`

## 問題概要
`UpdateCageInput`、`const colCage*`、`buildCageUpdateFields` が
`CageService interface` とメソッド実装の **後** に定義されている。
規約の正しい定義順序は以下の通り:

```
CreateXxxInput
UpdateXxxInput  ← ここに置くべき
const colXxx*   ← ここに置くべき
buildXxxUpdateFields  ← ここに置くべき
type XxxService interface { ... }
type xxxService struct { ... }
func NewXxxService(...)
func (s *xxxService) methods...
```

## 現状コード（cage_service.go 抜粋）

```go
// line 16: CreateCageInput (OK)
type CreateCageInput struct { ... }

// line 26: CageService interface (NG: UpdateInput より先に来ている)
type CageService interface { ... }

// line 35: cageService struct
type cageService struct { ... }

// line 40〜139: メソッド実装

// line 143: UpdateCageInput (NG: interface/メソッドより後)
type UpdateCageInput struct {
    Name        *string
    CageType    *string
    CageSize    *string
    Price       *int64
    IsActive    *bool
    Description *string
    SortOrder   *int
}

// line 153: const colCage* (NG: interface/メソッドより後)
const (
    colCageName        = "name"
    colCageCageType    = "cage_type"
    colCageCageSize    = "cage_size"
    colCagePrice       = "price"
    colCageIsActive    = "is_active"
    colCageDescription = "description"
    colCageSortOrder   = "sort_order"
)

// line 163: buildCageUpdateFields (NG: interface/メソッドより後)
func buildCageUpdateFields(input *UpdateCageInput) map[string]any { ... }
```

## 修正後コード（定義順序）

```go
// ---- CageService ----

// CreateCageInput はケージ作成の入力DTO
type CreateCageInput struct {
    Name        string
    CageType    string
    CageSize    string
    Price       *int64
    IsActive    bool
    Description string
    SortOrder   int
}

// UpdateCageInput はケージ更新のサービス入力 DTO
type UpdateCageInput struct {
    Name        *string
    CageType    *string
    CageSize    *string
    Price       *int64
    IsActive    *bool
    Description *string
    SortOrder   *int
}

const (
    colCageName        = "name"
    colCageCageType    = "cage_type"
    colCageCageSize    = "cage_size"
    colCagePrice       = "price"
    colCageIsActive    = "is_active"
    colCageDescription = "description"
    colCageSortOrder   = "sort_order"
)

func buildCageUpdateFields(input *UpdateCageInput) map[string]any { ... }

type CageService interface { ... }

type cageService struct { ... }

func NewCageService(...) CageService { ... }

func (s *cageService) List(...) { ... }
// ... 他メソッド
```

## 影響範囲
コンパイル・動作への影響はなし。コードの可読性・規約統一のみ。

## 対応方針
`UpdateCageInput`、`const colCage*`、`buildCageUpdateFields` を
ファイル先頭（`CreateCageInput` の直後、`CageService interface` の前）に移動する。
