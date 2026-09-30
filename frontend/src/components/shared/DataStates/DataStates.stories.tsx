import type { Meta, StoryObj } from "@storybook/react-vite";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState, ErrorFallback, LoadingFallback } from "./DataStates";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DataStates",
  component: EmptyState,
  tags: ["autodocs"],
} satisfies Meta<typeof EmptyState>;

type Story = StoryObj<typeof EmptyState>;

export const Loading: Story = {
  render: () => <LoadingFallback />,
};

export const Error_: Story = {
  name: "Error",
  render: () => <ErrorFallback />,
};

export const ErrorCustomMessage: Story = {
  name: "Error（カスタムメッセージ）",
  render: () => <ErrorFallback message="診察記録の取得に失敗しました" />,
};

export const Empty: Story = {
  args: {
    message: "記録がありません",
    description: "条件を変更して再検索してください。",
    icon: <Search className="size-8" />,
  },
};

export const EmptyWithAction: Story = {
  args: {
    message: "カルテがありません",
    children: <Button size="sm">記録を作成</Button>,
  },
};
