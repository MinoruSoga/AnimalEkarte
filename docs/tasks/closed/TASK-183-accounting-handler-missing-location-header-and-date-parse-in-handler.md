# TASK-183: CreateAccounting POST 201 Location ヘッダー欠落 / GetDailySummary 日付パース in Handler

## 優先度: Medium

## 概要
2つの問題を含む:
1. `CreateAccounting` で 201 Created を返しているが `Location` ヘッダーがない
2. `GetDailySummary` Handler でタイムゾーン指定の日付パース処理が Handler 内にある（Service 委譲すべき）

## 対象ファイル
`backend/internal/handler/accounting_handler.go`

---

## 問題1: POST 201 に Location ヘッダー欠落

### 現状コード（L127〜128）

```go
c.JSON(http.StatusCreated, toAccountingResponse(created))
// ❌ Location ヘッダーがない
```

### 修正後コード

```go
c.Header("Location", fmt.Sprintf("/v1/accountings/%d", created.ID))
c.JSON(http.StatusCreated, toAccountingResponse(created))
```

---

## 問題2: GetDailySummary Handler での日付パース処理

### 現状コード（L258〜272）

```go
func (h *Handler) GetDailySummary(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }

    dateStr := c.Query("date")
    if dateStr == "" {
        dateStr = time.Now().Format("2006-01-02")
    }
    // ❌ Handler 内でタイムゾーン指定の日付パース処理
    date, err := time.ParseInLocation("2006-01-02", dateStr, time.FixedZone("Asia/Tokyo", 9*60*60))
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD"))
        return
    }

    result, err := h.svc.Accounting.GetDailySummary(c.Request.Context(), clinicID, date)
```

### 問題の詳細
- Handler はリクエスト解析（文字列の取り出し）のみを行い、パース・タイムゾーン設定は Service に委譲すべき
- `time.FixedZone("Asia/Tokyo", 9*60*60)` というタイムゾーン設定はビジネスロジック（JST 前提）であり Handler の責務外
- `parseDateQuery` を使えば統一された日付パースが可能

### 修正後コード

```go
func (h *Handler) GetDailySummary(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }

    // ✅ 日付文字列の取り出しのみ Handler で行う
    dateStr := c.Query("date")
    result, err := h.svc.Accounting.GetDailySummary(c.Request.Context(), clinicID, dateStr)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, result)
}
```

```go
// Service 側でパース・タイムゾーン設定を吸収する
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error) {
    if dateStr == "" {
        dateStr = time.Now().Format("2006-01-02")
    }
    jst := time.FixedZone("Asia/Tokyo", 9*60*60)
    date, err := time.ParseInLocation("2006-01-02", dateStr, jst)
    if err != nil {
        return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
    }
    result, err := s.repo.GetDailySummary(ctx, clinicID, date)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get daily summary")
    }
    return result, nil
}
```

`GetDailySummary` の Service インターフェース・シグネチャも合わせて変更する:
```go
GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error)
```
