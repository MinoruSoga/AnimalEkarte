# TASK-124: `diagnosis_service.go` — DB カラム定数が Input DTO より前に定義されている

## 優先度

**Low** — コードの一貫性の問題。機能には影響しない。

---

## 概要

`diagnosis_service.go` において DB カラム定数ブロック（行 13-29）が
`CreateDiagnosisTypeInput` DTO（行 33 以降）より前に定義されている。

プロジェクト規約では「Input DTOs → 定数 → buildUpdateFields → interface → impl」の順序が定められている。
この順序では DTO が必ず定数より先に定義される。

---

## 問題箇所

### `service/diagnosis_service.go:13-40`

```go
// ❌ 定数が DTO より前に定義されている
const (  // 行 13
    colDiagnosisTypeName        = "name"
    colDiagnosisTypeIsActive    = "is_active"
    colDiagnosisTypeDescription = "description"
    colDiagnosisTypeSortOrder   = "sort_order"
)

const (  // 行 22（DiagnosisName 用）
    colDiagnosisNameName            = "name"
    colDiagnosisNameIsActive        = "is_active"
    colDiagnosisNameDescription     = "description"
    colDiagnosisNameSortOrder       = "sort_order"
    colDiagnosisNameDiagnosisTypeID = "diagnosis_type_id"
)

// ← ここで初めて DTO が登場
type CreateDiagnosisTypeInput struct { ... }  // 行 33
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/reservation_type_service.go の正しい順序
// 1. DTO（先に定義）
type CreateReservationTypeInput struct { ... }
type UpdateReservationTypeInput struct { ... }

// 2. 定数（DTOの後）
const (
    colReservationTypeName = "name"
    ...
)
```

---

## 修正方針

2 つの `const` ブロック（行 13-29）を `CreateDiagnosisTypeInput`（行 33）より後ろに移動する。

```go
// ✅ 修正後の順序

// ---- DiagnosisType Input DTOs ----
type CreateDiagnosisTypeInput struct { ... }
type UpdateDiagnosisTypeInput struct { ... }

// ---- DiagnosisName Input DTOs ----
type CreateDiagnosisNameInput struct { ... }
type UpdateDiagnosisNameInput struct { ... }

// ---- DB カラム定数 ----
const (
    colDiagnosisTypeName = "name"
    ...
)
const (
    colDiagnosisNameName = "name"
    ...
)

// ---- builder ----
func buildDiagnosisTypeUpdateFields(...) { ... }
func buildDiagnosisNameUpdateFields(...) { ... }

// ---- interfaces & impl ----
type DiagnosisTypeService interface { ... }
...
```

コードの移動のみ。ロジック変更なし。

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/diagnosis_service.go:13-29` | 2つの const ブロック | ❌ DTO より前に配置 |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/reservation_type_service.go` — DTO → 定数 → builder → interface の正しい順序
- 関連タスク: TASK-119（consultation_service でも const が DTO より前）
