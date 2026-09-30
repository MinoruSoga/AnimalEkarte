import type { Meta, StoryObj } from "@storybook/react-vite";
import { FieldHelp } from "./FieldHelp";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/FieldHelp",
  component: FieldHelp,
  tags: ["autodocs"],
  args: {
    label: "単価",
    content: "税抜きの単価を入力します。\n0 円も許可されます。",
  },
} satisfies Meta<typeof FieldHelp>;

type Story = StoryObj<typeof FieldHelp>;

export const Default: Story = {};

/** フォーム行内での実使用イメージ（ラベルの外・兄弟要素として配置） */
export const InFormRow: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <span className="text-base">単価</span>
      <FieldHelp label="単価" content="税抜きの単価を入力します。" />
    </div>
  ),
};
