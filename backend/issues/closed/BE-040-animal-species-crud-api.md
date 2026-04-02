# BE-040: animal_species マスタ CRUD API 追加

**Status**: Open
**Priority**: Medium
**Affects**: マスタ設定 — 動物種類管理
**Date Created**: 2026-03-17
**Related**: TASK-001, FE-013

## Summary

animal_species マスタテーブルの CRUD API を追加する。現在は List（GET）のみ実装されており、Create/Update/Delete/Reorder が欠落している。マスタ設定画面から動物種類を管理可能にする。

## 現状のコード

### Model（変更不要）

```go
// backend/internal/model/animal_species.go:1-17
type AnimalSpecies struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"not null"                 json:"name"`
    IsActive  bool      `gorm:"default:true"             json:"is_active"`
    SortOrder int       `gorm:"default:0"                json:"sort_order"`
    CreatedAt time.Time `gorm:"autoCreateTime"           json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"           json:"updated_at"`
}
```

**注意: `clinic_id` なし（システム共通マスタ）。** diagnosis_category 等の clinic_id フィルタは不要。

### Handler（List のみ）

```go
// backend/internal/handler/animal_species_handler.go:10-18
func (h *Handler) ListAnimalSpecies(c *gin.Context) {
    species, err := h.svc.AnimalSpecies.List(c.Request.Context())
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, species)
}
```

### Service（List のみ）

```go
// backend/internal/service/animal_species_service.go:12-27
type AnimalSpeciesService interface {
    List(ctx context.Context) ([]model.AnimalSpecies, error)
}
```

### Repository（FindAll のみ）

```go
// backend/internal/repository/animal_species_repository.go:14-34
type AnimalSpeciesRepository interface {
    FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
}
```

### ルーティング（List のみ）

```go
// backend/internal/handler/staff_handler.go:155
masters.GET("/animal-species", h.ListAnimalSpecies)
```

## 参照実装

`diagnosis_category` の CRUD が完全実装されている。同パターンを animal_species に適用する。

- `backend/internal/handler/diagnosis_handler.go` — 302行（全メソッド）
- `backend/internal/handler/diagnosis_request.go` — 42行（Create/Update/Reorder 型）
- `backend/internal/service/diagnosis_service.go` — 270行（Input DTO + buildUpdateFields）
- `backend/internal/repository/diagnosis_repository.go` — 245行（CRUD + Reorder トランザクション）

## 必要な変更

### 1. Request 型追加（新規ファイル）

```go
// backend/internal/handler/animal_species_request.go（新規作成）
package handler

type createAnimalSpeciesRequest struct {
    Name      string `json:"name"      binding:"required"`
    IsActive  bool   `json:"is_active"`
    SortOrder int    `json:"sort_order"`
}

type updateAnimalSpeciesRequest struct {
    Name      *string `json:"name"`
    IsActive  *bool   `json:"is_active"`
    SortOrder *int    `json:"sort_order"`
}

type reorderAnimalSpeciesRequest struct {
    IDs []uint64 `json:"ids" binding:"required,min=1"`
}
```

### 2. Service 拡張

```go
// backend/internal/service/animal_species_service.go
// インターフェースに追加:
type AnimalSpeciesService interface {
    List(ctx context.Context) ([]model.AnimalSpecies, error)
    GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)         // 追加
    Create(ctx context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error)  // 追加
    Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error)  // 追加
    Delete(ctx context.Context, id uint64) error                                  // 追加
    Reorder(ctx context.Context, ids []uint64) error                              // 追加
}

// Input DTO
type CreateAnimalSpeciesInput struct {
    Name      string
    IsActive  bool
    SortOrder int
}

type UpdateAnimalSpeciesInput struct {
    Name      *string
    IsActive  *bool
    SortOrder *int
}

// buildAnimalSpeciesUpdateFields ヘルパー
func buildAnimalSpeciesUpdateFields(input *UpdateAnimalSpeciesInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive
    }
    if input.SortOrder != nil {
        fields["sort_order"] = *input.SortOrder
    }
    return fields
}
```

**注意: clinic_id パラメータは不要（システム共通マスタ）。** diagnosis パターンと異なる点。

### 3. Repository 拡張

```go
// backend/internal/repository/animal_species_repository.go
// インターフェースに追加:
type AnimalSpeciesRepository interface {
    FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
    FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)      // 追加
    Create(ctx context.Context, species *model.AnimalSpecies) error             // 追加
    Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error)  // 追加
    Delete(ctx context.Context, id uint64) error                                // 追加
    Reorder(ctx context.Context, ids []uint64) error                            // 追加
}

// Reorder はトランザクション内で sort_order を順番に更新
// Delete は pets テーブルで使用中の場合は ErrConflict を返す
```

### 4. Handler 拡張

```go
// backend/internal/handler/animal_species_handler.go に追加:
// GetAnimalSpecies     — GET    /masters/animal-species/:id
// CreateAnimalSpecies  — POST   /masters/animal-species
// UpdateAnimalSpecies  — PATCH  /masters/animal-species/:id
// DeleteAnimalSpecies  — DELETE /masters/animal-species/:id
// ReorderAnimalSpecies — PATCH  /masters/animal-species/reorder
```

### 5. ルーティング追加

```go
// backend/internal/handler/staff_handler.go の RegisterMasterRoutes() に追加:
masters.GET("/animal-species", h.ListAnimalSpecies)
masters.POST("/animal-species", h.CreateAnimalSpecies)
masters.PATCH("/animal-species/reorder", h.ReorderAnimalSpecies)  // 静的パス先行
masters.GET("/animal-species/:id", h.GetAnimalSpecies)
masters.PATCH("/animal-species/:id", h.UpdateAnimalSpecies)
masters.DELETE("/animal-species/:id", h.DeleteAnimalSpecies)
```

**重要: `/reorder` を `/:id` より前に登録すること。**

## API レスポンス形式

```json
// GET /v1/masters/animal-species
[
  { "id": 1, "name": "犬", "is_active": true, "sort_order": 1, "created_at": "...", "updated_at": "..." },
  { "id": 2, "name": "猫", "is_active": true, "sort_order": 2, "created_at": "...", "updated_at": "..." }
]

// POST /v1/masters/animal-species
{ "id": 7, "name": "フェレット", "is_active": true, "sort_order": 7, "created_at": "...", "updated_at": "..." }

// PATCH /v1/masters/animal-species/:id
{ "id": 1, "name": "犬（小型）", "is_active": true, "sort_order": 1, "created_at": "...", "updated_at": "..." }

// DELETE /v1/masters/animal-species/:id
204 No Content
// ※ pets で使用中の場合は 409 Conflict

// PATCH /v1/masters/animal-species/reorder
{ "ids": [2, 1, 3, 4, 5, 6] }
→ 204 No Content
```

## フロントエンド影響

- `make codegen` でモデル変更なし（model は変わらない）
- FE-013 で API hooks + マスタ設定画面 UI を追加

## 完了条件

- [ ] `animal_species_request.go` 新規作成（Create/Update/Reorder 型）
- [ ] Repository に FindByID/Create/Update/Delete/Reorder 追加
- [ ] Service に CRUD + Reorder + buildUpdateFields 追加
- [ ] Handler に Get/Create/Update/Delete/Reorder 追加
- [ ] ルーティング登録（`/reorder` を `/:id` より前に）
- [ ] Delete 時に pets テーブル参照チェック（使用中なら 409）
- [ ] 既存テストが通る
- [ ] `docker compose exec backend go test ./... -v` パス
