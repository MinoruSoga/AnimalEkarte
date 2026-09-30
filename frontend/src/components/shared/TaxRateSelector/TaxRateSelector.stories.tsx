import type { Meta, StoryObj } from "@storybook/react-vite";
import { TaxRateSelector } from "./TaxRateSelector";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/TaxRateSelector",
  component: TaxRateSelector,
  tags: ["autodocs"],
  args: { value: 10, onChange: () => {} },
} satisfies Meta<typeof TaxRateSelector>;

type Story = StoryObj<typeof TaxRateSelector>;

export const Default: Story = {};

export const ReducedRate: Story = {
  args: { value: 8 },
};

export const Disabled: Story = {
  args: { disabled: true },
};
