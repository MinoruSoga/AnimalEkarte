import type { Meta, StoryObj } from "@storybook/react-vite";

import { Separator } from "./separator";

const meta = {
  title: "UI/Separator",
  component: Separator,
  tags: ["autodocs"],
} satisfies Meta<typeof Separator>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Horizontal: Story = {
  render: () => (
    <div style={{ maxWidth: 320 }}>
      <p className="text-sm">上のセクション</p>
      <Separator className="my-4" />
      <p className="text-sm text-muted-foreground">下のセクション</p>
    </div>
  ),
};

export const Vertical: Story = {
  render: () => (
    <div style={{ display: "flex", alignItems: "center", gap: 16, height: 40 }}>
      <span className="text-sm">編集</span>
      <Separator orientation="vertical" />
      <span className="text-sm">印刷</span>
      <Separator orientation="vertical" />
      <span className="text-sm text-destructive">削除</span>
    </div>
  ),
};
