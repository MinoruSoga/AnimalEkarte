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

    expect(screen.getByPlaceholderText("開始日").closest("div.min-w-0")).not.toBeNull();
    expect(screen.getByPlaceholderText("終了日").closest("div.min-w-0")).not.toBeNull();
    expect(screen.getByLabelText("検索単語")).toHaveClass("h-11");
    expect(screen.getByRole("button", { name: "クリア" })).toHaveClass("h-11");
    const sortTrigger = screen.getByRole("combobox", { name: "並び順" });
    expect(sortTrigger).toHaveClass("h-11");
    // EMR-313: 80px固定で「新→古」が切れていたため内容幅+96px下限へ
    expect(sortTrigger).not.toHaveClass("w-[80px]");
    expect(sortTrigger).toHaveClass("w-auto", "min-w-24");
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
