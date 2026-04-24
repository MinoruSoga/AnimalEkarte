# TASK-122: `trimming_master_service.go` / `insurance_service.go` — UpdateInput・定数・buildUpdateFields の定義順序が逆

## 優先度

**Low** — 機能に影響はないが、ファイル内の構造が他マスタサービスと統一されておらず可読性が低下している。

---

## 概要

`trimming_master_service.go` と `insurance_service.go` の両ファイルにおいて、
`UpdateXxxInput` DTO・DB カラム定数・`buildXxxUpdateFields` ヘルパーが
すべて **ファイル末尾（メソッド実装の後）** に定義されている。

TASK-111（`exam_type_service.go`）・TASK-116（`checkup_type_service.go`）・TASK-119（`consultation_service.go`）と同一のパターン違反。

---

## 問題箇所

### `service/trimming_master_service.go`（TrimmingCourse 部分）

```
現状の順序:
1. CreateTrimmingCourseInput（16行）← 正しい
2. TrimmingCourseService インターフェース（26行）
3. Service 実装（35行〜）
4. const colTrimmingCourse*（128行）← 定数がメソッド実装より後
5. UpdateTrimmingCourseInput（138行）← UpdateDTO がメソッド実装より後
6. buildTrimmingCourseUpdateFields（149行）← builder がメソッド実装より後
```

### `service/insurance_service.go`

```
現状の順序:
1. CreateInsuranceInput（16行）← 正しい
2. InsuranceService インターフェース（25行）
3. Service 実装（34行〜）
4. UpdateInsuranceInput（132行）← UpdateDTO がメソッド実装より後
5. const colInsurance*（142行）← 定数がメソッド実装より後
6. buildInsuranceUpdateFields（151行）← builder がメソッド実装より後
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/reservation_type_service.go の正しい順序
type CreateReservationTypeInput struct { ... }    // 1. Create DTO
type UpdateReservationTypeInput struct { ... }    // 2. Update DTO
const (
    colReservationTypeName = "name"               // 3. 定数
    ...
)
func buildReservationTypeUpdateFields(...) { ... } // 4. builder
type ReservationTypeService interface { ... }      // 5. interface
type reservationTypeService struct { ... }         // 6. impl
func (s *reservationTypeService) List(...) { ... } // 7. メソッド群
```

---

## 修正方針

各ファイルで以下の順序に並び替える（ロジック変更なし）:

**`trimming_master_service.go`（TrimmingCourse 部分）:**
```
CreateTrimmingCourseInput → UpdateTrimmingCourseInput → const → buildUpdateFields → interface → impl
```

**`insurance_service.go`:**
```
CreateInsuranceInput → UpdateInsuranceInput → const → buildInsuranceUpdateFields → InsuranceService interface → impl
```

**注意:** `trimming_master_service.go` は TrimmingCourse と TrimmingOption の 2 ドメインを含む可能性がある。両方のセクションで同様に並び替えること。

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/trimming_master_service.go:126-147+` | const・UpdateDTO・builder | ❌ メソッド実装より後 |
| `service/insurance_service.go:132-172` | UpdateDTO・const・builder | ❌ メソッド実装より後 |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/reservation_type_service.go:15-82` — 正しい順序
- `service/cage_service.go` — 同上
- 関連タスク: TASK-111, TASK-116, TASK-119
