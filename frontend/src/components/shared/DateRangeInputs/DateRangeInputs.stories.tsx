import type { Meta, StoryObj } from "@storybook/react-vite";
import { DateRangeInputs } from "./DateRangeInputs";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DateRangeInputs",
  component: DateRangeInputs,
  tags: ["autodocs"],
  args: {
    fromValue: "2026-09-01",
    toValue: "2026-09-30",
    onFromChange: () => {},
    onToChange: () => {},
  },
} satisfies Meta<typeof DateRangeInputs>;

type Story = StoryObj<typeof DateRangeInputs>;

export const Default: Story = {};

export const EmptyState: Story = {
  args: { fromValue: "", toValue: "" },
};
