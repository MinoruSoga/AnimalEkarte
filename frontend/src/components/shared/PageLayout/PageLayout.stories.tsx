import type { Meta, StoryObj } from "@storybook/react-vite";
import { PawPrint } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageLayout } from "./PageLayout";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PageLayout",
  component: PageLayout,
  tags: ["autodocs"],
  args: {
    title: "カルテ一覧",
    children: (
      <div className="rounded-lg border border-dashed p-12 text-center text-muted-foreground">
        ページ本文
      </div>
    ),
  },
} satisfies Meta<typeof PageLayout>;

type Story = StoryObj<typeof PageLayout>;

export const Default: Story = {};

export const WithDescription: Story = {
  args: { description: "全患者の診察記録を表示します" },
};

export const WithBackAndAction: Story = {
  args: {
    onBack: () => {},
    icon: <PawPrint className="size-6" />,
    headerAction: <Button size="sm">新規作成</Button>,
  },
};

export const LeftAligned: Story = {
  args: { align: "left", description: "左寄せレイアウト" },
};
