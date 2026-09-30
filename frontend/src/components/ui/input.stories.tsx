import type { Meta, StoryObj } from "@storybook/react-vite";

import { Input } from "./input";
import { Label } from "./label";

/** design-states.md §3 の状態（hover/focus/disabled/invalid）を確認するカタログ。 */
const meta = {
  title: "UI/Input",
  component: Input,
  tags: ["autodocs"],
  argTypes: {
    disabled: { control: "boolean" },
    placeholder: { control: "text" },
  },
  args: { placeholder: "飼主名を入力" },
} satisfies Meta<typeof Input>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithLabel: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <Label htmlFor="owner-name">飼主名</Label>
      <Input id="owner-name" placeholder="例: 山田 太郎" />
    </div>
  ),
};

/** aria-invalid=true で destructive border + bg/5（design-states.md §3 invalid）。 */
export const Invalid: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <Label htmlFor="invalid-input">電話番号</Label>
      <Input id="invalid-input" aria-invalid="true" defaultValue="090-" />
      <p className="text-sm text-destructive">ハイフンを含めて入力してください</p>
    </div>
  ),
};

export const Disabled: Story = {
  args: { disabled: true, defaultValue: "読み取り専用" },
};
