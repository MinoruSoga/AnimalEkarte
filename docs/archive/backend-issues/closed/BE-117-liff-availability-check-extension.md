# BE-117: LIFF 予約可否チェックに不可時間 & 職種ガードを組み込む

**Status**: Open
**Priority**: High
**Affects**: LINE予約 日付・時間選択
**Date Created**: 2026-04-16
**Related**: TASK-001, BE-115（先に完了必要）, BE-116（CountWorkingStaff が必要）

## Summary

`liff_service.go` の `GetAvailableDates()` と `GetAvailableTimes()` に、
BE-115/116 で追加した「予約不可時間」と「職種出勤ガード」のロジックを組み込む。
LIFF フロントエンドへの API 変更はなし（既存レスポンス形式を維持）。

## 現状のコード

```go
// backend/internal/service/liff_service.go（概要）
// GetAvailableDates: CalcAvailableDates（available_dates.go）に委譲して予約可能日を計算する
//   戻り値型: ([]AvailableDateResult, BookingWindow, error)
//   AvailableDateResult は { Date string, Weekday int, Available bool, Reason string }
//
// GetAvailableTimes: GenerateTimeSlots（timeslot_engine.go）に委譲して時間枠を計算する
//   戻り値型: ([]TimeSlot, error)
//   TimeSlotsInput.DefaultBreaks に BreakPeriod を追加することで時間帯を除外できる
//
// backend/internal/service/timeslot_engine.go
//   - BreakPeriod.Start / End は "HHMM" 形式（コロンなし4文字）
//   - subtract(base, excl []interval) []interval  ← interval は内部型（外部から直接使わない）
//   - GenerateTimeSlots が DefaultBreaks を自動的に除外する
// ※ 予約区分毎の不可時間は考慮していない

// backend/internal/service/available_dates.go
//   CalcAvailableDates は AvailableDatesInput.StaffInputsFn / SlotSettingsFn をコールバックとして受け取り
//   日付ループを内部で実行する。GetAvailableDates はその結果を受け取って返すだけ。
```

```go
// liffService 構造体（現状 — liff_service.go:29-40）
type liffService struct {
    settingRepo     repository.LineReservationSettingRepository
    typeLiffRepo    repository.ReservationTypeLiffRepository
    staffRepo       repository.ReservationStaffRepository
    scheduleRepo    repository.ReservationScheduleRepository
    adminRepo       repository.ReservationAdminRepository
    reservationRepo repository.ReservationRepository
    customerRepo    repository.LineCustomerRepository
    ownerRepo       repository.OwnerRepository
    validators      ReservationValidators
    notifier        ReservationNotifier
    // ※ unavailableTimeRepo / occupationRepo は未存在
}
```

## 必要な変更

### 1. liffService に新 Repository を注入

```go
// liff_service.go の構造体に追加（既存フィールドの順序を変えない）
type liffService struct {
    settingRepo         repository.LineReservationSettingRepository
    typeLiffRepo        repository.ReservationTypeLiffRepository
    staffRepo           repository.ReservationStaffRepository
    scheduleRepo        repository.ReservationScheduleRepository
    adminRepo           repository.ReservationAdminRepository
    reservationRepo     repository.ReservationRepository
    customerRepo        repository.LineCustomerRepository
    ownerRepo           repository.OwnerRepository
    validators          ReservationValidators  // interface 型（* 不要）
    notifier            ReservationNotifier
    unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository // ★ NEW
    occupationRepo      repository.ReservationTypeOccupationRepository      // ★ NEW
}

// NewLiffService のシグネチャ変更（末尾に2引数追加）
// 現在の引数順: settingRepo, typeLiffRepo, staffRepo, scheduleRepo, adminRepo,
//               customerRepo, ownerRepo, tx, reservationRepo, notifier
func NewLiffService(
    settingRepo         repository.LineReservationSettingRepository,
    typeLiffRepo        repository.ReservationTypeLiffRepository,
    staffRepo           repository.ReservationStaffRepository,
    scheduleRepo        repository.ReservationScheduleRepository,
    adminRepo           repository.ReservationAdminRepository,
    customerRepo        repository.LineCustomerRepository,
    ownerRepo           repository.OwnerRepository,
    tx                  repository.Transactor,
    reservationRepo     repository.ReservationRepository,
    notifier            ReservationNotifier,
    unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository, // ★ NEW
    occupationRepo      repository.ReservationTypeOccupationRepository,      // ★ NEW
) LiffService { ... }
```

