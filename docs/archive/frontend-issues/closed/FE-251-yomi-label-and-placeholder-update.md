# FE-251: 飼主・ペット「よみ」UI ラベル統一（カナ → よみ + placeholder ひらがな化）

**Status**: Closed (2026-04-14)
**Priority**: Medium
**Affects**: 飼主編集 / ペット編集 / 検索フォーム / 予約フォーム
**Date Created**: 2026-04-14
**Related**: BUG-375, BE-113, BE-114

## Summary

UI ラベル「飼主名(カナ)」「ペット名(カナ)」を **「飼主名よみ」「ペット名よみ」** に変更し、
placeholder のカタカナ例（"ハヤシ フミアキ" / "イリス"）を ひらがな例 ("はやし ふみあき" / "いりす") に更新する。
バリデーション拒否は追加しない（C: 文字種制約なし）。

## 現状のコード

### 1. 飼主編集フォーム

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:361-373
<Label htmlFor="ownerNameKana" className={`text-sm ${C.text60}`}>
  飼主名(カナ) <span className={C.textRequired}>*</span>
</Label>
<Input
  id="ownerNameKana"
  placeholder="ハヤシ フミアキ"
  value={ownerData.ownerNameKana}
  ...
/>
```

### 2. ペット編集モーダル

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx:327-336
<Label htmlFor="petNameKana" className={LABEL_CLS}>
  ペット名(カナ)
</Label>
<Input
  id="petNameKana"
  placeholder=""  // ※要確認
  value={formData.petNameKana || ""}
  onChange={(e) =>
    setFormData(prev => ({ ...prev, petNameKana: e.target.value }))
  }
/>
```

### 3. バリデーションエラー

```typescript
// frontend/src/features/owners/hooks/use-owner-form.ts:126
if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名（カナ）を入力してください";
```

### 4. ペット選択検索フォーム

```typescript
// frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx:30, 33
{ id: "ownerNameKana", label: "飼主名(カナ)", placeholder: "例: ハヤシ フミアキ" },
{ id: "petNameKana",   label: "ペット名(カナ)", placeholder: "例: イリス" },
```

### 5. 予約フォーム患者選択

```typescript
// frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx:26, 29
{ key: "ownerNameKana", label: "飼主名(カナ)", placeholder: "例: ハヤシ" },
{ key: "petNameKana",   label: "ペット名(カナ)", placeholder: "例: イリス" },
```

## 必要な変更

### 1. 飼主編集フォーム

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:361
<Label htmlFor="ownerNameKana" className={`text-sm ${C.text60}`}>
  飼主名よみ <span className={C.textRequired}>*</span>
</Label>
<Input
  id="ownerNameKana"
  placeholder="はやし ふみあき"  // ★ 変更
  ...
/>
```

### 2. ペット編集モーダル

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx:327-336
<Label htmlFor="petNameKana" className={LABEL_CLS}>
  ペット名よみ
</Label>
<Input
  id="petNameKana"
  placeholder="いりす"  // ★ 追加 or 変更
  ...
/>
```

### 3. バリデーションエラー

```typescript
// frontend/src/features/owners/hooks/use-owner-form.ts:126
if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名よみを入力してください";
```

### 4. ペット選択検索フォーム

```typescript
// frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx
{ id: "ownerNameKana", label: "飼主名よみ", placeholder: "例: はやし ふみあき" },
{ id: "petNameKana",   label: "ペット名よみ", placeholder: "例: いりす" },
```

### 5. 予約フォーム患者選択

```typescript
// frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx
{ key: "ownerNameKana", label: "飼主名よみ", placeholder: "例: はやし" },
{ key: "petNameKana",   label: "ペット名よみ", placeholder: "例: いりす" },
```

### 6. その他検索: 確認漏れがないか grep で全件チェック

```bash
grep -rn "カナ\|ハヤシ\|イリス" frontend/src/ | grep -v node_modules
```

カラム名定義 (`ownerNameKana` / `petNameKana`) は変更しない（型互換維持）。

## UI 操作フロー

1. ユーザーが飼主新規登録画面を開く
2. ラベル「飼主名よみ」「ペット名よみ」が表示されている
3. placeholder「はやし ふみあき」「いりす」が表示
4. ひらがな・カタカナ・漢字・英数字 いずれを入力しても保存可能
5. 必須エラー文言は「飼主名よみを入力してください」

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（既存パターン踏襲）
- [ ] 型は `models.ts` から導出（変更なし）
- [ ] デザイントークン `C`, `STYLE` 使用

## 依存関係

- BE-113, BE-114 とは独立で着手可能（UI 文言変更のみ）
- BE-114 が完了していれば「ハヤシ」入力でも「はやし」DB レコードがヒット → 動作確認できる

## 完了条件

- [ ] 5 箇所すべてのラベル「カナ」→「よみ」変更（OwnerForm / PetEditModal / use-owner-form / PetSelectionSearchForm / PatientSelectionTable）
- [ ] placeholder ひらがな化（5 箇所）
- [ ] バリデーションエラー文言更新
- [ ] grep で「カナ」「ハヤシ」「イリス」残存 0 件
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし
- [ ] 既存テスト全件パス（owners 系 transforms.test.ts 等）
- [ ] 飼主新規・編集 → 保存できる
- [ ] 検索フォームでひらがな入力でカタカナ DB レコード（変換前）/ ひらがな DB レコード（変換後）双方ヒット確認
