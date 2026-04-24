# TASK-141: [High] accounting_report_handler.go — CSV生成ロジックがHandlerに存在

## 優先度
**High**（責務違反 + Handler 肥大化、テスト不可）

## 対象ファイル
- `backend/internal/handler/accounting_report_handler.go`
- `backend/internal/service/accounting_report_service.go`

## 問題

`ExportMonthlyCSV` ハンドラが CSV の行データ変換・BOM 書き込み・`csv.NewWriter` 操作を直接実装している。
Handler の責務は「リクエスト解析 + Service 委譲」のみであり、データ整形・CSV 構築はすべて Service（または専用の presenter/formatter）の責務。

このままでは：
- CSV フォーマット変更のたびに Handler を修正する必要がある
- Unit テストで CSV 出力をテストできない（`gin.Context.Writer` が必要になる）

## 現状コード（handler/accounting_report_handler.go）

```go
// ExportMonthlyCSV godoc
// GET /v1/reports/monthly/csv?year=2026&month=4
func (h *Handler) ExportMonthlyCSV(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    // ... year/month パース ...

    result, err := h.svc.AccountingReport.GetMonthly(c.Request.Context(), clinicID, year, month)
    // ...

    filename := fmt.Sprintf("monthly_report_%04d%02d.csv", year, month)
    c.Header("Content-Type", "text/csv; charset=utf-8")
    c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
    c.Status(http.StatusOK)

    // BOM（Excel の UTF-8 認識用）
    _, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))

    w := csv.NewWriter(c.Writer)
    _ = w.Write([]string{
        "日付", "曜日", "AM件数", "AM純売上(円)", ...
    })

    for _, d := range result.DailyDetails {
        amClosed := "未"
        if d.AMClosed { amClosed = "済" }
        // ... 変換ロジック ...
        _ = w.Write([]string{...})
    }
    // ... 合計行 ...
    w.Flush()
}
```

## 修正後コード

### service/accounting_report_service.go に追加

```go
// MonthlyReportCSVRow は月次レポート CSV 1行のデータ
type MonthlyReportCSVRow struct {
    Fields []string
}

// BuildMonthlyCSV は MonthlyReportResponse から CSV 用の行データを構築する
func BuildMonthlyCSV(r *MonthlyReportResponse) []MonthlyReportCSVRow {
    rows := make([]MonthlyReportCSVRow, 0, len(r.DailyDetails)+2)
    // ヘッダー行
    rows = append(rows, MonthlyReportCSVRow{Fields: []string{
        "日付", "曜日", "AM件数", "AM純売上(円)", "PM件数", "PM純売上(円)", "日計(円)", "返金(円)", "AM締め", "PM締め", "休診",
        "標準税率課税対象額(円)", "標準税率消費税額(円)", "軽減税率課税対象額(円)", "軽減税率消費税額(円)",
    }})
    // 明細行
    for _, d := range r.DailyDetails {
        amClosed := "未"
        if d.AMClosed { amClosed = "済" }
        pmClosed := "未"
        if d.PMClosed { pmClosed = "済" }
        holiday := ""
        if d.IsHoliday { holiday = "休" }
        rows = append(rows, MonthlyReportCSVRow{Fields: []string{
            d.Date, d.Weekday,
            fmt.Sprintf("%d", d.AMCount), fmt.Sprintf("%d", d.AMNet),
            fmt.Sprintf("%d", d.PMCount), fmt.Sprintf("%d", d.PMNet),
            fmt.Sprintf("%d", d.DayNet), fmt.Sprintf("%d", d.Refund),
            amClosed, pmClosed, holiday,
            "", "", "", "",
        }})
    }
    // 合計行
    tb := r.Summary.TaxBreakdown
    rows = append(rows, MonthlyReportCSVRow{Fields: []string{
        "合計", "", "", "", "", "",
        fmt.Sprintf("%d", r.Summary.NetAmount),
        fmt.Sprintf("%d", r.Summary.TotalRefund),
        "", "", "",
        fmt.Sprintf("%d", tb.Standard.TaxableAmount),
        fmt.Sprintf("%d", tb.Standard.TaxAmount),
        fmt.Sprintf("%d", tb.Reduced.TaxableAmount),
        fmt.Sprintf("%d", tb.Reduced.TaxAmount),
    }})
    return rows
}
```

### handler/accounting_report_handler.go の修正

```go
func (h *Handler) ExportMonthlyCSV(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }

    year, month, err := parseYearMonth(c)
    if err != nil {
        RespondError(c, err)
        return
    }

    result, err := h.svc.AccountingReport.GetMonthly(c.Request.Context(), clinicID, year, month)
    if err != nil {
        RespondError(c, err)
        return
    }

    filename := fmt.Sprintf("monthly_report_%04d%02d.csv", year, month)
    c.Header("Content-Type", "text/csv; charset=utf-8")
    c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
    c.Status(http.StatusOK)

    _, _ = c.Writer.Write([]byte("\xEF\xBB\xBF")) // BOM

    csvRows := service.BuildMonthlyCSV(result)
    w := csv.NewWriter(c.Writer)
    for _, row := range csvRows {
        _ = w.Write(row.Fields)
    }
    w.Flush()
}
```

## 補足
Handler に残してよいのは BOM 書き込みと `csv.NewWriter` の I/O 部分のみ。
行データの生成・文字列変換・ラベル定義はすべて Service（または presenter 層）に移動する。
