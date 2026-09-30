import type { Meta, StoryObj } from "@storybook/react-vite";
import { ConfirmDialog } from "./ConfirmDialog";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ConfirmDialog",
  component: ConfirmDialog,
  tags: ["autodocs"],
  args: {
    open: true,
    onClose: () => {},
    onConfirm: () => {},
    title: "この項目を削除しますか？",
    description: "この操作は取り消せません。",
  },
} satisfies Meta<typeof ConfirmDialog>;

type Story = StoryObj<typeof ConfirmDialog>;

export const Default: Story = {};

export const Destructive: Story = {
  args: { variant: "destructive", confirmLabel: "削除する" },
};

export const Pending: Story = {
  args: { isPending: true, confirmLabel: "削除する" },
};
