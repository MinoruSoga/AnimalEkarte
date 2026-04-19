# TASK-049: handler/DTO での model 型依存 — procedure / cage / vaccine（TASK-046 補完）

## 優先度

MEDIUM

---

## 概要

TASK-046 で `trimming_master_handler.go` の `model.TargetSize` 変換をハンドラ層から排除した。同パターンが `procedure_handler.go`・`cage_handler.go` にも存在し、さらに `vaccine_service.go` の Create DTO にも `*model.VaccineSpecies` 依存が残っている。

**ルール**: ハンドラ層は `model` パッケージへの型変換を行ってはならない。Input DTO は `*string` で受け取り、service 層（`buildXxxUpdateFields` または Create 関数）で `model.XxxType` に変換する。

---

## 問題 1: procedure_handler — Update で model 型変換

### ファイル
`backend/internal/handler/procedure_handler.go` L119-124（概算）

```go
// ❌ ハンドラが model 型に変換
if req.Anesthesia != nil {
    a := model.AnesthesiaType(*req.Anesthesia)
    svcInput.Anesthesia = &a
}
if req.TaxType != nil {
    t := model.TaxType(*req.TaxType)
    svcInput.TaxType = &t
}
```

**根本原因**: `UpdateProcedureInput.Anesthesia` が `*model.AnesthesiaType`、`TaxType` が `*model.TaxType` であるため、ハンドラが model パッケージに依存せざるを得ない。

---

## 問題 2: cage_handler — Update で model 型変換

### ファイル
`backend/internal/handler/cage_handler.go` L70, L104, L109（概算）

```go
// ❌ ハンドラが model 型に変換
if req.CageType != nil {
    ct := model.CageType(*req.CageType)
    svcInput.CageType = &ct
}
if req.CageSize != nil {
    cs := model.CageSize(*req.CageSize)
    svcInput.CageSize = &cs
}
```

**根本原因**: `UpdateCageInput.CageType` が `*model.CageType`、`CageSize` が `*model.CageSize`。

---

## 問題 3: vaccine_service — CreateVaccineInput.Species が model 型

### ファイル
`backend/internal/service/vaccine_service.go` L93（概算）

```go
// ❌ Create Input DTO が model パッケージに依存
type CreateVaccineInput struct {
    Name    string
    Species *model.VaccineSpecies  // ← model 型
    // ...
}
```

`CreateTrimmingCourseInput.TargetSize` は `string`、`CreateMedicineInput.TaxType` も `string` であるのに、vaccine だけ Create DTO で model 型を使っている。

---

## 修正方針

### 1. service DTO の ENUM フィールドを `*string` に変更

```go
// service/procedure_service.go
type UpdateProcedureInput struct {
    // ...
    Anesthesia *string  // *model.AnesthesiaType → *string
    TaxType    *string  // *model.TaxType → *string
}

// service/cage_service.go
type UpdateCageInput struct {
    // ...
    CageType *string  // *model.CageType → *string
    CageSize *string  // *model.CageSize → *string
}

// service/vaccine_service.go
type CreateVaccineInput struct {
    // ...
    Species string  // *model.VaccineSpecies → string
}
```

### 2. service 層で変換（buildUpdateFields または Create 関数内）

```go
// buildProcedureUpdateFields 内
if input.Anesthesia != nil {
    fields["anesthesia"] = model.AnesthesiaType(*input.Anesthesia)
}
if input.TaxType != nil {
    fields["tax_type"] = model.TaxType(*input.TaxType)
}

// buildCageUpdateFields 内
if input.CageType != nil {
    fields["cage_type"] = model.CageType(*input.CageType)
}
```

### 3. ハンドラを単純化

```go
// procedure_handler.go — if ブロック不要
svcInput := service.UpdateProcedureInput{
    Anesthesia: req.Anesthesia,  // *string のまま渡す
    TaxType:    req.TaxType,     // *string のまま渡す
    // ...
}
```

---

## 備考

- TASK-048（buildUpdateFields 定数化）と同時実施すると procedure / cage の修正が 1 コミットで完結する。
- `CreateVaccineInput.Species` が `*model.VaccineSpecies` である場合、handler 側の Create でも同様の変換が発生している可能性がある。handler も併せて確認・修正すること。
