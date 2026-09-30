import type { Meta, StoryObj } from "@storybook/react-vite";
import { TaxTypeSelector } from "./TaxTypeSelector";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/TaxTypeSelector",
  component: TaxTypeSelector,
  tags: ["autodocs"],
  args: { value: "included", onChange: () => {} },
} satisfies Meta<typeof TaxTypeSelector>;

type Story = StoryObj<typeof TaxTypeSelector>;

export const Included: Story = {};

export const Excluded: Story = {
  args: { value: "excluded" },
};

export const Exempt: Story = {
  args: { value: "exempt" },
};

export const Disabled: Story = {
  args: { disabled: true },
};
