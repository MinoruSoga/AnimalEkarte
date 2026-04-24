# TASK-144: [Low] accounting_report_service.go — DTO・レスポンス型の定義順序違反

## 優先度
**Low**（コード整序）

## 対象ファイル
- `backend/internal/service/accounting_report_service.go`

## 問題

プロジェクト規約では Service ファイルの定義順序は以下の通り:

```
1. Input DTO (CreateXxxInput, UpdateXxxInput)
2. const colXxx* = "..." （あれば）
3. func buildXxxUpdateFields(...) （あれば）
4. type XxxService interface { ... }
5. type xxxService struct { ... }
6. メソッド実装
```

現状の `accounting_report_service.go` では、レスポンス型（`MonthlyReportResponse` 等）が
インターフェース・struct の**前**に定義されている点は正しいが、
`validateMonth` 関数（L88〜94）が `accountingReportService` struct（L69〜73）の定義より後に
独立した関数として定義されており、インターフェース前に置くべき validator/helper が
struct 定義と実装の間に挟まっている。

また `buildMonthlyReportResponse`（L131〜252）、`resolvePaymentMethodName`（L255〜263）、
`paymentMethodNameForClose`（L267〜269）、`buildTaxBreakdown`（L273〜284）、
`buildPayMethodNameMap`（L288〜294）などのヘルパー関数がすべてメソッド実装の後に定義されており、
ファイルの最下部まで読まないと helper の存在を把握できない。

## 現状の定義順序

```
L14  AccountingReportService interface
L18  MonthlyReportResponse（レスポンス型）
L29  MonthlyReportSummary
L41  TaxBreakdownSummary
L46  TaxBreakdownEntry
L52  DailyReportDetail
L69  accountingReportService struct
L76  NewAccountingReportService
L88  validateMonth（← ここに置くべき）
L96  GetMonthly（メソッド実装）
L131 buildMonthlyReportResponse（← helper は実装の前に置くべき）
L255 resolvePaymentMethodName
L267 paymentMethodNameForClose
L273 buildTaxBreakdown
L288 buildPayMethodNameMap
```

## 修正後の定義順序（推奨）

```
1. Input DTO（今回は入力なし）
2. Response DTO（MonthlyReportResponse, MonthlyReportSummary, ...）
3. const / 定数（タイムゾーン等 - TASK-143 対応後）
4. validateMonth
5. buildMonthlyReportResponse
6. resolvePaymentMethodName
7. buildTaxBreakdown
8. buildPayMethodNameMap
9. AccountingReportService interface
10. accountingReportService struct
11. NewAccountingReportService
12. GetMonthly（メソッド実装）
```

## 修正方法

ファイル内の関数・型ブロックを上記順序に並べ替える。
動作変更は伴わないため、テスト通過を確認するだけでよい。

```bash
docker compose exec backend go test ./internal/service/... -v
```
