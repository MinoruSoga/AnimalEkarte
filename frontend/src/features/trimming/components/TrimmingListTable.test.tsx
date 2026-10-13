import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router";
import { describe, expect, it, vi } from "vitest";
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

function LocationProbe() {
  const { state } = useLocation();
  const from =
    state && typeof state === "object" && "from" in state && typeof state.from === "string"
      ? state.from
      : "";
  return <output data-testid="location-from">{from}</output>;
}

interface RenderTableOptions {
  onEdit?: (id: string) => void;
  isValidStaff?: (staff: string) => boolean;
}

function renderTable(
  records: TrimmingUI[],
  { onEdit = noop, isValidStaff = () => true }: RenderTableOptions = {},
) {
  return render(
    <MemoryRouter initialEntries={["/trimming"]}>
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
        isValidStaff={isValidStaff}
        directionFor={() => "none"}
        onSearchChange={noop}
        onFilterChange={noop}
        onSortChange={noop}
        onToggleSort={noop}
        onEdit={onEdit}
        onDeleteClick={noop}
        onPageChange={noop}
      />
      <LocationProbe />
    </MemoryRouter>,
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

describe("TrimmingListTable row navigation accessibility", () => {
  it("ペット名・診療日を含む44px以上のnative detail linkを行内に表示する", () => {
    renderTable([makeRecord({ id: "trim-1" })]);

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute("href", "/trimming/trim-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/trim-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
    fireEvent.click(detailLink);
    expect(screen.getByTestId("location-from")).toHaveTextContent(/^\/trimming$/);
  });

  it("detail link以外のセルclickでは編集遷移を起動しない", () => {
    const onEdit = vi.fn();
    renderTable([makeRecord({})], { onEdit });

    fireEvent.click(screen.getByText("ヤマダタロウ"));

    expect(onEdit).not.toHaveBeenCalled();
  });

  it("無効な担当staffの警告をscreen readerへ文脈付きで説明する", () => {
    renderTable([makeRecord({ staff: "退職者" })], {
      isValidStaff: () => false,
    });

    const warning = screen.getByLabelText(/無効/);
    expect(warning).toHaveAccessibleName(/退職者/);
  });

  it("行操作buttonのaccessible nameにトリミングIDを含める", () => {
    renderTable([makeRecord({ id: "trim-1" })]);

    expect(screen.getByRole("button", { name: /trim-1/ })).toBeInTheDocument();
  });

  it("petNumberをDESIGN body-smで表示する", () => {
    renderTable([makeRecord({ petNumber: "P-1" })]);

    expect(screen.getByText("P-1")).toHaveClass("text-sm");
    expect(screen.getByText("P-1")).not.toHaveClass("text-base");
  });
});

describe("TrimmingListTable 列レイアウト (EMR-378)", () => {
  it("担当列は読める最小幅を持ちスタッフ名が1字縦積みにならない", () => {
    renderTable([makeRecord({})]);

    const staffHeader = screen.getByRole("columnheader", { name: /担当/ });
    expect(staffHeader.className).toContain("min-w-[120px]");
    expect(staffHeader.className).toContain("whitespace-nowrap");
  });

  it("ステータス列は共有のnowrap+min-width契約を使う", () => {
    renderTable([makeRecord({})]);

    const statusHeader = screen.getByRole("columnheader", { name: /ステータス/ });
    expect(statusHeader.className).toContain("min-w-[100px]");
    expect(statusHeader.className).toContain("whitespace-nowrap");
  });

  it("診療日セルは折り返さない", () => {
    renderTable([makeRecord({})]);

    const dateCell = screen.getByText("2026-07-13").closest("td");
    expect(dateCell?.className).toContain("whitespace-nowrap");
  });

  it("飼主名・ペット名ヘッダは折り返さず読める最小幅を持つ (T002)", () => {
    renderTable([makeRecord({})]);

    const ownerHeader = screen.getByRole("columnheader", { name: /飼主名/ });
    expect(ownerHeader.className).toContain("whitespace-nowrap");
    expect(ownerHeader.className).toContain("min-w-[140px]");

    const petHeader = screen.getByRole("columnheader", { name: /ペット名/ });
    expect(petHeader.className).toContain("whitespace-nowrap");
    expect(petHeader.className).toContain("min-w-[160px]");
  });

  it("飼主名・ペット名セルはヘッダと同じ最小幅契約を持ち、氏名は折り返せる", () => {
    renderTable([makeRecord({})]);

    const ownerCell = screen.getByText("ヤマダタロウ").closest("td");
    expect(ownerCell?.className).toContain("min-w-[140px]");

    const petCell = screen.getByRole("link", { name: /ポチ/ }).closest("td");
    expect(petCell?.className).toContain("min-w-[160px]");
  });

  it("petNumber のような短い識別子はハイフンで折り返さない (T002)", () => {
    renderTable([makeRecord({ petNumber: "SYN-PET-990003" })]);

    const petNumber = screen.getByText("SYN-PET-990003");
    expect(petNumber.className).toContain("whitespace-nowrap");
  });
});
