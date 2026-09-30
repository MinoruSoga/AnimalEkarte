import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { CategoryChipsFilter } from "./CategoryChipsFilter";

function Demo({ chipRounded }: { chipRounded?: "sm" | "md" }) {
  const [active, setActive] = useState<string | null>(null);
  return (
    <CategoryChipsFilter
      categories={["診療", "処置", "ワクチン", "検査"]}
      activeCategory={active}
      onSelectCategory={setActive}
      chipRounded={chipRounded}
    />
  );
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/CategoryChipsFilter",
  component: CategoryChipsFilter,
  tags: ["autodocs"],
} satisfies Meta<typeof CategoryChipsFilter>;

type Story = StoryObj<typeof CategoryChipsFilter>;

export const Default: Story = {
  render: () => <Demo />,
};

export const RoundedSm: Story = {
  render: () => <Demo chipRounded="sm" />,
};
