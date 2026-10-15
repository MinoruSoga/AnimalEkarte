import type { ComponentProps } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { VaccinationListContent } from "./VaccinationListPanels";

type RecordRow = ComponentProps<
  typeof VaccinationListContent
>["pagination"]["paginatedData"][number];
function show(date: string, nextDate: string) {
  const record = {
    id: "1",
    date,
    nextDate,
    ownerName: "合成飼主",
    petName: "合成患者",
    vaccineName: "",
    vaccineId: "1",
    doctor: "",
    version: 1,
  } as RecordRow;
  render(
    <MemoryRouter>
      <VaccinationListContent
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
        pagination={{
          paginatedData: [record],
          currentPage: 1,
          totalPages: 1,
          totalCount: 1,
          startIndex: 1,
          endIndex: 1,
        }}
        canEdit={false}
        canDelete={false}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
        detailState={{ from: "/vaccinations" }}
      />
    </MemoryRouter>,
  );
}

describe("Vaccination displayed date and name", () => {
  it("formats date and overdue next date without changing source fields", () => {
    show("2026-10-14", "2000-01-01");
    expect(screen.getByRole("cell", { name: "2026/10/14" })).toHaveClass("whitespace-nowrap");
    expect(screen.getByText("2000/01/01")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "—" })).toBeInTheDocument();
  });
  it("missing dates and vaccine name are em dashes", () => {
    show("", "");
    expect(screen.getAllByRole("cell", { name: "—" })).toHaveLength(3);
  });
});
