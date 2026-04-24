# TASK-094: vaccine_response — Species が *string 型（*model.VaccineSpecies に統一すべき）

## 優先度

LOW

---

## 概要

`vaccine_response.go` の `Species` フィールドが生の `*string` 型で定義されており、
モデルの `*model.VaccineSpecies` 型と不統一。TASK-086（merchandise_item）と同パターン。

変換関数内で不要なポインタ変換コードが発生している。

---

## 問題箇所

### vaccine_response.go

```go
// ❌ 生の *string 型
type vaccineResponse struct {
    // ...
    Species *string `json:"species,omitempty"` // ← *model.VaccineSpecies ではない
    // ...
}

// ❌ 変換時に不要なキャスト
func toVaccineResponse(v *model.Vaccine) vaccineResponse {
    var species *string
    if v.Species != nil {
        s := string(*v.Species)  // ← 冗長なキャスト
        species = &s
    }
    return vaccineResponse{
        // ...
        Species: species,
        // ...
    }
}
```

---

## モデル定義

```go
// model/vaccine.go
type VaccineSpecies string

const (
    VaccineSpeciesDog  VaccineSpecies = "dog"
    VaccineSpeciesCat  VaccineSpecies = "cat"
    VaccineSpeciesBoth VaccineSpecies = "both"
)

type Vaccine struct {
    Species *VaccineSpecies `gorm:"type:vaccine_species" json:"species,omitempty"`
}
```

---

## 参照実装（TASK-086 で修正予定の procedure_response.go パターン）

```go
// ✅ model 型を直接使用
type procedureResponse struct {
    TaxType model.TaxType `json:"tax_type"`
}

func toProcedureResponse(p *model.Procedure) procedureResponse {
    return procedureResponse{
        TaxType: p.TaxType,  // キャスト不要
    }
}
```

---

## 修正方針

```go
// ✅ 修正後
type vaccineResponse struct {
    // ...
    Species *model.VaccineSpecies `json:"species,omitempty"`
    // ...
}

func toVaccineResponse(v *model.Vaccine) vaccineResponse {
    return vaccineResponse{
        // ...
        Species: v.Species,  // キャスト不要、直接代入
        // ...
    }
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/vaccine_response.go` | `Species *string` → `*model.VaccineSpecies`、変換コードのポインタキャスト処理を削除 |

---

## 関連

- TASK-086: merchandise_item_response の TaxType/Category 同種問題