`backend/internal/service/service.go` の DI 配線も更新（`NewLiffService` 呼び出し箇所）:

```go
// service.go の NewServices 関数内
Liff: NewLiffService(
    repos.LineReservationSetting,
    repos.ReservationTypeLiff,
    repos.ReservationStaff,
    repos.ReservationSchedule,
    repos.ReservationAdmin,
    repos.LineCustomerMgr,               // 実際のフィールド名（repositories.go:76）
    repos.Owner,
    tx,
    repos.Reservation,
    notifier,
    repos.ReservationTypeUnavailableTime, // ★ NEW（BE-115 で追加したリポジトリ）
    repos.ReservationTypeOccupation,      // ★ NEW
),
```

### 2. GetAvailableDates に職種ガードを追加

`GetAvailableDates` は内部で `CalcAvailableDates`（`available_dates.go:87`）に委譲する。
`CalcAvailableDates` が日付ループを持つため、職種ガードは **CalcAvailableDates の戻り値を後処理** する形で追加する。

```go
// 実際のシグネチャ（変更なし）:
// func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)

func (s *liffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
    // ... 既存ロジック（CalcAvailableDates 呼び出しまで変更なし） ...
    results, window, err := CalcAvailableDates(ctx, &AvailableDatesInput{
        Settings:       datesSettings,
        TypeID:         typeID,
        StaffID:        staffID,
        StaffInputsFn:  staffInputsFn,
        SlotSettingsFn: slotSettingsFn,
    })
    if err != nil {
        return nil, window, err
    }

    // ★ 新規: 職種ガード（CalcAvailableDates の結果を後処理）
    // 職種紐付けが1件以上ある場合のみチェック（0件は従来通り → 後方互換）
    occupations, err := s.occupationRepo.FindAll(ctx, clinicID, typeID)
    if err != nil {
        return nil, window, apperrors.Wrap(err, "failed to get occupation guard")
    }
    if len(occupations) > 0 {
        for i, r := range results {
            if !r.Available {
                continue // すでに不可の日はスキップ
            }
            // AvailableDateResult.Date は "2006-01-02" 形式（JST）
            date, err := time.ParseInLocation("2006-01-02", r.Date, jstLocation())
            if err != nil {
                return nil, window, apperrors.Wrap(err, "failed to parse date")
            }
            count, err := s.occupationRepo.CountWorkingStaff(ctx, clinicID, typeID, date)
            if err != nil {
                return nil, window, apperrors.Wrap(err, "failed to count working staff")
            }
            if count == 0 {
                results[i].Available = false
                results[i].Reason = "staff_off" // 既存の Reason 値と統一
            }
        }
    }

    return results, window, nil
}
```

### 3. GetAvailableTimes に不可時間を追加

`GenerateTimeSlots` は `TimeSlotsInput.DefaultBreaks` に `BreakPeriod` を追加するだけで時間帯を除外できる。
`BreakPeriod.Start/End` は `"HHMM"` 形式（コロンなし4文字）。モデルの `StartTime/EndTime` は `"HH:MM"` 形式のためコロン除去が必要。

