# BUG-432: 複数マスタ request.go で Enum フィールドの binding:"oneof" タグが欠落・不整合

## 概要

複数のマスタリクエスト構造体で、Enum フィールドに `binding:"required,oneof=..."` または
`binding:"omitempty,oneof=..."` タグが欠落または不整合になっている。
バリデーションが Service 層のみに依存する状態で、Handler 層での早期弾きが効いていない。

---

## 問題箇所一覧

### 1. `consultation_request.go`

**Create リクエスト**:
```go
// consultation_request.go（問題行）
TaxType       string `json:"tax_type"`           // ← binding:"required,oneof=..." 欠落
TimeCondition string `json:"time_condition"`      // ← binding タグが完全に欠落
```

**Update リクエスト**:
```go
TaxType       *string `json:"tax_type"`           // ← binding:"omitempty,oneof=..." 欠落
```

**修正案**:
```go
// Create
TaxType       string `json:"tax_type"       binding:"required,oneof=included excluded exempt"`
TimeCondition string `json:"time_condition"  binding:"required,oneof=anytime morning afternoon evening"`

// Update
TaxType       *string `json:"tax_type"      binding:"omitempty,oneof=included excluded exempt"`
```

---

### 2. `reservation_type_request.go`

**Update リクエスト**:
```go
// reservation_type_request.go（問題行）
ReservationDayOption *string `json:"reservation_day_option"` // ← binding:"omitempty,oneof=..." 欠落
```

**Create リクエスト**:
```go
// Create では binding:"omitempty,oneof=general trimming" と設定されているが
// required ではなくオプショナル扱いになっている（設計意図の確認が必要）
Category string `json:"category" binding:"omitempty,oneof=general trimming"`
```

**修正案（Update）**:
```go
ReservationDayOption *string `json:"reservation_day_option" binding:"omitempty,oneof=everyday weekday weekend"`
// ← 実際の oneof 値はモデル定義から確認すること
```

---

### 3. `trimming_master_request.go`

**Update リクエスト**:
```go
// trimming_master_request.go（問題行）
TargetSize *string `json:"target_size"` // ← binding:"omitempty,oneof=..." 欠落
```

**Create リクエスト** では正しく `binding:"omitempty,oneof=..."` が設定されている。

**修正案（Update）**:
```go
TargetSize *string `json:"target_size" binding:"omitempty,oneof=small medium large"`
// ← 実際の oneof 値は Create リクエストの oneof 値と揃えること
```

---

### 4. `vaccine_request.go`

**Create リクエスト**:
```go
// vaccine_request.go（問題行）
Species string `json:"species" binding:"omitempty,oneof=dog cat both"` // ← Create で omitempty は不整合
```

**問題**: Create リクエストで `Species` が `omitempty` のため、空文字での登録が許可されてしまう。
必須 Enum フィールドなら `binding:"required,oneof=..."` にすべき。設計意図の確認が必要。

**修正案**:
```go
// Species が必須の場合
Species string `json:"species" binding:"required,oneof=dog cat both"`

// Species がオプショナルの場合（現状維持）
Species string `json:"species" binding:"omitempty,oneof=dog cat both"`
```

---

### 5. `medicine_request.go`

**Create リクエスト**:
```go
// medicine_request.go（問題行）
DosageForm string `json:"dosage_form"` // ← binding タグが完全に欠落
TaxType    string `json:"tax_type"`    // ← binding タグが完全に欠落
TaxRate    *int   `json:"tax_rate"`    // ← binding:"omitempty,min=0,max=100" 等が欠落
```

**Update リクエスト**:
```go
DosageForm *string `json:"dosage_form"` // ← binding:"omitempty,oneof=..." 欠落
TaxType    *string `json:"tax_type"`    // ← binding:"omitempty,oneof=..." 欠落
```

**修正案（Create）**:
```go
DosageForm string `json:"dosage_form" binding:"required,oneof=tablet capsule liquid powder injection"`
TaxType    string `json:"tax_type"    binding:"required,oneof=included excluded exempt"`
TaxRate    *int   `json:"tax_rate"    binding:"omitempty,min=0,max=100"`
```

**修正案（Update）**:
```go
DosageForm *string `json:"dosage_form" binding:"omitempty,oneof=tablet capsule liquid powder injection"`
TaxType    *string `json:"tax_type"    binding:"omitempty,oneof=included excluded exempt"`
```

---

## 影響ファイル

| ファイル | 問題フィールド | 重要度 |
|---------|-------------|--------|
| consultation_request.go | TaxType（Create/Update）、TimeCondition（Create） | Medium |
| reservation_type_request.go | ReservationDayOption（Update） | Low |
| trimming_master_request.go | TargetSize（Update） | Low |
| vaccine_request.go | Species（Create の omitempty 設計確認） | Low |
| medicine_request.go | DosageForm・TaxType（Create/Update）、TaxRate（Create） | Medium |

## 修正時の注意事項

1. 実際の `oneof=` 値は対応する Go モデルの定数定義（`model/*.go` の const ブロック）から取得すること
2. Service 層でもバリデーションは実施しているため、binding タグ追加は多層防御（defense in depth）として機能
3. Update リクエストの Enum フィールドは必ず `binding:"omitempty,oneof=..."` 形式にすること（nil を許容するため）

## 優先度

**Medium** — Handler 層でのバリデーション欠落。Service 層でも弾かれるが、不正なリクエストが Service まで到達し余分な処理が発生する。medicine の DosageForm・TaxType は必須フィールドのため影響度が高い。
