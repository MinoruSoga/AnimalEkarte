import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { PropertyFilter } from "./PropertyFilter";
import type { ActiveFilter } from "./types";
import type { FilterProperty } from "./types";

const PROPERTIES: FilterProperty[] = [
  {
    key: "species",
    label: "種別",
    type: "select",
    options: [
      { value: "dog", label: "犬" },
      { value: "cat", label: "猫" },
      { value: "other", label: "その他" },
    ],
  },
  {
    key: "visit_date",
    label: "来院日",
    type: "date-range",
  },
];

function ControlledPropertyFilter({ initialFilters = [] }: { initialFilters?: ActiveFilter[] }) {
  const [filters, setFilters] = useState<ActiveFilter[]>(initialFilters);
  return (
    <PropertyFilter properties={PROPERTIES} activeFilters={filters} onFilterChange={setFilters} />
  );
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PropertyFilter",
  component: PropertyFilter,
  tags: ["autodocs"],
} satisfies Meta<typeof PropertyFilter>;

type Story = StoryObj<typeof PropertyFilter>;

export const Empty: Story = {
  render: () => <ControlledPropertyFilter />,
};

export const WithActiveFilter: Story = {
  render: () => (
    <ControlledPropertyFilter
      initialFilters={[{ key: "species", condition: "is", value: "dog", displayValue: "犬" }]}
    />
  ),
};
