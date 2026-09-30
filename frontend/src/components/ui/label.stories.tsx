import type { Meta, StoryObj } from "@storybook/react-vite";

import { Label } from "./label";
import { Input } from "./input";

const meta = {
  title: "UI/Label",
  component: Label,
  tags: ["autodocs"],
} satisfies Meta<typeof Label>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <Label htmlFor="label-demo">飼主名</Label>
      <Input id="label-demo" placeholder="姓 名" />
    </div>
  ),
};
