# TASK-073: 全 Create ハンドラ — Location ヘッダ欠落（medicine 参照実装との不統一）

## 優先度

MEDIUM

---

## 概要

`medicine_handler.go` の CreateMedicine は HTTP 201 レスポンスに `Location` ヘッダを付与しているが、
他の全マスタドメインの Create ハンドラには Location ヘッダが実装されていない。
REST ベストプラクティスとしてリソース作成時には Location ヘッダで新規 URI を示すべきであり、
参照実装との統一が必要。

---

## 参照実装

```go
// medicine_handler.go L87-88
c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
c.JSON(http.StatusCreated, toMedicineResponse(medicine))
```

---

## 対象ファイル（全 11 ドメイン）

| ファイル | 修正箇所 | 修正内容 |
|---------|---------|---------|
| `exam_type_handler.go` | CreateExaminationType | Location ヘッダ追加 |
| `checkup_type_handler.go` | CreateCheckupType | Location ヘッダ追加 |
| `procedure_handler.go` | CreateProcedure | Location ヘッダ追加 |
| `vaccine_handler.go` | CreateVaccine | Location ヘッダ追加 |
| `occupation_handler.go` | CreateOccupation | Location ヘッダ追加 |
| `inquiry_template_handler.go` | CreateInquiryTemplate | Location ヘッダ追加 |
| `reservation_type_group_handler.go` | CreateReservationTypeGroup | Location ヘッダ追加 |
| `cage_handler.go` | CreateCage | Location ヘッダ追加 |
| `chief_complaint_handler.go` | CreateChiefComplaint | Location ヘッダ追加 |
| `diagnosis_handler.go` | CreateDiagnosisType, CreateDiagnosisName | Location ヘッダ追加（2か所） |
| `insurance_handler.go` | CreateInsurance | Location ヘッダ追加 |

---

## 修正パターン

```go
// ❌ 修正前
c.JSON(http.StatusCreated, toXxxResponse(entity))

// ✅ 修正後
c.Header("Location", fmt.Sprintf("/v1/masters/xxx/%d", entity.ID))
c.JSON(http.StatusCreated, toXxxResponse(entity))
```
