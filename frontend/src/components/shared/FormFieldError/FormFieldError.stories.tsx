import type { Meta, StoryObj } from "@storybook/react-vite";
import { FormFieldError } from "./FormFieldError";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/FormFieldError",
  component: FormFieldError,
  tags: ["autodocs"],
  args: { message: "必須項目です" },
} satisfies Meta<typeof FormFieldError>;

type Story = StoryObj<typeof FormFieldError>;

export const WithMessage: Story = {};

export const NoMessage: Story = {
  args: { message: undefined },
};
