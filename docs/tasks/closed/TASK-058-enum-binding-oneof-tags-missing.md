# TASK-058: ENUM フィールドに binding `oneof` タグが未設定 — 8ドメイン

## 優先度

MEDIUM

---

## 概要

Create/Update リクエスト struct の ENUM string フィールドに `binding:"required,oneof=val1 val2"` タグが設定されていない。  
バリデーションなしでは、不正な ENUM 値がそのまま DB に到達し、GORM が PostgreSQL エラーを返す（500 Internal Server Error になる可能性がある）。

参照実装 `medicine_request.go` では ENUM フィールドを `oneof` で正しく制約している。

---

## 参照実装（medicine_request.go）

```go
// ✅ ENUM フィールドに oneof で制約
type CreateMedicineRequest struct {
    Name        string  `json:"name" binding:"required"`
    DosageForm  string  `json:"dosage_form" binding:"required,oneof=tablet liquid injection powder cream"`
    TaxType     string  `json:"tax_type" binding:"required,oneof=included excluded"`
    // ...
}
```

---

## 違反箇所一覧

| リクエスト struct | フィールド | ENUM 値 | 現状 |
|-----------------|-----------|---------|------|
| `CreateVaccineRequest` | `Species` | `dog cat rabbit ...` | `binding:"required"` のみ |
| `CreateVaccineRequest` | `Interval` | `monthly yearly ...` | `binding:"required"` のみ |
| `CreateProcedureRequest` | `Anesthesia` | `local general none` | `binding:"required"` のみ |
| `CreateProcedureRequest` | `TaxType` | `included excluded` | `binding:"required"` のみ |
| `CreateCageRequest` | `CageType` | `dog cat small_animal ...` | `binding:"required"` のみ |
| `CreateCageRequest` | `CageSize` | `ss s m l ll` | `binding:"required"` のみ |
| `CreateCheckupTypeRequest` | `Interval` | `monthly yearly ...` | `binding:"required"` のみ |
| `CreateCheckupTypeRequest` | `TargetAge` | `puppy adult senior ...` | `binding:"required"` のみ |
| `CreateTrimmingCourseRequest` | `TargetSize` | `ss s m l ll` | `binding:"required"` のみ |
| `CreateHospitalizationPlanRequest` | `BodySize` | `ss s m l ll` | タグなし |
| `CreateHospitalizationPlanRequest` | `BillingUnit` | `per_day per_night` | タグなし |
| `CreateHospitalizationPlanRequest` | `TaxType` | `included excluded` | タグなし |
| `CreateReservationTypeRequest` | `ReservationDayOption` | `weekday weekend ...` | `binding:"required"` のみ |
| `CreateMerchandiseItemRequest` | `TaxType` | `included excluded` | `binding:"required"` のみ |

---

## 修正方針

各 ENUM フィールドの `binding` タグに `oneof=<値リスト>` を追加する。  
Update 系 Request で `omitempty` を使っている場合は `binding:"omitempty,oneof=..."` とする。

```go
// ❌ 修正前
Species  string `json:"species" binding:"required"`

// ✅ 修正後
Species  string `json:"species" binding:"required,oneof=dog cat rabbit ferret bird hamster other"`
```

**ENUM 値リストの確認方法:** `backend/internal/model/` の対応する model ファイルの const 定義を参照する。

---

## 修正ファイル

- `backend/internal/handler/vaccine_request.go`
- `backend/internal/handler/procedure_request.go`
- `backend/internal/handler/cage_request.go`
- `backend/internal/handler/checkup_type_request.go`
- `backend/internal/handler/trimming_master_request.go`
- `backend/internal/handler/hospitalization_plan_request.go`
- `backend/internal/handler/reservation_type_request.go`
- `backend/internal/handler/merchandise_item_request.go`

---

## 備考

- TASK-052（service 層の ENUM バリデーター追加）と対になる修正。handler 層の binding タグでまず弾き、service 層でも念のため検証する二重防御が理想。
- `oneof` 値は model の const 定義と完全一致させること。
