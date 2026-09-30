import type { Meta, StoryObj } from "@storybook/react-vite";

import { ScrollArea } from "./scroll-area";

const meta = {
  title: "UI/ScrollArea",
  component: ScrollArea,
  tags: ["autodocs"],
} satisfies Meta<typeof ScrollArea>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <ScrollArea
      style={{ height: 160, maxWidth: 320, border: "1px solid var(--border)", borderRadius: 6 }}
    >
      <div style={{ padding: 12, display: "grid", gap: 8 }}>
        {Array.from({ length: 15 }, (_, i) => (
          <p key={i} className="text-sm">
            履歴項目 {i + 1}
          </p>
        ))}
      </div>
    </ScrollArea>
  ),
};
