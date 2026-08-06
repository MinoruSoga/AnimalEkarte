# S08: 会計訂正系（クレジット訂正・未収金）

> **目的**: 確定後のカード金額訂正が理由必須・権限制御・監査記録つきで機能すること、および現行 UI では部分入金を保存できず未収金の把握→消し込みフローが到達不能であることを証明する。
> **所要目安**: 25分 / **深度**: 深い
> **仕様正本**: [会計一覧](../../../spec/screens/10-accounting-list.md) / [会計精算](../../../spec/screens/11-accounting-detail.md) / [未納者一覧](../../../spec/screens/30-unpaid-list.md) / [レジ締め業務](../../../spec/cash-register.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。
- ログイン: 権限グループ「執行」のスタッフ（accounting 全操作＋ accounting-post-close-edit の edit を保有）。負例（#5）は権限グループ「一般」のスタッフ（accounting-post-close-edit の edit なし）。#10 レジ締めは執行でも cash-register-close 付与済みのため執行で可。
- 検証用飼主: 飼主一覧から任意の飼主 1 名を選び、以降の会計をすべて同一飼主で作成する（未収金の突合のため）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:---|:---|
| 1 | 「新規会計登録」から検証用飼主の会計（1 件目）を作成し、手動項目を追加して支払方法「クレジットカード」で全額精算・確定 | 会計一覧でステータス「精算済」（`ACCOUNTING_STATUS_LABELS.completed`）、支払方法にカードが表示される。新規確定は `POST /api/v1/accountings/complete`（BUG-018 原子 complete・Idempotency-Key 付き） |
| 2 | 当該会計の詳細 `/accounting/:id` でクレジット訂正の導線を開き、訂正理由を空のまま実行 | 「訂正理由を入力してください」と表示され、訂正は実行されない（`CreditCorrectionDialog`） |
| 3 | 訂正理由を入力し、カード金額を変更して訂正を実行 | 訂正後の金額が会計に反映される。ダイアログに「訂正理由は監査ログに記録されます」の説明があること。API は `POST /api/v1/accountings/:id/credit-correction` |
| 4 | 会計一覧へ戻り、1 件目の行を確認 | 訂正後の金額・支払方法（カード）で表示され、ステータスは「精算済」のまま |
| 5 | 権限グループ「一般」のスタッフでログインし直し、同じ会計詳細を開く | クレジット訂正の導線が表示されない |
| 6 | 「執行」で再ログインし、同じ飼主の会計（2 件目）を作成。請求額より少ない現金額のみ入力して保存（部分入金） | **部分入金は現行 UI/BE とも拒否**: `PaymentCard` は `remaining !== 0` で確定 disabled（「残り ¥… 未入力」）。すり抜けても BE `validatePaymentSplits` が合計不一致で 400。部分入金→未納 #7–#9 は **BLOCKED**（runtime 2026-08-01: 未入力 disabled を観測） |
| 7 | 手順 6 のまま会計一覧の「未納者一覧」タブ（`?tab=unpaid`）を開く | 2 件目は確定送信されていないため、検証用飼主の未納額として新規計上されない。部分入金からの未納計上確認は現行 UI では BLOCKED |
| 8 | 検証用飼主の 1 件目の会計詳細を開く | 手順 6 由来の未納残高カードは表示されない。未納者一覧と既存未納残高カードの金額整合は、別経路で未納会計（status=waiting）を準備できる場合に限り確認する |
| 9 | 手順 6 の 2 件目へ未納者一覧から遷移しようとする | 2 件目は作成されず未納者一覧にも存在しないため、残額精算・消し込みの受入確認は現行 UI では BLOCKED |
| 10 | 対象日の該当区分（1 件目の完了時刻が属する区分）のレジ締めを確定した後、1 件目のクレジット訂正を理由つきで再実行 | 訂正は成功し、締め済み期間への訂正（締め後訂正）として区別記録される（確認観点参照） |

## 確認観点

- **監査証跡（fail-closed）**: クレジット訂正は監査ログと同一トランザクションで記録され、監査記録に失敗すると訂正ごとロールバックされる（`backend/internal/billing/accounting_service_correction.go` / SOLO-03）。#10 の締め後訂正は監査メタデータに `post_close: true` と対象日が記録される（#189）。
- **訂正対象の限定**: 訂正できるのはカード系（`credit_card`・`electronic_money`）の内訳のみ。現金の誤りは別フロー（お釣り上書き #188）の管轄で、本訂正機能の対象外。
- **権限の分離**: 訂正導線は「確定済み（completed）かつ accounting-post-close-edit の edit 保有」でのみ表示される（BE: `POST /accountings/:id/credit-correction` が同権限を要求）。
- **未収金の定義整合**: 未納残高カードの金額は未納者一覧（飼主単位）・月次繰越と同一の残高定義（status=waiting の合計 — `get-owner-unpaid-balance.ts`）。ただし現行 UI から部分入金会計を作成できないため、#7〜#9 の統合確認は BLOCKED。
- **拠点横断時の訂正対象**: 訂正リクエストはグローバル選択クリニックではなく会計自体のクリニックに対して送られる（`frontend/src/features/accounting/api/correct-credit-payment.ts`、X-Clinic-ID 明示指定）。拠点横断で開いた会計でも誤テナントへ書き込まれないこと。
- **原子 complete（BUG-018）**: 新規精算は `completeAccounting` → `POST /v1/accountings/complete`。明細・支払・監査を同一 TX で確定し、Idempotency-Key で再送を安全化する。#1 の成功時に一覧が「精算済」になること。
- **回帰確認（#197）**: 現金入金額 NULL の非対称による集計ずれは修正済みだが、#6〜#9 の部分入金統合経路は現行 UI では実行できない。

## 異常系

- **理由なし訂正**: #2 で検証（クライアント側の必須制御に加え、API も理由必須のためすり抜け不可）。
- **締め済み期間の通常編集**: 締め済み期間の会計を通常の編集で再保存する場合も修正理由（`post_close_reason`）の入力（必須）が要求され、監査ログに記録される（仕様正本 11 §2.1・#115）。警告表示の検証は S09 #8 で行う。

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - 精算経路を `POST /accountings/complete`（BUG-018 原子 complete）として明記
  - 部分入金拒否を FE `remaining !== 0` + BE `validatePaymentSplits` で再確認
  - クレジット訂正 API パス・カード系 method・未納定義（waiting）を現行実装に合わせて整理
  - 監査 fail-closed / X-Clinic-ID 拠点横断は変更なし

- runtime 2026-08-07: **PARTIAL** — auth OK; Playwright accounting list/unpaid tab/reports PASS; `GET /accountings` 200 total=163. Partial-payment UI still expected disabled (`remaining !== 0`) — full correction/complete path not walked end-to-end this session. Kana search e2e for「Iris」2 cases FAIL (seed/UI locator drift, not auth).
