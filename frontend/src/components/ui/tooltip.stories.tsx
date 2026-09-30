import type { Meta, StoryObj } from "@storybook/react-vite";

import { Tooltip } from "./tooltip";
import { Button } from "./button";

/** 独自実装 Tooltip — content 文字列をホバー/フォーカスで portal 表示。 */
const meta = {
  title: "UI/Tooltip",
  component: Tooltip,
  tags: ["autodocs"],
  argTypes: {
    content: { control: "text" },
  },
  args: {
    content: "診察を開始します",
    children: <Button variant="outline">ホバーで表示</Button>,
  },
} satisfies Meta<typeof Tooltip>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
