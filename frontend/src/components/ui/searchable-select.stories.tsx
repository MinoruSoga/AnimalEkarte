import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";

import { SearchableSelect, type SearchableSelectOption } from "./searchable-select";
import { Label } from "./label";

const OPTIONS: SearchableSelectOption[] = [
  { value: "shoinsatsu", label: "初診料", keywords: ["しょしん"] },
  { value: "saiinsatsu", label: "再診料", keywords: ["さいしん"] },
  { value: "vaccine", label: "ワクチン接種" },
  { value: "nail", label: "爪切り" },
  { value: "ear", label: "耳掃除", disabled: true },
];

const meta = {
  title: "UI/SearchableSelect",
  component: SearchableSelect,
  tags: ["autodocs"],
  args: { value: "", onValueChange: () => {} },
} satisfies Meta<typeof SearchableSelect>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

function ControlledDemo(props: { placeholder?: string; disabled?: boolean; clearable?: boolean }) {
  const [value, setValue] = useState("");
  return (
    <div style={{ maxWidth: 320 }}>
      <SearchableSelect
        value={value}
        onValueChange={setValue}
        options={OPTIONS}
        placeholder={props.placeholder}
        disabled={props.disabled}
        clearable={props.clearable}
      />
    </div>
  );
}

export const Default: Story = {
  render: () => <ControlledDemo />,
};

function WithLabelDemo() {
  const [value, setValue] = useState("vaccine");
  return (
    <div style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <Label htmlFor="treatment">診療項目</Label>
      <SearchableSelect
        id="treatment"
        value={value}
        onValueChange={setValue}
        options={OPTIONS}
        clearable
      />
    </div>
  );
}

export const WithLabel: Story = {
  render: () => <WithLabelDemo />,
};

export const Disabled: Story = {
  render: () => <ControlledDemo disabled />,
};

/** キーワード検索（かな別名）が効くことを確認する story。 */
export const KeywordSearch: Story = {
  render: () => <ControlledDemo placeholder="かな検索も可（例: しょしん）" />,
};
