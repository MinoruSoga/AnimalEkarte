import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { OwnersListTable } from "./OwnersListTable";
import type { Pet } from "@/types";

// 行アクション(RowActionDropdown)の表示・発火に焦点を当てるため、無関係な重い子は無効化する。
vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({ PropertyFilter: () => null }));
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
  status: "死亡",
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

describe("OwnersListTable レポート行アクション (#158)", () => {
  it("canReport=true で『レポート』が出て、click で onReport(ownerId, petId) が発火する", async () => {
    const user = userEvent.setup();
    const onReport = vi.fn();
    renderTable({ onReport });

    await user.click(screen.getByRole("button", {
      name: /飼主.*山田太郎.*ID: 42.*ペット.*ポチ.*ID: 7.*操作/,
    }));
    await user.click(await screen.findByRole("menuitem", { name: /レポート/ }));

    expect(onReport).toHaveBeenCalledWith("42", "7");
  });

  it("canReport=false では『レポート』を出さない(編集・削除は残る)", async () => {
    const user = userEvent.setup();
    renderTable({ canReport: false });

    await user.click(screen.getByRole("button", {
      name: /飼主.*山田太郎.*ID: 42.*ペット.*ポチ.*ID: 7.*操作/,
    }));

    expect(await screen.findByRole("menuitem", { name: /編集/ })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /レポート/ })).not.toBeInTheDocument();
  });
});

describe("OwnersListTable header (#266: サーバサイドページネーション化でソート列は撤去)", () => {
  it("列見出しはプレーンテキストで表示される（ソートボタンは存在しない）", () => {
    renderTable();

    expect(screen.getByText("飼主名")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "飼主名でソート" })).not.toBeInTheDocument();
  });
});

describe("OwnersListTable 臨床ステータス", () => {
  it("生死列は狭幅でも非表示にせず、死亡状態を内部スクロールで確認できる", () => {
    renderTable();

    expect(screen.getByRole("columnheader", { name: "生死" })).not.toHaveClass("hidden");
    expect(screen.getByText("死亡").closest("td")).not.toHaveClass("hidden");
  });
});

describe("OwnersListTable row navigation accessibility", () => {
  it("同一医院かつ編集権限がある行に飼主・ペット・ID付き44px native detail linkを表示する", () => {
    renderTable({ currentClinicId: "clinic-1", canEdit: true });

    const detailLink = screen.getByRole("link", { name: /山田太郎/ });
    expect(detailLink).toHaveAttribute("href", "/owners/42");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/ID: 42/);
    expect(detailLink).toHaveAccessibleName(/ID: 7/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it.each([
    {
      label: "編集権限なし",
      overrides: { canEdit: false, currentClinicId: "clinic-1" },
    },
    {
      label: "別医院行",
      overrides: {
        canEdit: true,
        currentClinicId: "clinic-1",
        pets: [{ ...pet, clinicId: "clinic-2" }],
      },
    },
  ])("$label は飼主名をdetail linkにしない", ({ overrides }) => {
    renderTable(overrides);

    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /山田太郎/ })).not.toBeInTheDocument();
  });
});
