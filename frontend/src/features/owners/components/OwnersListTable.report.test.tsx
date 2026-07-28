import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { C } from "@/lib/design-tokens";
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

describe("OwnersListTable 危険理由 Popover (#234)", () => {
  it("危険度 high の保存済み理由を click で開示し、同じ trigger の再 click で閉じる", async () => {
    const user = userEvent.setup();
    renderTable({
      pets: [{ ...pet, dangerLevel: "高", dangerReason: "保定時に噛む" }],
    });

    const trigger = screen.getByRole("button", {
      name: "ポチの危険理由を表示",
    });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(trigger).toHaveAttribute("aria-controls");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText("保定時に噛む")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it.each([
    { caseName: "undefined", dangerReason: undefined },
    { caseName: "空文字", dangerReason: "" },
    { caseName: "空白のみ", dangerReason: " \t\n " },
  ])(
    "dangerReason が $caseName の high 個体は理由未登録を開示する",
    async ({ dangerReason }) => {
      const user = userEvent.setup();
      renderTable({
        pets: [{ ...pet, dangerLevel: "高", dangerReason }],
      });

      await user.click(
        screen.getByRole("button", { name: "ポチの危険理由を表示" }),
      );

      expect(await screen.findByText("理由未登録")).toBeInTheDocument();
    },
  );

  it.each([
    { keyName: "Enter", key: "{Enter}" },
    { keyName: "Space", key: " " },
  ])(
    "$keyName の同一キー操作で危険理由を開閉する",
    async ({ key }) => {
      const user = userEvent.setup();
      renderTable({
        pets: [{ ...pet, dangerLevel: "高", dangerReason: "診察台で噛む" }],
      });
      const trigger = screen.getByRole("button", {
        name: "ポチの危険理由を表示",
      });

      trigger.focus();
      await user.keyboard(key);
      expect(await screen.findByText("診察台で噛む")).toBeInTheDocument();

      trigger.focus();
      await user.keyboard(key);
      await waitFor(() => {
        expect(screen.queryByText("診察台で噛む")).not.toBeInTheDocument();
      });
    },
  );

  it("危険度 high だけ既存の警告文言と視覚クラスを trigger に維持する", () => {
    renderTable({
      pets: [
        { ...pet, id: "high", name: "ポチ", dangerLevel: "高" },
        { ...pet, id: "medium", name: "ミケ", dangerLevel: "中" },
        { ...pet, id: "low", name: "コタロウ", dangerLevel: "低" },
      ],
    });

    const trigger = screen.getByRole("button", {
      name: "ポチの危険理由を表示",
    });
    expect(trigger).toHaveTextContent("⚠ 危険");
    expect(trigger).toHaveClass(
      "inline-flex",
      "items-center",
      "rounded",
      "px-1.5",
      "py-0.5",
      "text-xs",
      "font-semibold",
      C.bgDanger10,
      C.danger,
      C.borderDanger20,
    );
    expect(screen.getAllByText("⚠ 危険")).toHaveLength(1);
    expect(
      screen.queryByRole("button", { name: "ミケの危険理由を表示" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "コタロウの危険理由を表示" }),
    ).not.toBeInTheDocument();
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

  it("同一医院のview-onlyユーザーにも飼主detailへのnative read linkを維持する", () => {
    renderTable({
      canEdit: false,
      canDelete: false,
      canReport: false,
      currentClinicId: "clinic-1",
    });

    expect(screen.getByRole("link", { name: /山田太郎/ })).toHaveAttribute("href", "/owners/42");
  });

  it("別医院行は編集権限があっても飼主名をdetail linkにしない", () => {
    renderTable({
      canEdit: true,
      currentClinicId: "clinic-1",
      pets: [{ ...pet, clinicId: "clinic-2" }],
    });

    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /山田太郎/ })).not.toBeInTheDocument();
  });
});
