# レジ締め・履歴 仕様書 (Cash Register Close)

## 概要
- **画面の目的**: 日次の売上集計、レジ内の現金の突合、および過去の締め記録の管理。
- **URLパターン**: 
  - レジ締め: `/accounting/close`
  - 締め履歴: `/accounting/close/history`
- **アクセス権限**: 会計担当者・医院管理者（`ResourceCashRegisterClose` 権限が必要）

---

## 画面構成

### 1. 当日の売上集計 (`/close`)
- **集計単位選択**: 「午前」「午後」「緊急」のいずれかの区分で集計を実行可能。
- **理論売上表示**: システムが自動計算した、支払方法別（現金、カード等）の売上合計を表示。
- **越日「緊急」の帰属**: 「緊急」区分は当日の診療終了時刻（PM終了）〜翌日の診療開始時刻（AM開始）までの越日レンジで集計される（`resolvePeriodRange`、#215）。そのため深夜0:00〜AM開始前に発生した会計は、日付が変わっていても「前日の緊急」区分に帰属する。過去の締め済みレコードは境界設定（AM開始・PM終了時刻）を変更しても再計算されない。

### 2. 現金突合フォーム (`CashReconciliationCard`)
- **実レジ金額入力**: レジ内の現金の合計額を1つの有限な 0 以上の数値で入力（金種別内訳の入力欄はない）。負値または有限でない値では送信を無効化し、インラインエラーを表示する。
- **過不足算出**: 理論上の現金売上と実際の入力値との差異をリアルタイムに算出。
- **メモ入力（任意）**: 差異が発生した理由や特記事項を記載。

### 3. 締め履歴一覧 (`/history`)
- **過去の締め記録**: 日付・区分・理論現金・実際の現金・差額・担当者・締め時刻を一覧表示。既定期間は JST 当月。レポートからのドリルダウンは `?date=YYYY-MM-DD` で当該日のみ。
- **区分フィルタ**: AM/PM/緊急は取得済みページ内のクライアント絞り込み。
- **詳細参照**: 各レコードの日付セルのボタンをクリックすると、当時の売上内訳を再確認可能。詳細ダイアログは一覧の snapshot を使い、GET `/:id` は呼び出さない。

---

## 主要な機能

### 1. 会計レコードのロック
「締める」を実行すると、該当日（同日のいずれかの区分が締め済みなら日単位）の会計レコードが編集不可の状態にロックされます。これにより、確定後の不用意な書き換えによる売上不整合を物理的に防止します。修正が必要な場合は、`ResourceAccountingPostCloseEdit`（`accounting-post-close-edit`）権限を持つ担当者が編集理由（`post_close_reason` 必須）を添えて編集します（ロック解除操作は存在しない）。

### 2. 日次報告書の出力
レジ締め画面のプレビュー表示（未締め時）から、「印刷 / PDF出力」ボタンで A4 横形式の「レジ締めサマリー」を印刷・PDF保存できます（`PrintPortal` 基盤を再利用）。院内保管用や本部への提出資料として使用します。締め履歴の詳細ダイアログには印刷導線はありません。

### 3. 部門×支払方法マトリクス（#247 / DEC-16⑥）
- **金額基準**: 支払実額（割引適用後）。期間全体の payment 比率による擬似按分は行わない。
- **配賦**: 会計（billing）単位で明細金額比例・最大剰余法によりカテゴリへ配賦。行合計=列合計=総計を円単位で保存する。
- **返金**: 発生日（`refunded_at`）の負値としてマトリクスに載せる。
- **件数**: 会計 distinct（payment split による二重計上なし）。`other` 行は DEC-40 の独立 distinct。
- **支払方法列**: 医院 master の display_order。期間内にデータがある inactive / 削除済み / 不明 method は末尾表示。
- **画面と印刷**: `UnifiedClosingSummaryTable` / `ClosePrintArea` は同一の配賦マトリクスを描画源とする。

#### DEC-16⑥ 返金・総額契約（#247 review MEDIUM / W15）

> ⑥#247 配賦契約: matrix 総額=支払実額基準（割引適用後・締め合計と一致）・割引は明細金額比例で配賦・返金は発生日の負値行・端数は最大剰余法で行合計=列合計=総計を円単位保存・「件数」=会計 distinct・支払方法列は医院 master 順で構築し期間内データを持つ無効/削除済み method は末尾表示。

##### (a) 返金は発生日の負値行（pre-aggregation 禁止）

