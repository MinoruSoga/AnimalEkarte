import type { Meta, StoryObj } from "@storybook/react-vite";
import { DeleteIconButton } from "./DeleteIconButton";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DeleteIconButton",
  component: DeleteIconButton,
  tags: ["autodocs"],
  args: { onClick: () => {} },
} satisfies Meta<typeof DeleteIconButton>;

type Story = StoryObj<typeof DeleteIconButton>;

export const Default: Story = {};

export const Disabled: Story = {
  args: { disabled: true },
};
