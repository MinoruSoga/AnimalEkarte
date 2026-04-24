# TASK-047: buildUpdateFields ベタ文字列 — trimming ドメイン（TASK-044 補完）

## 優先度

MEDIUM

---

## 概要

TASK-044 で checkup_type / chief_complaint_type / reservation_type_group / reservation_type / company / hospitalization_plan の 6 ドメインで `buildXxxUpdateFields` 内のカラム名がベタ文字列リテラルであることを指摘した。

trimming_course と trimming_option も同パターンであり TASK-044 の対象から漏れていた。本タスクで補完する。

---

## 対象ファイル

`backend/internal/service/trimming_master_service.go`

- `buildTrimmingCourseUpdateFields`（L134-158）
- `buildTrimmingOptionUpdateFields`（L278-302）

---

## 問題

```go
// ❌ buildTrimmingCourseUpdateFields（L134-158）
func buildTrimmingCourseUpdateFields(input UpdateTrimmingCourseInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name           // ← ベタ文字列
    }
    if input.Price != nil {
        fields["price"] = *input.Price         // ← ベタ文字列
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive  // ← ベタ文字列
    }
    // ...
}

// ❌ buildTrimmingOptionUpdateFields（L278-302）
func buildTrimmingOptionUpdateFields(input UpdateTrimmingOptionInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name           // ← ベタ文字列
    }
    // ...
}
```

---

## 修正方針

参照実装 `medicine_service.go` と同様に、ファイル先頭に定数ブロックを追加してすべて定数参照に置き換える。

```go
// ✅ 修正後
const (
    colTrimmingCourseName        = "name"
    colTrimmingCoursePrice       = "price"
    colTrimmingCourseIsActive    = "is_active"
    colTrimmingCourseDescription = "description"
    colTrimmingCourseTargetSize  = "target_size"
    colTrimmingCourseDuration    = "duration"
    colTrimmingCourseSortOrder   = "sort_order"

    colTrimmingOptionName         = "name"
    colTrimmingOptionPrice        = "price"
    colTrimmingOptionIsActive     = "is_active"
    colTrimmingOptionDescription  = "description"
    colTrimmingOptionDuration     = "duration"
    colTrimmingOptionIsCombinable = "is_combinable"
    colTrimmingOptionSortOrder    = "sort_order"
)

func buildTrimmingCourseUpdateFields(input UpdateTrimmingCourseInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil        { fields[colTrimmingCourseName]        = *input.Name }
    if input.Price != nil       { fields[colTrimmingCoursePrice]       = *input.Price }
    if input.IsActive != nil    { fields[colTrimmingCourseIsActive]    = *input.IsActive }
    if input.Description != nil { fields[colTrimmingCourseDescription] = *input.Description }
    if input.TargetSize != nil  { fields[colTrimmingCourseTargetSize]  = *input.TargetSize }
    if input.Duration != nil    { fields[colTrimmingCourseDuration]    = *input.Duration }
    if input.SortOrder != nil   { fields[colTrimmingCourseSortOrder]   = *input.SortOrder }
    return fields
}
```

---

## 備考

TASK-046 の修正（`UpdateTrimmingCourseInput.TargetSize` を `*string` に変更）と合わせて実施することで、`buildTrimmingCourseUpdateFields` 内の型変換も同時に整理できる。
