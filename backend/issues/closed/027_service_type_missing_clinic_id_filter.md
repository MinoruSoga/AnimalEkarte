# [BE-027] ServiceType CRUD の複合的な問題修正

## 概要

`service_types` の CRUD 実装に、clinic_id フィルタ欠落・GORMゼロ値スキップ・デッドコードを含む
複数の問題が存在する。マルチテナント要件違反かつ Update が実質壊れている状態。

---

## 問題一覧

### 1. clinic_id フィルタ未適用（マルチテナント違反）

`backend/internal/repository/service_type_repository.go`

| メソッド | 現状 | 影響 |
|---------|------|------|
| `FindAll` | `WHERE` なし | 全クリニックのデータを返す |
| `FindByID` | `WHERE id = ?` のみ | 他クリニックの ID を参照できる |
| `Delete` | `WHERE id = ?` のみ | 他クリニックのレコードを削除できる |

`Reorder` / `Update` は `clinic_id` フィルタ適用済み。

---

### 2. UpdateServiceType が実質壊れている（重大）

`backend/internal/handler/service_type_handler.go` の `UpdateServiceType` が `extractClinicID` を呼ばない。

```go
// ❌ 現状: ClinicID が 0 のまま渡される
serviceType := &model.ServiceType{
    ID:   id,
    Name: input.Name,
    // ClinicID: 未設定 → 0
}
```

リポジトリの `Update` は `WHERE id = ? AND clinic_id = ?` で絞るため、
`clinic_id = 0` に一致するレコードは存在せず、**常に RowsAffected=0 → NotFound** を返す。

---

### 3. Update リポジトリが GORM ゼロ値スキップ問題を抱えている

```go
// ❌ 現状: struct 渡しで GORM がゼロ値フィールドをスキップ
r.db.WithContext(ctx).Model(&model.ServiceType{}).
    Where("id = ? AND clinic_id = ?", serviceType.ID, serviceType.ClinicID).
    Updates(serviceType)
```

以下のフィールドは更新できない:
- `IsActive = false`（bool のゼロ値）
- `SortOrder = 0`（int のゼロ値）
- `Color = ""`（string のゼロ値）
- `Description = ""`（string のゼロ値）

`map[string]any` パターンに変更すべき。

---

### 4. ハンドラファイル内のインライン構造体定義（デッドコード）

`service_type_handler.go` に `createServiceTypeInput` / `updateServiceTypeInput` がインライン定義されており、
`service_type_request.go` の `createServiceTypeRequest` / `updateServiceTypeRequest` が使われていない。

```go
// service_type_handler.go 内（使われている）
type createServiceTypeInput struct { ... }
type updateServiceTypeInput struct { ... }

// service_type_request.go 内（デッドコード）
type createServiceTypeRequest struct { ... }
type updateServiceTypeRequest struct { ... }
```

ハンドラ内のインライン定義を削除し、request ファイルの型を使うよう統一する。

---

### 5. バインドエラーの非統一

`CreateServiceType` / `UpdateServiceType` のバインドエラーが `err.Error()` 直接出力になっており、
`parseBindError(err)` を使っていない。

```go
// ❌ 現状
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

// ✅ 統一すべき
c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
```

---

## 修正方針

### repository シグネチャ変更

```go
type ServiceTypeRepository interface {
    FindAll(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
    Create(ctx context.Context, serviceType *model.ServiceType) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error  // map[string]any へ変更
    Delete(ctx context.Context, clinicID, id uint64) error
    Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}
```

### service シグネチャ変更・Input DTO 追加

```go
type ServiceTypeService interface {
    List(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
    Create(ctx context.Context, clinicID uint64, input *CreateServiceTypeInput) (*model.ServiceType, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateServiceTypeInput) (*model.ServiceType, error)
    Delete(ctx context.Context, clinicID, id uint64) error
    Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type CreateServiceTypeInput struct {
    Name        string
    Color       string
    IsActive    bool
    Description string
    SortOrder   int
}

type UpdateServiceTypeInput struct {
    Name        *string
    Color       *string
    IsActive    *bool
    Description *string
    SortOrder   *int
}
```

### handler 修正点

- `ListServiceTypes`: `extractClinicID` 追加、`clinicID` を service に渡す
- `CreateServiceType`: インライン struct を削除し request 型に統一、`parseBindError` 使用
- `UpdateServiceType`: `extractClinicID` 追加、インライン struct 削除、`parseBindError` 使用
- `DeleteServiceType`: `extractClinicID` 追加、`clinicID` を service に渡す

---

## 影響範囲

- `backend/internal/repository/service_type_repository.go`
- `backend/internal/service/service_type_service.go`
- `backend/internal/handler/service_type_handler.go`
- `backend/internal/handler/service_type_request.go`（インライン定義を移動）
- `backend/internal/service/service_type_service_test.go`

## ステータス

- [ ] 実装
- [ ] テスト
- [ ] レビュー
