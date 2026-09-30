import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { UnifiedTabs } from "./UnifiedTabs";

function Demo() {
  const [value, setValue] = useState("basic");
  return (
    <UnifiedTabs
      value={value}
      onValueChange={setValue}
      items={[
        { value: "basic", label: "基本情報" },
        { value: "medical", label: "診療履歴" },
        { value: "billing", label: "会計" },
      ]}
    >
      <div className="p-4">選択中: {value}</div>
    </UnifiedTabs>
  );
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/UnifiedTabs",
  component: UnifiedTabs,
  tags: ["autodocs"],
  args: {
    value: "basic",
    onValueChange: () => {},
    items: [{ value: "basic", label: "基本情報" }],
  },
} satisfies Meta<typeof UnifiedTabs>;

type Story = StoryObj<typeof UnifiedTabs>;

export const Interactive: Story = {
  render: () => <Demo />,
};
