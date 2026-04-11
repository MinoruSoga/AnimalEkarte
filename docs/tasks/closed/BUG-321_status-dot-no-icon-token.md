# BUG-321: ステータスドット (`size-2`/`size-2.5`/`size-1.5`) の ICON トークン未定義

## 概要

ステータスインジケーターや色ドット（丸いラベル）に `size-2`, `size-2.5`, `size-1.5` が直接指定されているが、
これらに対応する ICON トークンが存在しない。
`design-tokens.ts` の "すべてのアイコンサイズはここで一元管理する" の原則に反している。

## 再現手順

1. 下記ファイルの該当行を参照
2. `rounded-full` のドット要素に `size-2` / `size-2.5` / `size-1.5` が直接書かれている

## 期待する動作

- `ICON` オブジェクトに `dot`, `dotMd`, `dotSm` 等のトークンを追加
- 各コンポーネントでトークンを使用する

## 現状コード

### `size-2` パターン (8px) — 2箇所

#### `frontend/src/features/reception/components/KanbanColumn.tsx:62`
```tsx
<div className={`size-2 rounded-full ${colors.dot}`} aria-hidden="true" />
```

#### `frontend/src/features/medical-records/components/VitalsTab/VitalsGraph.tsx:138`
```tsx
className="size-2 rounded-full shrink-0"
```

---

### `size-2.5` パターン (10px) — 5箇所

#### `frontend/src/features/master/routes/ReservationTypeSettings.tsx:136`
```tsx
<span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: group.color }} />
```

#### `frontend/src/features/master/routes/ReservationTypeSettings.tsx:201`
```tsx
<span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: PALETTE.grayMedium }} />
```

#### `frontend/src/features/master/routes/StaffSettings.tsx:441`
```tsx
className="size-2.5 rounded-full shrink-0"
```

#### `frontend/src/features/master/routes/ReservationTypeSidePanel.tsx:113`
```tsx
<span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: g.color }} />
```

---

### `size-1.5` パターン (6px) — 1箇所

#### `frontend/src/features/master/routes/StaffSettings.tsx:712`
```tsx
className="size-1.5 rounded-full shrink-0"
```

---

### `size-2` サイドバー インジケーター — 2箇所

#### `frontend/src/components/shared/Layout/Sidebar.tsx:230,257`
```tsx
// 「」マーク等の通知バッジ相当
className="size-2 rounded-full"  // （推定。要確認）
```

---

### 比較: 既存の ICON トークン（最小は `size-3` = 12px）
```ts
// frontend/src/lib/design-tokens.ts
xxs:  "size-3",    // 12px — 極小インジケーター
smXs: "size-3.5",  // 14px
sm:   "size-4",    // 16px
```

`size-2` (8px) / `size-2.5` (10px) / `size-1.5` (6px) に対応するトークンが存在しない。

## 影響範囲

| 対象 | 行番号 | サイズ | 用途 |
|------|--------|-------|------|
| `features/reception/components/KanbanColumn.tsx` | 62 | `size-2` | カンバンカラムの色ドット |
| `features/medical-records/components/VitalsTab/VitalsGraph.tsx` | 138 | `size-2` | バイタルグラフの凡例ドット |
| `features/master/routes/ReservationTypeSettings.tsx` | 136,201 | `size-2.5` | 予約種別カラードット |
| `features/master/routes/StaffSettings.tsx` | 441,712 | `size-2.5`,`size-1.5` | スタッフ設定の状態インジケーター |
| `features/master/routes/ReservationTypeSidePanel.tsx` | 113 | `size-2.5` | 予約種別グループのカラードット |
| `components/shared/Layout/Sidebar.tsx` | 230,257 | `size-2` | サイドバー通知バッジ |

**合計**: 11箇所

## 修正方針

### Phase 1: `design-tokens.ts` に ICON ドットトークンを追加

```ts
// ICON オブジェクト内に追加（`xxs` の後に）
/** 極小ステータスドット / 通知バッジ (8px) */
dot:    "size-2",
/** 小ステータスドット / 色ラベル (10px) */
dotMd:  "size-2.5",
/** 最小インジケーター (6px) */
dotSm:  "size-1.5",
```

### Phase 2: 各コンポーネントで ICON トークンを使用

```tsx
// Before
<div className={`size-2 rounded-full ${colors.dot}`} aria-hidden="true" />

// After
import { ICON } from "@/lib/design-tokens";
<div className={`${ICON.dot} rounded-full ${colors.dot}`} aria-hidden="true" />
```

```tsx
// Before
<span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: group.color }} />

// After
<span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: group.color }} />
```

## 準拠すべきプロジェクト規約

### `design-tokens.ts` コメント — ICON サイズ一元管理
> ```
> /*  3. Icon Sizes
>  *  すべてのアイコンサイズはここで一元管理する。
>  *  直接 size-N / h-N w-N を書かず、このトークンを使うこと。
>  */
> ```

### `.claude/rules/code-style.md` — デザイントークン必須
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`, `ICON`) を使用する。

### プロジェクト内参照実装
- `frontend/src/lib/design-tokens.ts:648` — `ICON.xxs` = `"size-3"` が正しく定義されている（最小トークン）

## 優先度

**Low** — 機能への影響なし。ICON トークン体系の完全性向上のため、`ICON.dot`/`ICON.dotMd`/`ICON.dotSm` を追加して統一する。

## 関連チケット

- BUG-320: `size-7`/`size-8` アイコンボタントークン欠如（同種）

## 関連ファイル

- `frontend/src/lib/design-tokens.ts:630-649` — ICON トークン定義箇所（追加対象）
- `frontend/src/features/reception/components/KanbanColumn.tsx:62`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsGraph.tsx:138`
- `frontend/src/features/master/routes/ReservationTypeSettings.tsx:136,201`
- `frontend/src/features/master/routes/StaffSettings.tsx:441,712`
- `frontend/src/features/master/routes/ReservationTypeSidePanel.tsx:113`
- `frontend/src/components/shared/Layout/Sidebar.tsx:230,257`
