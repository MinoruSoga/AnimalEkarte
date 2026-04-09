# BUG-169: 必須フィールドマーカー `*` に text-red-500 ハードコード（10+ファイル）

## 概要

フォームの必須フィールドマーカー（`*`）に `text-red-500` / `text-red-600` を直接 Tailwind クラスとして指定しているケースが 10 ファイル以上に存在する。プロジェクト規約では `C.textRequired`（`text-[#E03E3E]`）の使用が必須。バラバラな red 系クラスを使うと将来の色変更時に全ファイルを手動修正する必要が生じる。

## 再現手順

1. フォームページ（在庫・ワクチン・見積・入院等）を開く
2. 必須フィールドのラベル横の `*` を確認
3. **結果**: `text-red-500` / `text-red-600` クラスで赤くなっているが、`C.textRequired` トークンを使っていない

## 期待する動作

- 必須フィールドマーカーはすべて `style={{ color: C.textRequired }}` または `className="text-[#E03E3E]"` を使用する
- 好ましくは `<span style={{ color: C.textRequired }}>*</span>` のような共通パターン

## 現状コード

### `frontend/src/features/inventory/routes/InventoryForm.tsx:65,78,101,146,161`
```tsx
// ❌ ハードコード（5箇所）
<span className="text-red-500">*</span>
```

### `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:196,208`
```tsx
// ❌ ハードコード（2箇所）
<span className="text-red-500">*</span>
```

### `frontend/src/features/estimates/routes/EstimateForm.tsx:63`
```tsx
// ❌ ハードコード
<span className="text-red-500">*</span>
```

### `frontend/src/features/hospitalization/` — 複数ファイル
```tsx
// ❌ ハードコード（複数箇所）
<span className="text-red-500">*</span>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ トークン使用
<span style={{ color: C.textRequired }}>*</span>

// または共通コンポーネント化
// components/shared/RequiredMark/RequiredMark.tsx
export function RequiredMark() {
  return <span style={{ color: C.textRequired }} aria-hidden="true">*</span>;
}
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 状態 |
|---|---|---|
| `features/inventory/routes/InventoryForm.tsx` | 5箇所 (L65,78,101,146,161) | 未修正 |
| `features/vaccinations/routes/VaccinationForm.tsx` | 2箇所 (L196,208) | 未修正 |
| `features/estimates/routes/EstimateForm.tsx` | 1箇所 (L63) | 未修正 |
| `features/hospitalization/routes/` 配下複数ファイル | 複数箇所 | 未修正 |
| その他フォームファイル（要調査） | 未調査 | 未修正 |

## 修正方針

### Option A: 直接置換（最小変更）
各ファイルで以下を置換する:
```tsx
// Before
<span className="text-red-500">*</span>

// After
import { C } from '@/lib/design-tokens';
<span style={{ color: C.textRequired }}>*</span>
```

### Option B: 共通コンポーネント化（推奨・再利用性高）
```tsx
// components/shared/RequiredMark/RequiredMark.tsx (新規作成)
import { C } from '@/lib/design-tokens';

export function RequiredMark() {
  return (
    <span
      style={{ color: C.textRequired }}
      aria-hidden="true"
      className="ml-0.5"
    >
      *
    </span>
  );
}

// 各フォームで
import { RequiredMark } from '@/components/shared/RequiredMark/RequiredMark';

<label>ラベル<RequiredMark /></label>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

`text-red-500` は Tailwind プリセットカラーであり、`C.textRequired` (#E03E3E) と微妙に異なる可能性がある。トークンに統一することで将来の色変更に一箇所の修正で対応できる。

### プロジェクト内参照実装
- `features/medical-records/routes/` — `C.textRequired` を正しく使用している箇所（存在する場合）

## 優先度
**Medium** — UX の見た目に大きな差はないが、デザイン整合性のために修正が必要。10+ ファイルに渡る規約違反。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反
- BUG-168: 共有コンポーネントのハードコードカラー違反

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — `C.textRequired` 定義
- `frontend/src/features/inventory/routes/InventoryForm.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`
- `frontend/src/features/estimates/routes/EstimateForm.tsx`
