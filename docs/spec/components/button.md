# Button

> 汎用ボタン。配置: `frontend/src/components/ui/button.tsx`（variants: `button-variants.ts`）
> 使用箇所: 約 120 ファイル（最大消費プリミティブ）/ Storybook: `button.stories.tsx`

## Anatomy

`<button>`（`asChild` で Slot 差替可）。`data-slot="button"`。`type` 既定 `button`（`asChild` 時は注入しない）。

## Props / Variants

| prop | 型 | 既定 | 説明 |
|---|---|---|---|
| `variant` | `default \| destructive \| outline \| secondary \| ghost \| link \| primary \| ghost-danger` | `default` | `default`/`primary` = pill + `#027078` 塗り（`bgActionPrimarySolid`、AA 適合） |
| `size` | `default \| sm \| lg \| icon` | `default` | 全サイズ `min-h-11 min-w-11`（44px ターゲット保証） |
| `loading` | `boolean` | `false` | design-states.md §3 — spinner + aria-busy + 操作抑止 |
| `asChild` | `boolean` | `false` | Radix Slot |

## 状態

| 状態 | 実装 | 備考 |
|---|---|---|
| hover | ○ | `C.hoverBgActionPrimarySolid`（`#025F66`）等 variant 毎 |
| active | ○ | `C.activeBgActionPrimarySolid`（`#025F66`） |
| focus-visible | ○ | ring-2 `#038B94` + offset-1（EMR-207 回帰テスト済み） |
| disabled | ○ | `opacity-50` + `pointer-events-none` |
| loading | ○ | FE-DS1 で追加。`aria-busy` |
| invalid | — | 非入力要素 |

## アクセシビリティ要件

- キーボードフォーカスで ring が必ず見える（`button.test.tsx` で全 variant × size を検証）
- `disabled` 時はフォーカス不可 + 操作不可の両立
- アイコン単独使用（`size="icon"`）は `aria-label` 必須

## 既知のギャップ

- 旧塗り `#038B94` は 4.10:1 で AA 未達だったため、決裁により `#027078`（5.85:1）へ移行済み（design-states.md §2.3）
