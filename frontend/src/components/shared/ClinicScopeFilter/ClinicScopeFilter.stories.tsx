import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import type { ClinicMembership } from "@/types/auth";
import { ClinicScopeFilter } from "./ClinicScopeFilter";

const CLINICS: ClinicMembership[] = [
  { clinicId: "1", clinicName: "本院", isMain: true },
  { clinicId: "2", clinicName: "分院（北）", isMain: false },
  { clinicId: "3", clinicName: "分院（南）", isMain: false },
];

function Demo({ initial = ["1"] }: { initial?: string[] }) {
  const [selected, setSelected] = useState(initial);
  return (
    <ClinicScopeFilter
      clinics={CLINICS}
      selectedIds={selected}
      onToggle={(id) =>
        setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
      }
    />
  );
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ClinicScopeFilter",
  component: ClinicScopeFilter,
  tags: ["autodocs"],
} satisfies Meta<typeof ClinicScopeFilter>;

type Story = StoryObj<typeof ClinicScopeFilter>;

export const Default: Story = {
  render: () => <Demo />,
};

export const AllSelected: Story = {
  render: () => <Demo initial={["1", "2", "3"]} />,
};
