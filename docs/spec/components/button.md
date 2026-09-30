# Button

> 汎用ボタン。配置: `frontend/src/components/ui/button.tsx`（variants: `button-variants.ts`）
> 使用箇所: 約 120 ファイル（最大消費プリミティブ）/ Storybook: `button.stories.tsx`

## Anatomy

`<button>`（`asChild` で Slot 差替可）。`data-slot="button"`。`type` 既定 `button`（`asChild` 時は注入しない）。

## Props / Variants

| prop | 型 | 既定 | 説明 |
|---|---|---|---|
| `variant` | `default \| destructive \| outline \| secondary \| ghost \| link \| primary \| ghost-danger` | `default` | `default`/`primary` = pill + `#038B94` 塗り（button-primary 仕様） |
| `size` | `default \| sm \| lg \| icon` | `default` | 全サイズ `min-h-11 min-w-11`（44px ターゲット保証） |
| `loading` | `boolean` | `false` | design-states.md §3 — spinner + aria-busy + 操作抑止 |
| `asChild` | `boolean` | `false` | Radix Slot |

## 状態

| 状態 | 実装 | 備考 |
|---|---|---|
| hover | ○ | `C.hoverBgActionPrimary`（`#027078`）等 variant 毎 |
| active | ○ | `C.activeBgActionPrimary` |
| focus-visible | ○ | ring-2 `#038B94` + offset-1（EMR-207 回帰テスト済み） |
| disabled | ○ | `opacity-50` + `pointer-events-none` |
| loading | ○ | FE-DS1 で追加。`aria-busy` |
| invalid | — | 非入力要素 |

## アクセシビリティ要件

- キーボードフォーカスで ring が必ず見える（`button.test.tsx` で全 variant × size を検証）
- `disabled` 時はフォーカス不可 + 操作不可の両立
- アイコン単独使用（`size="icon"`）は `aria-label` 必須

## 既知のギャップ

- white on `#038B94` は 4.10:1 で AA 通常文字未達（design-states.md §2.3 — 要デザイン決裁）
