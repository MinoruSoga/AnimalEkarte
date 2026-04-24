# TASK-152: medicine_handler.go — ポインタリテラルを使わず中間変数を介している

**優先度**: Low
**対象ファイル**: `backend/internal/handler/medicine_handler.go`
**チェック項目**: 3（Handler の Input 型 — ポインタリテラル）

---

## 問題

プロジェクト規約では Handler から Service に Input を渡す際、中間変数を作らずポインタリテラルで直接渡すことを求めている。

`CreateMedicine` および `UpdateMedicine` ハンドラでは、`input :=` として変数に代入してから `&input` を渡しており、規約と一致していない。

---

## 現状コード

### CreateMedicine（medicine_handler.go 67〜83行）

```go
input := service.CreateMedicineInput{
    Name:            req.Name,
    ParentID:        req.ParentID,
    Price:           req.Price,
    IsActive:        req.IsActive,
    Description:     req.Description,
    DosageForm:      req.DosageForm,
    MedicineUnit:    req.MedicineUnit,
    InventoryID:     req.InventoryID,
    DefaultQuantity: req.DefaultQuantity,
    SortOrder:       req.SortOrder,
    TaxType:         req.TaxType,
    TaxRate:         req.TaxRate,
}

medicine, err := h.svc.Medicine.Create(c.Request.Context(), clinicID, &input)
```

### UpdateMedicine（medicine_handler.go 108〜124行）

```go
input := service.UpdateMedicineInput{
    Name:            req.Name,
    ParentID:        req.ParentID,
    ClearParentID:   req.ClearParentID,
    Price:           req.Price,
    IsActive:        req.IsActive,
    Description:     req.Description,
    DosageForm:      req.DosageForm,
    MedicineUnit:    req.MedicineUnit,
    InventoryID:     req.InventoryID,
    DefaultQuantity: req.DefaultQuantity,
    SortOrder:       req.SortOrder,
    TaxType:         req.TaxType,
    TaxRate:         req.TaxRate,
}

medicine, err := h.svc.Medicine.Update(c.Request.Context(), clinicID, id, &input)
```

---

## 修正後コード

### CreateMedicine

```go
medicine, err := h.svc.Medicine.Create(c.Request.Context(), clinicID, &service.CreateMedicineInput{
    Name:            req.Name,
    ParentID:        req.ParentID,
    Price:           req.Price,
    IsActive:        req.IsActive,
    Description:     req.Description,
    DosageForm:      req.DosageForm,
    MedicineUnit:    req.MedicineUnit,
    InventoryID:     req.InventoryID,
    DefaultQuantity: req.DefaultQuantity,
    SortOrder:       req.SortOrder,
    TaxType:         req.TaxType,
    TaxRate:         req.TaxRate,
})
```

### UpdateMedicine

```go
medicine, err := h.svc.Medicine.Update(c.Request.Context(), clinicID, id, &service.UpdateMedicineInput{
    Name:            req.Name,
    ParentID:        req.ParentID,
    ClearParentID:   req.ClearParentID,
    Price:           req.Price,
    IsActive:        req.IsActive,
    Description:     req.Description,
    DosageForm:      req.DosageForm,
    MedicineUnit:    req.MedicineUnit,
    InventoryID:     req.InventoryID,
    DefaultQuantity: req.DefaultQuantity,
    SortOrder:       req.SortOrder,
    TaxType:         req.TaxType,
    TaxRate:         req.TaxRate,
})
```

---

## 修正手順

`medicine_handler.go` の `CreateMedicine`・`UpdateMedicine` で中間変数 `input :=` を削除し、`h.svc.Medicine.Create/Update(...)` の引数にポインタリテラルを直接渡すよう変更する。

