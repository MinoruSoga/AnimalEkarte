# TASK-065: reorderRequest struct — 23ドメインで同一定義が重複

## 優先度

LOW

---

## 概要

`handler/` 配下に Reorder 系エンドポイント用の request struct が**23個、すべて同一構造**で重複定義されている。
共通の struct 1つにまとめることで冗長コードを削減できる。

---

## 現状

各ドメインの `*_request.go` が独自の `reorderXxxRequest` を定義している:

```go
// occupation_request.go
type reorderOccupationRequest struct {
    IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// vaccine_request.go
type reorderVaccineRequest struct {
    IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// ... 以下、全 23 ドメイン同様
```

**定義一覧（全 23 個）:**

| ファイル | struct 名 |
|---------|----------|
| animal_species_request.go | reorderAnimalSpeciesRequest |
| cage_request.go | reorderCageRequest |
| checkup_type_request.go | reorderCheckupTypeRequest |
| chief_complaint_request.go | reorderChiefComplaintRequest |
| consultation_request.go | reorderConsultationRequest |
| diagnosis_request.go | reorderDiagnosisTypeRequest / reorderDiagnosisNameRequest |
| exam_type_request.go | reorderExaminationTypeRequest |
| hospitalization_plan_request.go | reorderHospitalizationPlanRequest |
| inquiry_template_request.go | reorderInquiryTemplateRequest |
| insurance_request.go | reorderInsuranceRequest |
| medicine_request.go | reorderMedicineRequest |
| merchandise_item_request.go | reorderMerchandiseItemRequest |
| occupation_request.go | reorderOccupationRequest |
| permission_group_request.go | reorderPermissionGroupRequest |
| procedure_request.go | reorderProcedureRequest |
| reservation_type_group_request.go | reorderReservationTypeGroupRequest |
| reservation_type_request.go | reorderReservationTypeRequest |
| shift_template_handler.go | reorderShiftTemplateRequest |
| staff_request.go | reorderStaffRequest |
| trimming_master_request.go | reorderTrimmingCourseRequest / reorderTrimmingOptionRequest |
| vaccine_request.go | reorderVaccineRequest |

---

## 修正方針

`handler/request_helpers.go`（または `handler/slice_helpers.go`）に共通 struct を 1 つ定義し、
各ドメインの定義を削除して共通 struct を直接使う。

```go
// ✅ handler/request_helpers.go に一元定義
type reorderRequest struct {
    IDs []uint64 `json:"ids" binding:"required,min=1"`
}
```

各 handler ファイルでは `reorderRequest` を直接使用:
```go
// ✅ 修正後（occupation_handler.go）
var req reorderRequest
if err := c.ShouldBindJSON(&req); err != nil { ... }
```

---

## 削減効果

- 削除対象: 23 struct 定義 × 3行 = **約 69 行**
- 将来的な binding タグ変更時の修正箇所が 1 箇所に集約

---

## 備考

- `mapSlice` と同様に、共通ユーティリティとして一元管理する設計が望ましい
- binding タグが `binding:"required"` のみのものと `binding:"required,min=1"` のものが混在しているため、統一後は `min=1` を採用することを推奨
