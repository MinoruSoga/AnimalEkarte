# 混在会計 (Mixed Payment) スモークテストチェックリスト

> **目的**: 混在会計(payment_splits)の動作確認手順を定義する。
> **読者**: デプロイ担当・QA。
> **タイミング**: 混在会計関連の変更時。

## 前提データ

- 動物病院クリニックがログイン済み
- テスト用ペット・飼主が登録済み
- カルテ → 会計作成済み (status: waiting)
- 確認する会計の `billing_amount` が設定済みであること

---

## 1. 単一支払い (現金のみ)

**操作:** 会計詳細ページで支払方法「現金」のみ選択、金額を入力して保存

| チェック項目 | 期待値 |
|------------|--------|
| 保存が成功する | status: completed |
| `payment_splits` に1行 | method=cash, amount=billing_amount |
| `payments.method` | cash |
| 受取金・お釣りが正しく反映 | received_amount - amount = change_amount |
| 会計詳細再表示で値が復元される | 入力値と一致 |

---

## 2. 2種混在 (現金 + クレジットカード)

**操作:** 「支払い方法を追加」で2行目を追加し、現金とクレジットカードに金額を分割して保存

| チェック項目 | 期待値 |
|------------|--------|
| 保存が成功する | status: completed |
| `payment_splits` に2行 | cash + credit_card |
| 両方の金額合計 = billing_amount | バリデーション通過 |
| `payments.method` (legacy) | cash (優先度: cash > credit_card > bank_transfer > electronic_money) |
| 会計詳細再表示で2行が表示される | 各 method・amount が復元 |
| 本日会計タブの支払方法列 | 「現金 / クレジットカード」表示 |

---

## 3. 3種混在 (現金 + クレジットカード + 電子マネー)

**操作:** 3行入力し、合計を`billing_amount`へ一致させて保存する。

| チェック項目 | 期待値 |
|---|---|
| 保存 | status: completed |
| `payment_splits` | cash + credit_card + electronic_money の3行 |
| 合計 | `sum(amounts) == billing_amount` |
| 表示/集計 | 3手段それぞれに復元・計上 |

### 3.1 bank_transferを含む代表case

`bank_transfer`は対応済み。`cash + bank_transfer`または4種のcaseでsplit保存、再表示、日次集計を確認する。legacy `payments.method`の代表選択優先度は`cash > credit_card > bank_transfer > electronic_money`。

---

## 4. 現金お釣り計算

**操作:** 現金1種で、受取金 > billing_amount の場合に保存

| チェック項目 | 期待値 |
|------------|--------|
| お釣り = 受取金 - billing_amount | 自動計算済みで保存 |
| change_amount が正しく保存される | payment_splits.change_amount 確認 |

---

## 5. バリデーション — 保存不可ケース

| 操作 | 期待されるエラー |
|------|----------------|
| 支払い合計 < billing_amount | 合計不足エラー、保存ボタン無効 |
| 支払い合計 > billing_amount | 合計超過エラー、保存ボタン無効 |
| 同じ支払方法を2行追加 | 重複エラー (HTTP 400) |
| 現金の受取金 < 現金 amount | 預り金不足エラー |
| 現金のお釣り計算不正 | お釣り計算エラー |

---

## 6. 再表示 (保存済み payment_splits の復元)

**操作:** 混在会計を保存後、同ページを再度開く (または再読み込み)

| チェック項目 | 期待値 |
|------------|--------|
| 保存した splits の行数が一致 | 保存時と同数 |
| 各行の method が復元される | ドロップダウン値が一致 |
| 各行の amount が復元される | 入力値が一致 |
| 現金の受取金・お釣りが復元される | `receivedAmount`, `changeAmount` 一致 |

---

## 7. 本日会計タブ (DailyAccountingTab)

**操作:** 本日会計タブを開き、混在支払いの会計を確認

| チェック項目 | 期待値 |
|------------|--------|
| 単一支払いの会計 | 支払方法列に1種のラベル |
| 2種混在の会計 | 「現金 / カード」など「/」区切り (DailyAccountingTab の credit_card ラベルは "カード") |
| 集計サマリーカード | 支払方法別に金額が分かれて表示 |
| 日次サマリー API | payment_totals に複数手段のエントリ |

---

## 8. レジ締め (CloseAggregate) プレビュー

**操作:** レジ締め画面を開き、混在会計が含まれる期間でプレビュー

| チェック項目 | 期待値 |
|------------|--------|
| 支払方法別集計 | payment_splits.amount を参照 (payments.billing_amount 非依存) |
| 現金欄の金額 | 混在会計の cash 行 amount のみ加算 |
| クレジットカード欄 | 混在会計の credit_card 行 amount が正しく加算 |

---

## 9. 返金 (Refund)

返金APIはoptional `payment_method`を実装済み。許可値は`cash` / `credit_card` / `electronic_money` / `bank_transfer`。

| case | 期待 |
|---|---|
| splitに存在するmethodを指定し、そのmethodの残額以内 | success。保存/listの`payment_method`が指定値と一致 |
| method省略、billing全体の残額以内 | success。保存値は省略/null |
| splitに無いmethodを指定 | `400` |
| 指定methodの残額を超える | `400` |
| billing全体の残額を超える | `400` |

返金後は`billing_refunds`、`total_refunded_amount`、method別残額を確認する。

> **既知のreporting制約:** close-report detailはbilling-levelの`refund_amount`と`net_amount = billing_amount - refund_amount`を計算する。method未指定返金をpayment splitsへ比例配分する表示ではない。この制約は返金作成・保存が未実装という意味ではない。

---

## 10. 既存会計 (payment_splits なし) の後方互換

**対象:** 混在会計機能追加前に作成された会計 (payment_splits 行なし)

| チェック項目 | 期待値 |
|------------|--------|
| 会計詳細を開ける | エラーなし |
| payments[0] の内容で表示される | method, billing_amount が正しく表示 |
| paymentSplits は undefined | FE で payment フォールバック表示 |
| 再保存時 (単一支払い) | backward compat: 1行 split が自動生成される |
| 月次レポートに含まれる | payment_splits JOIN で集計される (seed backfill 済み) |

---

## 11. NG 切り分け

| 現象 | 確認箇所 |
|------|---------|
| 保存後 payment_splits が0件 | `SavePaymentSplits` のトランザクション / DELETE 後 INSERT 漏れ |
| 集計に混在会計が反映されない | `GetDailySummary` / `GetCloseAggregate` の JOIN 条件を確認 |
| 再表示で splits が消える | API response の `payment_splits` フィールド / `transformToAccounting` の fallback 条件 |
| 全額が現金欄に集計される | 旧 `payments.billing_amount` 参照コードが残存していないか確認 |
| 合計一致なのに保存ボタンが無効 | FE の `isDisabled` 計算 / `parseInt(s.amount)` のパース誤り |

---

*作成: 2026-05-24 — 混在会計 (payment_splits) 実装追加時のスモークテスト手順*
