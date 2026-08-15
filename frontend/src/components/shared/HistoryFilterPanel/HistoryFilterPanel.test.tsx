import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { HistoryFilterPanel } from "./HistoryFilterPanel";

describe("HistoryFilterPanel", () => {
  it("日付欄は狭い親幅を押し広げず、検索操作は44px以上を保つ", () => {
    render(
      <HistoryFilterPanel
        filterStartDate=""
        onFilterStartDateChange={vi.fn()}
        filterEndDate=""
        onFilterEndDateChange={vi.fn()}
        searchTerm=""
        onSearchTermChange={vi.fn()}
        sortOrder="desc"
        onSortOrderChange={vi.fn()}
        onClear={vi.fn()}
      />,
    );

    expect(screen.getByPlaceholderText("開始日").parentElement).toHaveClass("min-w-0");
    expect(screen.getByPlaceholderText("終了日").parentElement).toHaveClass("min-w-0");
    expect(screen.getByLabelText("検索単語")).toHaveClass("h-11");
    expect(screen.getByRole("button", { name: "クリア" })).toHaveClass("h-11");
    expect(screen.getByRole("combobox", { name: "並び順" })).toHaveClass("h-11");
  });

  it("開始日・終了日の実inputに明示labelとid/nameを接続する", () => {
    render(
      <HistoryFilterPanel
        filterStartDate=""
        onFilterStartDateChange={vi.fn()}
        filterEndDate=""
        onFilterEndDateChange={vi.fn()}
        searchTerm=""
        onSearchTermChange={vi.fn()}
        sortOrder="desc"
        onSortOrderChange={vi.fn()}
        onClear={vi.fn()}
      />,
    );

    const startDate = screen.getByRole("textbox", { name: "開始日" });
    const endDate = screen.getByRole("textbox", { name: "終了日" });

    expect(startDate).toHaveAttribute("id");
    expect(startDate).toHaveAttribute("name", "historyStartDate");
    expect(endDate).toHaveAttribute("id");
    expect(endDate).toHaveAttribute("name", "historyEndDate");
    expect(startDate.id).not.toBe(endDate.id);
  });
});
