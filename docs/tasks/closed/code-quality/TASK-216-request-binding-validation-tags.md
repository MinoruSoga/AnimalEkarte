# TASK-216: handler request.go — binding バリデーションタグの欠落・不統一（複数ファイル）

## 優先度
Medium

## 対象ファイル（5ファイル）
- `backend/internal/handler/insurance_request.go`
- `backend/internal/handler/reservation_type_liff_request.go`
- `backend/internal/handler/reservation_staff_request.go`
- `backend/internal/handler/staff_request.go`
- `backend/internal/handler/procedure_request.go`

## 問題概要
複数の request struct でビジネスルール上必要な `binding` バリデーションタグが欠落しており、
不正値が handler の ShouldBindJSON をすり抜けて service/DB まで到達する。

## 各ファイルの修正箇所

### 1. `insurance_request.go` — CoverageRate の範囲制約なし
```go
// 現状（NG）
CoverageRate *int `json:"coverage_rate"`

// あるべき姿（0〜100の整数）
CoverageRate *int `json:"coverage_rate" binding:"omitempty,min=0,max=100"`
```

### 2. `reservation_type_liff_request.go` — ReservationDayOption に oneof なし
管理側 `createReservationTypeRequest` は `binding:"omitempty,oneof=none saturday weekday anyday"` を持つが LIFF 側にない。

```go
// 現状（NG）
ReservationDayOption string `json:"reservation_day_option"`

// あるべき姿
ReservationDayOption string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
```

Update 用ポインタ型も同様に修正。

### 3. `reservation_staff_request.go` — StaffType に oneof なし
DB の ENUM 値に対応したバリデーションが欠落している。

```go
// 現状（NG）
StaffType string `json:"staff_type"`

// あるべき姿（実際の ENUM 値に合わせること）
StaffType string `json:"staff_type" binding:"omitempty,oneof=doctor nurse groomer other"`
```

### 4. `staff_request.go` — Email フォーマット・Password 最小長なし
```go
// 現状（NG）
Email    string `json:"email"`
Password string `json:"password"`

// あるべき姿
Email    string `json:"email"    binding:"omitempty,email"`
Password string `json:"password" binding:"omitempty,min=8"`
```

`updateStaffRequest.Password *string` も同様。

### 5. `procedure_request.go` — TaxRate に上限なし
`medicine_request.go` には `binding:"omitempty,min=0,max=1"` があるが `procedure_request.go` には範囲制約がない。

```go
// 現状（NG）
TaxRate *float64 `json:"tax_rate"`

// あるべき姿
TaxRate *float64 `json:"tax_rate" binding:"omitempty,min=0,max=1"`
```

`updateProcedureRequest.TaxRate` も同様。

## 完了条件
- [ ] 上記5ファイルの binding タグをすべて修正
- [ ] `go test ./backend/internal/...` がパス
