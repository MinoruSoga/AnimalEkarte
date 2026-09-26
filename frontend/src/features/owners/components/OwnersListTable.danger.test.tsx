import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { C } from "@/lib/design-tokens";
import { OwnersListTable } from "./OwnersListTable";
import type { Pet } from "@/types";

// 危険マーク表示に焦点を当てるため、無関係な重い子は無効化する（report.test.tsx と同型）。
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

describe("OwnersListTable 危険度バッジ (EMR-173)", () => {
  it("危険度 高 は赤い ⚠ 危険 バッジを飼主名列に出し、理由 Popover を開閉できる", async () => {
    const user = userEvent.setup();
    renderTable({
      pets: [{ ...pet, dangerLevel: "高", dangerReason: "保定時に噛む" }],
    });

    const trigger = screen.getByRole("button", { name: "ポチの危険理由を表示" });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);

    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText("保定時に噛む")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("危険度 中 は黄色 ⚠ 注意 バッジを出し、注意理由 Popover を開く", async () => {
    const user = userEvent.setup();
    renderTable({
      pets: [{ ...pet, dangerLevel: "中", dangerReason: "興奮しやすい" }],
    });

    const trigger = screen.getByRole("button", { name: "ポチの注意理由を表示" });
    expect(trigger).toHaveTextContent("⚠ 注意");
    expect(trigger).toHaveClass(C.bgNotice, C.textNotice, C.borderNotice);
    expect(screen.queryByRole("button", { name: "ポチの危険理由を表示" })).not.toBeInTheDocument();

    await user.click(trigger);
    expect(await screen.findByText("興奮しやすい")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
  });

  it("危険度 低・未設定の行は危険バッジを何も出さない", () => {
    renderTable({
      pets: [
        { ...pet, id: "low", name: "ロウ", dangerLevel: "低" },
        { ...pet, id: "unset", name: "ミセッテイ" },
      ],
    });

    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();
    expect(screen.queryByText("⚠ 注意")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /の(危険|注意)理由を表示/ }),
    ).not.toBeInTheDocument();
  });

  it("飼主が危険人物なら飼主名リンクの横に ⚠ 危険人物 を出す", () => {
    renderTable({
      pets: [{ ...pet, ownerIsDangerous: true }],
    });

    const ownerCell = screen.getByRole("link", { name: /山田太郎/ }).parentElement;
    expect(ownerCell).toHaveTextContent("⚠ 危険人物");
  });

  it.each([
    { caseName: "false", ownerIsDangerous: false },
    { caseName: "未設定", ownerIsDangerous: undefined },
  ])("飼主の is_dangerous が $caseName なら危険人物マークを出さない", ({ ownerIsDangerous }) => {
    renderTable({
      pets: [{ ...pet, ownerIsDangerous }],
    });

    expect(screen.queryByText("⚠ 危険人物")).not.toBeInTheDocument();
  });
});
