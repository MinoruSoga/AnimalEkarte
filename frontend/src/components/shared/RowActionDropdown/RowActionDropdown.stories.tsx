import type { Meta, StoryObj } from "@storybook/react-vite";
import { Pencil, Trash2, Copy } from "lucide-react";
import { RowActionDropdown } from "./RowActionDropdown";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/RowActionDropdown",
  component: RowActionDropdown,
  tags: ["autodocs"],
  args: {
    ariaLabel: "行操作",
    actions: [
      { label: "編集", icon: Pencil, onClick: () => {} },
      { label: "複製", icon: Copy, onClick: () => {} },
      { label: "削除", icon: Trash2, onClick: () => {}, variant: "destructive" },
    ],
  },
} satisfies Meta<typeof RowActionDropdown>;

type Story = StoryObj<typeof RowActionDropdown>;

export const Default: Story = {};

export const SingleAction: Story = {
  args: { actions: [{ label: "編集", icon: Pencil, onClick: () => {} }] },
};
