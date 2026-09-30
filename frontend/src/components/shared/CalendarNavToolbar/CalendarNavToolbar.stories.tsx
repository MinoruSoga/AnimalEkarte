import type { Meta, StoryObj } from "@storybook/react-vite";
import { CalendarNavToolbar } from "./CalendarNavToolbar";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/CalendarNavToolbar",
  component: CalendarNavToolbar,
  tags: ["autodocs"],
  args: {
    label: "2026年9月",
    onPrev: () => {},
    onNext: () => {},
    onToday: () => {},
    prevAriaLabel: "前月",
    nextAriaLabel: "翌月",
  },
} satisfies Meta<typeof CalendarNavToolbar>;

type Story = StoryObj<typeof CalendarNavToolbar>;

export const Default: Story = {};

export const WithoutToday: Story = {
  args: { onToday: undefined },
};

export const DayLabel: Story = {
  args: { label: "2026-09-30（水）", todayLabel: "今日" },
};
