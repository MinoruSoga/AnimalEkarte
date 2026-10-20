import type { Preview } from "@storybook/react-vite";

// Tailwind v4 は CSS import でテーマ全体（@theme inline 含む）が入る
import "../src/index.css";

const preview: Preview = {
  parameters: {
    controls: { expanded: true },
    layout: "centered",
    backgrounds: {
      options: {
        // design-system.md §2.2 — ページ canvas（暖色オフホワイト）と白 surface
        canvasSoft: { name: "canvas-soft", value: "#F6F5F4" },
        canvas: { name: "canvas (white)", value: "#FFFFFF" },
      },
    },
    a11y: {
      // axe を全 story に適用（design-states.md §2 の実測表と併用）。
      // "error": 違反を a11y パネルで失敗として表示し、Chromatic の Accessibility
      // tests と Storybook の Vitest 連携（導入時）では story を fail させる。
      // 正当な例外は story 側で parameters.a11y.config.rules に理由付きで記録する。
      test: "error",
    },
  },
  initialGlobals: {
    backgrounds: { value: "canvasSoft" },
  },
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default preview` が必須形式（dev-only カタログ設定）
export default preview;
