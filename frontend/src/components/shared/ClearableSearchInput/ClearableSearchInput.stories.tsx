import type { Meta, StoryObj } from "@storybook/react-vite";
import { ClearableSearchInput } from "./ClearableSearchInput";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ClearableSearchInput",
  component: ClearableSearchInput,
  tags: ["autodocs"],
  args: {
    value: "ポチ",
    onChange: () => {},
    placeholder: "ペット名・飼主名で検索",
  },
} satisfies Meta<typeof ClearableSearchInput>;

type Story = StoryObj<typeof ClearableSearchInput>;

/** 入力済み — クリアボタンが表示される */
export const WithValue: Story = {};

export const EmptyState: Story = {
  args: { value: "" },
};
