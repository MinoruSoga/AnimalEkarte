# TASK-053: inquiry_template — sort_order あり Reorder 未実装

## 優先度

MEDIUM

---

## 概要

`inquiry_template` モデルは `sort_order int` フィールドを持つが、Reorder エンドポイント（並び順変更 API）が **handler・route の両方で未実装**である。  
同じ構造を持つ `occupation` / `vaccine` / `chief_complaint_type` などはすべて Reorder を実装しており、inquiry_template だけが欠落している。

---

## 確認事項

| 確認項目 | 状態 |
|---------|------|
| `model.InquiryTemplate.SortOrder` フィールド | ✅ あり |
| `InquiryTemplateService.Reorder` メソッド | **要確認**（service 実装が存在する可能性あり） |
| `InquiryTemplateRepository.Reorder` メソッド | **要確認** |
| `inquiry_template_handler.go` の `ReorderInquiryTemplates` | ❌ なし |
| `main.go` の `PATCH /inquiry-templates/reorder` route | ❌ なし |

---

## 修正方針

### Step 1: service 層の確認・追加

`inquiry_template_service.go` に `Reorder` が未実装であれば追加する。

```go
// Interface
Reorder(ctx context.Context, clinicID uint64, ids []uint64) error

// 実装（occupation_service.go の Reorder を参照）
func (s *inquiryTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder inquiry templates")
    }
    slog.InfoContext(ctx, "inquiry templates reordered",
        slog.Uint64("clinic_id", clinicID),
        slog.Int("count", len(ids)))
    return nil
}
```

### Step 2: repository 層の確認・追加

`inquiry_template_repository.go` に `Reorder` が未実装であれば `reorderByClinicID` ヘルパーを使って追加する。

```go
func (r *inquiryTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    return reorderByClinicID(ctx, r.db, "inquiry_templates", clinicID, ids)
}
```

### Step 3: handler に `ReorderInquiryTemplates` を追加

```go
// inquiry_template_handler.go
func (h *InquiryTemplateHandler) ReorderInquiryTemplates(c *gin.Context) {
    ctx := c.Request.Context()
    clinicID, err := extractClinicID(c)
    if err != nil {
        RespondError(c, err)
        return
    }
    var req reorderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    if err := h.service.Reorder(ctx, clinicID, req.IDs); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}
```

### Step 4: route 登録

```go
// main.go (RegisterMasterRoutes)
masters.PATCH("/inquiry-templates/reorder", inquiryTemplateHandler.ReorderInquiryTemplates)
```

---

## 参照実装

`occupation_handler.go` / `occupation_service.go` / `occupation_repository.go` の Reorder 実装をそのままテンプレートとして使用できる。
