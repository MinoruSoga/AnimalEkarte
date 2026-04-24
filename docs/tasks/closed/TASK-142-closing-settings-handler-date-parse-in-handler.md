# TASK-142: [Medium] closing_settings_handler.go — UpdateSpecialPeriod の日付パース・Input構築がHandlerに存在

## 優先度
**Medium**（責務分離違反）

## 対象ファイル
- `backend/internal/handler/closing_settings_handler.go`

## 問題

`UpdateSpecialPeriod` ハンドラが `time.Parse("2006-01-02", ...)` による日付パースと
`service.UpdateSpecialPeriodInput` の条件付き組み立てロジックを直接実装している。

Handler はリクエストの「受け口」に徹すべきであり、日付文字列からの `time.Time` 変換と
Input 型の部分的な条件組み立ては Service の入力として `string` のまま渡すか、
専用の変換ロジックを Service/DTO 側に持たせるべき。

また、`input := service.UpdateSpecialPeriodInput{...}` → value 型に代入してから `&input` を
Service に渡しているが、プロジェクト規約では**ポインタリテラル**（`&service.UpdateSpecialPeriodInput{...}`）での
渡し方を推奨している（CreateSpecialPeriod は正しくポインタリテラルで渡している）。

## 現状コード（handler/closing_settings_handler.go L150〜173）

```go
input := service.UpdateSpecialPeriodInput{
    AmPmBoundary: req.AmPmBoundary,
    PmEnd:        req.PmEnd,
    Note:         req.Note,
}
if req.StartDate != nil {
    t, err := time.Parse("2006-01-02", *req.StartDate)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください"))
        return
    }
    input.StartDate = &t
}
if req.EndDate != nil {
    t, err := time.Parse("2006-01-02", *req.EndDate)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください"))
        return
    }
    input.EndDate = &t
}

period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(c.Request.Context(), clinicID, id, input)
```

## 修正後コード

### 方針A：Input フィールドを `*string` に変更してパースを Service に委譲（推奨）

`UpdateSpecialPeriodInput` の `StartDate`/`EndDate` を `*time.Time` → `*string` に変更し、
Service 側で日付文字列のパースとバリデーションを行う。

#### service/closing_settings_service.go の Input 変更

```go
// UpdateSpecialPeriodInput は特別期間更新の入力
type UpdateSpecialPeriodInput struct {
    StartDate    *string // YYYY-MM-DD 形式（Service 内でパース）
    EndDate      *string // YYYY-MM-DD 形式（Service 内でパース）
    AmPmBoundary *string
    PmEnd        *string
    Note         *string
}
```

#### handler/closing_settings_handler.go の修正

```go
period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(
    c.Request.Context(), clinicID, id,
    &service.UpdateSpecialPeriodInput{  // ← ポインタリテラルで渡す
        StartDate:    req.StartDate,
        EndDate:      req.EndDate,
        AmPmBoundary: req.AmPmBoundary,
        PmEnd:        req.PmEnd,
        Note:         req.Note,
    },
)
```

#### service 側でのパース（closing_settings_service.go）

```go
func (s *closingSettingsService) UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input *UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
    current, err := s.periodRepo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get special period")
    }

    var startDate, endDate time.Time
    if input.StartDate != nil {
        t, err := time.Parse("2006-01-02", *input.StartDate)
        if err != nil {
            return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
        }
        startDate = t
    } else {
        startDate = current.StartDate
    }
    // ... EndDate 同様 ...
```

### 方針B：ポインタリテラルのみ修正（最小変更）

日付パースロジックの移動が難しい場合、少なくとも input 渡しをポインタリテラルに変更する。

```go
// ❌ 現状
input := service.UpdateSpecialPeriodInput{...}
// ...
period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(..., input)  // value渡し

// ✅ 修正後（ポインタリテラル）
period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(..., &service.UpdateSpecialPeriodInput{...})
```

## 補足

`CreateSpecialPeriod` では正しくポインタリテラル渡しになっている（L119〜125）。
`UpdateSpecialPeriod` も同じスタイルに揃えること。

日付パース処理の Service への移動は方針A が推奨（テスト容易性の向上）。
