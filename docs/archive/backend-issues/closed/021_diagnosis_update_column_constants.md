---
status: open
---

# diagnosis service: buildUpdateFields の列名を定数化

## 背景

`diagnosis_service.go` の `buildDiagnosisCategoryUpdateFields` / `buildDiagnosisNameUpdateFields` では
GORM update map のキーに文字列リテラルを直接使用している。

`pet_service.go` では列名を定数で管理しており、タイポによるサイレントバグを防いでいる。

## 問題

```go
// 現在の診断 service（文字列リテラル）
func buildDiagnosisCategoryUpdateFields(input *UpdateDiagnosisCategoryInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name          // ← タイポしてもコンパイルエラーにならない
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive  // ← 同上
    }
    // ...
}
```

## 修正方針

```go
// pet_service.go と同じパターンで定数化
const (
    colDiagnosisCategoryName        = "name"
    colDiagnosisCategoryIsActive    = "is_active"
    colDiagnosisCategoryDescription = "description"
    colDiagnosisCategorySortOrder   = "sort_order"

    colDiagnosisNameName               = "name"
    colDiagnosisNameIsActive           = "is_active"
    colDiagnosisNameDescription        = "description"
    colDiagnosisNameSortOrder          = "sort_order"
    colDiagnosisNameDiagnosisCategoryID = "diagnosis_category_id"
)

func buildDiagnosisCategoryUpdateFields(input *UpdateDiagnosisCategoryInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colDiagnosisCategoryName] = *input.Name
    }
    // ...
}
```

## 完了条件

- [ ] `diagnosis_service.go` にカラム名定数を定義
- [ ] `buildDiagnosisCategoryUpdateFields` を定数参照に変更
- [ ] `buildDiagnosisNameUpdateFields` を定数参照に変更
