import type { StorybookConfig } from "@storybook/react-vite";

/**
 * Component catalog (引き継ぎ資産) — ui/* プリミティブと shared コンポーネントの
 * 全 variant / 状態をブラウズ・検証するための Storybook 構成。
 * 状態仕様の正本は docs/spec/design-states.md、トークン層は design-token-layers.md。
 */
const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(ts|tsx)"],
  // essentials（controls/actions/backgrounds/viewport）は SB9+ で core 統合済みのため
  // addons 配列には列挙しない（v10 に addon-essentials パッケージは存在しない）。
  addons: ["@storybook/addon-a11y", "@storybook/addon-docs"],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  core: {
    // ローカル dev カタログのため匿名 telemetry を無効化し、Docker 経由の
    // ホスト接続を loopback に限定する（dev server の allowedHosts 警告対応）。
    disableTelemetry: true,
    allowedHosts: ["localhost", "127.0.0.1", "0.0.0.0"],
  },
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default config` が必須形式（dev-only カタログ設定）
export default config;
