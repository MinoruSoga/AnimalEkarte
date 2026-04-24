# FE-202: 入院管理コンポーネント群でデザイントークン未使用（Tailwind直接色指定）

## 概要

`frontend/src/features/hospitalization/` 配下の複数コンポーネントで、
`C.*` デザイントークンを使わず Tailwind の直接色クラスを大量に使用している。
プロジェクト規約「Hex カラー・Tailwind 直接色指定禁止。`C`, `STYLE` 定数を使用」に違反。

## 影響範囲

| ファイルパス | 行番号 | 違反コード | 置換方法 |
|------------|--------|-----------|---------|
| `DischargeAlertDialog.tsx` | 52 | `bg-red-600 hover:bg-red-700` | `C.bgDanger` + `C.hoverBgDanger90` |
| `CarePlanItemRow.tsx` | 62 | `border-gray-200` | `C.borderMedium` |
| `CarePlanItemRow.tsx` | 63 | `bg-gray-50 px-2 py-0.5 rounded` | `C.bgPage px-2 py-0.5 rounded` |
| `CarePlanItemRow.tsx` | 71 | `bg-gray-100 px-2 py-0.5 rounded text-gray-600` | `C.bgLight px-2 py-0.5 rounded ${C.text60}` |
| `CarePlanItemRow.tsx` | 80 | `bg-green-500` | 新トークン必要（後述） |
| `CarePlanItemRow.tsx` | 82 | `bg-gray-50 hover:bg-gray-100` | `${C.bgPage} ${C.hoverBgLight}` |
| `DailyRecordTimeline.tsx` | 57 | `bg-gray-50` | `C.bgPage` |
| `DailyCareNoteForm.tsx` | 49,51,52 | `text-gray-400`, `border-gray-100`, `text-gray-500` | FE-199 で対応済み |
| `HospitalizationExpandedView.tsx` | 54 | `bg-gray-50/50` | `C.bgPage30` |
| `button-variants.ts` | 20 | `text-red-600 hover:bg-red-50 hover:text-red-700` | `C.danger ${C.hoverBgDanger5}` |

## カテゴリ1: 既存トークンへの置換（即修正可能）

### `DischargeAlertDialog.tsx:52`
```tsx
// Before
className="... bg-red-600 hover:bg-red-700 ..."
// After
className={`... ${C.bgDanger} ${C.hoverBgDanger90} ...`}
```

### `CarePlanItemRow.tsx:62-63,71,82`
```tsx
// Before
<span className={`... border-gray-200`}>{getTypeLabel(plan.type)}</span>
<span className={`... bg-gray-50 px-2 py-0.5 rounded`}>{plan.description}</span>
<span className={`... bg-gray-100 px-2 py-0.5 rounded text-gray-600`}>{t}</span>
<Button ... className="h-9 w-9 p-0 bg-gray-50 hover:bg-gray-100">

// After
<span className={`... ${C.borderMedium}`}>{getTypeLabel(plan.type)}</span>
<span className={`... ${C.bgPage} px-2 py-0.5 rounded`}>{plan.description}</span>
<span className={`... ${C.bgLight} px-2 py-0.5 rounded ${C.text60}`}>{t}</span>
<Button ... className={`h-9 w-9 p-0 ${C.bgPage} ${C.hoverBgLight}`}>
```

### `DailyRecordTimeline.tsx:57`
```tsx
// Before
className={`... bg-gray-50 ...`}
// After
className={`... ${C.bgPage} ...`}
```

### `HospitalizationExpandedView.tsx:54`
```tsx
// Before: bg-gray-50/50
// After: ${C.bgPage30}  （または ${C.bgPage}/50）
```

### `button-variants.ts:20`（ghost-danger バリアント）
```ts
// Before
"ghost-danger": "text-red-600 hover:bg-red-50 hover:text-red-700",

// After — C.danger = "text-[#C0392B]" に相当するクラスを使用
"ghost-danger": `${C.danger} ${C.hoverBgDanger5}`,
```

## カテゴリ2: 新規デザイントークンが必要なもの

以下の医療ステータス色は `design-tokens.ts` に対応するトークンがなく、新規追加が必要。

### `CarePlanItemRow.tsx:80` — アクティブステータスドット
```tsx
// Before
bg-green-500  // アクティブ（緑）
bg-gray-300   // 非アクティブ（グレー）

// 必要なトークン
C.bgActive      // bg-[#4CAF50] または同等
C.bgInactive    // = C.bgLight 等で代替可能
```

### `DailyRecordTimeline.tsx:34-42` と `CarePlanItemRow.tsx:44-50` — 医療タイプバッジ色
```
食事:   bg-orange-50 text-orange-600 border-orange-200
投薬:   bg-blue-50   text-blue-600   border-blue-200
処置:   bg-purple-50 text-purple-600 border-purple-200
バイタル: bg-rose-50  text-rose-600   border-rose-200
排泄:   bg-amber-50  text-amber-600  border-amber-200
その他:  bg-gray-50  text-gray-600   border-gray-200
```

→ `design-tokens.ts` に `C.careType` オブジェクトとして追加、または個別定数を追加する方針を決定してから修正すること。

### `TimingSection.tsx:46,62` — 完了ステータスバッジ
```
完了: bg-green-100 text-green-600
未完: bg-gray-100  text-gray-400
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> Tailwind 直接色（`gray-*`, `red-*`, `blue-*` 等）の直接指定は禁止。

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/` — デザイントークンを徹底使用しているコンポーネント群

## 優先度
**Medium** — カテゴリ1（既存トークンへの置換）は即修正可能。カテゴリ2（新規トークン追加）は設計判断が必要。

## 関連ファイル
- `frontend/src/features/hospitalization/components/DischargeAlertDialog.tsx:52`
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx:46-83`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordTimeline.tsx:34-57`
- `frontend/src/features/hospitalization/components/DailyRecord/TimingSection.tsx:46,62`
- `frontend/src/components/ui/button-variants.ts:20`
- `frontend/src/lib/design-tokens.ts` — 置換先・追加先