| 項目 | 内容 |
|:---|:---|
| DEC 引用 | 返金は発生日の負値行 |
| 実装 | `GetCategoryPaymentAllocationData` が `billing_refunds` を `refunded_at ∈ [periodStart, periodEnd)` で **行単位** 取得（`GROUP BY` / 期間内 net 合算なし）。`BuildAllocationBillings` が各返金を `Amount: -ref.Amount` の独立 `AllocationPayment` として `Payments` に append。`AllocateBillingPayments` は payment / 負値 refund を **1 行ずつ** 最大剰余法で配賦する。 |
| 参照 | `backend/internal/billing/accounting_repository_reports_allocation.go`（refund SQL / `BuildAllocationBillings`）、`allocation.go`（`AllocateBillingPayments` / `AggregateCategoryPaymentMatrix`） |
| 判定 | **match** — 同一 billing×method に支払と返金が同居しても、支払へ事前相殺（pre-net）してから配賦しない。親会計が期間外完了でも `RefundParentWeights` でカテゴリ重みだけ補い、返金は発生日の負値行のまま。billing 単位の 1 allocation pass は可（負値行セマンティクスを壊さない）。 |
| 回帰 | `TestBuildAllocationBillings_RefundsRemainSeparateNegativeRows`、`TestAllocateBillingPayments_RefundNegative`、`TestBuildAllocationBillings_Conservation` |

##### (b) matrix 総額 = 締め合計（KPI NetAmount との二重定義）

| 定義 | 算式 | 用途 |
|:---|:---|:---|
| **締め合計 / matrix grand（DEC-16⑥）** | `Σ payment_splits(completed_at ∈ 期間)` − `Σ billing_refunds(refunded_at ∈ 期間)` | レジ締めプレビューの部門×支払マトリクス、締め snapshot の `category_breakdown`、月次 `category_payment_matrix.totals.grand_total`。close / monthly は同一 helper（`GetCategoryPaymentAllocationData` → `BuildAllocationBillings` → `AggregateCategoryPaymentMatrix`）。 |
| **KPI NetAmount（完了会計帰属の返金）** | `Σ payment_splits(completed_at ∈ 期間)` − `Σ refunds(親 billing.completed_at ∈ 期間)` | 月次 `summary.net_amount`（`MonthlyReportResult.GrandTotal`）、`GetCloseAggregate.TotalRefund` 経由の理論現金控除。`sumRefundsForCompletedBillings` / close Query3。 |

| 項目 | 内容 |
|:---|:---|
| DEC 引用 | matrix 総額=支払実額基準（割引適用後・締め合計と一致） |
| 実装 | matrix grand は上記「締め合計」定義そのもの。支払は割引適用後の `payment_splits.amount`（税・保険・会計割引は支払実額側に折り込み済み）。行合計=列合計=総計は最大剰余法で円単位保存。 |
| 判定 | **match（matrix == 締め合計）**。同一期間・同一 clinic では close マトリクス総額と monthly `category_payment_matrix.totals.grand_total` が一致する。 |
| KPI との差 | 返金発生日 ≠ 親会計完了日のとき、`summary.net_amount` / `TotalRefund`（完了会計帰属）と matrix grand（発生日帰属）は **意図的に不一致** になり得る。DEC-16⑥ が正とするのは matrix / 締め合計側。KPI 側は「完了会計に紐づく返金」のレガシー集計であり、matrix を KPI に合わせて書き換えない。 |
| 回帰 | `TestBuildAllocationBillings_Conservation`、`TestBuildAllocationBillings_MatrixGrandEqualsOccurrenceNetNotCompletedAttachedKPI`、`TestAggregateCategoryPaymentMatrix_MultiBillingConservation` |

---


## 技術仕様

### 使用コンポーネント
- **`CashRegisterClosePage`**: レジ締めメイン画面。
- **`CashReconciliationCard`**: 理論現金・実際現金・差額の表示モジュール。
- **`ClosePrintArea`**: 帳票用テンプレートコンポーネント（`cash-register/components/`）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/cash-register/preview` | 指定期間の未確定売上の集計取得 | `cash-register-close` | `view` |
| POST | `/api/v1/cash-register/closes` | レジ締めの確定保存とロックの実行 | `cash-register-close` | `create` |
| GET | `/api/v1/cash-register/closes` | 過去の締め履歴一覧の取得 | `cash-register-close` | `view` |
| GET | `/api/v1/cash-register/closes/:id` | 特定の締め記録詳細の取得（BE実装済みだが本画面からは未呼出） | `cash-register-close` | `view` |

---

