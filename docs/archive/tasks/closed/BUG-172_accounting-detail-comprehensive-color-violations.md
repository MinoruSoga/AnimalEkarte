# BUG-172: AccountingDetail.tsx の包括的ハードコードカラー違反（~20箇所）

## 概要

`features/accounting/routes/AccountingDetail.tsx` で Tailwind プリセットカラーが約 20 箇所ハードコードされている。タブナビゲーションの境界線色、バッジの背景色、金額テキストの条件付き色変更など、多岐にわたる。すべてが `C.*` / `STYLE.*` / `BADGE.*` トークンに置換されるべき。

## 再現手順

1. 会計詳細画面（`/accounting/:id`）を開く
2. タブナビゲーション、バッジ、金額表示エリアを確認
3. **結果**: `border-blue-600 text-blue-600`（アクティブタブ）、`text-green-700 bg-green-50`（合計額）、`text-red-500`（割引・マイナス値）等のハードコード色が使われている

## 期待する動作

- アクティブタブの border/text 色: `C.bgAccent` トークンを使用
- バッジ色: `BADGE.blue` / `BADGE.green` 等のトークンを使用
- 金額の条件付き色: `C.bgDanger` / `C.bgStatusGreenDot` 等を使用

## 現状コード

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:147`
```tsx
// ❌ ハードコードバッジ
<span className="text-[10px] text-blue-500 bg-blue-50 px-1.5 py-0.5 rounded">
  カルテ連携
</span>
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:229-241`
```tsx
// ❌ タブナビゲーションの border/text ハードコード
className={cn(
  "pb-2 px-1 text-sm",
  activeTab === tab.id
    ? "border-b-2 border-blue-600 text-blue-600 font-medium"
    : "text-gray-500 hover:text-gray-700"
)}
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:440`
```tsx
// ❌ 金額色ハードコード
<span className={cn(
  totalAmount < 0 ? "text-red-500" : "text-gray-900"
)}>
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:582`
```tsx
// ❌ 条件付き色ハードコード
className={amount < 0 ? "text-red-500" : "text-gray-900"}
```

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:440付近`
```tsx
// ❌ 合計エリアのハードコード
<div className="text-green-700 bg-green-50 ...">
```

### 比較: 正しい実装
```tsx
import { C, STYLE, BADGE } from '@/lib/design-tokens';

// ✅ バッジ
<span style={BADGE.blue} className="text-[10px] px-1.5 py-0.5 rounded">
  カルテ連携
</span>

// ✅ アクティブタブ
style={activeTab === tab.id ? { borderBottomColor: C.bgAccent, color: C.bgAccent } : {}}

// ✅ 金額条件付き色
style={{ color: amount < 0 ? C.bgDanger : C.textMain }}

// ✅ 合計エリア
style={{ color: C.bgStatusGreenDot, backgroundColor: C.bgStatusGreen }}
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 状態 |
|---|---|---|
| `features/accounting/routes/AccountingDetail.tsx` | ~20箇所 (L147, L229-241, L440, L582, 他) | 未修正 |
| `features/accounting/routes/AccountingDocument.tsx` | 別チケット BUG-174 参照 | 未修正 |

## 修正方針

### 1. バッジ系 (L147) → BADGE トークン
```tsx
import { BADGE } from '@/lib/design-tokens';
<span style={BADGE.blue} className="text-[10px] px-1.5 py-0.5 rounded">カルテ連携</span>
```

### 2. タブナビゲーション (L229-241) → C.bgAccent
```tsx
import { C } from '@/lib/design-tokens';

const tabStyle = activeTab === tab.id
  ? { borderBottomWidth: '2px', borderBottomColor: C.bgAccent, color: C.bgAccent }
  : { color: C.textSecondary };

<button style={tabStyle} className="pb-2 px-1 text-sm font-medium">
```

### 3. 金額条件付き色 (L440, L582) → C.bgDanger / C.textMain
```tsx
import { C } from '@/lib/design-tokens';

<span style={{ color: amount < 0 ? C.bgDanger : C.textMain }}>
  {formatAmount(amount)}
</span>
```

### 4. 合計エリア → C.bgStatusGreen / C.bgStatusGreenDot
```tsx
import { C } from '@/lib/design-tokens';

<div style={{ color: C.bgStatusGreenDot, backgroundColor: C.bgStatusGreen }}>
  合計: {total}
</div>
```

### 5. bg-gray-50/100, text-gray-500/600 → C.bgHover / C.textSecondary
```tsx
style={{ backgroundColor: C.bgHover }}
style={{ color: C.textSecondary }}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

### プロジェクト内参照実装
- `features/accounting/routes/AccountingList.tsx` — 修正済み箇所を参照

## 優先度
**Medium** — 会計詳細は業務上重要な画面であり、色の一貫性が信頼性に影響する。20箇所と多いが機能的な問題はない。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反（accounting も含む）
- BUG-174: AccountingDocument.tsx 印刷ビューの色違反

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/lib/design-tokens.ts` — C, BADGE トークン定義
