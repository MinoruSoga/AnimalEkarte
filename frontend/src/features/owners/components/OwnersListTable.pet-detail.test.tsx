import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, it, expect, vi } from "vitest";
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

describe("OwnersListTable 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月のペット行はペット名セルに Cake バッジを表示する", () => {
    renderTable({ pets: [{ ...pet, birthDate: "2018-10-05" }] });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birthDate 未登録", undefined],
  ])("%s のペットには Cake バッジを表示しない", (_label, birthDate) => {
    renderTable({ pets: [{ ...pet, birthDate }] });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderTable({ pets: [{ ...pet, status: "死亡", birthDate: "2018-10-05" }] });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("OwnersListTable 列レイアウト (EMR-378)", () => {
  it("日付・固定ラベル列のヘッダは折り返さない", () => {
    renderTable();

    for (const name of [
      "飼主No",
      "ペット番号",
      "生死",
      "種",
      "生年月日",
      "体重",
      "環境",
      "前回来院",
      "操作",
    ]) {
      const header = screen.getByRole("columnheader", { name });
      expect(header.className).toContain("whitespace-nowrap");
    }
  });

  it("飼主名・ペット名セルは読める最小幅の中で折り返す（nowrap強制しない）", () => {
    renderTable();

    const ownerCell = screen.getByRole("link", { name: /飼主「山田太郎」/ }).closest("td");
    expect(ownerCell?.className).toContain("min-w-[140px]");
    expect(ownerCell?.className).not.toContain("whitespace-nowrap");

    const petCell = screen
      .getByRole("link", { name: "ペット「ポチ」(ID: 7) の詳細を開く" })
      .closest("td");
    expect(petCell?.className).toContain("min-w-[120px]");
    expect(petCell?.className).not.toContain("whitespace-nowrap");
  });
});
