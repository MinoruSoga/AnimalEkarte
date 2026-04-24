# FE-250: 割引入力フォームを `discount` 権限で disabled 制御（5 画面）

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: owners / medical-records / hospitalization / estimates / accounting の各フォーム
**Date Created**: 2026-04-14
**Related**: BUG-372, BE-112

## Summary

5 画面の割引入力欄（値引率・値引額）を `usePermission("discount")` で `disabled` 制御。権限なし時は値表示のまま編集不可。

## 現状のコード

### 飼主 - 値引率（権限制御なし）
```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:425-440
<div className="space-y-1.5">
  <Label htmlFor="discountRate" className={`text-sm ${C.text60}`}>値引率 (%)</Label>
  <NumberInput
    id="discountRate"
    min={0}
    max={100}
    step={1}
    value={ownerData.discountRate || ""}
    aria-invalid={!!fieldErrors.discountRate}
    aria-describedby={fieldErrors.discountRate ? "discountRate-error" : undefined}
    onChange={(v) => { onChange("discountRate", Number(v)); onClearError("discountRate"); }}
    suffix="%"
    className={`${STYLE.formInput} ${fieldErrors.discountRate ? STYLE.formInputError : ""}`}
  />
  <FormFieldError id="discountRate-error" message={fieldErrors.discountRate} />
</div>
```

### 既存 usePermission パターン
```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:30
import { usePermission } from "@/features/auth";

// 既存使用例
const { canEdit, canCreate } = usePermission("owners");
```

## 必要な変更

### 1. 飼主フォーム

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx

// 新規 hook 呼び出し追加（既存 usePermission("owners") の隣）
const { canEdit: canEditDiscount } = usePermission("discount");

// NumberInput に disabled 追加
<NumberInput
  id="discountRate"
  min={0}
  max={100}
  step={1}
  value={ownerData.discountRate || ""}
  disabled={!canEditDiscount}                 // ★ 追加
  aria-invalid={!!fieldErrors.discountRate}
  aria-describedby={
    fieldErrors.discountRate ? "discountRate-error" :
    !canEditDiscount ? "discountRate-permission" : undefined
  }
  onChange={(v) => { onChange("discountRate", Number(v)); onClearError("discountRate"); }}
  suffix="%"
  className={`${STYLE.formInput} ${fieldErrors.discountRate ? STYLE.formInputError : ""}`}
/>
{!canEditDiscount ? (
  <p id="discountRate-permission" className={`text-xs ${C.text50}`}>
    値引率の変更には権限が必要です
  </p>
) : null}
```

### 2. 治療（明細）フォーム

`frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx` 等で `discount_rate` / `discount_amount` を編集している箇所すべてに同パターン適用。

```typescript
const { canEdit: canEditDiscount } = usePermission("discount");
// 該当 Input に disabled={!canEditDiscount} 追加
```

### 3. 入院フォーム

`frontend/src/features/hospitalization/` 配下で割引入力欄を持つコンポーネントを特定し、`disabled={!canEditDiscount}` 追加。

### 4. 見積フォーム

`frontend/src/features/estimates/` 配下の見積本体・明細の割引入力欄に同様に追加。

### 5. 会計フォーム

`frontend/src/features/accounting/routes/AccountingDetail.tsx` の支払割引入力欄に同様に追加。

## UI 操作フロー

### 権限あり (`discount:edit=true`)
1. フォームを開く
2. 割引入力欄が通常通り編集可能
3. 値変更して保存 → 200 OK で更新

### 権限なし
1. フォームを開く
2. 割引入力欄に既存値が表示されているが `disabled` 状態
3. 入力欄下に「値引率の変更には権限が必要です」と表示
4. 他のフィールド（住所等）を編集して保存 → 通常通り 200 OK
5. ブラウザの DevTools 等で disabled を外して値変更 → BE-112 の権限チェックで 403

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（feature 間 import なし）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（既存 form 動作変更なし）
- [ ] 型は `models.ts` から導出（`Resource` enum を使用）
- [ ] デザイントークン `C`, `STYLE` 使用
- [ ] `useCallback` で onChange 安定化済み（既存パターン踏襲）

## 依存関係

- BE-112 が先に完了している必要がある（`ResourceDiscount` が `models.ts` に追加される）
- `make codegen` 完了後に `usePermission("discount")` が型エラーなく呼べる

## 完了条件

- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] AC-11〜AC-13（BUG-372 参照）すべて達成
- [ ] 5 画面すべての割引入力欄に `disabled` 制御適用
- [ ] 権限あり時は従来通り編集可能
- [ ] 権限なし時の保存（割引以外の項目変更）が成功
- [ ] ブラウザ DevTools で disabled を外してリクエスト送信した場合、BE で 403 が返ることを確認
