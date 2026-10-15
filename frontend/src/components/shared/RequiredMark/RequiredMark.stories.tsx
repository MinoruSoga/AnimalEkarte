import type { Meta, StoryObj } from "@storybook/react-vite";
import { RequiredMark } from "./RequiredMark";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/RequiredMark",
  component: RequiredMark,
  tags: ["autodocs"],
} satisfies Meta<typeof RequiredMark>;

type Story = StoryObj<typeof RequiredMark>;

export const Default: Story = {};

/** フォームのラベル末尾に置く実使用イメージ（`*` は aria-hidden でラベルの読み上げに含まれない） */
export const InLabel: Story = {
  render: () => (
    <span className="text-base">
      品目名 <RequiredMark />
    </span>
  ),
};
