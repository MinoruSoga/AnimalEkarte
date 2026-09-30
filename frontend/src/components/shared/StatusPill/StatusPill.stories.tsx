import type { Meta, StoryObj } from "@storybook/react-vite";
import { StatusPill } from "./StatusPill";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/StatusPill",
  component: StatusPill,
  tags: ["autodocs"],
  args: { isActive: true },
} satisfies Meta<typeof StatusPill>;

type Story = StoryObj<typeof StatusPill>;

export const Active: Story = {};

export const Inactive: Story = {
  args: { isActive: false },
};

export const Both: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <StatusPill isActive />
      <StatusPill isActive={false} />
    </div>
  ),
};
