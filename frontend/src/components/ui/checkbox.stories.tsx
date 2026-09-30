import type { Meta, StoryObj } from "@storybook/react-vite";

import { Checkbox } from "./checkbox";
import { Label } from "./label";

const meta = {
  title: "UI/Checkbox",
  component: Checkbox,
  tags: ["autodocs"],
  argTypes: {
    checked: { control: "boolean" },
    touchTarget: { control: "boolean" },
    disabled: { control: "boolean" },
  },
} satisfies Meta<typeof Checkbox>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <Checkbox id="cb-default" />
      <Label htmlFor="cb-default" className="mb-0">
        ワクチン済み
      </Label>
    </div>
  ),
};

export const States: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 12 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Checkbox id="cb-off" />
        <Label htmlFor="cb-off" className="mb-0">
          unchecked
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Checkbox id="cb-on" defaultChecked />
        <Label htmlFor="cb-on" className="mb-0">
          checked
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Checkbox id="cb-disabled" disabled />
        <Label htmlFor="cb-disabled" className="mb-0">
          disabled
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Checkbox id="cb-invalid" aria-invalid="true" />
        <Label htmlFor="cb-invalid" className="mb-0">
          aria-invalid
        </Label>
      </div>
    </div>
  ),
};

/** touchTarget=true — 44px ヒットエリアラッパー（design-states.md §1）。 */
export const TouchTarget: Story = {
  render: () => (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <Checkbox id="cb-touch" touchTarget defaultChecked />
      <Label htmlFor="cb-touch" className="mb-0">
        タッチ領域 44px（touchTarget）
      </Label>
    </div>
  ),
};
