# TASK-048: buildUpdateFields ベタ文字列 — 残存 7 ドメイン（TASK-044/047 補完）

## 優先度

MEDIUM

---

## 概要

TASK-044（6 ドメイン）・TASK-047（trimming 2 ドメイン）で `buildXxxUpdateFields` 内カラム名の定数化を進めたが、以下の 7 ドメインが未対応のまま残存している。本タスクで補完する。

参照実装: `medicine_service.go`（`colMedicineName = "name"` 等の定数ブロックを先頭に定義）

---

## 対象ファイルと該当箇所

| ドメイン | ファイル | 関数 | 概算行 |
|---------|---------|------|------|
| vaccine | `backend/internal/service/vaccine_service.go` | `buildVaccineUpdateFields` | L100-129 |
| occupation | `backend/internal/service/occupation_service.go` | `buildOccupationUpdateFields` | L133-148 |
| inquiry_template | `backend/internal/service/inquiry_template_service.go` | `buildInquiryTemplateUpdateFields` | L118-136 |
| insurance | `backend/internal/service/insurance_service.go` | `buildInsuranceUpdateFields` | L128-149 |
| exam_type | `backend/internal/service/exam_type_service.go` | `buildExamTypeUpdateFields` | L136-159 |
| procedure | `backend/internal/service/procedure_service.go` | `buildProcedureUpdateFields` | L154-189 |
| cage | `backend/internal/service/cage_service.go` | `buildCageUpdateFields` | L134-158 |

---

## 問題

```go
// ❌ 現状（例: insurance_service.go）
func buildInsuranceUpdateFields(input UpdateInsuranceInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name           // ← ベタ文字列
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive  // ← ベタ文字列
    }
    // ...
}
```

---

## 修正方針

各 service ファイルの先頭に定数ブロックを追加し、`buildXxxUpdateFields` 内のすべてのベタ文字列を定数参照に置き換える。

```go
// ✅ 修正後（例: insurance_service.go）
const (
    colInsuranceName      = "name"
    colInsuranceIsActive  = "is_active"
    // 実際のフィールドに合わせて追加
)

func buildInsuranceUpdateFields(input UpdateInsuranceInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil     { fields[colInsuranceName]     = *input.Name }
    if input.IsActive != nil { fields[colInsuranceIsActive] = *input.IsActive }
    return fields
}
```

---

## 備考

- TASK-049 で procedure / cage の ENUM フィールド型変更も行う場合、`buildProcedureUpdateFields` / `buildCageUpdateFields` の修正と同時実施が効率的。
- 7 ドメインをまとめて 1 コミットで対応してよい。
