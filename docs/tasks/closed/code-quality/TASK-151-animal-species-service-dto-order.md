# TASK-151: animal_species_service.go — DTO/定数/builder の定義順序違反

**優先度**: Low
**対象ファイル**: `backend/internal/service/animal_species_service.go`
**チェック項目**: 6（Service の DTO・定数・helper の定義順序）

---

## 問題

プロジェクト規約では Service ファイルの定義順序を以下のように定めている。

```
CreateXxxInput
UpdateXxxInput
const colXxx* = "..."
func buildXxxUpdateFields(...)
type XxxService interface { ... }
type xxxService struct { ... }
func (s *xxxService) ...メソッド...
```

`animal_species_service.go` では `const` と `buildAnimalSpeciesUpdateFields` がインターフェース・実装より**後ろ**（128行目以降）に定義されており、規約違反となっている。

---

## 現状コード（抜粋）

```go
// 29行目: インターフェース（DTO の後に正しく定義）
type AnimalSpeciesService interface { ... }

type animalSpeciesService struct { ... }

func NewAnimalSpeciesService(...) AnimalSpeciesService { ... }

func (s *animalSpeciesService) List(...) { ... }
// ... 実装メソッドが続く ...
func (s *animalSpeciesService) Reorder(...) { ... }

// 128行目以降: const と builder がメソッド実装の後ろに配置されている
const (
    colAnimalSpeciesName      = "name"
    colAnimalSpeciesIsActive  = "is_active"
    colAnimalSpeciesSortOrder = "sort_order"
)

func buildAnimalSpeciesUpdateFields(input *UpdateAnimalSpeciesInput) map[string]any { ... }
```

---

## 修正後コード（ファイル全体の構造）

```go
// ---- Input DTOs ----
type CreateAnimalSpeciesInput struct { ... }
type UpdateAnimalSpeciesInput struct { ... }

// ---- DB column constants ----
const (
    colAnimalSpeciesName      = "name"
    colAnimalSpeciesIsActive  = "is_active"
    colAnimalSpeciesSortOrder = "sort_order"
)

// buildAnimalSpeciesUpdateFields はポインタが非 nil のフィールドのみ map に追加する
func buildAnimalSpeciesUpdateFields(input *UpdateAnimalSpeciesInput) map[string]any { ... }

// ---- AnimalSpeciesService ----
type AnimalSpeciesService interface { ... }

type animalSpeciesService struct { ... }

func NewAnimalSpeciesService(...) AnimalSpeciesService { ... }

func (s *animalSpeciesService) List(...) { ... }
// ... 実装メソッド ...
```

---

## 修正手順

`animal_species_service.go` で `const (colAnimalSpecies...)` ブロックと `buildAnimalSpeciesUpdateFields` 関数を、`UpdateAnimalSpeciesInput` 定義の直後（インターフェース定義より前）に移動する。

