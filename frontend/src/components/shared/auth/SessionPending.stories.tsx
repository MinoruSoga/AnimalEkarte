import type { Meta, StoryObj } from "@storybook/react-vite";
import { SessionPending } from "./SessionPending";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/auth/SessionPending",
  component: SessionPending,
  tags: ["autodocs"],
  args: { message: "セッションを確認しています…" },
} satisfies Meta<typeof SessionPending>;

type Story = StoryObj<typeof SessionPending>;

export const Default: Story = {};
