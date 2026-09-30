import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MoneyInput } from "./MoneyInput";
import { PropertyRow } from "./PropertyRow";
import { StatusToggleButton } from "./StatusToggleButton";

function MoneyInputDemo() {
  const [value, setValue] = useState(12000);
  return <MoneyInput label="単価" value={value} onChange={setValue} />;
}

function StatusToggleDemo() {
  const [active, setActive] = useState(true);
  return <StatusToggleButton isActive={active} onToggle={() => setActive((v) => !v)} />;
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/SidePeek",
  component: PropertyRow,
  tags: ["autodocs"],
} satisfies Meta<typeof PropertyRow>;

type Story = StoryObj<typeof PropertyRow>;

export const Rows: Story = {
  render: () => (
    <div className="w-[360px]">
      <PropertyRow label="名前">ポチ</PropertyRow>
      <PropertyRow label="単価">
        <MoneyInputDemo />
      </PropertyRow>
      <PropertyRow label="ステータス">
        <StatusToggleDemo />
      </PropertyRow>
    </div>
  ),
};

export const MoneyInputError: Story = {
  render: () => (
    <MoneyInput
      label="単価"
      value={-100}
      onChange={() => {}}
      error="0 以上の値を入力してください"
    />
  ),
};
