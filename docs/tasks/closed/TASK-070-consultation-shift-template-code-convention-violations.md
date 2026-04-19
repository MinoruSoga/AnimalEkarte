# TASK-070: consultation / shift_template — 複数のコード規約違反（パターン系）

## 優先度

MEDIUM

---

## 概要

`consultation` と `shift_template` ドメインは今回の監査で初めて対象となったが、
他ドメインで TASK-048〜057 として修正対象となった同じパターンの違反が複数確認された。

---

## 違反一覧

### consultation_handler.go

#### 1. ListConsultations — mapSlice 未使用（TASK-056/060 パターン）

```go
// ❌ 手動ループ（L45-49 概算）
resp := make([]consultationResponse, len(consultations))
for i := range consultations {
    resp[i] = toConsultationResponse(&consultations[i])
}

// ✅ 修正後
c.JSON(http.StatusOK, mapSlice(consultations, toConsultationResponse))
```

#### 2. UpdateConsultation — handler 内で model.TaxType 型変換（TASK-049 パターン）

```go
// ❌ handler 内で model 型変換（L107-111 概算）
var taxType *model.TaxType
if input.TaxType != nil {
    tt := model.TaxType(*input.TaxType)
    taxType = &tt
}

// ✅ 変換は service 層（buildConsultationUpdateFields 内）で行う
```

### consultation_service.go

#### 3. slog clinic_id 順序不統一（TASK-057 パターン）

```go
// ❌ Create, Delete で entity_id が先頭（L88, L120 概算）
slog.InfoContext(ctx, "consultation created",
    slog.Uint64("consultation_id", consultation.ID),
    slog.Uint64("clinic_id", clinicID))  // ← clinic_id が後

// ✅ 修正後
slog.InfoContext(ctx, "consultation created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("consultation_id", consultation.ID))
```

#### 4. buildConsultationUpdateFields — bare string literals（TASK-048 パターン）

```go
// ❌ 文字列リテラル直書き（L152-187 概算）
fields["name"] = *input.Name
fields["price"] = *input.Price
fields["is_active"] = *input.IsActive
// ... 計 10 個のリテラル

// ✅ 定数化
const (
    colConsultationName          = "name"
    colConsultationPrice         = "price"
    colConsultationIsActive      = "is_active"
    colConsultationDescription   = "description"
    colConsultationTimeCondition = "time_condition"
    colConsultationDuration      = "duration"
    colConsultationParentID      = "parent_id"
    colConsultationSortOrder     = "sort_order"
    colConsultationTaxType       = "tax_type"
    colConsultationTaxRate       = "tax_rate"
)
```

### consultation_request.go

#### 5. TaxType に oneof バリデーション欠落（TASK-058 パターン）

```go
// ❌ binding タグなし
TaxType string `json:"tax_type"`

// ✅ 修正後
TaxType string `json:"tax_type" binding:"omitempty,oneof=included excluded exempt"`
```

---

### shift_template_handler.go

#### 6. List — mapSlice 未使用（TASK-056/060 パターン）

```go
// ❌ 手動ループ（L112-115 概算）
resp := make([]shiftTemplateResponse, len(templates))
for i := range templates {
    resp[i] = toShiftTemplateResponse(&templates[i])
}

// ✅ 修正後
c.JSON(http.StatusOK, mapSlice(templates, toShiftTemplateResponse))
```

---

### shift_template_service.go

#### 7. Delete 前の FK 依存チェック未確認

`shift_template` が他テーブルから FK 参照されている場合、Delete 前に CountUsage チェックが必要。
`shift_template_service.go` の Delete メソッドに依存チェックがないことを確認。

**確認事項:** `shift_entries` テーブルが `shift_template_id` FK を持つ場合は WrapConflict が必要。
マイグレーション SQL（`001_init.sql`）で `shift_entries.shift_template_id` の存在を確認すること。

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| consultation_handler.go | mapSlice 使用（List）/ model 型変換削除（Update） |
| consultation_service.go | slog 順序修正 / buildUpdateFields 定数化 |
| consultation_request.go | TaxType binding oneof 追加 |
| shift_template_handler.go | mapSlice 使用（List） |
| shift_template_service.go | FK 依存チェック追加（要仕様確認） |
