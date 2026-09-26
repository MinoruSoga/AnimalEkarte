import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { OwnersListTable } from "./OwnersListTable";
import type { Pet } from "@/types";

// EMR-174: ペット名クリック → 飼主詳細 ?pet= deep link（既存ペットモーダルを開く）。
// 飼主名リンクと同じく、別医院行にはリンクを出さない（医院横断の閲覧専用境界を保持）。

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));
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
  status: "生存",
  clinicId: "clinic-1",
} as unknown as Pet;

type Props = React.ComponentProps<typeof OwnersListTable>;

function baseProps(overrides: Partial<Props> = {}): Props {
  return {
    pets: [pet],
    pagination: {
      totalPages: 1,
      totalCount: 1,
      startIndex: 0,
      endIndex: 1,
      currentPage: 1,
    },
    searchTerm: "",
    activeFilters: [],
    filterProperties: [],
    isFiltering: false,
    canEdit: true,
    canDelete: true,
    canReport: true,
    onSearchChange: vi.fn(),
    onFilterChange: vi.fn(),
    onEdit: vi.fn(),
    onDeleteRequest: vi.fn(),
    onReport: vi.fn(),
    onPageChange: vi.fn(),
    ...overrides,
  };
}

function renderTable(overrides: Partial<Props> = {}) {
  return render(
    <MemoryRouter initialEntries={["/owners"]}>
      <OwnersListTable {...baseProps(overrides)} />
    </MemoryRouter>,
  );
}

describe("OwnersListTable pet detail deep link (EMR-174)", () => {
  it("ペット名は /owners/:ownerId?pet=:petId の native link になる", () => {
    renderTable({ currentClinicId: "clinic-1" });

    const petLink = screen.getByRole("link", { name: "ペット「ポチ」(ID: 7) の詳細を開く" });
    expect(petLink).toHaveAttribute("href", "/owners/42?pet=7");
  });

  it("飼主名 link とペット名 link は別の遷移先を持つ", () => {
    renderTable({ currentClinicId: "clinic-1" });

    const ownerLink = screen.getByRole("link", { name: /飼主「山田太郎」/ });
    expect(ownerLink).toHaveAttribute("href", "/owners/42");
  });

  it("view-only ユーザーにもペット名の detail link を表示する", () => {
    renderTable({
      canEdit: false,
      canDelete: false,
      canReport: false,
      currentClinicId: "clinic-1",
    });

    expect(
      screen.getByRole("link", { name: "ペット「ポチ」(ID: 7) の詳細を開く" }),
    ).toHaveAttribute("href", "/owners/42?pet=7");
  });

  it("別医院行はペット名を detail link にしない（閲覧専用境界）", () => {
    renderTable({
      canEdit: true,
      currentClinicId: "clinic-1",
      pets: [{ ...pet, clinicId: "clinic-2" }],
    });

    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "ペット「ポチ」(ID: 7) の詳細を開く" }),
    ).not.toBeInTheDocument();
  });
});
