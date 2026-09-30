import type { Meta, StoryObj } from "@storybook/react-vite";
import { MasterSelectModal } from "./MasterSelectModal";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/MasterSelectModal",
  component: MasterSelectModal,
  tags: ["autodocs"],
  args: {
    open: true,
    onOpenChange: () => {},
    onSelect: () => {},
    title: "診療項目を選択",
    items: [
      { id: 1, name: "初診料", price: 1500 },
      { id: 2, name: "再診料", price: 800 },
      { id: 3, name: "ワクチン接種", price: 5500 },
      { id: 4, name: "血液検査", price: 8000 },
    ],
  },
} satisfies Meta<typeof MasterSelectModal>;

type Story = StoryObj<typeof MasterSelectModal>;

export const Default: Story = {};

export const Selected: Story = {
  args: { selectedValue: "ワクチン接種" },
};

export const EmptyItems: Story = {
  args: { items: [] },
};
