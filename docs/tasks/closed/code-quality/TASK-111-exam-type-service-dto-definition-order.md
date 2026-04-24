# TASK-111: `exam_type_service.go` — UpdateExamTypeInput・定数・buildUpdateFields の定義順序が逆

## 優先度

**Low** — 機能に影響はないが、ファイル内の構造が他マスタサービスと統一されておらず可読性が低下している。

---

## 概要

`exam_type_service.go` において `UpdateExamTypeInput` DTO、DB カラム定数、
`buildExamTypeUpdateFields` ヘルパーがすべて **ファイル末尾（128〜171行）** に定義されている。

他のマスタサービス（`reservation_type_service.go`, `cage_service.go`, `medicine_service.go` 等）は
以下の統一順序に従っている:

```
1. Input DTOs（CreateXxxInput, UpdateXxxInput）
2. DB カラム定数（const colXxx = "..."）
3. buildXxxUpdateFields ヘルパー
4. Service インターフェース
5. Service 実装（構造体 + メソッド群）
```

しかし `exam_type_service.go` は:

```
1. CreateExamTypeInput（16行）← 正しい
2. Service インターフェース（25行）
3. Service 実装（34行〜）     ← Update/Reorder 実装が先に来る
4. UpdateExamTypeInput（128行）← DTOs がメソッド実装より後に定義
5. DB カラム定数（139行）
6. buildExamTypeUpdateFields（148行）
```

---

## 問題箇所

### `service/exam_type_service.go:128-171`

```go
// ❌ メソッド実装（Reorder:115〜126）の後に DTO と定数が定義されている
func (s *examTypeService) Reorder(...) error {
    ...
    return nil  // 行 126
}

// ← ここで DTO と定数が突然登場
type UpdateExamTypeInput struct {  // 行 128
    Name          *string
    ...
}

const (  // 行 139
    colExamTypeName = "name"
    ...
)

func buildExamTypeUpdateFields(...) map[string]any {  // 行 148
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

`UpdateExamTypeInput`（行 128-137）・定数（行 139-146）・`buildExamTypeUpdateFields`（行 148-171）
を `CreateExamTypeInput`（行 16-23）の直後（行 24 以降）に移動する。

```go
// ✅ 修正後の順序

// ---- Input DTOs ----
type CreateExamTypeInput struct { ... }  // 既存位置（行 16）

// ← ここに UpdateExamTypeInput を移動
type UpdateExamTypeInput struct {
    Name          *string
    Price         *int64
    IsActive      *bool
    Description   *string
    ParentID      *uint64
    ClearParentID bool
    SortOrder     *int
}

// ← ここに定数を移動
const (
    colExamTypeName        = "name"
    colExamTypePrice       = "price"
    colExamTypeIsActive    = "is_active"
    colExamTypeDescription = "description"
    colExamTypeParentID    = "parent_id"
    colExamTypeSortOrder   = "sort_order"
)

// ← ここに buildExamTypeUpdateFields を移動
func buildExamTypeUpdateFields(input *UpdateExamTypeInput) map[string]any { ... }

// ---- ExamTypeService ----
type ExamTypeService interface { ... }  // 以降は既存通り
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/exam_type_service.go:128-171` | UpdateExamTypeInput・const・buildUpdateFields | ❌ 定義位置が後方（メソッド実装後） |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — 命名規則

> パッケージ lowercase、エクスポート PascalCase...

Go の慣用的なコードでは、型定義は利用箇所より前に配置する。DTO をファイル先頭に集約することで
「このサービスで何を入力として受け取るか」が一目でわかる。

### プロジェクト内参照実装

- `service/reservation_type_service.go:15-82` — DTO → 定数 → ビルダー → インターフェース の正しい順序
- `service/cage_service.go` — 同上
- `service/medicine_service.go` — 同上
