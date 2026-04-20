# TASK-130: 複数 service ファイル — Input DTO / 定数 / buildUpdateFields の定義順序違反

## 優先度

**Low** — 機能に影響はないが、プロジェクト全体の統一規約から外れており可読性が低下している。

---

## 概要

以下の service ファイルで「Input DTO → 定数 → buildUpdateFields → interface → impl」の
統一順序規約から外れている箇所が複数ある。
TASK-111/116/119/122/123/124 と同一のパターン違反が残りのファイルにも存在する。

---

## 問題箇所

### 1. `service/inventory_service.go` — UpdateDTO と builder がインターフェース・実装の後

```
現状の順序:
1. CreateInventoryInput（13行）← 正しい
2. InventoryService インターフェース（27行）← UpdateDTO がまだ未定義
3. Service 実装（35行〜）
4. UpdateInventoryInput（115行）← UpdateDTO がメソッド実装より後
5. buildInventoryUpdateFields（129行）← builder がメソッド実装より後
```

### 2. `service/reservation_schedule_service.go` — インターフェースが Input DTO より前

```
現状の順序:
1. ReservationScheduleService インターフェース（14行）← DTO がまだ未定義
2. BreakInput（20行）← 後に DTO が登場
3. UpsertScheduleInput（27行）
4. Service 実装（以降）
```

### 3. `service/hospitalization_plan_service.go` — 定数が UpdateDTO より前

```
現状の順序（後半部分）:
1. const colHospitalizationPlan*（144行）← 定数が先
2. UpdateHospitalizationPlanInput（158行）← UpdateDTO が後
3. buildHospitalizationPlanUpdateFields（166行）
```

正しい順序: UpdateDTO → const → buildUpdateFields

### 4. `service/line_reservation_setting_service.go` — インターフェースが Input DTO より前

```
現状の順序:
1. LineReservationSettingService インターフェース（14行）
2. UpsertLineReservationSettingInput（19行）← DTO がインターフェース後
```

### 5. `service/chief_complaint_service.go` — 定数が buildUpdateFields より後

```
現状の順序:
1. Input DTOs（15-39行）← 正しい
2. buildXxxUpdateFields ヘルパー（定数の前に来ている）
3. const colChiefComplaintType*（137-144行）← 定数が builder より後
```

正しい順序: DTOs → const → buildUpdateFields

---

## 修正方針

各ファイルで以下の統一順序に並び替える（ロジック変更なし）:

```
CreateXxxInput
UpdateXxxInput
const colXxx* = "..."
func buildXxxUpdateFields(...) map[string]any { ... }
type XxxService interface { ... }
type xxxService struct { ... }
func (s *xxxService) List(...) { ... }
...
```

---

## 影響範囲

| ファイル | 問題 |
|---------|------|
| `service/inventory_service.go` | UpdateDTO・builder がインターフェース・実装より後（行 115, 129） |
| `service/reservation_schedule_service.go` | インターフェースが BreakInput・UpsertScheduleInput より前（行 14） |
| `service/hospitalization_plan_service.go` | const が UpdateDTO より前（行 144 vs 158） |
| `service/line_reservation_setting_service.go` | インターフェースが UpsertDTO より前（行 14 vs 19） |
| `service/chief_complaint_service.go` | const がbuildUpdateFields より後（行 137） |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/reservation_type_service.go:15-82` — 完全に正しい順序の参照実装
- 関連タスク: TASK-111, TASK-116, TASK-119, TASK-122, TASK-123, TASK-124
