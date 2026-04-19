# BUG-374: ブラウザテスト — 会計系 4 イシュー実装検証 (commit 8fcd1382)

**作成日**: 2026-04-14
**Status**: Partial — コードレビュー追検証完了 / ブラウザ E2E 残
**Priority**: HIGH（リリース前検証）

## 2026-04-15 追補: コードレビュー検証

ブラウザ Agent 起動前に、未確認 TC をコードレビューで検証可能な範囲を確認した結果：

| TC | 結果 | 根拠 |
|----|------|------|
| TC-367-03 | **FIXED** | `accountingOwnerSummary.OwnerName` JSON タグが `owner_name` だったため FE の `data.owner?.name` が常に undefined → 帳票で「様」のみ表示。BE タグを `name` に修正 + FE 側にフォールバック「（飼主名未取得）」を追加 |
| TC-367-04/05 | RESOLVED | seed 005 で軽減税率 8% 品目を追加済（commit `a2cec54d`）。Document 側は `approxEqual(taxRate, 0.08)` で「※」マーク表示済 (`AccountingDocument.tsx`) |
| TC-371-06 | **VERIFIED** | `AccountingDetail.tsx:494,510,606` で `canSubmit = isEditMode ? canEdit : canCreate`、`canEdit ? <input> : <readonly>`、`canSubmit ? <SubmitButton> : null` の三段ガード。権限なしは確実に編集不可 |
| TC-372-09 | **VERIFIED** | `requireDiscountEditFloat/Int` が owner/treatment/treatment_plan/estimate/accounting handler の Create/Update で呼ばれており、`discount.edit` 権限なし + 値変更ありで 403 を返す。DevTools バイパス不可 |
| TC-372-02/03/06/07/10 | NEEDS_BROWSER | UI 側の disabled 属性確認は実機検証必須 |

### コミット
- `(本コミット)`: TC-367-03 root cause 修正（BE/FE）

## 実施結果サマリ (2026-04-14)

Haiku Agent による Chrome DevTools MCP 経由の E2E 検証を 2 ラウンド実施。Agent の UI 観察と Backend API 実動作に乖離があるケースを複数発見。**Agent の NG 判定は複数が false positive。**

### 追検証済み

| Agent 報告 | 追検証結果 | 根拠 |
|-----------|-----------|------|
| TC-371-10 cancel 済への再 cancel が 404 | **REFUTED** | curl で 409 Conflict 正常返却を確認 (`"既にキャンセル済みの会計です"`) |
| TC-372-01 admin@noavet.jp で値引率 disabled | **REFUTED** | `/me` で `is_system_admin=true` + `discount.edit=true` 返却、`PATCH /owners/1 {discount_rate:5}` が 200 OK。Agent の UI セッション誤認が濃厚 |
| TC-370-11 基準日 URL 復元 | **CONFIRMED** | `UnpaidCustomerList.tsx:44-60` で `reference_date` が `searchParams` 未同期。BUG-375 として起票 |

### 未確認（追検証必要）

| TC | 内容 | メモ |
|----|------|------|
| TC-367-03 | 帳票に飼主名が「様」のみ | UI レベル確認必要（統合帳票コンポーネント） |
| TC-367-04/05 | 軽減税率 8% 表示 | シードに 8% 品目不足 → seed 追加検討 |
| TC-371-06 | 権限なしで完了済 disabled | admin@example.com での再検証必要 |
| TC-372-02/03/06/07/09/10 | 割引権限の partial 検証 | Agent 中断、再実行要 |

### 新規起票

- **BUG-376**: 未納者一覧の基準日が URL 同期されない（LOW）
  ※ BUG-375 は先行起票済み（よみ ひらがな化・検索両対応）のため採番変更

### Agent 運用の課題

- ログインセッションの切替（admin@noavet.jp ↔ admin@example.com）でキャッシュ UI を見ていた可能性
- 次回は各テスト前に `evaluate_script` で JWT デコード → `is_system_admin` ログ出力を強制すること
- clinic_id コンテキストの同期ミス（owner id=1 は clinic 3 所属だが、Agent は current_clinic_id の切替を怠った）

---


**Affects**: `features/accounting`, `features/owners`, `features/medical-records`, `features/estimates`, RBAC

**依頼元**:

