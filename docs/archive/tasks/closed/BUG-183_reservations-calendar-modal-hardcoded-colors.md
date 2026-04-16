# BUG-183: 予約カレンダー（MonthView）・予約詳細モーダルのハードコードカラー違反

## 概要

`features/reservations/` 配下の `MonthView.tsx` と `ReservationDetailModal.tsx` で Tailwind プリセットカラーが大量使用されている。曜日の色分け（日曜 red / 土曜 blue）、当日強調（blue）、初診・再診バッジ（red/blue）、指名バッジ（amber）がすべてハードコード。BUG-162 で reservations の一部が指摘済みだが、カレンダービュー固有の違反として追加起票。

## 再現手順

1. 予約管理画面（`/reservations`）を開き、月表示に切り替える
2. 曜日ヘッダー・今日の日付・予約カードの初診/再診バッジを確認する
3. 予約をクリックして詳細モーダルを表示する
4. **結果**: 日曜=赤・土曜=青・当日=青・初診=赤・再診=青・指名=amber がすべて Tailwind プリセットでハードコード

## 期待する動作

- 曜日色・バッジ色はすべて `C.*` / `BADGE.*` トークンを使用する
- 訪問タイプ（初診/再診）の色は `status-helpers.ts` や定数マップとして定義する

## 現状コード

### `frontend/src/features/reservations/components/MonthView.tsx:39`
```tsx
// ❌ 曜日色ハードコード
i === 0 ? "text-red-500" : i === 6 ? "text-blue-500" : C.text60
```

### `frontend/src/features/reservations/components/MonthView.tsx:73`
```tsx
// ❌ 当日ハイライトハードコード
"bg-blue-50/30"
```

### `frontend/src/features/reservations/components/MonthView.tsx:79`
```tsx
// ❌ 選択日・ホバー色ハードコード
"bg-blue-600 text-white shadow-sm" : "hover:bg-blue-100 hover:text-blue-700"
```

### `frontend/src/features/reservations/components/MonthView.tsx:105-106`
```tsx
// ❌ 初診・再診バッジハードコード
"bg-red-100 text-red-600"   // 初診
"bg-blue-100 text-blue-600" // 再診
```

### `frontend/src/features/reservations/components/ReservationDetailModal.tsx:177`
```tsx
// ❌ 指名バッジ — Amber ハードコード
"bg-amber-50 text-amber-700 border-amber-200"
```

### `frontend/src/features/reservations/components/ReservationDetailModal.tsx:195-196`
```tsx
// ❌ メモセクション枠線・アイコン — Amber ハードコード
"border border-amber-100 bg-amber-50/50"
"text-amber-700"
```

### 比較: 正しい実装
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// ✅ 曜日色
const getDayColor = (dayIndex: number): string => {
  if (dayIndex === 0) return `text-[${C.bgDanger}]`;   // 日曜
  if (dayIndex === 6) return `text-[${C.bgAccent}]`;   // 土曜
  return C.text60;
};

// ✅ 当日ハイライト
style={{ backgroundColor: `${C.bgAccent}1A` }}  // 10% 不透明度

// ✅ 初診・再診バッジ
const visitTypeStyle = isFirstVisit ? BADGE.red : BADGE.blue;
<span style={visitTypeStyle}>初診</span>
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 状態 |
|---|---|---|
| `features/reservations/components/MonthView.tsx` | 5箇所 (L39, L73, L79, L105, L106) | 未修正 |
| `features/reservations/components/ReservationDetailModal.tsx` | 3箇所 (L177, L195, L196) | 未修正 |

## 修正方針

### 1. `MonthView.tsx` 曜日・日付色をトークンに置換
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// 曜日色
const dayLabelStyle = (i: number): React.CSSProperties => {
  if (i === 0) return { color: C.bgDanger };
  if (i === 6) return { color: C.bgAccent };
  return {};
};

// 当日ハイライト
style={{ backgroundColor: `${C.bgAccent}1A` }}  // #2383E2 の 10% 不透明

// 選択日ボタン
style={isSelected ? { backgroundColor: C.bgAccent, color: '#fff' } : {}}
className="hover:bg-[${C.bgAccent}]/10"

// 初診・再診バッジ
<span style={isFirstVisit ? BADGE.red : BADGE.blue} className="text-xs px-1 rounded">
```

### 2. `ReservationDetailModal.tsx` 指名・メモ色をトークンに置換
```tsx
// 指名バッジ → BADGE.yellow 近似
<span style={BADGE.yellow} className="border rounded px-2 py-0.5 text-xs">
  指名
</span>

// メモセクション枠線 → C.bgStatusYellow 系
<div style={{ backgroundColor: C.bgStatusYellow, borderColor: C.bgStatusYellowDot }} className="border rounded p-3">
  <MessageSquare style={{ color: C.bgStatusYellowDot }} />
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### プロジェクト内参照実装
- `features/shifts/components/ShiftCalendar/ShiftCalendar.tsx` — カレンダー曜日色の同様パターン（BUG-171 で指摘済み、合わせて修正するとよい）

## 優先度
**Medium** — 予約管理画面は日常業務の中核機能。カレンダーの色ルールが設計トークンと乖離しており、将来のテーマ変更時に乗り遅れる。BUG-171（シフトカレンダー）と合わせて一括修正が効率的。

## 関連チケット
- BUG-162: 予約 feature のその他ハードコード違反
- BUG-168: ReservationFormModal の共有コンポーネント違反
- BUG-171: ShiftCalendar の曜日色ハードコード（同パターン）
- BUG-182: Amber 色系ハードコード（指名バッジの amber と同類）

## 関連ファイル
- `frontend/src/features/reservations/components/MonthView.tsx`
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx`
- `frontend/src/lib/design-tokens.ts`
