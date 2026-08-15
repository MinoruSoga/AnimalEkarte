# BUG-007 回帰メモ（クレジット訂正差額の未収追跡）

## 修正内容（要約）

- 未収定義を拡張: `status=waiting` 全額に加え、`completed` かつ `patient_due > billing_amount` の差額を未収とする
- `patient_due = payments.total_amount - insurance_amount - discount_amount`（complete 時の請求額と同式）
- API: `outstanding_amount` を会計レスポンスに追加
- UI: 会計詳細に「この会計の未収残高」、未納者一覧（会計単位）は `outstanding_amount` を表示
- クレジット訂正成功時に unpaid / owner unpaid balance を invalidate

## 回帰マトリクス

| シナリオ | 期待 | 自動検証 |
|:---|:---|:---|
| 通常カード全額精算（due=collected） | 未収一覧に出ない / outstanding=0 | `TestUnpaidIncludesCreditCorrectionResidual_BUG007` full paid 行 + `TestOutstandingAmount` |
| クレジット訂正 1100→900 | outstanding=200、未納一覧・飼主残高に載る | 同上 residual 行 |
| waiting のみ（payment 無し） | 従来どおり total_amount 全額 | 既存 Unpaid 系 + SumUnpaidByOwner |
| 保険あり due=5000・collected=4000 | residual=1000（medical 全額ではない） | `TestOutstandingAmount` insurance |
| 過入金（collected>due） | residual=0 | `TestOutstandingAmount` over-collection |
| 部分入金 UI 経路 | 現行どおり `remaining !== 0` で確定 disabled（本 BUG 対象外・S08 BLOCKED） | 変更なし・回帰メモのみ |

## 手動確認（Needs Human / QA）

1. カード全額精算 → クレジット訂正で減額 → 会計詳細に未収残高カード
2. `/accounting?tab=unpaid` 飼主単位・会計単位で差額が件数・金額に載る
3. 再訂正で due まで戻すと未収から消える
4. 通常精算のみの会計が未納一覧に誤爆しない

## 非対象

- merge/push/migrate
- 部分入金 UI の解放（S08 #6–9）
- 他 BUG
