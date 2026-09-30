import type { Meta, StoryObj } from "@storybook/react-vite";
import { NumberInput } from "./NumberInput";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/NumberInput",
  component: NumberInput,
  tags: ["autodocs"],
  args: {
    value: 1500,
    onChange: () => {},
    suffix: "円",
  },
} satisfies Meta<typeof NumberInput>;

type Story = StoryObj<typeof NumberInput>;

export const Default: Story = {};

export const WithoutSuffix: Story = {
  args: { suffix: undefined, placeholder: "数量" },
};

export const Disabled: Story = {
  args: { disabled: true },
};

export const Invalid: Story = {
  args: { "aria-invalid": true },
};
