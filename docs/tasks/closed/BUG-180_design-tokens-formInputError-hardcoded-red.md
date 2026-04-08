# BUG-180: `design-tokens.ts` の `STYLE.formInputError` に Tailwind red プリセット（全フォームに影響）

## 概要

`src/lib/design-tokens.ts:855` の `STYLE.formInputError` 定義内に `ring-red-300 border-red-400` という Tailwind プリセットカラーが直接ハードコードされている。このトークンは OwnerForm・PetEditModal・ExaminationForm 等の全フォームのエラー状態に使用されているため、**デザイントークンファイル自体が規約違反**という根本的な問題。

## 再現手順

1. `frontend/src/lib/design-tokens.ts:855` を開く
2. `STYLE.formInputError` の値を確認する
3. **結果**: `"ring-2 ring-red-300 border-red-400"` と Tailwind プリセットがハードコード
4. さらに OwnerForm.tsx 等で `STYLE.formInputError` を検索すると 10+ 箇所で使用されていることを確認

## 期待する動作

```typescript
// ✅ トークン経由
formInputError: `ring-2 ring-[${PALETTE.danger}]/30 border-[${PALETTE.danger}]`,
```

## 現状コード

### `frontend/src/lib/design-tokens.ts:855`
```typescript
// ❌ Tailwind プリセット直接ハードコード
formInputError: "ring-2 ring-red-300 border-red-400",
```

### 影響を受けるファイル（使用箇所）
```tsx
// OwnerForm.tsx — 7箇所 (L260, L300, L318, L366, L393, L414, L434)
className={`... ${hasError ? STYLE.formInputError : ''}`}

// PetEditModal.tsx — 複数箇所
// ExaminationForm.tsx — 複数箇所
```

### 比較: 正しい実装
```typescript
// design-tokens.ts 内の他の正しい定義
export const C = {
  bgDanger: "bg-[#C0392B]",  // PALETTE.danger 参照
};

// 修正後
formInputError: `ring-2 ring-[#C0392B]/30 border-[#C0392B]`,
// または PALETTE を使って
formInputError: `ring-2 ring-[${PALETTE.danger}]/30 border-[${PALETTE.danger}]`,
```

## 影響範囲

| 対象ファイル | 使用箇所数 | 状態 |
|---|---|---|
| `src/lib/design-tokens.ts` | 定義元 (L855) | 未修正 |
| `features/owners/routes/OwnerForm.tsx` | 7箇所 | 違反継承 |
| `features/pets/components/PetEditModal.tsx` | 複数 | 違反継承 |
| `features/examinations/routes/ExaminationForm.tsx` | 複数 | 違反継承 |
| その他 `STYLE.formInputError` 参照ファイル | 未調査 | 違反継承 |

## 修正方針

### `design-tokens.ts:855` — PALETTE 参照に変更

```typescript
// Before
formInputError: "ring-2 ring-red-300 border-red-400",

// After
formInputError: `ring-2 ring-[${PALETTE.danger}]/30 border-[${PALETTE.danger}]`,
```

**注意**: `PALETTE.danger` の値が `#C0392B` であることを確認してから変更する。変更後は全フォームのエラー表示が自動的に修正される。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

`design-tokens.ts` の **定義内** でも Tailwind プリセットは禁止。`PALETTE.*` 定数を使って定義すること。

### プロジェクト内参照実装
- `design-tokens.ts` の他の `STYLE.*` 定義 — `PALETTE` 参照の正しいパターン

## 優先度
**High** — デザイントークンファイル自体の違反であり、全フォームのエラー表示スタイルに影響する。修正箇所は1行だが影響範囲が広い。

## 関連チケット
- BUG-169: 必須フィールドマーカーのハードコード
- BUG-173: エラーメッセージの text-red-600 ハードコード

## 関連ファイル
- `frontend/src/lib/design-tokens.ts:855`
- `frontend/src/features/owners/routes/OwnerForm.tsx`