> ブラウザテストタスクを起票して

---

## 概要

commit `8fcd1382 feat: BUG-367/370/371/372` で実装した会計系 4 イシューの動作を Chrome DevTools MCP 経由でブラウザ E2E 検証する。BE 単体テストはパス済み（TestAccountingService_Cancel + TestFloatEquals）だが、UI フロー・権限挙動・帳票印字等は未検証。

## テスト対象

| ID | 検証内容 | 重要度 |
|----|---------|-------|
| BUG-372 | 割引フィールド権限制御 | CRITICAL |
| BUG-367 | 明細兼領収書（インボイス対応） | CRITICAL |
| BUG-371 | 精算済会計の修正・論理削除 | HIGH |
| BUG-370 | 月末未納者一覧 | HIGH |

## 実行方法

```
/browser-test 6        # Section 6: 会計管理（メイン）
/browser-test 1        # Section 1: 飼主・ペット管理（BUG-372 値引率欄）
/browser-test 4        # Section 4: カルテ管理（BUG-372 治療値引額）
/browser-test 12       # Section 12: 見積管理（BUG-372 割引額）
```

実行は **Haiku Agent** に委譲（`docs/CLAUDE.md` の browser-test スキル定義）。

## テストケース（BUG 別）

### BUG-372: 割引フィールド権限制御

#### 権限あり（`is_system_admin=true` のスタッフでログイン）

- [ ] **TC-372-01**: 飼主編集画面 → 値引率入力欄が **編集可能**
- [ ] **TC-372-02**: カルテ詳細 → 治療明細の値引額編集が **可能**
- [ ] **TC-372-03**: 見積編集画面 → 割引額入力欄が **編集可能**
- [ ] **TC-372-04**: 値引率を 10% に変更して保存 → 200 OK・反映確認

#### 権限なし（一般スタッフ = `is_system_admin=false` でログイン）

- [ ] **TC-372-05**: 飼主編集画面 → 値引率欄が `disabled`、説明テキスト「値引率の変更には権限が必要です」表示
- [ ] **TC-372-06**: カルテ治療明細 → 値引額セルが灰色（cursor-not-allowed）、クリックしても編集モード入らない
- [ ] **TC-372-07**: 見積編集 → 割引額が `disabled`
- [ ] **TC-372-08**: 飼主編集（値引率以外を変更）→ 保存成功
- [ ] **TC-372-09**: DevTools で `disabled` 属性外し、`PATCH /owners/:id` に `discount_rate=15` 送信 → **403 Forbidden**
- [ ] **TC-372-10**: 値引率と現在値が同じ値（再送）→ 200 OK（権限不要）

### BUG-367: 明細兼領収書

- [ ] **TC-367-01**: 完了済み会計 → 「明細兼領収書」ボタン 1 個だけ表示（旧 2 個ボタン削除確認）
- [ ] **TC-367-02**: ボタンクリック → A4 サイズの統合帳票が表示
- [ ] **TC-367-03**: 帳票内容確認: タイトル・発行日・発行No・登録番号・宛名・ペット名・品目テーブル・税率列・小計・税率別内訳・請求金額・お預かり・お釣り
- [ ] **TC-367-04**: 軽減税率（8%）品目を含むケース → 品目名横に「※」マーク + 「8%対象 ※軽減税率」行表示
- [ ] **TC-367-05**: 軽減税率品目が 0 件のケース → 8% 行は非表示、10% のみ表示
- [ ] **TC-367-06**: 病院設定で `companies.invoice_registration_number` が未設定 → 画面上部に赤い警告バナー + 帳票内に「登録番号: 未設定」
- [ ] **TC-367-07**: 印刷ダイアログ → A4 サイズで印刷プレビュー（実機印刷は環境次第）

### BUG-371: 精算済会計修正・論理削除

#### 修正

- [ ] **TC-371-01**: `accounting:edit` 権限あり → 完了済み会計の明細・支払方法・受領額・保険率が **編集可能**（disabled 解除）
- [ ] **TC-371-02**: 「精算完了済み」固定ボタンが「修正を保存する」に変わる
- [ ] **TC-371-03**: 保存ボタン押下 → 確認モーダル「精算済みの会計を修正します」表示
- [ ] **TC-371-04**: モーダル「修正する」 → API 200、status は `completed` のまま、再描画
- [ ] **TC-371-05**: モーダル「キャンセル」 → 操作中止、データ未変更
- [ ] **TC-371-06**: 権限なし → 完了済みは従来通り disabled

