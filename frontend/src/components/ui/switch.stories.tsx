import type { Meta, StoryObj } from "@storybook/react-vite";

import { Switch } from "./switch";
import { Label } from "./label";

const meta = {
  title: "UI/Switch",
  component: Switch,
  tags: ["autodocs"],
  argTypes: {
    checked: { control: "boolean" },
    disabled: { control: "boolean" },
  },
} satisfies Meta<typeof Switch>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <Switch id="sw-default" />
      <Label htmlFor="sw-default" className="mb-0">
        リマインダー送信
      </Label>
    </div>
  ),
};

export const States: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 12 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Switch id="sw-on" defaultChecked />
        <Label htmlFor="sw-on" className="mb-0">
          on
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Switch id="sw-off" />
        <Label htmlFor="sw-off" className="mb-0">
          off
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Switch id="sw-dis" disabled defaultChecked />
        <Label htmlFor="sw-dis" className="mb-0">
          disabled
        </Label>
      </div>
    </div>
  ),
};
