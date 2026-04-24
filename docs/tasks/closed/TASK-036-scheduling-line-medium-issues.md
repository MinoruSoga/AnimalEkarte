# TASK-036: スケジューリング・LINE系 MEDIUM 問題 3件

## 優先度

MEDIUM

---

## 問題 1: reservation_schedule_service で validateShiftTimes 呼び出しが欠落

### ファイル
`backend/internal/service/reservation_schedule_service.go:74-106`

### 問題
予約スケジュール生成時にシフト時間のバリデーション（`validateShiftTimes` 相当）が呼ばれていない。`shift_service.go` 側では時間バリデーションが実装されているにもかかわらず、スケジュール生成フローでは省略されており、StartTime > EndTime の矛盾した時間範囲でもスケジュールが生成されてしまう。

### 修正案
スケジュール生成前に start_time < end_time の整合性チェックを追加する。

```go
if schedule.StartTime != nil && schedule.EndTime != nil {
    if *schedule.StartTime >= *schedule.EndTime {
        return nil, apperrors.WrapInvalidInput("start_time は end_time より前である必要があります")
    }
}
```

---

## 問題 2: shift_template_service の no-op Update に slog.InfoContext なし

### ファイル
`backend/internal/service/shift_template_service.go`（Update メソッド）

### 問題
シフトテンプレートの更新操作（`Update`）に `slog.InfoContext` がない。他の mutation（Create / Delete）にはログがある。Reorder・Update の slog 欠落は TASK-023/TASK-027/TASK-029 で繰り返し指摘されているパターンと同根。

### 修正案
```go
slog.InfoContext(ctx, "shift_template updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("shift_template_id", id))
```

---

## 問題 3: line_reservation_setting_service の []byte フィールドが JSON バリデーションなし

### ファイル
`backend/internal/service/line_reservation_setting_service.go`

### 問題
`UpsertLineReservationSettingInput` の以下のフィールドは JSON バイト列として DB に保存されるが、service 層でバリデーションがない。

```go
ClosedWeekdays         []byte
ClosedDates            []byte
BusinessHours          []byte
BusinessHoursByWeekday []byte
BreakHours             []byte
AdditionalFields       []byte
```

不正な JSON（`nil`、`[]`、`{"key":}` 等）が渡されると DB には保存されるが、クライアントで読み取り時にパースエラーになる可能性がある。

### 修正案
各 []byte フィールドに対して `json.Valid()` チェックを追加する。

```go
// service 内でバリデーション
for _, field := range []struct{ name string; val []byte }{
    {"closed_weekdays", input.ClosedWeekdays},
    {"closed_dates", input.ClosedDates},
    {"business_hours", input.BusinessHours},
    // ...
} {
    if len(field.val) > 0 && !json.Valid(field.val) {
        return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は有効な JSON である必要があります", field.name))
    }
}
```
