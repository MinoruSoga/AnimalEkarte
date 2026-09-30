import type { Meta, StoryObj } from "@storybook/react-vite";

import { Popover, PopoverContent, PopoverTrigger } from "./popover";
import { Button } from "./button";
import { Input } from "./input";
import { Label } from "./label";

const meta = {
  title: "UI/Popover",
  component: Popover,
  tags: ["autodocs"],
} satisfies Meta<typeof Popover>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">フィルタ</Button>
      </PopoverTrigger>
      <PopoverContent>
        <div style={{ display: "grid", gap: 8 }}>
          <Label htmlFor="pop-filter">表示件数</Label>
          <Input id="pop-filter" defaultValue="20" />
        </div>
      </PopoverContent>
    </Popover>
  ),
};

export const Open: Story = {
  render: () => (
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">フィルタ</Button>
      </PopoverTrigger>
      <PopoverContent>
        <p className="text-sm text-muted-foreground">Popover コンテンツ</p>
      </PopoverContent>
    </Popover>
  ),
};
