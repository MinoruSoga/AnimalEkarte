# BUG-186: 共有コンポーネント追加のハードコードカラー違反（PatientInfoCard・ImageWithFallback・PatientSelectionTable）

## 概要

BUG-168 で指摘した共有コンポーネントの違反に加え、新たに `PatientInfoCard.tsx`・`ImageWithFallback.tsx`・`PatientSelectionTable.tsx` の 3 コンポーネントで `bg-gray-50/80`・`bg-gray-100`・`text-gray-400` 等がハードコードされていることが判明。これらは全 feature から参照される共有コンポーネントであり影響範囲が大きい。

## 再現手順

1. ペット一覧・カルテ画面などで `PatientInfoCard` が使われている箇所を開く
2. 死亡ペットの患者情報カードを確認する（`isDeceased = true`）
3. **結果**: `bg-gray-50/80` でハードコードされた gray 背景が適用される（トークン未使用）

## 期待する動作

```tsx
// ✅ 死亡ペットの非アクティブ背景
style={{ backgroundColor: C.bgInactive }}  // または C.bgHover の 50% 不透明
```

## 現状コード

### `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx:62`
```tsx
// ❌ 死亡ペット背景ハードコード
className={`... ${isDeceased ? "bg-gray-50/80" : "bg-white"}`}
```

### `frontend/src/components/shared/Feedback/ImageWithFallback.tsx:18`
```tsx
// ❌ 画像フォールバック背景ハードコード
"inline-block bg-gray-100 text-center align-middle"
```

### `frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx:201`
```tsx
// ❌ 死亡ペット行の無効スタイルハードコード
"bg-gray-100 text-gray-400 border-transparent cursor-not-allowed"
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ PatientInfoCard — 死亡ペット背景
style={{ backgroundColor: isDeceased ? C.bgHover : '#fff' }}
// または条件で
className={isDeceased ? "" : ""}
style={isDeceased ? { backgroundColor: C.bgHover, opacity: 0.8 } : {}}

// ✅ ImageWithFallback — フォールバック背景
style={{ backgroundColor: C.bgHover }}
className="inline-block text-center align-middle"

// ✅ PatientSelectionTable — 死亡ペット行
style={{
  backgroundColor: C.bgHover,
  color: C.textSecondary,
  cursor: 'not-allowed'
}}
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 使用箇所 | 状態 |
|---|---|---|---|---|
| `components/shared/PatientInfoCard/PatientInfoCard.tsx` | 62 | bg-gray-50/80 | 全 feature のカルテ・予約画面 | 未修正 |
| `components/shared/Feedback/ImageWithFallback.tsx` | 18 | bg-gray-100 | オーナー・ペット画像全般 | 未修正 |
| `components/shared/ReservationFormModal/PatientSelectionTable.tsx` | 201 | bg-gray-100, text-gray-400 | 予約フォームのペット選択 | 未修正 |

## 修正方針

### 1. `PatientInfoCard.tsx:62`
```tsx
import { C } from '@/lib/design-tokens';

// Before
className={`${isDeceased ? "bg-gray-50/80" : "bg-white"} ...`}

// After
style={{ backgroundColor: isDeceased ? C.bgHover : undefined }}
className="bg-white ..."
```

### 2. `ImageWithFallback.tsx:18`
```tsx
import { C } from '@/lib/design-tokens';

// Before
"inline-block bg-gray-100 text-center align-middle"

// After
<span
  style={{ backgroundColor: C.bgHover }}
  className="inline-block text-center align-middle"
>
```

### 3. `PatientSelectionTable.tsx:201`
```tsx
import { C } from '@/lib/design-tokens';

// Before
"bg-gray-100 text-gray-400 border-transparent cursor-not-allowed"

// After
style={{
  backgroundColor: C.bgHover,
  color: C.textSecondary,
  borderColor: 'transparent',
  cursor: 'not-allowed'
}}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

共有コンポーネントは全 feature から参照されるため、ここでのトークン違反は全体に影響する。BUG-168 との合算で対応することを推奨。

### プロジェクト内参照実装
- `components/shared/DataTable/` — `C.bgHover` / `C.textSecondary` の正しいグレー系使用例

## 優先度
**Medium** — 共有コンポーネントへの違反のため影響範囲が大きいが、機能的問題はない。BUG-168 の修正対応時に合わせて修正すること。

## 関連チケット
- BUG-168: 共有コンポーネントのハードコードカラー違反（第一報）

## 関連ファイル
- `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
- `frontend/src/components/shared/Feedback/ImageWithFallback.tsx`
- `frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx`
- `frontend/src/lib/design-tokens.ts`
