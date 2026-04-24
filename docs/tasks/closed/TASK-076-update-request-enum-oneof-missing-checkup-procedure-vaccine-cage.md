# TASK-076: Update リクエスト — ENUM フィールドの oneof バリデーション欠落 + checkup_type UpdateInput 値型

## 優先度

HIGH

---

## 概要

複数ドメインの Update リクエスト構造体で ENUM フィールドに `binding:"omitempty,oneof=..."` が付与されていない。
また、`checkup_type_service.go` の Update メソッドが値型 Input を受け取っており、
参照実装（medicine）との不統一がある。

---

## 違反一覧

### 1. checkup_type_service.go — UpdateInput 値型（TASK-050 パターン）

```go
// ❌ 値型で定義されている
type CheckupTypeService interface {
    Update(ctx context.Context, clinicID, id uint64, input UpdateCheckupTypeInput) (*model.CheckupType, error)
}

func (s *checkupTypeService) Update(ctx context.Context, clinicID, id uint64, input UpdateCheckupTypeInput) (*model.CheckupType, error) {
```

```go
// ✅ ポインタ型に統一
type CheckupTypeService interface {
    Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error)
}

func (s *checkupTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
```

**handler 側も合わせて修正（値型で構造体を作って `&svcInput` を渡す → `&service.UpdateCheckupTypeInput{...}` で統一）。**

---

### 2. procedure_request.go — updateProcedureRequest ENUM バリデーション欠落

```go
// ❌ Anesthesia, TaxType に oneof なし
type updateProcedureRequest struct {
    Name       *string  `json:"name"`
    Anesthesia *string  `json:"anesthesia"`          // binding なし
    TaxType    *string  `json:"tax_type"`             // binding なし
    // ...
}

// ✅ 修正後
type updateProcedureRequest struct {
    Name       *string  `json:"name"`
    Anesthesia *string  `json:"anesthesia"  binding:"omitempty,oneof=none local sedation general"`
    TaxType    *string  `json:"tax_type"    binding:"omitempty,oneof=included excluded exempt"`
    // ...
}
```

---

### 3. vaccine_request.go — updateVaccineRequest ENUM バリデーション欠落

```go
// ❌ Species に oneof なし
type updateVaccineRequest struct {
    Species *string `json:"species"`   // binding なし
    // ...
}

// ✅ 修正後
type updateVaccineRequest struct {
    Species *string `json:"species"  binding:"omitempty,oneof=dog cat both"`
    // ...
}
```

---

### 4. cage_request.go — updateCageRequest ENUM バリデーション欠落

```go
// ❌ CageType, CageSize に oneof なし
type updateCageRequest struct {
    CageType *string `json:"cage_type"`   // binding なし
    CageSize *string `json:"cage_size"`   // binding なし
    // ...
}

// ✅ 修正後（create request の oneof と同じ値を使用）
type updateCageRequest struct {
    CageType *string `json:"cage_type"  binding:"omitempty,oneof=icu dog cat general"`
    CageSize *string `json:"cage_size"  binding:"omitempty,oneof=small medium large"`
    // ...
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `checkup_type_service.go` | UpdateInput を値型 → ポインタ型に変更（Interface + 実装） |
| `checkup_type_handler.go` | Update 呼び出しを `&service.UpdateCheckupTypeInput{...}` に統一 |
| `procedure_request.go` | `Anesthesia`, `TaxType` に `binding:"omitempty,oneof=..."` 追加 |
| `vaccine_request.go` | `Species` に `binding:"omitempty,oneof=..."` 追加 |
| `cage_request.go` | `CageType`, `CageSize` に `binding:"omitempty,oneof=..."` 追加 |
