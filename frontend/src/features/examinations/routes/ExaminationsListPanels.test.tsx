import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { transformExamination } from "../api/transforms";
import { ExaminationsListContent } from "./ExaminationsListPanels";

function show(date: string) {
  const record = transformExamination({ id: 1, date, status: "pending" } as Parameters<
    typeof transformExamination
  >[0]);
  return render(
    <MemoryRouter>
      <ExaminationsListContent
        filterProperties={[]}
        activeFilters={[]}
        onFilterChange={vi.fn()}
        searchTerm=""
        onSearchChange={vi.fn()}
        count={1}
        activeSorts={[]}
        onSortChange={vi.fn()}
        directionFor={() => "none"}
        toggleSort={vi.fn()}
        isFiltering={false}
        hasPageScopedFilter={false}
        pagination={{
          paginatedData: [record],
          currentPage: 1,
          totalPages: 1,
          totalCount: 1,
          startIndex: 1,
          endIndex: 1,
        }}
        canEdit={false}
        canViewMedicalRecords={false}
        onEdit={vi.fn()}
        onPageChange={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("Examinations display-only date and missing values", () => {
  it("displays ISO date with slashes and keeps missing doctor/result as em dash", () => {
    show("2026-10-14");
    expect(screen.getByRole("cell", { name: "2026/10/14" })).toHaveClass("whitespace-nowrap");
    expect(screen.getAllByRole("cell", { name: "-" }).length).toBeGreaterThanOrEqual(2);
  });
  it("missing date displays em dash", () => {
    show("");
    expect(screen.getAllByRole("cell")[0]).toHaveTextContent("-");
  });
});

describe("EMR-415 X2: 一覧ヘッダの文字色", () => {
  it("ソート可能ヘッダは eyebrow variant（#615D59・uppercase）で描画する", () => {
    show("2026-10-14");

    for (const button of screen.getAllByRole("button", { name: /でソート/ })) {
      expect(button).toHaveClass("text-[#615D59]");
      expect(button).toHaveClass("uppercase");
    }
  });
});
