# TASK-059: service テスト欠落 — clinic_holiday / reservation_type_group

## 優先度

LOW

---

## 概要

マスタ系 service の大部分はテーブル駆動テストが実装されている（`occupation_service_test.go`、`vaccine_service_test.go` 等）が、  
以下の 2 サービスはテストファイル自体が存在しない。

| サービス | テストファイル | 状態 |
|---------|-------------|------|
| `clinic_holiday_service.go` | `clinic_holiday_service_test.go` | ❌ 未作成 |
| `reservation_type_group_service.go` | `reservation_type_group_service_test.go` | ❌ 未作成 |

---

## 参照実装

`backend/internal/service/occupation_service_test.go` をテンプレートとして使用する。  
モック repository を用いたテーブル駆動テスト構成が整っている。

```go
func TestClinicHolidayService_SetHoliday(t *testing.T) {
    tests := []struct {
        name      string
        input     SetClinicHolidayInput
        mockSetup func(*MockClinicHolidayRepository)
        wantErr   bool
    }{
        {
            name: "valid holiday",
            input: SetClinicHolidayInput{/* ... */},
            mockSetup: func(m *MockClinicHolidayRepository) {
                m.On("Set", mock.Anything, mock.Anything).Return(nil)
            },
            wantErr: false,
        },
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // AAA
        })
    }
}
```

---

## 修正方針

### clinic_holiday_service_test.go

対象メソッド:
- `SetClinicHoliday` — 正常系・repository エラー系
- `ListClinicHolidays` — 正常系・空リスト系
- `DeleteClinicHoliday` — 正常系・存在しない ID 系

### reservation_type_group_service_test.go

対象メソッド:
- `CreateReservationTypeGroup` — 正常系・バリデーションエラー系・repository エラー系
- `UpdateReservationTypeGroup` — 正常系・存在しない ID 系
- `DeleteReservationTypeGroup` — 正常系・FK 依存あり系（409 Conflict）
- `ReorderReservationTypeGroups` — 正常系

---

## 作成ファイル

- `backend/internal/service/clinic_holiday_service_test.go`
- `backend/internal/service/reservation_type_group_service_test.go`

---

## 備考

- `-race` フラグでの実行を確認すること（`go test -race ./internal/service/...`）
- mock は `github.com/stretchr/testify/mock` を使用する（他テストと統一）
- 80% 以上のカバレッジを目標とする
