import type { Meta, StoryObj } from "@storybook/react-vite";
import { ChangePasswordDialog } from "./ChangePasswordDialog";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ChangePasswordDialog",
  component: ChangePasswordDialog,
  tags: ["autodocs"],
  args: {
    open: true,
    onOpenChange: () => {},
  },
} satisfies Meta<typeof ChangePasswordDialog>;

type Story = StoryObj<typeof ChangePasswordDialog>;

/** フォーム初期状態（送信は API 呼び出しのため catalog 上では実行しない） */
export const Default: Story = {};
