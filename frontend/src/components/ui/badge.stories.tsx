import type { Meta, StoryObj } from "@storybook/react-vite";

import { Badge } from "./badge";

/**
 * ui/badge.tsx の4 variant。臨床ステータスのパステル BADGE.* トークンは
 * components/shared 側の責務 — コントラスト既知問題は design-states.md §2 参照。
 */
const meta = {
  title: "UI/Badge",
  component: Badge,
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "secondary", "destructive", "outline"],
    },
  },
  args: { children: "予約済" },
} satisfies Meta<typeof Badge>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const AllVariants: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
      <Badge variant="default">default</Badge>
      <Badge variant="secondary">secondary</Badge>
      <Badge variant="destructive">destructive</Badge>
      <Badge variant="outline">outline</Badge>
    </div>
  ),
};
