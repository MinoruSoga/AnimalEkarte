# BUG-205: VitalsTab・TrimmingForm の追加 gray 系 Tailwind 違反（補足）

## 概要

BUG-184（VitalsTab L158-159,173-174）・BUG-181/BUG-198（TrimmingForm）で報告した違反に加え、同一ファイルの別行に追加の gray 系 Tailwind プリセット違反が発見された。

## 現状コード

### `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:195,485,560`
```tsx
// ❌ L195: セルに gray ハードコード（BUG-184 の L158-159 とは別箇所）
<td className={`... border ${C.borderMedium} bg-gray-50 hover:bg-gray-100 min-w-[24px]`}>

// ❌ L485: 単位テキストに gray ハードコード
<span className="ml-0.5 text-[10px] text-gray-400">{vital.weight_unit}</span>

// ❌ L560: セルに gray ハードコード（L195 と同パターン）
<td className={`... border ${C.borderMedium} bg-gray-50 hover:bg-gray-100 min-w-[24px]`}>
```

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:173,286`
```tsx
// ❌ L173: 画像削除ボタンの hover 状態に gray ハードコード
<button className="absolute top-1 right-1 p-1 bg-white rounded-full shadow-sm hover:bg-gray-100">

// ❌ L286: 完了済み画像削除ボタンの hover 状態に gray ハードコード（同パターン）
<button className="absolute top-1 right-1 p-1 bg-white rounded-full shadow-sm hover:bg-gray-100">
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 195, 560 | bg-gray-50, hover:bg-gray-100 | 未修正 |
| `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 485 | text-gray-400 | 未修正 |
| `features/trimming/routes/TrimmingForm.tsx` | 173, 286 | hover:bg-gray-100 | 未修正 |

## 修正方針

```tsx
import { C } from '@/lib/design-tokens';

// VitalsTab.tsx L195,560
<td style={{ backgroundColor: C.bgPage }} className={`... border ... min-w-[24px] hover:bg-[${C.bgHover}]`}>
// または onMouseEnter/Leave で style 切り替え

// VitalsTab.tsx L485
<span style={{ color: C.textSecondary }} className="ml-0.5 text-[10px]">
  {vital.weight_unit}
</span>

// TrimmingForm.tsx L173,286
<button
  style={{}}
  className="absolute top-1 right-1 p-1 bg-white rounded-full shadow-sm"
  onMouseEnter={e => { e.currentTarget.style.backgroundColor = C.bgHover; }}
  onMouseLeave={e => { e.currentTarget.style.backgroundColor = ''; }}
>
// またはTailwind 4の CSS変数記法:
// hover:bg-[var(--color-bg-hover)]
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Low** — 既存の BUG-184（VitalsTab）・BUG-198（TrimmingForm）の対応時に合わせて修正すること。

## 関連チケット
- BUG-184: VitalsTab L158-159,173-174（同ファイルの先行チケット）
- BUG-198: TrimmingForm loading UI（同ファイルの先行チケット）

## 関連ファイル
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
