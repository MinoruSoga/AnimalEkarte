# TASK-108: CSV 生成ロジックが Service 層に配置されている（accounting_report_service.go）

## 優先度

**Medium** — 責務分離違反。CSV のフォーマット（列名・BOM・列構造）はプレゼンテーション層の関心事。

---

## 概要

`accounting_report_service.go` の `ExportCSV` メソッドが
CSV フォーマット（BOM 付与・ヘッダー行・列構成・合計行）を直接生成し `io.Reader` を返している。

CSV の列名やフォーマット変更はビジネスロジックではなくプレゼンテーション（レスポンス形式）の変更であり、
これを Service 層に持つと、フォーマット変更のたびに Service のテスト・ユニットテストを修正する必要がある。

正しい責務分離:
- **Service 層**: データを取得し DTO で返す
- **Handler 層**: DTO を CSV にシリアライズしてレスポンスに書き出す

---

## 問題箇所

### `service/accounting_report_service.go:40-88`

```go
// ❌ Service 層に CSV フォーマットロジックが混在
func (s *accountingReportService) ExportCSV(ctx context.Context, clinicID uint64, year, month int) (io.Reader, error) {
    result, err := s.repo.GetMonthlyReport(...)
    ...
    var buf bytes.Buffer
    w := csv.NewWriter(&buf)
    buf.WriteString("\xEF\xBB\xBF")  // BOM — プレゼンテーションの関心事
    header := []string{"日付", "区分(AM/PM)", "カテゴリ", "支払方法ID", "純売上(円)"}  // 列名 — プレゼンテーション
    w.Write(header)
    for _, row := range result.Rows { ... }
    w.Write([]string{"合計", "", "", "", ...})  // 合計行レイアウト — プレゼンテーション
    ...
    return &buf, nil
}
```

---

## 修正方針

`ExportCSV` を Service から削除し、`AccountingReportService` は `GetMonthly` のみを持つ。
CSV 生成は handler で行う。

### 1. `service/accounting_report_service.go`

```go
// ✅ ExportCSV メソッドを削除
// AccountingReportService インターフェースから ExportCSV を除去
type AccountingReportService interface {
    GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error)
    // ExportCSV は削除
}
```

### 2. `handler/accounting_report_handler.go`

```go
// ✅ CSV 生成ロジックをハンドラに移動
func (h *Handler) ExportMonthlyCSV(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
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

    w := csv.NewWriter(c.Writer)
    // BOM（Excel の UTF-8 認識用）
    c.Writer.Write([]byte("\xEF\xBB\xBF"))

    w.Write([]string{"日付", "区分(AM/PM)", "カテゴリ", "支払方法ID", "純売上(円)"})
    for _, row := range result.Rows {
        pmID := ""
        if row.PaymentMethodID != nil {
            pmID = fmt.Sprintf("%d", *row.PaymentMethodID)
        }
        w.Write([]string{
            row.Date, row.Period, row.Category, pmID,
            fmt.Sprintf("%d", row.NetAmount),
        })
    }
    w.Write([]string{"合計", "", "", "", fmt.Sprintf("%d", result.GrandTotal)})
    w.Flush()
}
```

ハンドラで `c.Writer` に直接書き出すことで `bytes.Buffer` の中間コピーも不要になる。

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/accounting_report_service.go:15-17` | AccountingReportService インターフェース | ExportCSV 削除 |
| `service/accounting_report_service.go:40-88` | ExportCSV 実装 | 削除 |
| `handler/accounting_report_handler.go:37-75` | ExportMonthlyCSV | CSV 生成ロジック移動 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — 依存関係の方向

> handler (プレゼンテーション層) → service (ビジネスロジック層) → repository (データアクセス層)

CSV フォーマットは「どう表示するか」の関心事であり、プレゼンテーション層（handler）が担うべき。
Service 層に置くと、フォーマット変更時に「ビジネスロジックの変更」として扱われ、
Service テストの更新が必要になる（不要なコスト）。

### プロジェクト内参照実装

他のエクスポート系ハンドラが存在する場合は、同様のパターンを参照すること。
