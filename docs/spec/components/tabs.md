# Tabs

> Radix Tabs ベースのタブ切替。配置: `frontend/src/components/ui/tabs.tsx`
> Storybook: `tabs.stories.tsx`

## Anatomy

- `Tabs`（Root）→ `TabsList` → `TabsTrigger` + `TabsContent`（Root の直下にも置ける）

## Props / Variants

| prop | 型 | 既定 | 説明 |
|------|----|----|------|
| `Tabs.defaultValue` / `value` | string | — | 選択タブ（controlled/uncontrolled 両対応） |
| `TabsTrigger.value` | string | — | TabsContent の value と対応 |

## 状態

| 状態 | 実装 | 備考 |
|------|------|------|
| selected | ○ | `data-[state=active]:bg-card` — 選択中のみ背景カード色 |
| focus-visible | ○ | `focus-visible:ring-[3px] ring-ring` + outline-1 |
| disabled | ○ | `disabled:opacity-50` + pointer-events-none |
| hover | ○ | muted 背景内で非選択は muted-foreground |

## アクセシビリティ要件

- Radix が tablist/tab/tabpanel ロールと矢印キー巡回を提供。
- `TabsTrigger` は `min-h-11 min-w-11` で 44px タッチターゲット確保済み。
- disabled タブは「権限なし」等の理由を隣接テキストかツールチップで補う（design-states.md §3 read-only/権限）。

## 使い方（Do / Don't）

- Do: タブ = 同一データの切替。別ページ遷移にはリンク/ナビを使う。
- Don't: `TabsList` を垂直・多段にしない（折返し前提の設計ではない）。

## 既知のギャップ

- モバイル時の scrollable tablist は未整備（現状 `w-fit` inline-flex のみ）。
