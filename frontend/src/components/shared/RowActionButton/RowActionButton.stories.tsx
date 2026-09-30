import type { Meta, StoryObj } from "@storybook/react-vite";
import { Trash2, Eye } from "lucide-react";
import { RowActionButton } from "./RowActionButton";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/RowActionButton",
  component: RowActionButton,
  tags: ["autodocs"],
  args: {
    "aria-label": "編集",
    onClick: () => {},
  },
} satisfies Meta<typeof RowActionButton>;

type Story = StoryObj<typeof RowActionButton>;

export const Default: Story = {};

export const CustomIcon: Story = {
  args: { icon: Eye, "aria-label": "詳細を見る" },
};

export const DeleteIcon: Story = {
  args: { icon: Trash2, "aria-label": "削除" },
};
