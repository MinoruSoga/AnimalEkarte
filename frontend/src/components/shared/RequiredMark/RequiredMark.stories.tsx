import type { Meta, StoryObj } from "@storybook/react-vite";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { RequiredMark } from "./RequiredMark";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/RequiredMark",
  component: RequiredMark,
  tags: ["autodocs"],
} satisfies Meta<typeof RequiredMark>;

type Story = StoryObj<typeof RequiredMark>;

export const Default: Story = {};

export const RequiredField: Story = {
  render: () => (
    <div className="space-y-2">
      <Label htmlFor="required-mark-owner">
        飼主名
        <RequiredMark />
      </Label>
      <Input id="required-mark-owner" required placeholder="入力例" />
    </div>
  ),
};
