import type { Meta, StoryObj } from "@storybook/react-vite";
import { CharCountTextarea } from "./CharCountTextarea";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/CharCountTextarea",
  component: CharCountTextarea,
  tags: ["autodocs"],
  args: {
    value: "食欲不振が3日続いている。",
    onChange: () => {},
    maxLength: 200,
    placeholder: "症状を入力",
  },
} satisfies Meta<typeof CharCountTextarea>;

type Story = StoryObj<typeof CharCountTextarea>;

export const Default: Story = {};

export const NearLimit: Story = {
  args: { value: "あ".repeat(190) },
};

export const AtLimit: Story = {
  args: { value: "あ".repeat(200) },
};

export const Disabled: Story = {
  args: { disabled: true },
};
