# TASK-107: `GetPreview` と `Close` の重複集計ロジック（cash_register_service.go）

## 優先度

**Medium** — 同一ロジックが2箇所に存在し、変更時の修正漏れリスクがある。

---

## 概要

`cash_register_service.go` の `GetPreview` と `Close` メソッドが
以下の同一処理シーケンスを重複して実装している:

1. `s.closingsSvc.ResolveSchedule(...)` — スケジュール取得
2. `time.FixedZone(...)` + `time.Date(...)` — JST 日付構築
3. `resolvePeriodRange(...)` — 集計期間（start/end）算出
4. `s.accountingRepo.GetCloseAggregate(...)` — 会計集計クエリ
5. `calcTheoreticalCash(...)` — 理論現金計算

`Close` はこの結果に加えて DB への記録を行うだけで、集計部分のロジックは完全に同一。

---

## 問題箇所

### `service/cash_register_service.go:64-101` (GetPreview)

```go
func (s *cashRegisterService) GetPreview(...) (*CashRegisterPreview, error) {
    // 以下は Close と完全に重複
    schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, date)
    ...
    jst := time.FixedZone("Asia/Tokyo", 9*60*60)
    dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jst)
    periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule, jst)
    ...
    aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, repository.GetCloseAggregateInput{...})
    ...
    theoreticalCash := calcTheoreticalCash(aggregate.AggregateRows)
    ...
}
```

### `service/cash_register_service.go:104-175` (Close)

```go
func (s *cashRegisterService) Close(...) (*model.CashRegisterClose, error) {
    // STEP1: period バリデーション
    // STEP2: 二重締めチェック
    // STEP3: 以下は GetPreview と完全に重複
    schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, input.Date)
    ...
    jst := time.FixedZone("Asia/Tokyo", 9*60*60)
    dateJST := ...
    periodStart, periodEnd, err := resolvePeriodRange(...)
    ...
    aggregate, err := s.accountingRepo.GetCloseAggregate(...)
    ...
    theoreticalCash := calcTheoreticalCash(...)
    // STEP4: DB 登録（Close 固有）
}
```

---

## 修正方針

集計処理をプライベートメソッド `fetchAggregate` として抽出し、
両メソッドから呼び出す。

### 集計結果型を定義

```go
// ✅ 内部型
type periodAggregate struct {
    Schedule        *DaySchedule
    AggregateRows   []repository.BillingAggregateRow
    BillingDetails  []repository.CloseBillingDetail
    TheoreticalCash int64
    PeriodStart     time.Time
    PeriodEnd       time.Time
}
```

### プライベートメソッド抽出

```go
// ✅ 共通集計ロジック
func (s *cashRegisterService) fetchAggregate(ctx context.Context, clinicID uint64, date time.Time, period string) (*periodAggregate, error) {
    schedule, err := s.closingsSvc.ResolveSchedule(ctx, clinicID, date)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to resolve schedule")
    }

    dateJST := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, jstLocation)
    periodStart, periodEnd, err := resolvePeriodRange(dateJST, period, schedule)
    if err != nil {
        return nil, err
    }

    aggregate, err := s.accountingRepo.GetCloseAggregate(ctx, repository.GetCloseAggregateInput{
        ClinicID:    clinicID,
        PeriodStart: periodStart,
        PeriodEnd:   periodEnd,
    })
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to aggregate billings")
    }

    return &periodAggregate{
        Schedule:        schedule,
        AggregateRows:   aggregate.AggregateRows,
        BillingDetails:  aggregate.BillingDetails,
        TheoreticalCash: calcTheoreticalCash(aggregate.AggregateRows),
        PeriodStart:     periodStart,
        PeriodEnd:       periodEnd,
    }, nil
}
```

### GetPreview の簡略化

```go
// ✅ 修正後
func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, date time.Time, period string) (*CashRegisterPreview, error) {
    if err := validatePeriod(period); err != nil {
        return nil, err
    }
    agg, err := s.fetchAggregate(ctx, clinicID, date, period)
    if err != nil {
        return nil, err
    }
    return &CashRegisterPreview{
        Date:            date.Format("2006-01-02"),
        Period:          period,
        Schedule:        agg.Schedule,
        AggregateRows:   agg.AggregateRows,
        BillingDetails:  agg.BillingDetails,
        TheoreticalCash: agg.TheoreticalCash,
    }, nil
}
```

### Close の簡略化

```go
// ✅ 修正後
func (s *cashRegisterService) Close(ctx context.Context, clinicID uint64, input CloseRegisterInput) (*model.CashRegisterClose, error) {
    if err := validatePeriod(input.Period); err != nil {
        return nil, err
    }
    existing, err := s.closeRepo.FindByDateAndPeriod(ctx, clinicID, input.Date, input.Period)
    ...
    agg, err := s.fetchAggregate(ctx, clinicID, input.Date, input.Period)
    if err != nil {
        return nil, err
    }
    cashDifference := input.ActualCash - agg.TheoreticalCash
    ...
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/cash_register_service.go:64-101` | GetPreview | ❌ 重複ロジック |
| `service/cash_register_service.go:104-175` | Close | ❌ 重複ロジック |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — インターフェース設計

DRY 原則。重複するビジネスロジックはプライベートヘルパーに抽出する。

TASK-105（JST ハードコード重複）と同時修正することで、`jstLocation` 変数を
`fetchAggregate` 内で一か所だけ参照できる。

---

## 関連チケット

- TASK-104: バリデーション重複（`validatePeriod` 抽出）
- TASK-105: JST タイムゾーンハードコード重複
