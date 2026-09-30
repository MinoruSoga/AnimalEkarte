import type { Meta, StoryObj } from "@storybook/react-vite";

import { Calendar } from "./calendar";

/** react-day-picker ラッパー。予約・来院日選択に使う月カレンダー。 */
const meta = {
  title: "UI/Calendar",
  component: Calendar,
  tags: ["autodocs"],
} satisfies Meta<typeof Calendar>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <Calendar mode="single" defaultMonth={new Date(2026, 8, 1)} />,
};

export const WithSelection: Story = {
  render: () => (
    <Calendar mode="single" selected={new Date(2026, 8, 15)} defaultMonth={new Date(2026, 8, 1)} />
  ),
};

/** 診療日など範囲外を選択不可にするパターン（休診日 = 水曜の例）。 */
export const WithDisabledDays: Story = {
  render: () => (
    <Calendar mode="single" defaultMonth={new Date(2026, 8, 1)} disabled={{ dayOfWeek: [3] }} />
  ),
};
