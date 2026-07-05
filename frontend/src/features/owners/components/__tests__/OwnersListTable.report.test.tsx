import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { OwnersListTable } from "../OwnersListTable";
import { STYLE } from "@/lib/design-tokens";
import type { Pet } from "@/types";

// 行アクション(RowActionDropdown)の表示・発火に焦点を当てるため、無関係な重い子は無効化する。
vi.mock("@/components/shared/NotionFilter/NotionFilter", () => ({ NotionFilter: () => null }));
vi.mock("@/components/shared/Pagination", () => ({ Pagination: () => null }));
vi.mock("@/components/shared/FilteringIndicator/FilteringIndicator", () => ({
  FilteringIndicator: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

const pet = {
  id: "7",
  ownerId: "42",
  ownerName: "山田太郎",
  name: "ポチ",
  species: "犬",
} as unknown as Pet;

type Props = React.ComponentProps<typeof OwnersListTable>;

function baseProps(overrides: Partial<Props> = {}): Props {
  return {
    filteredCount: 1,
    pagination: {
      paginatedData: [pet],
      totalPages: 1,
      totalCount: 1,
      startIndex: 0,
      endIndex: 1,
      currentPage: 1,
    },
    searchTerm: "",
    activeFilters: [],
    activeSorts: [],
    isFiltering: false,
    canEdit: true,
    canDelete: true,
    canReport: true,
    directionFor: () => "none",
    onSearchChange: vi.fn(),
    onFilterChange: vi.fn(),
    onSortChange: vi.fn(),
    onToggleSort: vi.fn(),
    onRowClick: vi.fn(),
    onEdit: vi.fn(),
    onDeleteRequest: vi.fn(),
    onReport: vi.fn(),
    onPageChange: vi.fn(),
    ...overrides,
  };
}

describe("OwnersListTable レポート行アクション (#158)", () => {
  it("canReport=true で『レポート』が出て、click で onReport(ownerId, petId) が発火する", async () => {
    const user = userEvent.setup();
    const onReport = vi.fn();
    render(<OwnersListTable {...baseProps({ onReport })} />);

    await user.click(screen.getByRole("button", { name: "操作" }));
    await user.click(await screen.findByRole("menuitem", { name: /レポート/ }));

    expect(onReport).toHaveBeenCalledWith("42", "7");
  });

  it("canReport=false では『レポート』を出さない(編集・削除は残る)", async () => {
    const user = userEvent.setup();
    render(<OwnersListTable {...baseProps({ canReport: false })} />);

    await user.click(screen.getByRole("button", { name: "操作" }));

    expect(await screen.findByRole("menuitem", { name: /編集/ })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /レポート/ })).not.toBeInTheDocument();
  });
});

describe("OwnersListTable header (DESIGN.md ex-data-table-cell)", () => {
  it("ソート可能な列見出しが eyebrow 相当（STYLE.sectionLabel）で表示される", () => {
    render(<OwnersListTable {...baseProps()} />);

    const sortButton = screen.getByRole("button", { name: "飼主名でソート" });
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(sortButton.className).toContain(cls);
    }
  });
});
