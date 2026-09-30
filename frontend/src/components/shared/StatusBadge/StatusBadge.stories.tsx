import type { Meta, StoryObj } from "@storybook/react-vite";
import { BADGE } from "@/lib/design-tokens";
import { StatusBadge } from "./StatusBadge";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/StatusBadge",
  component: StatusBadge,
  tags: ["autodocs"],
  args: {
    children: "受付中",
    colorClass: BADGE.green,
  },
} satisfies Meta<typeof StatusBadge>;

type Story = StoryObj<typeof StatusBadge>;

export const Green: Story = {};

export const Yellow: Story = {
  args: { children: "依頼中", colorClass: BADGE.yellow },
};

export const Red: Story = {
  args: { children: "期限切れ", colorClass: BADGE.red },
};

export const Gray: Story = {
  args: { children: "終了", colorClass: BADGE.gray },
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-2">
      <StatusBadge colorClass={BADGE.green}>受付中</StatusBadge>
      <StatusBadge colorClass={BADGE.yellow}>依頼中</StatusBadge>
      <StatusBadge colorClass={BADGE.orange}>進行中</StatusBadge>
      <StatusBadge colorClass={BADGE.red}>期限切れ</StatusBadge>
      <StatusBadge colorClass={BADGE.muted}>未対応</StatusBadge>
      <StatusBadge colorClass={BADGE.gray}>終了</StatusBadge>
    </div>
  ),
};
