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

describe("OwnersListTable 特記バッジ (EMR-173/231)", () => {
  it("特記レベル 高 は赤いアイコンバッジを飼主名列に出し、メモ Popover を開閉できる", async () => {
    const user = userEvent.setup();
    renderTable({
      pets: [{ ...pet, dangerLevel: "高", dangerReason: "保定時に噛む" }],
    });

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger.textContent).toBe("");
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

  it("特記レベル 中 は黄色アイコンバッジを出し、同じ中立名の Popover を開く", async () => {
    const user = userEvent.setup();
    renderTable({
      pets: [{ ...pet, dangerLevel: "中", dangerReason: "興奮しやすい" }],
    });

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.textContent).toBe("");
    expect(trigger).toHaveClass(C.bgNotice, C.textBadgeYellow, C.borderNotice);

    await user.click(trigger);
    expect(await screen.findByText("興奮しやすい")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
  });

  it("特記レベル 低・未設定の行はバッジを何も出さず、直接文言も出ない", () => {
    renderTable({
      pets: [
        { ...pet, id: "low", name: "ロウ", dangerLevel: "低" },
        { ...pet, id: "unset", name: "ミセッテイ" },
      ],
    });

    expect(screen.queryByText(/危険|注意/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /の詳細を表示/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("img", { name: "特記" })).not.toBeInTheDocument();
  });

  it("飼主に特記フラグがあれば飼主名リンクの横にアイコンマークを出す", () => {
    renderTable({
      pets: [{ ...pet, ownerIsDangerous: true }],
    });

    const ownerCell = screen.getByRole("link", { name: /山田太郎/ }).parentElement;
    expect(ownerCell?.querySelector("svg")).not.toBeNull();
    expect(screen.getByRole("img", { name: "特記" })).toBeInTheDocument();
    expect(ownerCell).not.toHaveTextContent(/危険|注意/);
  });

  it.each([
    { caseName: "false", ownerIsDangerous: false },
    { caseName: "未設定", ownerIsDangerous: undefined },
  ])("飼主の is_dangerous が $caseName なら特記マークを出さない", ({ ownerIsDangerous }) => {
    renderTable({
      pets: [{ ...pet, ownerIsDangerous }],
    });

    expect(screen.queryByRole("img", { name: "特記" })).not.toBeInTheDocument();
  });
});
