# BUG-182: Amber 色系ハードコードと危険バッジの Tailwind 直接使用（InventoryList・TrimmingList・OwnersList）

## 概要

`InventoryList`、`TrimmingList`、`OwnersList` の 3 ファイルで `text-amber-600`・`bg-amber-50`・`border-amber-200`・`bg-red-100 text-red-700 border-red-300` 等を直接 Tailwind クラスとして使用している。Amber 色はデザイントークン (`C.*` / `BADGE.*`) に未定義のため、既存トークンで近似するか、新規トークンを追加する必要がある。

## 再現手順

1. 在庫一覧（`/inventory`）を開く → 警告バナー（在庫切れ・残少）の色を確認
2. トリミング一覧（`/trimming`）を開く → アラートアイコンの色を確認
3. オーナー一覧（`/owners`）を開く → ペットの「危険」バッジを確認
4. **結果**: 各所で `text-amber-*`・`bg-amber-*`・`text-red-700`・`bg-red-100` 等がハードコード使用

## 期待する動作

- 警告（Amber/Yellow 系): `BADGE.yellow`（`C.bgStatusYellow` / `C.bgStatusYellowDot`）を使用
- 危険（Red 系バッジ): `BADGE.red` を使用

## 現状コード

### `frontend/src/features/inventory/routes/InventoryList.tsx:269-278`
```tsx
// ❌ 警告バナー — Amber ハードコード
<div className="flex items-center gap-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
  <AlertTriangle className={`${ICON.page} text-amber-600`} />
  <span className="text-red-600 font-medium">在庫切れ: {summary.outOfStock}件</span>
  <span className="text-amber-600 font-medium">残少: {summary.lowStock}件</span>
</div>
```

### `frontend/src/features/trimming/routes/TrimmingList.tsx:86`
```tsx
// ❌ 削除確認アラートアイコン — Amber ハードコード
<AlertTriangle className={`${ICON.action} text-amber-500`} />
```

### `frontend/src/features/owners/routes/OwnersList.tsx:333`
```tsx
// ❌ ペットの危険度「高」バッジ — Red ハードコード
<span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold bg-red-100 text-red-700 border border-red-300">
  ⚠ 危険
</span>
```

### 比較: 正しい実装
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// ✅ 警告バナー（Yellow/Amber 近似）
<div style={{ backgroundColor: C.bgStatusYellow, borderColor: C.bgStatusYellowDot }}>
  <AlertTriangle style={{ color: C.bgStatusYellowDot }} />
  <span style={{ color: C.bgDanger }}>在庫切れ: {summary.outOfStock}件</span>
  <span style={{ color: C.bgStatusYellowDot }}>残少: {summary.lowStock}件</span>
</div>

// ✅ 危険バッジ
<span style={BADGE.red}>⚠ 危険</span>
```

## 影響範囲

| 対象ファイル | 行番号 | 色 | 状態 |
|---|---|---|---|
| `features/inventory/routes/InventoryList.tsx` | 269, 273, 278 | bg-amber-50, border-amber-200, text-amber-600, text-red-600 | 未修正 |
| `features/trimming/routes/TrimmingList.tsx` | 86 | text-amber-500 | 未修正 |
| `features/owners/routes/OwnersList.tsx` | 333 | bg-red-100, text-red-700, border-red-300 | 未修正 |

## 修正方針

### 1. `InventoryList.tsx:269-278` — Amber → Yellow トークン
```tsx
import { C, BADGE } from '@/lib/design-tokens';

<div
  style={{
    backgroundColor: C.bgStatusYellow,
    borderColor: C.bgStatusYellowDot,
  }}
  className="flex items-center gap-4 p-3 rounded-lg border"
>
  <AlertTriangle style={{ color: C.bgStatusYellowDot }} className={ICON.page} />
  <span style={{ color: C.bgDanger }} className="font-medium">
    在庫切れ: {summary.outOfStock}件
  </span>
  <span style={{ color: C.bgStatusYellowDot }} className="font-medium">
    残少: {summary.lowStock}件
  </span>
</div>
```

### 2. `TrimmingList.tsx:86` — Amber → Yellow トークン
```tsx
import { C } from '@/lib/design-tokens';

<AlertTriangle style={{ color: C.bgStatusYellowDot }} className={ICON.action} />
```

### 3. `OwnersList.tsx:333` — Red バッジ → BADGE.red トークン
```tsx
import { BADGE } from '@/lib/design-tokens';

<span
  style={BADGE.red}
  className="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold"
>
  ⚠ 危険
</span>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

Amber 色が `C.*` に直接定義されていない場合は `C.bgStatusYellow` / `C.bgStatusYellowDot` で代替するか、新規に `C.bgStatusAmber` トークンを追加すること。

### プロジェクト内参照実装
- `utils/status-helpers.ts` — `BADGE.yellow` / `BADGE.red` の正しい使用例

## 優先度
**Medium** — 機能的な問題はないが、在庫警告バナーは重要な業務情報表示であり、色の一貫性は信頼性に関わる。

## 関連チケット
- BUG-162: 複数 feature のハードコードカラー違反
- BUG-173: エラーメッセージの text-red-600 ハードコード

## 関連ファイル
- `frontend/src/features/inventory/routes/InventoryList.tsx`
- `frontend/src/features/trimming/routes/TrimmingList.tsx`
- `frontend/src/features/owners/routes/OwnersList.tsx`
- `frontend/src/lib/design-tokens.ts` — BADGE, C トークン定義
