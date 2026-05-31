import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TrimmingRightColumn, type TrimmingHistoryItem } from "./TrimmingFormColumns";

describe("TrimmingRightColumn", () => {
  it("履歴クリック時にコース・オプション・使用シャンプー・使用リボン・備考をコピー対象に含める", async () => {
    const onHistoryClick = vi.fn();
    const history: TrimmingHistoryItem = {
      id: "1",
      date: "2026-05-01",
      styleRequest: "短め",
      courseId: "10",
      optionIds: ["20", "21"],
      usedShampoo: "薬用A",
      usedRibbon: "赤",
      remarks: "皮膚注意",
      staff: "山田",
    };

    render(
      <TrimmingRightColumn
        sortedHistory={[history]}
        isHistoryLoading={false}
        historySearchTerm=""
        historySortOrder="desc"
        historyDateRange={{ from: "", to: "" }}
        onSearchTermChange={vi.fn()}
        onSortOrderChange={vi.fn()}
        onClear={vi.fn()}
        onFilterStartDateChange={vi.fn()}
        onFilterEndDateChange={vi.fn()}
        onHistoryClick={onHistoryClick}
      />,
    );

    await userEvent.click(screen.getByRole("button", { name: /短め/ }));

    expect(onHistoryClick).toHaveBeenCalledWith({
      courseId: "10",
      optionIds: ["20", "21"],
      styleRequest: "短め",
      usedShampoo: "薬用A",
      usedRibbon: "赤",
      remarks: "皮膚注意",
      staffName: "山田",
    });
  });
});
