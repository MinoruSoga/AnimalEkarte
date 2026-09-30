import type { Meta, StoryObj } from "@storybook/react-vite";
import { SortableHeader } from "./SortableHeader";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/SortableHeader",
  component: SortableHeader,
  tags: ["autodocs"],
  args: {
    label: "受診日",
    direction: "ascending",
    onToggle: () => {},
  },
} satisfies Meta<typeof SortableHeader>;

type Story = StoryObj<typeof SortableHeader>;

export const Asc: Story = {};

export const Desc: Story = {
  args: { direction: "descending" },
};

export const Eyebrow: Story = {
  args: { variant: "eyebrow" },
};
