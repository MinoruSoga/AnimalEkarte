# BUG-320: アイコンボタン `size-7`/`size-8` の STYLE トークン未定義 — 繰り返しパターンが直接指定されている

## 概要

`size-8 flex items-center justify-center rounded-[3px]` / `size-7 flex items-center justify-center rounded-[3px]` というアイコンボタンコンテナパターンが
複数ファイルに繰り返し直接記述されているが、対応する STYLE/LAYOUT トークンが存在しない。
既存の `STYLE.sidePeekToolbarBtn`（`size-9` 版）と同じ責務を持つ 32px/28px バリアントのトークン化漏れ。

## 再現手順

1. 下記ファイルの該当行を参照
2. `size-8 flex items-center justify-center rounded-[3px]` または `size-7` 版が直接書かれている

## 期待する動作

- `design-tokens.ts` に `STYLE.iconBtn32` / `STYLE.iconBtn28`（または類似命名）を追加
- 各コンポーネントでトークンを使用する

## 現状コード

### `size-8` パターン (32px) — 9箇所

#### `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx:144,152,321`
```tsx
// 3箇所すべて同一ベースパターン（カラーのみ異なる）
className={`size-8 flex items-center justify-center rounded-[3px] ${C.textStatusGreen} ${C.hoverBgStatusGreen} transition-colors`}
className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverBgLight} transition-colors`}
className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverText} ${C.hoverBgLight} transition-colors`}
```

#### `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:221,229,504`
```tsx
// 3箇所すべて同一ベースパターン（カラーのみ異なる）
className={`size-8 flex items-center justify-center rounded-[3px] ${C.textStatusGreen} ${C.hoverBgStatusGreen} transition-colors`}
className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverBgLight} transition-colors`}
className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverText} ${C.hoverBgLight} transition-colors`}
```

#### `frontend/src/features/auth/components/LoginForm.tsx:214`
```tsx
// パスワード表示トグルボタン（absolute 配置付き）
className={`absolute right-1 top-1/2 -translate-y-1/2 size-8 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} transition-colors`}
```

#### `frontend/src/features/auth/routes/ResetPasswordPage.tsx:124,151`
```tsx
// 同上（absolute 配置付き）
className={`absolute right-1 top-1/2 -translate-y-1/2 size-8 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} transition-colors`}
```

---

### `size-7` パターン (28px) — 5箇所

#### `frontend/src/components/shared/Layout/Sidebar.tsx:385,394`
```tsx
className={`size-7 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} ${C.hoverBgMedium} transition-colors shrink-0`}
```

#### `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:367,377`
```tsx
className={`size-7 ${C.text40} ${C.hoverText} disabled:opacity-20`}
```

#### `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:387`
```tsx
className="size-7"
```

---

### 比較: 既存の正しいトークン（`size-9`版）
```ts
// frontend/src/lib/design-tokens.ts — STYLE.sidePeekToolbarBtn
sidePeekToolbarBtn:
  `size-9 flex items-center justify-center rounded-[3px] ${C.text45} ${C.hoverBgMedium} transition-colors`,
```

`size-7` / `size-8` の同様のトークンが存在しないため、直接指定になっている。

## 影響範囲

| 対象 | 行番号 | パターン | 件数 |
|------|--------|---------|------|
| `features/medical-records/components/CheckupsTab/CheckupsTab.tsx` | 144,152,321 | `size-8` アイコンボタン | 3 |
| `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 221,229,504 | `size-8` アイコンボタン | 3 |
| `features/auth/components/LoginForm.tsx` | 214 | `size-8` パスワードトグル | 1 |
| `features/auth/routes/ResetPasswordPage.tsx` | 124,151 | `size-8` パスワードトグル | 2 |
| `components/shared/Layout/Sidebar.tsx` | 385,394 | `size-7` アイコンボタン | 2 |
| `features/medical-records/components/TreatmentsTab/TreatmentRow.tsx` | 367,377,387 | `size-7` アイコンボタン | 3 |

**合計**: 14箇所

## 修正方針

### Phase 1: `design-tokens.ts` にトークンを追加

```ts
// STYLE オブジェクト内に追加
/* ── Compact Icon Button (size-8 / 32px) ── */
/** 32px アイコンボタン基底クラス (医療記録タブ・認証フォーム) */
iconBtn32:
  `size-8 flex items-center justify-center rounded-[3px] transition-colors`,

/* ── Compact Icon Button (size-7 / 28px) ── */
/** 28px アイコンボタン基底クラス (サイドバー・TreatmentRow) */
iconBtn28:
  `size-7 flex items-center justify-center rounded-[3px] transition-colors`,
```

### Phase 2: 各コンポーネントでトークンを使用

カラークラスは用途ごとに異なるため、基底クラスをトークンで管理し、カラーは `cn()` で合成する：

```tsx
// Before
className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverBgLight} transition-colors`}

// After
className={cn(STYLE.iconBtn32, C.text60, C.hoverBgLight)}
```

```tsx
// Before (Sidebar)
className={`size-7 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} ${C.hoverBgMedium} transition-colors shrink-0`}

// After
className={cn(STYLE.iconBtn28, C.text35, C.hoverText, C.hoverBgMedium, "shrink-0")}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — デザイントークン必須
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### `design-tokens.ts` — ICON サイズ一元管理の原則
> すべてのアイコンサイズはここで一元管理する。直接 size-N / h-N w-N を書かず、このトークンを使うこと。

### プロジェクト内参照実装
- `frontend/src/lib/design-tokens.ts` — `STYLE.sidePeekToolbarBtn` = `size-9` バージョンが正しく定義済み
- `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:251` — `ICON.xxs` / `ICON.smXs` の正しい使用例

## 優先度

**Medium** — 機能への影響はないが、トークン体系の一貫性が損なわれており、14箇所に渡る繰り返しが将来のデザイン変更コストを高める。

## 関連チケット

- BUG-315: デザイントークン参照ガイド（closed）

## 関連ファイル

- `frontend/src/lib/design-tokens.ts` — トークン追加対象
- `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx:144,152,321`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:221,229,504`
- `frontend/src/features/auth/components/LoginForm.tsx:214`
- `frontend/src/features/auth/routes/ResetPasswordPage.tsx:124,151`
- `frontend/src/components/shared/Layout/Sidebar.tsx:385,394`
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:367,377,387`
