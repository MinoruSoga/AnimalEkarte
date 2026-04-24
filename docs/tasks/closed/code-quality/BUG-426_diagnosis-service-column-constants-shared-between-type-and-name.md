# BUG-426: diagnosis_service で DiagnosisType と DiagnosisName の列名定数が共用されている

## 概要

`diagnosis_service.go` の `buildDiagnosisNameUpdateFields` 関数が、
DiagnosisName 用の独自定数を持たず DiagnosisType 用の定数（`colDiagnosisTypeName` 等）を流用している。
両テーブルのカラム名が現時点で一致しているため動作はするが、意図が不明確で保守性を損なう。
DiagnosisType または DiagnosisName のカラム名が将来変更された場合にバグが潜在する。

## 問題箇所

```go
// diagnosis_service.go:13-23（定数定義）
const (
    colDiagnosisTypeName        = "name"        // DiagnosisType 用として命名
    colDiagnosisTypeIsActive    = "is_active"
    colDiagnosisTypeDescription = "description"
    colDiagnosisTypeSortOrder   = "sort_order"
    colDiagnosisTypeIsSelected  = "is_selected"
    colDiagnosisNameDiagnosisTypeID = "diagnosis_type_id"
)
```

```go
// diagnosis_service.go:316-333（buildDiagnosisNameUpdateFields）
func buildDiagnosisNameUpdateFields(input *UpdateDiagnosisNameInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colDiagnosisTypeName] = *input.Name       // ← Type 定数を Name で使用
    }
    if input.IsActive != nil {
        fields[colDiagnosisTypeIsActive] = *input.IsActive  // ← 同上
    }
    if input.Description != nil {
        fields[colDiagnosisTypeDescription] = *input.Description  // ← 同上
    }
    if input.SortOrder != nil {
        fields[colDiagnosisTypeSortOrder] = *input.SortOrder  // ← 同上
    }
    // ...
}
```

## 修正方針

DiagnosisName 専用の定数ブロックを追加し、明示的に分離する。

```go
// diagnosis_service.go — 修正後
const (
    // DiagnosisType 列名
    colDiagnosisTypeName        = "name"
    colDiagnosisTypeIsActive    = "is_active"
    colDiagnosisTypeDescription = "description"
    colDiagnosisTypeSortOrder   = "sort_order"
    colDiagnosisTypeIsSelected  = "is_selected"
)

const (
    // DiagnosisName 列名
    colDiagnosisNameName           = "name"
    colDiagnosisNameIsActive       = "is_active"
    colDiagnosisNameDescription    = "description"
    colDiagnosisNameSortOrder      = "sort_order"
    colDiagnosisNameDiagnosisTypeID = "diagnosis_type_id"
)
```

```go
// buildDiagnosisNameUpdateFields — 修正後
func buildDiagnosisNameUpdateFields(input *UpdateDiagnosisNameInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colDiagnosisNameName] = *input.Name       // ← Name 専用定数を使用
    }
    if input.IsActive != nil {
        fields[colDiagnosisNameIsActive] = *input.IsActive
    }
    // ...
}
```

## 影響ファイル

- `backend/internal/service/diagnosis_service.go` — 行 13-23（定数定義）、行 316-333（buildDiagnosisNameUpdateFields）

## 優先度

**Medium** — 現時点では動作するが、テーブル分岐時にサイレントバグの原因になる。定数の命名意図が不明確。
