import type { Meta, StoryObj } from "@storybook/react-vite";

import { ToggleGroup, ToggleGroupItem } from "./toggle-group";

const meta = {
  title: "UI/ToggleGroup",
  component: ToggleGroup,
  tags: ["autodocs"],
  args: { type: "single" },
} satisfies Meta<typeof ToggleGroup>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Single: Story = {
  render: () => (
    <ToggleGroup type="single" defaultValue="list">
      <ToggleGroupItem value="list" aria-label="リスト表示">
        リスト
      </ToggleGroupItem>
      <ToggleGroupItem value="kanban" aria-label="カンバン表示">
        カンバン
      </ToggleGroupItem>
      <ToggleGroupItem value="calendar" aria-label="カレンダー表示">
        カレンダー
      </ToggleGroupItem>
    </ToggleGroup>
  ),
};
