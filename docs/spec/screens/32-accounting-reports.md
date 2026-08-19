# 月次集計レポート 仕様書 (Monthly Accounting Reports)

## 概要
- **画面の目的**: 月単位または任意期間の売上実績、支払方法内訳、消費税明細を把握し、経営分析や確定申告・会計処理の基礎データを提供する。
- **URLパターン**: `/accounting/reports`
- **アクセス権限**: 経理・医院管理者権限が必要（`ResourceAccountingReports`）

---

## 画面構成

画面は次のゾーンで構成します。上部の「集計単位」セレクタで、月次（年・月の選択）と期間指定（開始日・終了日の入力）を切り替えられます。初期表示は JST 当月（`reportMode = "month"`）。

### ゾーン1. 操作（フィルタ）
集計単位・年月（または開始日・終了日）の選択行。ヘッダー右の「印刷 / PDF出力」「CSV出力」は現状維持。

### ゾーン2. 日次明細
見出しは `{periodLabel} の日次明細`（例:「2026年7月 の日次明細」）。一ヶ月（または指定期間）の推移を日別に詳細表示します。
- **項目**: 日付、曜日、午前件数、午前売上、午後件数、午後売上、日計、返金、AM締/PM締ステータス。
- **ドリルダウン**: レジ締めの閲覧権限（`ResourceCashRegisterClose`）を保持し、かつ AM/PM いずれかの締めが実行済みの行は、クリックまたは Enter キーで締め履歴（`/accounting/close/history`）の該当日へ遷移できます。

### ゾーン3. 結論（KPI + 内訳）— `MonthlySummaryCards`
日次明細より下に配置する（#179 ④-b）。同一データの二重表示を避けるため、上部の支払方法別・消費税セクションは置かない。
- **KPI 4枚**: 診療日数 / 会計件数 / 売上合計（返金併記） / 純売上。
- **内訳3枚**: 支払方法別合計 / 部門別合計 / 消費税内訳（医院設定マスタの標準・軽減税率ごと。病院設定の閲覧権限保持時のみ消費税内訳ヘッダ右に「税率設定を変更」リンクを `taxSettingsLink` として表示）。

### ゾーン4. 部門×支払方法統合表（#247 / DEC-16⑥）— `CategoryPaymentMatrixTable`
- **目的**: カテゴリ×支払方法を 1 表で確認し、別集計・手計算を削除する。
- **金額**: 支払実額基準（割引適用後・締め合計と一致する配賦 helper を共有）。
- **配賦**: 会計単位・明細金額比例・最大剰余法。返金は発生日の負値。
- **件数**: 会計 distinct（フッタ件数は期間内完了会計の総 distinct。行件数の単純合算ではない）。
- **列順**: 医院 payment method master 順。期間内データ付き inactive / unknown は末尾。
- **印刷**: `MonthlyReportPrintArea` が同一 `category_payment_matrix` を描画（列欠落・横切れを避けるため A4 横）。

---

## 主要な機能

### 1. CSVエクスポート
画面に表示されている集計データ（月次または期間指定）を CSV 形式でダウンロード可能です。外部の会計ソフト（弥生、freee、マネーフォワード等）へのデータ取り込みの補助として利用できます。

### 2. 日別集計とレジ締めステータスの連動
日次明細の各行は `completed_at`（JST変換）の暦日単位で集計されます（医院設定に基づく夜間の日界カットオフは実装されていません）。「AM締」「PM締」列は、当該日の午前・午後区分でレジ締め（`/accounting/close`）が実行済みかどうかを表示します。

### 3. 印刷 / PDF出力
データ取得後、ヘッダーの「印刷 / PDF出力」ボタンからブラウザ印刷（PDFとして保存）で帳票を出力できます。印刷面（`MonthlyReportPrintArea`）は画面表示と同一の集計データを描画源とし、A4 横向きで対象期間・病院名・KPI・支払方法別・消費税・部門×支払マトリクス（`category_payment_matrix`）・日次明細を出力します。

---

## 技術仕様

### 使用コンポーネント
- **`AccountingReportsPage`**: メインページ。
- **`MonthlySummaryCards`**: 指標サマリーのカード部品。
- **`DailyBreakdownTable`**: 日次推移の一覧テーブル。
- **`CategoryPaymentMatrixTable`**: #247 部門×支払方法統合表。
- **`MonthlyReportPrintArea`**: 印刷 / PDF出力用の帳票ビュー（`PrintPortal` 経由・印刷時のみ表示）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/reports/monthly` | 指定年月（`year`/`month`）または期間（`start_date`/`end_date`）の集計データ取得。`category_payment_matrix` を含む | `accounting-reports` | `view` |
| GET | `/api/v1/reports/monthly/csv` | 集計データの CSV 出力（同じく年月・期間の両指定に対応） | `accounting-reports` | `view` |

---
