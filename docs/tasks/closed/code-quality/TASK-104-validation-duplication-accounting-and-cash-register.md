# TASK-104: バリデーション重複（accounting_report_service / cash_register_service）

## 優先度

**Low** — 機能は正常だが、同じバリデーションロジックが複数箇所に散在しておりメンテナンス性が低い。

---

## 概要

### 問題1: `accounting_report_service.go` の month バリデーション重複

`GetMonthly` と `ExportCSV` の両メソッドで同一の月バリデーションを重複記述。
さらに handler の `parseYearMonth` でも同じバリデーションをするため、実質3重チェックになっている。

### 問題2: `cash_register_service.go` の period バリデーション重複

`GetPreview` と `Close` の両メソッドで `period != "am" && period != "pm"` を重複記述。
さらに `resolvePeriodRange` の default ケースでも同じバリデーションが存在し、3重チェック。

---

## 問題箇所

### `service/accounting_report_service.go:31-33` と `:42-44`

```go
// GetMonthly 内
if month < 1 || month > 12 {
    return nil, apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
}

// ExportCSV 内（同一ロジック）
if month < 1 || month > 12 {
    return nil, apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
}
```

### `service/cash_register_service.go:65-67` と `:105-107`

```go
// GetPreview 内
if period != "am" && period != "pm" {
    return nil, apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
}

// Close 内（同一ロジック）
if period != "am" && period != "pm" {
    return nil, apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
}
```

---

## 修正方針

### 1. `service/accounting_report_service.go`

`validateMonth` ヘルパー関数を抽出する。または service layer のバリデーションを削除し
handler の `parseYearMonth` のみで保証する（handler がすでに同じバリデーションを実施）。

```go
// ✅ ヘルパー抽出
func validateMonth(month int) error {
    if month < 1 || month > 12 {
        return apperrors.WrapInvalidInput("month は 1〜12 の範囲で指定してください")
    }
    return nil
}

func (s *accountingReportService) GetMonthly(ctx context.Context, clinicID uint64, year, month int) (*repository.MonthlyReportResult, error) {
    if err := validateMonth(month); err != nil {
        return nil, err
    }
    ...
}

func (s *accountingReportService) ExportCSV(ctx context.Context, clinicID uint64, year, month int) (io.Reader, error) {
    if err := validateMonth(month); err != nil {
        return nil, err
    }
    ...
}
```

### 2. `service/cash_register_service.go`

`validatePeriod` ヘルパー関数を抽出する。

```go
// ✅ ヘルパー抽出
func validatePeriod(period string) error {
    if period != "am" && period != "pm" {
        return apperrors.WrapInvalidInput("period は 'am' または 'pm' を指定してください")
    }
    return nil
}

func (s *cashRegisterService) GetPreview(ctx context.Context, ...) (*CashRegisterPreview, error) {
    if err := validatePeriod(period); err != nil {
        return nil, err
    }
    ...
}

func (s *cashRegisterService) Close(ctx context.Context, ...) (*model.CashRegisterClose, error) {
    if err := validatePeriod(input.Period); err != nil {
        return nil, err
    }
    ...
}
```

`resolvePeriodRange` の default ケースは period が事前バリデーション済みなら到達しないため、
`panic("unreachable")` か削除しても良い。

---

## 影響範囲

| ファイル | 行 | 問題 |
|---------|---|------|
| `service/accounting_report_service.go` | 31-33, 42-44 | month バリデーション重複 |
| `service/cash_register_service.go` | 65-67, 105-107 | period バリデーション重複 |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — インターフェース設計

> インターフェース最小化（3-5メソッド）

重複コードの排除は DRY 原則の基本。Go では小さいヘルパー関数として抽出するのが慣用的。
