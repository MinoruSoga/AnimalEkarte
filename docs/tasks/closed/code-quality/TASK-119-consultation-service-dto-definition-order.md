# TASK-119: `consultation_service.go` — 定数が DTO より前・UpdateConsultationInput と buildUpdateFields がメソッド実装より後に定義

## 優先度

**Low** — 機能に影響はないが、ファイル内の構造が他マスタサービスと統一されておらず可読性が低下している。

---

## 概要

`consultation_service.go` において定義順序が統一規約から外れている:

```
現状の順序:
1. const（13行）← 定数が DTO より前
2. CreateConsultationInput（29行）
3. ConsultationService インターフェース（42行）← UpdateDTO が未定義のままインターフェースに使用
4. Service 実装（51行〜）
5. UpdateConsultationInput（153行）← DTO がメソッド実装より後に定義
6. buildConsultationUpdateFields（168行）

正しい順序:
1. Input DTOs（CreateXxxInput → UpdateXxxInput）
2. DB カラム定数
3. buildXxxUpdateFields ヘルパー
4. Service インターフェース
5. Service 実装（構造体 + メソッド群）
```

2 つの問題が混在している:
- `const` ブロックが `CreateConsultationInput` より前（行 13）に定義されている
- `UpdateConsultationInput` と `buildConsultationUpdateFields` がすべてのメソッド実装より後（行 153・168）に定義されている

---

## 問題箇所

### `service/consultation_service.go:13-24`（定数が DTO より前）

```go
// ❌ DTO 定義（行 29）より前に定数が配置されている
const (  // 行 13
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

// CreateConsultationInput は診察種別作成のサービス入力 DTO
type CreateConsultationInput struct { ... }  // 行 29
```

### `service/consultation_service.go:153-203`（UpdateDTO・builder がメソッド実装より後）

```go
// ❌ Reorder メソッド実装（行 140-151）の後に UpdateDTO と builder が定義
func (s *consultationService) Reorder(...) { ... }  // 行 140-151

type UpdateConsultationInput struct { ... }        // 行 153
func buildConsultationUpdateFields(...) { ... }    // 行 168
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/reservation_type_service.go の正しい順序
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
```

---

## 修正方針

以下の順序に並び替える（ロジック変更なし）:

```
1. CreateConsultationInput（行 29 を維持）
2. UpdateConsultationInput（行 153 から移動）
3. const ブロック（行 13 から移動）
4. buildConsultationUpdateFields（行 168 から移動）
5. ConsultationService インターフェース（行 42 を維持）
6. Service 実装（行 51〜 を維持）
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/consultation_service.go:13-24` | const ブロック | ❌ DTO より前に配置 |
| `service/consultation_service.go:153-203` | UpdateConsultationInput・buildUpdateFields | ❌ メソッド実装より後に定義 |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/reservation_type_service.go:15-82` — DTO → 定数 → ビルダー → インターフェース の正しい順序
- `service/cage_service.go` — 同上
- `service/medicine_service.go` — 同上
