# TASK-217: service — buildXxxUpdateFields で列名定数未定義（文字列リテラル使用・5ファイル）

## 優先度
Medium

## 対象ファイル（5ファイル）
- `backend/internal/service/reservation_staff_service.go`
- `backend/internal/service/reservation_type_liff_service.go`
- `backend/internal/service/staff_service.go`
- `backend/internal/service/clinic_service.go`
- `backend/internal/service/closing_settings_service.go`

## 問題概要
`buildXxxUpdateFields` 内でマップキー（DB 列名）に文字列リテラルを直接使用している。

他の多数のドメイン（animal_species, cage, exam_type, insurance, medicine, payment_method_master, permission_group, procedure, reservation_type, reservation_type_group, vaccine, company 等）は列名を `const` で定義して使用している。

文字列リテラルは：
- タイポしても**コンパイルエラーにならない**（実行時に列名不一致で更新失敗）
- 同じ列名が複数箇所に散在してリファクタリング時に見落としが発生しやすい

## 現状コード（例：reservation_staff_service.go:32-49）

```go
// 現状（NG）
func buildReservationStaffUpdateFields(input *UpdateReservationStaffInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name     // 文字列リテラル直書き
    }
    if input.StaffType != nil {
        fields["staff_type"] = *input.StaffType  // 文字列リテラル直書き
    }
    ...
}
```

## あるべき姿（他ドメインの統一パターン）

```go
// あるべき姿
const (
    colReservationStaffName      = "name"
    colReservationStaffStaffType = "staff_type"
    // ...
)

func buildReservationStaffUpdateFields(input *UpdateReservationStaffInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colReservationStaffName] = *input.Name
    }
    if input.StaffType != nil {
        fields[colReservationStaffStaffType] = *input.StaffType
    }
    ...
}
```

## 対象5ファイルの対応列名

| ファイル | 定数化が必要な列名 |
|---------|-----------------|
| reservation_staff_service.go | name, staff_type, color, is_available_all_day, ... |
| reservation_type_liff_service.go | name, duration_minutes, price, max_reservations, ... |
| staff_service.go | name, license_number, occupation_id, sort_order, ... |
| clinic_service.go | name, address, phone_number, tax_rate, ... |
| closing_settings_service.go | start_date, end_date, am_pm_boundary, ... |

## 完了条件
- [ ] 5ファイルそれぞれに `const (colXxx = "xxx")` ブロックを追加
- [ ] `buildXxxUpdateFields` 内の文字列リテラルを定数参照に置換
- [ ] `go test ./backend/internal/...` がパス
