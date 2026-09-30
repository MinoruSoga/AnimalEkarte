import type { Meta, StoryObj } from "@storybook/react-vite";
import { FilteringIndicator } from "./FilteringIndicator";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/FilteringIndicator",
  component: FilteringIndicator,
  tags: ["autodocs"],
  args: {
    isFiltering: false,
    children: (
      <div className="rounded-md border p-6">
        <p>フィルタ対象のリスト内容</p>
        <p>1,234 件中 20 件を表示</p>
      </div>
    ),
  },
} satisfies Meta<typeof FilteringIndicator>;

type Story = StoryObj<typeof FilteringIndicator>;

/** 遅延レンダリング中: 内容が薄くなる（useDeferredValue 連携用） */
export const Filtering: Story = {
  args: { isFiltering: true },
};

export const Idle: Story = {};
