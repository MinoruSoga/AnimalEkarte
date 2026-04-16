# BUG-339: Backend Go 規約監査 #8

- **種別**: コード品質
- **優先度**: HIGH
- **対象**: `backend/internal/service/`, `backend/internal/repository/`

---

## 検出された問題

### HIGH-1: `timeslot_engine.go` — `fmt.Errorf` が `apperrors.ErrInvalidInput` をラップしない

**ファイル**: `backend/internal/service/timeslot_engine.go:12-18`

`minutesSinceMidnight` が返す `fmt.Errorf(...)` は `apperrors.ErrInvalidInput` を含まない。
呼び出し元 `resolveWorkIntervals` → `GenerateTimeSlots` → `liff_service.go` → `RespondError` のパスで
`errors.Is(err, ErrInvalidInput)` が false となり、400 ではなく 500 を返す。

```go
// 現状
return 0, fmt.Errorf("invalid time format %q: must be HHMM", hhmm)

// 修正
return 0, apperrors.WrapInvalidInput(fmt.Sprintf("invalid time format %q: must be HHMM", hhmm))
```

### HIGH-2: `timeslot_engine.go:216-217` — エラーを `_` で無視

`minutesSinceMidnight` のエラーを黙って無視。失敗時に `defaultBreakStart/End` がゼロ値（0分 = 00:00）になり、
誤った勤務時間スロット計算が行われる。

```go
// 現状
defaultBreakStart, _ = minutesSinceMidnight(input.DefaultBreaks[0].Start)
defaultBreakEnd, _ = minutesSinceMidnight(input.DefaultBreaks[0].End)

// 修正
var errBreak error
defaultBreakStart, errBreak = minutesSinceMidnight(input.DefaultBreaks[0].Start)
if errBreak != nil {
    return nil, apperrors.Wrap(errBreak, "default_breaks.start")
}
defaultBreakEnd, errBreak = minutesSinceMidnight(input.DefaultBreaks[0].End)
if errBreak != nil {
    return nil, apperrors.Wrap(errBreak, "default_breaks.end")
}
```

### HIGH-3: `reservation_staff_service.go:63-76` — `GetByID` が全件取得後にメモリフィルタ

`Update`, `Delete`, `PatchStatus` の全操作が `GetByID` を経由し、毎回 `FindAllByClinicID` で全スタッフを
DB から取得してメモリフィルタ。スタッフ数が多い場合に N 倍の無駄クエリが発生する。

Repository に `FindByIDAndClinicID` を追加して 1 クエリで解決すべき。

### MEDIUM-1: `reservation_schedule_repository.go:113` — `updated_at: entry.UpdatedAt` がゼロ値

Upsert の更新フィールドマップに `"updated_at": entry.UpdatedAt` を手動設定しているが、
`entry.UpdatedAt` は呼び出し時点でゼロ値のまま。`gorm.Expr("NOW()")` に変更するか、
フィールドを削除して GORM の自動タイムスタンプ更新に委ねる。

---

## 修正方針

- HIGH-1/2: `timeslot_engine.go` の `minutesSinceMidnight` + `resolveWorkIntervals` を修正
- HIGH-3: `ReservationStaffRepository` に `FindByIDAndClinicID` 追加 + service の `GetByID` を更新
- MEDIUM-1: `updated_at: entry.UpdatedAt` を `"updated_at": gorm.Expr("NOW()")` に変更
