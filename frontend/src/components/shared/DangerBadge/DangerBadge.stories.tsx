import type { Meta, StoryObj } from "@storybook/react-vite";
import { DangerBadge } from "./DangerBadge";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DangerBadge",
  component: DangerBadge,
  tags: ["autodocs"],
  args: {
    variant: "pet",
    level: "high",
    subjectName: "ポチ",
    reason: "噛みつき歴あり（2025年受診時）",
  },
} satisfies Meta<typeof DangerBadge>;

type Story = StoryObj<typeof DangerBadge>;

export const PetHigh: Story = {};

export const PetMedium: Story = {
  args: { level: "medium" },
};

/** 低・未設定は何も描画しない（臨床 sentinel の契約） */
export const PetLow: Story = {
  args: { level: "low" },
};

export const PetNoReason: Story = {
  args: { level: "high", reason: null },
};

export const Owner: Story = {
  args: { variant: "owner" },
};
