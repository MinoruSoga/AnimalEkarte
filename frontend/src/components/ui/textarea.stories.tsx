import type { Meta, StoryObj } from "@storybook/react-vite";

import { Textarea } from "./textarea";
import { Label } from "./label";

const meta = {
  title: "UI/Textarea",
  component: Textarea,
  tags: ["autodocs"],
  argTypes: {
    disabled: { control: "boolean" },
    placeholder: { control: "text" },
  },
  args: { placeholder: "所見を入力" },
} satisfies Meta<typeof Textarea>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Invalid: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 8, maxWidth: 420 }}>
      <Label htmlFor="invalid-textarea">症状メモ</Label>
      <Textarea id="invalid-textarea" aria-invalid="true" defaultValue="（必須項目が未入力）" />
    </div>
  ),
};

export const Disabled: Story = {
  args: { disabled: true, defaultValue: "確定済みのカルテテキスト" },
};
