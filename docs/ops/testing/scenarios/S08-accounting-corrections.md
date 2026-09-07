# S08: 会計訂正系（クレジット訂正・未収金）

> **目的**: 確定後のカード金額訂正が理由必須・権限制御・監査記録つきで機能すること、および未納が waiting 全額とクレジット訂正 residual の両方から把握できることを証明する。部分入金 UI は現行では拒否される。
> **所要目安**: 25分 / **深度**: 深い
> **仕様正本**: [会計一覧](../../../spec/screens/10-accounting-list.md) / [会計精算](../../../spec/screens/11-accounting-detail.md) / [未納者一覧](../../../spec/screens/30-unpaid-list.md) / [レジ締め業務](../../../spec/cash-register.md)

## 前提条件

- ローカルの使い捨て clinic に合成 owner/pet と attached account を 2 つ作成する。正例アカウントは accounting 操作、`accounting-post-close-edit:edit`、cash-register-close を持ち、負例アカウントは post-close edit を持たない。
- 本シナリオで作成する全会計を同じ合成 owner に紐付ける。自動 seed の役割名・会計データを仮定しない。
- append-only 会計/締め記録を含むため、専用 clinic/date と cleanup/破棄手順が確認できない環境では BLOCKED とする。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:---|:---|
| 1 | 「新規会計登録」から検証用飼主の会計（1 件目）を作成し、手動項目を追加して支払方法「クレジットカード」で全額精算・確定 | 会計一覧でステータス「精算済」（`ACCOUNTING_STATUS_LABELS.completed`）、支払方法にカードが表示される。新規確定は `POST /api/v1/accountings/complete`（BUG-018 原子 complete・Idempotency-Key 付き） |
| 2 | 当該会計の詳細 `/accounting/:id` でクレジット訂正の導線を開き、訂正理由を空のまま実行 | 「訂正理由を入力してください」と表示され、訂正は実行されない（`CreditCorrectionDialog`） |
| 3 | 訂正理由を入力し、カード金額を変更して訂正を実行 | 訂正後の金額が会計に反映される。ダイアログに「訂正理由は監査ログに記録されます」の説明があること。API は `POST /api/v1/accountings/:id/credit-correction` |
| 4 | 会計一覧へ戻り、1 件目の行を確認 | 訂正後の金額・支払方法（カード）で表示され、ステータスは「精算済」のまま |
| 5 | 権限グループ「一般」のスタッフでログインし直し、同じ会計詳細を開く | クレジット訂正の導線が表示されない |
| 6 | 「執行」で再ログインし、同じ飼主の会計（2 件目）を作成。請求額より少ない現金額のみ入力して保存（部分入金） | **部分入金は現行 UI/BE とも拒否**: `PaymentCard` は `remaining !== 0` で確定 disabled。すり抜けても BE が合計不一致で 400 |
| 7 | 会計一覧の「未納者一覧」タブ（`?tab=unpaid`）を開く | 期間未指定でも JST 当月で API が発火する（空のまま「未納者はいません」にしない）。手順 3 のクレジット訂正 residual が未納として載る（`unpaidAmountSQL > 0`。waiting 全額 + completed の訂正差額） |
| 8 | 検証用飼主の 1 件目の会計詳細を開く | クレジット訂正 residual がある場合、未納残高カードと未納者一覧の金額が一致する |
| 9 | 未納者一覧から 1 件目へ遷移する | 詳細が開き、訂正後の残額を確認できる。部分入金で作った 2 件目は存在しない |
| 10 | 対象日の該当区分（1 件目の完了時刻が属する区分）のレジ締めを確定した後、1 件目のクレジット訂正を理由つきで再実行 | 訂正は成功し、締め済み期間への訂正（締め後訂正）として区別記録される（確認観点参照） |

## 確認観点

- **監査証跡（fail-closed）**: クレジット訂正は監査ログと同一トランザクションで記録され、監査記録に失敗すると訂正ごとロールバックされる（`backend/internal/billing/accounting_service_correction.go` / SOLO-03）。#10 の締め後訂正は監査メタデータに `post_close: true` と対象日が記録される（#189）。
- **訂正対象の限定**: 訂正できるのはカード系（`credit_card`・`electronic_money`）の内訳のみ。現金の誤りは別フロー（お釣り上書き #188）の管轄で、本訂正機能の対象外。
- **権限の分離**: 訂正導線は「確定済み（completed）かつ accounting-post-close-edit の edit 保有」でのみ表示される（BE: `POST /accountings/:id/credit-correction` が同権限を要求）。
- **未収金の定義**: waiting 全額に加え、クレジット訂正由来の completed 差額（residual）も含む。部分入金 UI は無いが、#3 の訂正後に未納タブで突合できる。
- **精算済み再保存**: 詳細は「精算済みの会計を修正します」ConfirmDialog を出してから保存する。締め済み日の新規 `POST /accountings/complete` は `post_close_reason` を FK より先に見る。
- **拠点横断時の訂正対象**: 訂正リクエストはグローバル選択クリニックではなく会計自体のクリニックに対して送られる（`frontend/src/features/accounting/api/correct-credit-payment.ts`、X-Clinic-ID 明示指定）。拠点横断で開いた会計でも誤テナントへ書き込まれないこと。
- **原子 complete（BUG-018）**: 新規精算は `completeAccounting` → `POST /v1/accountings/complete`。明細・支払・監査を同一 TX で確定し、Idempotency-Key で再送を安全化する。#1 の成功時に一覧が「精算済」になること。
- **支払方法 rename 回帰（#197）**: 標準行は表示名を変更しても `system_key` で解決され、精算・現金集計から消えないこと（V04 支払方法マスタと連携）。#6〜#9 の部分入金統合経路は現行 UI では実行できない。

## 異常系

- **理由なし訂正**: #2 で検証（クライアント側の必須制御に加え、API も理由必須のためすり抜け不可）。
- **締め済み期間の通常編集**: 締め済み期間の会計を通常の編集で再保存する場合も修正理由（`post_close_reason`）の入力（必須）が要求され、監査ログに記録される（仕様正本 11 §2.1・#115）。警告表示の検証は S09 #8 で行う。

## 実装突合
- 変更:
  - 精算経路を `POST /accountings/complete`（BUG-018 原子 complete）として明記
  - 部分入金拒否を FE `remaining !== 0` + BE `validatePaymentSplits` で再確認
  - クレジット訂正 API パス・カード系 method・未納定義（waiting）を現行実装に合わせて整理
  - 監査 fail-closed / X-Clinic-ID 拠点横断は変更なし
