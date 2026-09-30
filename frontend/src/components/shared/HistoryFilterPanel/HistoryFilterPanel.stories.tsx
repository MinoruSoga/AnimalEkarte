import type { Meta, StoryObj } from "@storybook/react-vite";
import { HistoryFilterPanel } from "./HistoryFilterPanel";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/HistoryFilterPanel",
  component: HistoryFilterPanel,
  tags: ["autodocs"],
  args: {
    showDateRange: true,
    filterStartDate: "2026-09-01",
    filterEndDate: "2026-09-30",
    onFilterStartDateChange: () => {},
    onFilterEndDateChange: () => {},
    searchTerm: "",
    onSearchTermChange: () => {},
    searchPlaceholder: "タイトル・メモで検索",
    sortOrder: "desc",
    onSortOrderChange: () => {},
    onClear: () => {},
  },
} satisfies Meta<typeof HistoryFilterPanel>;

type Story = StoryObj<typeof HistoryFilterPanel>;

export const Default: Story = {};

export const SearchOnly: Story = {
  args: { showDateRange: false },
};
