# Badge

> ステータス表示バッジ。配置: `frontend/src/components/ui/badge.tsx`（variants: `badge-variants.ts`）
> 使用箇所: 約 17 ファイル + `BADGE.*` トークンが status-helpers 経由で数十ステータスに波及 / Storybook: 未作成

## Anatomy

`<span>`。`data-slot="badge"`。

## Props / Variants

| prop | 型 | 既定 |
|---|---|---|
| `variant` | `default \| secondary \| destructive \| outline` | `default` |

業務ステータス色はこの variant ではなく `BADGE.*` コンボトークン（blue/green/purple/orange/yellow/red/gray/muted）を使う — `src/lib/status-helpers.ts` が status → BADGE の写像。

## 状態 / 規則

- インタラクティブではない（装飾・識別用途）。クリック可能なら Badge ではなく Button/Chip を使う。
- **非色 cue 必須**: 臨床 sentinel（死亡・危険）を表す場合はテキスト/アイコンを必ず併置し色のみに依存しない（design-states.md §1.1）。
- `gray` は死亡 sentinel に使われる最重要色（`getPetStatusColor`）。

## 既知のギャップ（重要）

- **コントラスト実測で 6/8 コンボが AA 通常文字未達** — design-states.md §2.1 の表と §2.3 の是正案を参照。`BADGE.gray`（死亡）は 2.38:1 と最悪値であり、臨床 sentinel の可読性問題として優先度高。是正は背景維持・文字暗色化の方針で提案済み（要決裁）。
