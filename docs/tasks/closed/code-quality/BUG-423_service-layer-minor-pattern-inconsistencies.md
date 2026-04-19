# BUG-423: サービス層の微細な実装パターン不統一群

## 概要

複数のマスタサービスで、単独チケット化するほどではないが統一性に欠ける実装パターンが散見される。
将来の保守性低下を防ぐため、一括対応を推奨する。

---

## 問題 1: map 初期化パターンの不統一

### 問題箇所
```go
// occupation_service.go:141
func buildOccupationUpdateFields(input *UpdateOccupationInput) map[string]any {
    fields := map[string]any{}   // ← 直接初期化（非推奨）

// chief_complaint_service.go:144
func buildChiefComplaintTypeUpdateFields(...) map[string]any {
    fields := map[string]any{}   // ← 直接初期化（非推奨）
```

### 他サービスの標準パターン
```go
// cage_service.go:160
fields := make(map[string]any)   // ← make() 使用（推奨）

// insurance_service.go:149
fields := make(map[string]any)   // ← make() 使用
```

### 修正
`map[string]any{}` → `make(map[string]any)` に統一。

---

## 問題 2: Input DTO 定義順序の逆転

### 問題箇所
```go
// occupation_service.go:15-22
type UpdateOccupationInput struct { ... }  // Update が先（逆）
// ...
type CreateOccupationInput struct { ... }  // Create が後
```

### 標準パターン（全他サービス）
```go
// cage_service.go
type CreateCageInput struct { ... }    // Create が先
// ...
type UpdateCageInput struct { ... }   // Update が後
```

### 修正
`occupation_service.go` で CreateOccupationInput と UpdateOccupationInput の定義順を入れ替える。

---

## 問題 3: slog 属性の出力順序不統一

### 問題箇所
```go
// trimming_service.go:167-169（Create ログ）
slog.InfoContext(ctx, "trimming appointment created",
    slog.Uint64("appointment_id", apptID),   // ← appointment_id が先（逆）
    slog.Uint64("clinic_id", clinicID))

// trimming_service.go:250-252（Update ログ）
slog.InfoContext(ctx, "trimming appointment updated",
    slog.Uint64("appointment_id", id),        // ← appointment_id が先（逆）
    slog.Uint64("clinic_id", clinicID))
```

### 標準パターン（全他サービス）
```go
// cage_service.go:81-83
slog.InfoContext(ctx, "cage created",
    slog.Uint64("clinic_id", clinicID),       // ← clinic_id が先（標準）
    slog.Uint64("cage_id", cage.ID))
```

### 修正
`trimming_service.go` の2箇所で `clinic_id` を最初の属性に移動。

---

## 問題 4: 列名定数の定義位置不統一

### 問題箇所
```go
// consultation_service.go:13-24（ファイル先頭）
const (
    colConsultationName  = "name"
    colConsultationPrice = "price"
    ...
)
// ← import 直後に定数定義（型・関数定義より前）

// diagnosis_service.go:13-23（同様）
const (
    colDiagnosisTypeName = "name"
    ...
)
```

### 標準パターン（多くのサービス）
```go
// cage_service.go
// Service type 定義 → 各メソッド → buildUpdateFields → const 定数（ファイル末尾）
```

### 修正方針
定数定義の位置を、使用箇所（`buildXxxUpdateFields`）の直前またはファイル末尾に統一する。
（`consultation_service.go`、`diagnosis_service.go` が対象）

---

## 問題 5: merchandise_item_service の slog Create/Update ログ非対称

### 問題箇所
```go
// merchandise_item_service.go:145-149（Create）
slog.InfoContext(ctx, "merchandise item created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("merchandise_item_id", item.ID),
    slog.String("name", item.Name))          // ← name あり

// merchandise_item_service.go:169-172（Update）
slog.InfoContext(ctx, "merchandise item updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("merchandise_item_id", id))  // ← name なし
```

Create に `name` を含めているのに Update には含めない。一貫性がなく、判断基準が不明確。

### 修正
プロジェクト標準（`clinic_id` + エンティティ ID のみ）に合わせ、Create の `name` を削除する。

---

## 対象ファイル

| 問題 | ファイル | 行番号 |
|-----|---------|--------|
| 1. map 初期化 | occupation_service.go | 141 |
| 1. map 初期化 | chief_complaint_service.go | 144 |
| 2. DTO 定義順序 | occupation_service.go | 15-22 |
| 3. slog 順序 | trimming_service.go | 167-169, 250-252 |
| 4. 定数位置 | consultation_service.go | 13-24 |
| 4. 定数位置 | diagnosis_service.go | 13-23 |
| 5. slog 非対称 | merchandise_item_service.go | 145-149 |

## 優先度

**Low** — 動作への影響なし。コードの一貫性・保守性向上のためのリファクタリング対象。
一括で PR 1本にまとめて対応することを推奨。
