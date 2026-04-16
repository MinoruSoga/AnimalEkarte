# BUG-231: `ghost-danger` ボタンバリアントが設計トークンの danger 色と不一致

## 概要

`src/components/ui/button-variants.ts:20` で定義された `ghost-danger` バリアントが Tailwind の `text-red-600`（`#DC2626`）を直接使用している。プロジェクトの設計トークンでは危険色を `C.danger = "text-[#C0392B]"`（`#C0392B`、暗いクリムゾン）と定めており、`STYLE.btnDangerGhost` もこれを参照している。`variant="ghost-danger"` を使うコンポーネントと `variant="ghost" className={STYLE.btnDangerGhost}` を使うコンポーネントで削除ボタンの色が異なる視覚的不一致が生じている。

## 再現手順

1. トリミングフォーム（`/trimming/[id]`）を開く
2. 削除ボタン（赤）の色を確認する — `#DC2626`（Tailwind red-600）
3. 入院フォーム（`/hospitalization/[id]`）の削除ボタンと比較する — `#C0392B`（プロジェクト danger 色）
4. **結果**: 2つのページで削除ボタンの赤色のトーンが異なる

## 期待する動作

- すべての ghost 系削除ボタンが同一の danger 色 `#C0392B` を使用すること

## 現状コード

### `src/components/ui/button-variants.ts:20`
```typescript
// ❌ Tailwind red-600 (#DC2626) を直接使用
"ghost-danger": "text-red-600 hover:bg-red-50 hover:text-red-700",
```

### 設計トークンの正しい danger 色（`src/lib/design-tokens.ts`）
```typescript
danger:         "#C0392B",           // PALETTE
C.danger:       "text-[#C0392B]",   // Tailwind class
C.hoverBgDanger5: "hover:bg-[#C0392B]/5",
STYLE.btnDangerGhost: `${C.danger} ${C.hoverBgDanger5} transition-colors`,
// = "text-[#C0392B] hover:bg-[#C0392B]/5 transition-colors"
```

### `ghost-danger` バリアントを使用している箇所（4ファイル）

```
features/trimming/routes/TrimmingForm.tsx:585
features/estimates/routes/EstimateDetail.tsx:68
features/medical-records/routes/MedicalRecordForm.tsx:414
features/hospitalization/components/HospitalizationDetailActions.tsx:31
```

### 比較: 正しい実装（`STYLE.btnDangerGhost` 使用）

```tsx
// ✅ 正しい: variant="ghost" + STYLE.btnDangerGhost
// features/examinations/routes/ExaminationForm.tsx
<Button
  variant="ghost"
  type="button"
  className={`h-10 text-sm ${STYLE.btnDangerGhost} mr-auto`}
>
  <Trash2 className={`mr-1.5 ${ICON.action}`} />
  削除
</Button>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `src/components/ui/button-variants.ts:20` | `ghost-danger` 定義で `text-red-600 hover:bg-red-50 hover:text-red-700` | 未修正 |
| `features/trimming/routes/TrimmingForm.tsx:585` | `variant="ghost-danger"` 使用 | 未修正 |
| `features/estimates/routes/EstimateDetail.tsx:68` | `variant="ghost-danger"` 使用 | 未修正 |
| `features/medical-records/routes/MedicalRecordForm.tsx:414` | `variant="ghost-danger"` 使用 | 未修正 |
| `features/hospitalization/components/HospitalizationDetailActions.tsx:31` | `variant="ghost-danger"` 使用 | 未修正 |

## 修正方針

### ステップ 1: `button-variants.ts:20` を設計トークンと整合させる

```typescript
// Before
"ghost-danger": "text-red-600 hover:bg-red-50 hover:text-red-700",

// After — 設計トークンの正しい値に合わせる
"ghost-danger": "text-[#C0392B] hover:bg-[#C0392B]/5 transition-colors",
```

または、各ファイルの `variant="ghost-danger"` を `variant="ghost" + STYLE.btnDangerGhost` に統一しバリアント定義自体を削除する（推奨）。

### ステップ 2（代替案・推奨）: `ghost-danger` バリアントを廃止し `STYLE.btnDangerGhost` に統一

```tsx
// Before
<Button variant="ghost-danger" className="h-10 rounded-[6px] text-sm px-4">

// After
<Button variant="ghost" className={`${STYLE.btnDangerGhost} h-10 rounded-[6px] text-sm px-4`}>
```

4ファイルすべてを修正し、`button-variants.ts` から `ghost-danger` エントリを削除する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用（`#37352F`等ハードコード禁止）

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

`text-red-600`・`hover:bg-red-50` は Tailwind 組み込み色クラスであり、プロジェクトの danger 色 `#C0392B` とは異なる。

### プロジェクト内参照実装
- `features/examinations/routes/ExaminationForm.tsx` — `variant="ghost" className={STYLE.btnDangerGhost}` の正しい使用例
- `features/hospitalization/routes/HospitalizationForm.tsx:163-169` — 同パターン
- `src/lib/design-tokens.ts:851-852` — `STYLE.btnDangerGhost` 定義

## 優先度
**Medium** — 削除ボタンの色が機能によって異なる視覚的不一致。ページによって Trash ボタンの赤色が違って見える。

## 関連チケット
- BUG-219: StaffSettings.tsx の hex カラー直接使用（同種の設計トークン違反）

## 関連ファイル
- `frontend/src/components/ui/button-variants.ts:20`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx:585`
- `frontend/src/features/estimates/routes/EstimateDetail.tsx:68`
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:414`
- `frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx:31`
- `frontend/src/lib/design-tokens.ts:309,314,851-852`
