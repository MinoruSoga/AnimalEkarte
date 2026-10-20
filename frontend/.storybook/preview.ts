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
      config: {
        rules: [
          // WCAG 2.2 AA（design-states.md / accessibility-rules.md の担保ライン）のうち
          // axe が機械判定できるのは SC 2.5.8 Target Size の target-size だけで、
          // axe-core 既定では無効なため明示的に有効化する。2.4.11 / 2.5.7 / 3.2.6 /
          // 3.3.7 / 3.3.8 は accessibility-rules.md §9 の手動レビューで担保する。
          { id: "target-size", enabled: true },
        ],
      },
    },
  },
  initialGlobals: {
    backgrounds: { value: "canvasSoft" },
  },
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default preview` が必須形式（dev-only カタログ設定）
export default preview;
