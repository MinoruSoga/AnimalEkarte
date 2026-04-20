# TASK-116: `checkup_type_service.go` — UpdateCheckupTypeInput・定数・buildUpdateFields の定義順序が逆

## 優先度

**Low** — 機能に影響はないが、ファイル内の構造が他マスタサービスと統一されておらず可読性が低下している。

---

## 概要

`checkup_type_service.go` において `UpdateCheckupTypeInput` DTO、DB カラム定数、
`buildCheckupTypeUpdateFields` ヘルパーがすべて **ファイル末尾（136〜189行）** に定義されている。

他のマスタサービス（`reservation_type_service.go`, `cage_service.go`, `medicine_service.go` 等）は
以下の統一順序に従っている:

```
1. Input DTOs（CreateXxxInput, UpdateXxxInput）
2. DB カラム定数（const colXxx = "..."）
3. buildXxxUpdateFields ヘルパー
4. Service インターフェース
5. Service 実装（構造体 + メソッド群）
```

しかし `checkup_type_service.go` は:

```
1. CreateCheckupTypeInput（16行）← 正しい
2. Service インターフェース（27行）
3. Service 実装（36行〜）     ← Update 実装が先に来る
4. DB カラム定数（136行）
5. UpdateCheckupTypeInput（147行）← UpdateDTO がメソッド実装より後に定義
6. buildCheckupTypeUpdateFields（160行）
```

---

## 問題箇所

### `service/checkup_type_service.go:136-189`

```go
// ❌ メソッド実装（Reorder:120〜131）の後に定数と UpdateDTO が定義されている
func (s *checkupTypeService) Reorder(...) error {
    ...
    return nil  // 行 131
}

// ← ここで定数・UpdateDTO・builder が突然登場
const (  // 行 136
    colCheckupTypeName        = "name"
    ...
)

type UpdateCheckupTypeInput struct {  // 行 147
    Name          *string
    ...
}

func buildCheckupTypeUpdateFields(...) map[string]any {  // 行 160
    ...
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ reservation_type_service.go の正しい順序
// 1. Input DTOs（15〜60行）
type CreateReservationTypeInput struct { ... }
type UpdateReservationTypeInput struct { ... }

// 2. DB カラム定数（62〜81行）
const (
    colReservationTypeName = "name"
    ...
)

// 3. buildUpdateFields ヘルパー（84〜134行）
func buildReservationTypeUpdateFields(...) map[string]any { ... }

// 4. Service インターフェース
type ReservationTypeService interface { ... }

// 5. Service 実装
type reservationTypeService struct { ... }
func (s *reservationTypeService) List(...) { ... }
```

---

## 修正方針

`const`（行 136-145）・`UpdateCheckupTypeInput`（行 147-158）・`buildCheckupTypeUpdateFields`（行 160-189）
を `CreateCheckupTypeInput`（行 15-25）の直後（行 26 以降）に移動する。

```go
// ✅ 修正後の順序

// ---- Input DTOs ----
type CreateCheckupTypeInput struct { ... }  // 既存位置（行 16）

// ← ここに UpdateCheckupTypeInput を移動
type UpdateCheckupTypeInput struct {
    Name          *string
    Price         *int64
    IsActive      *bool
    Description   *string
    Interval      *string
    TargetAge     *string
    ParentID      *uint64
    ClearParentID bool
    SortOrder     *int
}

// ← ここに定数を移動
const (
    colCheckupTypeName        = "name"
    colCheckupTypePrice       = "price"
    colCheckupTypeIsActive    = "is_active"
    colCheckupTypeDescription = "description"
    colCheckupTypeInterval    = "interval"
    colCheckupTypeTargetAge   = "target_age"
    colCheckupTypeParentID    = "parent_id"
    colCheckupTypeSortOrder   = "sort_order"
)

// ← ここに buildCheckupTypeUpdateFields を移動
func buildCheckupTypeUpdateFields(input *UpdateCheckupTypeInput) map[string]any { ... }

// ---- CheckupTypeService ----
type CheckupTypeService interface { ... }  // 以降は既存通り
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/checkup_type_service.go:136-189` | UpdateCheckupTypeInput・const・buildUpdateFields | ❌ 定義位置が後方（メソッド実装後） |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/reservation_type_service.go:15-82` — DTO → 定数 → ビルダー → インターフェース の正しい順序
- `service/cage_service.go` — 同上
- `service/medicine_service.go` — 同上
