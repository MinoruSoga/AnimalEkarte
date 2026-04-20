# TASK-125: `reservation_type_group_service.go` — 定数が DTO より前 + `IsActive` 入力値が無視されている

## 優先度

**Medium** — `IsActive` の入力値無視は機能バグに近い（作成時に必ず `true` が設定される）。

---

## 概要

`reservation_type_group_service.go` に 2 つの問題がある:

1. DB カラム定数（行 17-22）が Input DTO より前に定義されている（順序違反）
2. `Create` メソッド（行 84）で `input.IsActive` を使わず `IsActive: true` をハードコードしており、DTO の `IsActive` フィールドが事実上無効になっている

---

## 問題箇所

### 問題 1: `service/reservation_type_group_service.go:12-22` — 定数が DTO より前

```go
// ❌ 定数が DTO より前に定義されている
const defaultGroupColor = "#3B82F6"  // 行 12

const (  // 行 17
    colReservationTypeGroupName      = "name"
    colReservationTypeGroupColor     = "color"
    colReservationTypeGroupSortOrder = "sort_order"
    colReservationTypeGroupIsActive  = "is_active"
)

// ← ここで初めて DTO が登場
type CreateReservationTypeGroupInput struct {  // 行 24
    Name      string
    Color     string
    SortOrder int
    IsActive  bool  // ← フィールドは定義されているが...
}
```

### 問題 2: `service/reservation_type_group_service.go:84` — `IsActive` ハードコード

```go
// ❌ input.IsActive を無視して常に true を設定
g := &model.ReservationTypeGroup{
    ClinicID:  clinicID,
    Name:      input.Name,
    Color:     color,
    SortOrder: input.SortOrder,
    IsActive:  true,  // ← input.IsActive を使うべき
}
```

`CreateReservationTypeGroupInput.IsActive bool` フィールドが定義されているにもかかわらず、
Create 実装で使用されていない。これにより「非アクティブ状態で作成する」ユースケースが機能しない。

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/cage_service.go の正しい順序（DTO → const → builder → interface → impl）
type CreateCageInput struct { ... }
type UpdateCageInput struct { ... }
const ( colCageName = "name" ... )
func buildCageUpdateFields(...) { ... }
type CageService interface { ... }

// ✅ service/cage_service.go の Create（IsActive を input から取得）
cage := &model.Cage{
    ClinicID:  clinicID,
    Name:      input.Name,
    IsActive:  input.IsActive,  // ← input 値を使用
    ...
}
```

---

## 修正方針

### 問題 1: 定数の移動

```go
// ✅ 修正後の順序
type CreateReservationTypeGroupInput struct { ... }   // 先に DTO
type UpdateReservationTypeGroupInput struct { ... }

const defaultGroupColor = "#3B82F6"   // 後に定数
const (
    colReservationTypeGroupName = "name"
    ...
)

func buildReservationTypeGroupUpdateFields(...) { ... }

type ReservationTypeGroupService interface { ... }
type reservationTypeGroupService struct { ... }
```

### 問題 2: `IsActive` を input から取得

```go
// ✅ 修正後
g := &model.ReservationTypeGroup{
    ClinicID:  clinicID,
    Name:      input.Name,
    Color:     color,
    SortOrder: input.SortOrder,
    IsActive:  input.IsActive,  // ← input 値を使用
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/reservation_type_group_service.go:12-22` | const ブロック（2つ） | ❌ DTO より前に配置 |
| `service/reservation_type_group_service.go:84` | `IsActive: true` ハードコード | ❌ input.IsActive を無視 |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/cage_service.go` — DTO → const → builder の正しい順序 + `input.IsActive` を使用
- 関連タスク: TASK-124（diagnosis_service の const 前置問題）
