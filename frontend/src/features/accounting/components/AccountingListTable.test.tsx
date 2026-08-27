import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import type { Accounting } from "../types";
import { AccountingListTable } from "./AccountingListTable";

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));

vi.mock("@/components/shared/FilteringIndicator/FilteringIndicator", () => ({
  FilteringIndicator: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

const accounting: Accounting = {
  id: "acct-1",
  clinicId: "clinic-1",
  ownerId: "owner-1",
  ownerName: "山田太郎",
  petId: "pet-1",
  petName: "ポチ",
  status: "waiting",
  scheduledDate: "2026-07-13",
  items: [],
  totalRefundedAmount: 0,
};

function renderTable(canEdit = false) {
  const onEdit = vi.fn();
  render(
    <MemoryRouter initialEntries={["/accounting"]}>
      <AccountingListTable
        filteredCount={1}
        pagination={{
          paginatedData: [accounting],
          totalPages: 1,
          totalCount: 1,
          startIndex: 0,
          endIndex: 1,
          currentPage: 1,
        }}
        searchTerm=""
        activeFilters={[]}
        activeSorts={[]}
        isFiltering={false}
        canEdit={canEdit}
        directionFor={() => "none"}
        onSearchChange={vi.fn()}
        onFilterChange={vi.fn()}
        onSortChange={vi.fn()}
        onToggleSort={vi.fn()}
        onEdit={onEdit}
        onMedicalRecordOpen={vi.fn()}
        onPageChange={vi.fn()}
      />
    </MemoryRouter>,
  );

  return { onEdit };
}

describe("AccountingListTable recorded amounts", () => {
  it("負の入金合計を符号のまま表示する", () => {
    render(
      <MemoryRouter initialEntries={["/accounting"]}>
        <AccountingListTable
          filteredCount={1}
          pagination={{
            paginatedData: [{
              ...accounting,
              status: "completed",
              payment: {
                subtotal: -3000,
                taxTotal: 0,
                totalAmount: -3000,
                insuranceAmount: 0,
                discountAmount: 0,
                billingAmount: -3000,
                receivedAmount: 0,
                changeAmount: 0,
                method: "cash",
              },
            }],
            totalPages: 1,
            totalCount: 1,
            startIndex: 0,
            endIndex: 1,
            currentPage: 1,
          }}
          searchTerm=""
          activeFilters={[]}
          activeSorts={[]}
          isFiltering={false}
          canEdit={false}
          directionFor={() => "none"}
          onSearchChange={vi.fn()}
          onFilterChange={vi.fn()}
          onSortChange={vi.fn()}
          onToggleSort={vi.fn()}
          onEdit={vi.fn()}
          onMedicalRecordOpen={vi.fn()}
          onPageChange={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("¥-3,000")).toBeInTheDocument();
  });
});

describe("AccountingListTable row navigation accessibility", () => {
  it("編集権限に関係なく日付・飼主・ペット・ID付き44px native detail linkを表示する", () => {
    renderTable(false);

    const detailLink = screen.getByRole("link", { name: /2026-07-13/ });
    expect(detailLink).toHaveAttribute("href", "/accounting/acct-1");
    expect(detailLink).toHaveAccessibleName(/山田太郎/);
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/acct-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは編集callbackを発火しない", () => {
    const { onEdit } = renderTable(true);

    fireEvent.click(screen.getByText("山田太郎"));

    expect(onEdit).not.toHaveBeenCalled();
  });
});

describe("AccountingListTable status column (BUG-020)", () => {
  it("ステータスヘッダは min-width と whitespace-nowrap で1行表示する", () => {
    renderTable(false);

    const statusHeader = screen.getByRole("columnheader", { name: /ステータス/ });
    expect(statusHeader.className).toContain("w-[100px]");
    expect(statusHeader.className).toContain("min-w-[100px]");
    expect(statusHeader.className).toContain("whitespace-nowrap");
  });

  it("ステータスセルも whitespace-nowrap を持つ", () => {
    renderTable(false);

    const waitingBadge = screen.getByText("未精算");
    expect(waitingBadge.closest("td")?.className).toContain("whitespace-nowrap");
    expect(waitingBadge.closest("td")?.className).toContain("min-w-[100px]");
  });
});