```go
// 実際のシグネチャ（変更なし）:
// func (s *liffService) GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)

func (s *liffService) GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
    // ... 既存ロジック（input := &TimeSlotsInput{...} 構築まで変更なし） ...

    // ★ 新規: 予約不可時間を DefaultBreaks に追加して GenerateTimeSlots に渡す
    unavailableTimes, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, typeID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get unavailable times")
    }
    applicable := filterApplicable(unavailableTimes, date)
    for _, u := range applicable {
        // モデルの "HH:MM" → timeslot_engine の "HHMM" に変換（コロン除去）
        input.DefaultBreaks = append(input.DefaultBreaks, BreakPeriod{
            Start: strings.ReplaceAll(u.StartTime, ":", ""),
            End:   strings.ReplaceAll(u.EndTime, ":", ""),
        })
    }

    result, err := GenerateTimeSlots(input)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to generate time slots")
    }
    return result, nil
}
```

`strings` パッケージを import に追加すること（既存の import ブロックに `"strings"` が含まれているか確認）。

### 4. filterApplicable ヘルパー実装

```go
// filterApplicable は date に適用される不可時間帯を返す。
// 優先順位: specific > weekly（特定日設定が曜日設定を上書き）
//
// 例1: 毎週月曜 12:00-13:00 が設定されている日（月曜）
//   → specific がなければ weekly を適用
// 例2: 毎週月曜 12:00-13:00 + 2026-05-20（月曜）に specific 10:00-11:00
//   → specific のみ適用、weekly は無視
//   → 2026-05-20 は 10:00-11:00 のみ不可（12:00-13:00 は受け付ける）
//
// DayOfWeek の値は Go の time.Weekday() と同一: 0=Sun, 1=Mon, ..., 6=Sat
//
// 注意: GORM が DATE 型を time.Time に読み込む際は UTC 午前0時になる。
// date 引数も JST 日付文字列で比較するため、両辺を "2006-01-02" 形式に変換して比較する。
func filterApplicable(times []model.ReservationTypeUnavailableTime, date time.Time) []model.ReservationTypeUnavailableTime {
    dateStr := date.In(jstLocation()).Format("2006-01-02") // JST 日付文字列で統一
    var specific, weekly []model.ReservationTypeUnavailableTime
    for _, t := range times {
        switch t.UnavailableType {
        case model.UnavailableTypeSpecific:
            // SpecificDate（DATE型→UTC 0時 time.Time）を "2006-01-02" 形式で比較
            if t.SpecificDate != nil && t.SpecificDate.UTC().Format("2006-01-02") == dateStr {
                specific = append(specific, t)
            }
        case model.UnavailableTypeWeekly:
            if t.DayOfWeek != nil && int(*t.DayOfWeek) == int(date.In(jstLocation()).Weekday()) {
                weekly = append(weekly, t)
            }
        }
    }
    // specific がある日は weekly を無視
    if len(specific) > 0 {
        return specific
    }
    return weekly
}
```

### 5. Repository への FindAll 追加

BE-116 の `ReservationTypeUnavailableTimeRepository.FindAll()` を活用。
当日適用分のフィルタはアプリ層（`filterApplicable`）で行う（DB クエリは `weekly` + `specific(date)` の Union でも可だが、件数が少ないのでアプリ側フィルタで十分）。

## パフォーマンス考慮

- `reservation_type_unavailable_times` の件数は予約区分あたり高々数十件 → 全件取得してアプリフィルタで問題なし
- `CountWorkingStaff` は日付毎に1回実行 → 既存インデックス `idx_shift_entries_staff_date(staff_id, date)`（001_init.sql:1319）で十分高速

## 完了条件

- [ ] `GetAvailableDates` で職種ガードが機能する（職種スタッフ0人の日が除外される）
- [ ] 職種紐付けが0件の予約区分は従来通り動作する（後方互換）
- [ ] `GetAvailableTimes` で予約不可時間帯が除外される
- [ ] `weekly` / `specific` の両パターンが正しく除外される
- [ ] `specific` が存在する日は `weekly` が無視される
- [ ] `go build ./...` が通る
- [ ] `go test ./internal/service/...` が通る（新ロジックのユニットテスト追加）
