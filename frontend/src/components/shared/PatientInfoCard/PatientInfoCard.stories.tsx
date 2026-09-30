import type { Meta, StoryObj } from "@storybook/react-vite";
import { PatientInfoCard } from "./PatientInfoCard";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PatientInfoCard",
  component: PatientInfoCard,
  tags: ["autodocs"],
  args: {
    ownerName: "山田 花子",
    petName: "ポチ",
    petNumber: "P-0123",
    weight: "5.2kg",
    status: "alive",
    staffName: "佐藤 一郎",
    reservationType: "再診",
  },
} satisfies Meta<typeof PatientInfoCard>;

type Story = StoryObj<typeof PatientInfoCard>;

export const Default: Story = {};

export const Deceased: Story = {
  args: { status: "deceased" },
};

export const WithoutStaff: Story = {
  args: { hideStaff: true },
};

export const Minimal: Story = {
  args: { staffName: undefined, reservationType: undefined },
};
