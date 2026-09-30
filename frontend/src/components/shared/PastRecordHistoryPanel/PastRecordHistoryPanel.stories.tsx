import type { Meta, StoryObj } from "@storybook/react-vite";
import { PastRecordHistoryPanel } from "./PastRecordHistoryPanel";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PastRecordHistoryPanel",
  component: PastRecordHistoryPanel,
  tags: ["autodocs"],
  args: {
    title: "過去のカルテ",
    searchPlaceholder: "タイトルで検索",
    items: [
      { id: "1", date: "2026-08-12", title: "定期健診", subtitle: "体重 5.2kg" },
      { id: "2", date: "2026-05-03", title: "ワクチン接種", subtitle: "混合ワクチン" },
      { id: "3", date: "2025-11-20", title: "皮膚炎", subtitle: "外用薬処方" },
    ],
  },
} satisfies Meta<typeof PastRecordHistoryPanel>;

type Story = StoryObj<typeof PastRecordHistoryPanel>;

export const Default: Story = {};

export const Loading: Story = {
  args: { isLoading: true },
};

export const Empty: Story = {
  args: { items: [] },
};
