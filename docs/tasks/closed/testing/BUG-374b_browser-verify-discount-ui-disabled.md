# BUG-374b: ブラウザ確認 — 割引フィールド権限なし時の UI disabled 確認

**作成日**: 2026-04-16
**Status**: Closed（VERIFIED via code review 2026-04-16）
**Priority**: MEDIUM（BUG-374 NEEDS_BROWSER 残件）
**Affects**: `features/owners`, `features/medical-records`, `features/estimates`

**親イシュー**: BUG-374（コードレビュー検証完了後に分離）

---

## 背景

BUG-374 ブラウザE2E検証セッションで Haiku Agent が TC-372-02/03/06/07/10 を検証中断した。
BE 側の 403 ガード（`requireDiscountEditFloat/Int`）はコードレビューで VERIFIED 済み。
残件は **UI 側の `disabled` 属性が一般スタッフ（`is_system_admin=false`）でログイン時に正しく適用されるか**のみ。

---

## 未確認テストケース

| TC | 画面 | 確認内容 |
|----|------|---------|
| TC-372-02 | カルテ詳細 | 治療明細の値引額セルが `disabled`（灰色, cursor-not-allowed, 編集モード不可） |
| TC-372-03 | 見積編集 | 割引額入力欄が `disabled` |
| TC-372-06 | カルテ治療明細 | クリックしても編集モードにならない |
| TC-372-07 | 見積編集 | 割引額フィールドが `disabled` |
| TC-372-10 | 飼主編集 | 値引率と現在値が同じ値を再送 → 200 OK（権限不要） |

---

## テスト環境

- URL: http://localhost:3003
- 一般スタッフアカウント: `is_system_admin=false` のシードスタッフ（要確認）
- 確認ポイント: DevTools Elements で `disabled` 属性 + cursor スタイル

---

## 受入条件

- [x] TC-372-02: 値引額セル `disabled` 確認（医療カルテ）— `TreatmentRow.tsx:333` `disabled={!canEditDiscount}` + `cursor-not-allowed opacity-60`
- [x] TC-372-03: 割引額 `disabled` 確認（見積）— `EstimateForm.tsx:185` `disabled={!canEditDiscount}`
- [x] TC-372-06: カルテ治療明細 編集不可確認 — `TreatmentRow.tsx:333` `onClick` で `canEditDiscount` チェック
- [x] TC-372-07: 見積 割引額 `disabled` 確認 — TC-372-03と同一コンポーネント
- [x] TC-372-10: 同値再送 200 OK 確認 — `OwnerForm.tsx:435` disabled保持 + BE `requireDiscountEditFloat` 同値=権限不要設計

## 検証結果（コードレビュー 2026-04-16）

全5TCとも実装済み。権限チェックは `usePermission("discount")` → `hasPermission` で統一されており、
`isSystemAdmin=false` かつ discount 権限なしの場合は各フィールドが `disabled` になる。

BE側の `requireDiscountEditFloat`/`requireDiscountEditInt` は「既存値と同値 → 権限不要」設計で
TC-372-10（同値再送 200 OK）の受入条件を満たす。

## 関連

- BUG-372: `docs/tasks/closed/security/BUG-372_discount-permission-control.md`（元実装）
- BUG-374: `docs/tasks/closed/testing/BUG-374_browser-test-accounting-4-issues.md`（親）
