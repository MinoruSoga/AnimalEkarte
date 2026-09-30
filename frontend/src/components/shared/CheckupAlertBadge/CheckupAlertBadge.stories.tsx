import type { Meta, StoryObj } from "@storybook/react-vite";
import { CheckupAlertBadge } from "./CheckupAlertBadge";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/CheckupAlertBadge",
  component: CheckupAlertBadge,
  tags: ["autodocs"],
  args: { nextDate: "2026-10-01" },
} satisfies Meta<typeof CheckupAlertBadge>;

type Story = StoryObj<typeof CheckupAlertBadge>;

/** 次回予定が近い（期限間近の警告表示） */
export const DueSoon: Story = {
  args: { nextDate: "2026-10-01" },
};

export const Future: Story = {
  args: { nextDate: "2027-03-15" },
};

export const Overdue: Story = {
  args: { nextDate: "2025-01-10" },
};

export const NoDate: Story = {
  args: { nextDate: null },
};
