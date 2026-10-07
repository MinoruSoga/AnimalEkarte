# Badge

> ステータス表示バッジ。配置: `frontend/src/components/ui/badge.tsx`（variants: `badge-variants.ts`）
> 使用箇所: 約 17 ファイル + `BADGE.*` トークンが status-helpers 経由で数十ステータスに波及 / Storybook: `badge.stories.tsx`

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

- **コントラスト是正済み（2026-09-30 実装）**: 是正前の実測（6/8 コンボが AA 通常文字未達、`BADGE.gray` 死亡 sentinel は 2.38:1）は design-states.md §2.1 のベースライン表に記録済み。実装は §2.3 のとおり `C.textBadge*`（`BADGE.*` の text 参照を差替え、パステル背景は据え置き。`design-tokens.ts` の `C` セクションにメンバ定義）。機械検証は design-token-layers.md §6 C23（全 `BADGE.*` コンボ ≥4.5:1 を design-audit が検証）。
