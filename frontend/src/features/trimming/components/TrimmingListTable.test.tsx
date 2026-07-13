import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { TrimmingUI } from "@/types";
import { TrimmingListTable } from "./TrimmingListTable";

function makeRecord(overrides: Partial<TrimmingUI>): TrimmingUI {
  return {
    id: "1",
    reservationTypeId: "9",
    hasDetail: true,
    date: "2026-07-13",
    petId: "10",
    ownerId: "20",
    petNumber: "P-1",
    petName: "ポチ",
    ownerName: "ヤマダタロウ",
    species: "犬",
    breed: "",
    weight: "5.2",
    styleRequest: "",
    staff: "田中",
    status: "予約",
    staffId: "3",
    courseId: "4",
    courseName: "スタンダードコース",
    optionIds: [],
    bw: "",
    bwUnit: "Kg",
    bt: "",
    usedShampoo: "",
    usedRibbon: "",
    remarks: "",
    styleImage: undefined,
    completedImage: undefined,
    ...overrides,
  };
}

const noop = () => undefined;

function renderTable(records: TrimmingUI[]) {
  return render(
    <TrimmingListTable
      records={records}
      filteredCount={records.length}
      currentPage={1}
      totalPages={1}
      startIndex={0}
      endIndex={records.length}
      searchKeyword=""
      activeFilters={[]}
      activeSorts={[]}
      filterProperties={[]}
      isFiltering={false}
      canEdit
      canDelete
      isValidStaff={() => true}
      directionFor={() => "none"}
      onSearchChange={noop}
      onFilterChange={noop}
      onSortChange={noop}
      onToggleSort={noop}
      onEdit={noop}
      onDeleteClick={noop}
      onPageChange={noop}
    />,
  );
}

describe("TrimmingListTable 犬種列 (#231)", () => {
  it("犬種が設定されている場合はその値を表示する", () => {
    renderTable([makeRecord({ breed: "トイプードル" })]);
    expect(screen.getByText("トイプードル")).toBeInTheDocument();
  });

  it("犬種が空文字の場合は「-」を表示する", () => {
    renderTable([makeRecord({ breed: "" })]);
    expect(screen.getByText("-")).toBeInTheDocument();
  });
});