#### 論理削除（キャンセル）

- [ ] **TC-371-07**: `accounting:delete` 権限あり → 完了済み会計のヘッダーに「会計をキャンセル」ボタン表示
- [ ] **TC-371-08**: ボタン押下 → 確認モーダル「この会計をキャンセルします」表示（destructive 赤系）
- [ ] **TC-371-09**: 「キャンセルする」 → `POST /accountings/:id/cancel` → 204 → トースト「会計をキャンセルしました」 → 一覧画面遷移
- [ ] **TC-371-10**: cancelled の会計を再度キャンセル試行 → BE が **409 Conflict**（既にキャンセル済み）
- [ ] **TC-371-11**: 旧 `DELETE /accountings/:id` を直接呼ぶ → **404**（ルート撤去確認）

### BUG-370: 月末未納者一覧

- [ ] **TC-370-01**: サイドバーに「未納者一覧」項目表示（会計管理の下）
- [ ] **TC-370-02**: クリック → `/accounting/unpaid` 遷移
- [ ] **TC-370-03**: 基準日 = 今日 がデフォルト
- [ ] **TC-370-04**: 飼主単位タブ（デフォルト）→ サマリーカード（売掛金総額・件数・飼主数）+ テーブル表示
- [ ] **TC-370-05**: 飼主名クリック → `/owners/:id` 遷移
- [ ] **TC-370-06**: 会計単位タブ切替 → URL `?group_by=billing` 同期 + テーブル表示（飼主名・ペット名・診療日・未納額・経過日数）
- [ ] **TC-370-07**: 会計行クリック → `/accounting/:id` 遷移
- [ ] **TC-370-08**: 基準日を過去日に変更 → 該当データのみ表示
- [ ] **TC-370-09**: 未納 0 件の状態 → 「未納者はいません」空状態メッセージ
- [ ] **TC-370-10**: 21 件以上のデータ → ページネーション動作
- [ ] **TC-370-11**: ブラウザリロード → URL クエリパラメータ復元（タブ + 基準日）

## テスト環境

- URL: http://localhost:3003
- アカウント: admin@example.com / password（is_system_admin=true）
- 一般スタッフテスト用: シードに `is_system_admin=false` のスタッフが必要（要確認）
- ブラウザ: Chrome (DevTools MCP)

## 結果記録

`docs/FUNCTIONAL_TEST_REPORT.md` の Section 6 (会計管理) に結果列を更新。
NG 項目は別途 BUG-XXX として `docs/tasks/open/crash/` または `docs/tasks/open/accounting/` に起票。

## 完了条件

- [ ] 上記 30+ テストケースすべて実行・記録
- [ ] OK / NG / Partial / N/A の判定明記
- [ ] NG 項目はスクリーンショット添付（`/tmp/` 保存）
- [ ] 新規バグ発見時は別 BUG として起票
- [ ] サマリレポートをセッション終了時に出力

## 参照

- 実装コミット: `8fcd1382 feat: BUG-367/370/371/372 会計系 4 イシュー対応 + 関連テスト`
- BUG-372: `docs/tasks/closed/security/BUG-372_discount-permission-control.md`
- BUG-367: `docs/tasks/closed/accounting/BUG-367_invoice-compliant-statement-receipt.md`
- BUG-371: `docs/tasks/closed/accounting/BUG-371_completed-billing-edit-and-soft-delete.md`
- BUG-370: `docs/tasks/closed/accounting/BUG-370_unpaid-customer-list-month-end.md`

## リスク・懸念事項

| リスク | 影響 | 対策 |
|--------|------|------|
| 一般スタッフ用シードが存在しない | 権限なしテストが不可 | seed 確認 → 必要なら追加 |
| 軽減税率データが seed に存在しない | TC-367-04 検証不可 | sample 商品マスタを 8% 設定 |
| Chrome DevTools MCP の印刷ダイアログ操作不可 | TC-367-07 はスクショまで | 実機検証は別途人手 |
| pending 中の BUG-368/373 と機能干渉 | 影響軽微 | 飼主変更フローは pending のまま放置 |
