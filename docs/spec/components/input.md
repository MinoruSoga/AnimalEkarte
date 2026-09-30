# Input

> テキスト入力。配置: `frontend/src/components/ui/input.tsx`
> 使用箇所: 約 52 ファイル / Storybook: `input.stories.tsx`（未作成）

## Anatomy

`<input>` 単一要素。`rounded-xs`（4px — DESIGN.md text-input 仕様）。

## 状態

| 状態 | 実装 | 備考 |
|---|---|---|
| hover | ○ | `PALETTE.hoverBgInput`（暖色 neutral） |
| focus | ○ | `focus:bg-white` + focus ring（shadow-focus-primary） |
| focus-visible | ○ | ring-2 系 |
| disabled | ○ | `opacity-50` + `cursor-not-allowed` |
| invalid | ○ | `aria-invalid:border-destructive` + `bg-destructive/5` |
| readonly | △ | 素の `readOnly` 属性のみ。デザイン区別なし（design-states.md §1.1 は「ink 維持・hairline」を要求 — 実装側は未整備） |

## 既知のギャップ

- readonly の視覚仕様が未実装（薄い surface + hairline 化が望ましい）
- `PALETTE.*` の状態クラス直参照は層リーク（design-token-layers.md §5 — STATE バケット移行対象）
