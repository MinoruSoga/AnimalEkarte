# BUG-167: PALETTE オブジェクトを C.* セマンティックトークン経由せず直接使用

## 概要

`reservations/components/WeekView.tsx`、`master/components/StaffSettings.tsx` 等で
`design-tokens.ts` の内部プリミティブ `PALETTE` を直接コンポーネントで参照している。
`PALETTE` はセマンティックトークン（`C.*`、`STYLE.*`）の実装詳細であり、
コンポーネントから直接参照してはならない。
将来のテーマ変更時にセマンティック層を迂回した箇所だけ更新漏れが発生する。

## 再現手順

1. 対象ファイルで `PALETTE` を検索
2. `import { PALETTE } from "@/lib/design-tokens"` が存在する箇所を確認
3. **結果**: コンポーネントが `PALETTE.borderMedium`、`PALETTE.defaultGray` 等を直参照している

## 期待する動作

コンポーネントは `C.*` / `STYLE.*` セマンティックトークンのみを参照する。
`PALETTE` は `design-tokens.ts` 内部でのみ使用する。

## 現状コード

### `frontend/src/features/reservations/components/WeekView.tsx:133付近`
```tsx
// ❌ PALETTE 直参照
style={{ backgroundColor: PALETTE.borderMedium }}
// → C.* に対応するセマンティックトークン: C.borderMedium は STYLE 側に存在
```

### `frontend/src/features/master/components/StaffSettings.tsx:358,524-525,530付近`
```tsx
// ❌ fallback に PALETTE 直参照
style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}

// ❌ 透過背景の計算に PALETTE 直参照
backgroundColor: g.color ? `${g.color}18` : PALETTE.bgSkeleton,
color: g.color ?? PALETTE.primary,

style={{ backgroundColor: g.color ?? PALETTE.defaultGray }}
```

### `frontend/src/features/master/components/ServiceTypeSettings.tsx:40付近`
```tsx
// ❌ fallback に PALETTE 直参照
color: item?.color ?? PALETTE.defaultBlue
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// ✅ C.* セマンティックトークンを使用
import { C } from '@/lib/design-tokens';

// borderMedium の場合
style={{ borderColor: C.borderMedium }}  // C.borderMedium = "rgba(55,53,47,0.16)"

// グレーフォールバックの場合
style={{ backgroundColor: group.color ?? C.bgMuted }}
// または className を使用
className={group.color ? undefined : STYLE.bg.muted}
```

## 影響範囲

| ファイル | 該当箇所 | PALETTE 参照トークン |
|---------|---------|-------------------|
| `features/reservations/components/WeekView.tsx` | L133 | `PALETTE.borderMedium` |
| `features/master/components/StaffSettings.tsx` | L358 | `PALETTE.defaultGray` |
| `features/master/components/StaffSettings.tsx` | L524-525 | `PALETTE.bgSkeleton`, `PALETTE.primary` |
| `features/master/components/StaffSettings.tsx` | L530 | `PALETTE.defaultGray` |
| `features/master/components/ServiceTypeSettings.tsx` | L40 | `PALETTE.defaultBlue` |

## 修正方針

### 1. PALETTE → C.* 置換マッピング

```tsx
// WeekView.tsx L133
// Before
style={{ backgroundColor: PALETTE.borderMedium }}
// After
style={{ backgroundColor: C.borderMedium }}  // "rgba(55,53,47,0.16)"

// StaffSettings.tsx L358, L530
// Before
style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
// After — bgMuted は C.bgMuted = "rgba(55,53,47,0.06)" が最近傍
style={{ backgroundColor: group.color ?? "rgba(55,53,47,0.06)" }}
// または設計判断: PALETTE.defaultGray の実値を C.* にマッピングして定義する

// StaffSettings.tsx L524-525
// Before
backgroundColor: g.color ? `${g.color}18` : PALETTE.bgSkeleton,
color: g.color ?? PALETTE.primary,
// After
backgroundColor: g.color ? `${g.color}18` : C.bgSkeleton,  // C.bgSkeleton = "rgba(55,53,47,0.06)"
color: g.color ?? C.text,  // C.text = "#37352F"
```

### 2. PALETTE.defaultBlue / defaultGray に対応する C.* が未定義の場合

`design-tokens.ts` に以下を追加定義する:
```ts
// design-tokens.ts の C オブジェクトに追加
defaultGray: "rgba(55,53,47,0.09)",  // PALETTE.defaultGray の実値を確認して設定
defaultBlue: "#2383E2",              // C.bgAccent の生値と同値なら統一
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

`PALETTE` はセマンティック定数を定義するための内部詳細であり、コンポーネントが直接参照することは規約違反。

### プロジェクト内参照実装
- `frontend/src/features/owners/` — C.* / STYLE.* のみ使用、PALETTE 直参照なし

## 優先度
**Low** — 機能影響なし。テーマ一貫性とメンテナビリティの問題。

## 関連チケット
- BUG-162: ハードコード Tailwind カラー（同種・より広範な問題）

## 関連ファイル
- `frontend/src/features/reservations/components/WeekView.tsx:133`
- `frontend/src/features/master/components/StaffSettings.tsx:358,524-525,530`
- `frontend/src/features/master/components/ServiceTypeSettings.tsx:40`
- `frontend/src/lib/design-tokens.ts`
