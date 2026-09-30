import type { Meta, StoryObj } from "@storybook/react-vite";
import { DatePicker } from "./DatePicker";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DatePicker",
  component: DatePicker,
  tags: ["autodocs"],
  args: {
    mode: "single",
    value: "2026-09-30",
    onChange: () => {},
  },
} satisfies Meta<typeof DatePicker>;

type Story = StoryObj<typeof DatePicker>;

export const Single: Story = {};

export const SingleEmpty: Story = {
  args: { value: "", placeholder: "日付を選択" },
};

export const Range: Story = {
  args: { mode: "range", value: "2026-09-01 ~ 2026-09-30" },
};
