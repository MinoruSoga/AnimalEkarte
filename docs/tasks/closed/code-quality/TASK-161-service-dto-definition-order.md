# TASK-161: Service 層 DTO・定数・builder の定義順序違反

## 概要
`procedure_service.go`、`vaccine_service.go`、`reservation_staff_service.go` の3ファイルで、
`UpdateXxxInput` / `const colXxx*` / `buildXxxUpdateFields` がインターフェースおよび実装の**後に**定義されている。
プロジェクト規約では DTO → 定数 → builder → インターフェース → 実装 の順序が必須。

## 優先度
Low（コード整序）

## 対象ファイル
- `backend/internal/service/procedure_service.go`（行 172〜235）
- `backend/internal/service/vaccine_service.go`（行 117〜170）
- `backend/internal/service/reservation_staff_service.go`（行 189〜207）

---

## 1. procedure_service.go

### 現状コード（定義順序）
```go
// 行 16: CreateProcedureInput ← OK（インターフェース前に配置）
type CreateProcedureInput struct { ... }

// 行 29: ProcedureService interface ← ここまではOK
type ProcedureService interface { ... }

// 行 38: procedureService struct + メソッド群
type procedureService struct { ... }
func (s *procedureService) List(...) { ... }
// ...

// 行 172: UpdateProcedureInput ← ❌ インターフェース・実装の後に配置
type UpdateProcedureInput struct { ... }

// 行 187: const colProcedure* ← ❌ 後置
const (
    colProcedureName = "name"
    ...
)

// 行 200: buildProcedureUpdateFields ← ❌ 後置
func buildProcedureUpdateFields(input *UpdateProcedureInput) map[string]any { ... }
```

### 修正後コード（定義順序）
```go
// DTO を先に並べる
type CreateProcedureInput struct { ... }

type UpdateProcedureInput struct { ... }

const (
    colProcedureName        = "name"
    colProcedurePrice       = "price"
    colProcedureIsActive    = "is_active"
    colProcedureDescription = "description"
    colProcedureDuration    = "duration"
    colProcedureAnesthesia  = "anesthesia"
    colProcedureParentID    = "parent_id"
    colProcedureSortOrder   = "sort_order"
    colProcedureTaxType     = "tax_type"
    colProcedureTaxRate     = "tax_rate"
)

func buildProcedureUpdateFields(input *UpdateProcedureInput) map[string]any { ... }

type ProcedureService interface { ... }

type procedureService struct { ... }

func NewProcedureService(...) ProcedureService { ... }

func (s *procedureService) List(...) { ... }
// ...
```

---

## 2. vaccine_service.go

### 現状コード（定義順序）
```go
// 行 15: CreateVaccineInput ← OK
type CreateVaccineInput struct { ... }

// 行 27: VaccineService interface ← OK
type VaccineService interface { ... }

// 行 36: vaccineService struct + メソッド群
type vaccineService struct { ... }
func (s *vaccineService) List(...) { ... }
// ...

// 行 117: UpdateVaccineInput ← ❌ 後置
type UpdateVaccineInput struct { ... }

// 行 130: const colVaccine* ← ❌ 後置
const ( colVaccineName = "name" ... )

// 行 141: buildVaccineUpdateFields ← ❌ 後置
func buildVaccineUpdateFields(...) map[string]any { ... }
```

### 修正後コード（定義順序）
```go
type CreateVaccineInput struct { ... }

type UpdateVaccineInput struct { ... }

const (
    colVaccineName        = "name"
    colVaccinePrice       = "price"
    colVaccineIsActive    = "is_active"
    colVaccineDescription = "description"
    colVaccineSpecies     = "species"
    colVaccineInterval    = "interval"
    colVaccineParentID    = "parent_id"
    colVaccineSortOrder   = "sort_order"
)

func buildVaccineUpdateFields(input *UpdateVaccineInput) map[string]any { ... }

type VaccineService interface { ... }

type vaccineService struct { ... }

func NewVaccineService(...) VaccineService { ... }

func (s *vaccineService) List(...) { ... }
// ...
```

---

## 3. reservation_staff_service.go

### 現状コード（定義順序）
```go
// 行 13: ReservationStaffService interface ← ❌ DTO の前にインターフェース
type ReservationStaffService interface { ... }

// 行 28: CreateReservationStaffInput ← DTO がインターフェースの後
type CreateReservationStaffInput struct { ... }

// 行 38: UpdateReservationStaffInput ← 同様
type UpdateReservationStaffInput struct { ... }

// 行 47: reservationStaffService struct + メソッド群
type reservationStaffService struct { ... }
// ...

// 行 189: buildReservationStaffUpdateFields ← ❌ 実装の後に builder
func buildReservationStaffUpdateFields(...) map[string]any { ... }
```

### 修正後コード（定義順序）
```go
// DTO を先に
type CreateReservationStaffInput struct { ... }

type UpdateReservationStaffInput struct { ... }

func buildReservationStaffUpdateFields(input *UpdateReservationStaffInput) map[string]any { ... }

// その後にインターフェース・実装
type ReservationStaffService interface { ... }

type reservationStaffService struct { ... }

func NewReservationStaffService(...) ReservationStaffService { ... }

func (s *reservationStaffService) List(...) { ... }
// ...
```

## 修正方針
各ファイル内でのコードブロック移動のみ（ロジック変更なし）。
