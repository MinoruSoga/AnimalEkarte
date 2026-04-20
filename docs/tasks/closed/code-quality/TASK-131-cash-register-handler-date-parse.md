# TASK-131: cash_register_handler — Handler 内のビジネスロジック（日付パース・エラー握りつぶし）

## 優先度
**High**

## 対象ファイル
`backend/internal/handler/cash_register_handler.go`

---

## 問題 1: `GetCashRegisterPreview` — Handler で日付パース（ビジネスロジック）を実行している

### チェック項目
- **Handler の責務**: Handler は「リクエスト解析 + Service 委譲」のみ。型変換・バリデーションロジックは Service の責務。

### 現状コード（handler.go L35–56）
```go
func (h *Handler) GetCashRegisterPreview(c *gin.Context) {
    ...
    dateStr := c.Query("date")
    if dateStr == "" {
        RespondError(c, apperrors.WrapInvalidInput("date クエリパラメータは必須です"))
        return
    }
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください"))
        return
    }

    period := c.Query("period")
    if period == "" {
        RespondError(c, apperrors.WrapInvalidInput("period クエリパラメータは必須です"))
        return
    }

    preview, err := h.svc.CashRegister.GetPreview(c.Request.Context(), clinicID, date, period)
    ...
```

### 修正後コード
Handler は文字列をそのまま Service へ渡す。Service 側で `validatePeriod` と日付パースを実施する。

```go
// handler.go
func (h *Handler) GetCashRegisterPreview(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    dateStr := c.Query("date")
    period := c.Query("period")

    preview, err := h.svc.CashRegister.GetPreview(c.Request.Context(), clinicID, dateStr, period)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, preview)
}
```

```go
// service/cash_register_service.go
func (s *cashRegisterService) GetPreview(ctx context.Context, clinicID uint64, dateStr, period string) (*CashRegisterPreview, error) {
    if dateStr == "" {
        return nil, apperrors.WrapInvalidInput("date クエリパラメータは必須です")
    }
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください")
    }
    if err := validatePeriod(period); err != nil {
        return nil, err
    }
    ...
}
```

同様に `CashRegisterService` インターフェースのシグネチャも `dateStr, period string` に変更。

---

## 問題 2: `ListCashRegisterCloses` — エラーを `_` で握りつぶし（サイレントバリデーション失敗）

### チェック項目
- **サイレントバリデーション失敗**: バリデーション／変換失敗を無音でスキップしてはならない。

### 現状コード（handler.go L124–130）
```go
var startDate, endDate *time.Time
if startDateStr != nil {
    t, _ := time.Parse("2006-01-02", *startDateStr)  // ❌ エラー握りつぶし
    startDate = &t
}
if endDateStr != nil {
    t, _ := time.Parse("2006-01-02", *endDateStr)    // ❌ エラー握りつぶし
    endDate = &t
}
```

`parseDateQuery` が文字列を返したとき、不正形式（例: `"2024-13-99"`）でも `time.Zero` で処理が続行する。

### 修正後コード
```go
var startDate, endDate *time.Time
if startDateStr != nil {
    t, err := time.Parse("2006-01-02", *startDateStr)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください"))
        return
    }
    startDate = &t
}
if endDateStr != nil {
    t, err := time.Parse("2006-01-02", *endDateStr)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください"))
        return
    }
    endDate = &t
}
```

あるいは日付パース処理を Service 側へ委譲し、Handler は生文字列を渡す設計に変更する。

---

## 備考
問題 1 は「責務分離」の Medium だが、`GetCashRegisterPreview` と `CloseCashRegister` 両方で同パターンが存在する（`CloseCashRegister` L77–81 も同様の `time.Parse` を handler 内で実施）。問題 2 はバリデーション失敗が機能に影響するため High。
