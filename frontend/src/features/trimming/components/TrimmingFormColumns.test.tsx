import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { TrimmingFormData } from "@/types/trimming";
import { TrimmingLeftColumn, TrimmingRightColumn, type TrimmingHistoryItem } from "./trimming-form-columns";

const baseFormData: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

describe("TrimmingLeftColumn", () => {
  it("#233: showInitialStatusSelector=false のとき登録時ステータス選択を表示しない", () => {
    render(
      <MemoryRouter>
        <TrimmingLeftColumn
          formData={baseFormData}
          courses={[]}
          options={[]}
          styleImagePreview={null}
          onFormChange={vi.fn()}
          onCourseModalOpen={vi.fn()}
          onStyleImageChange={vi.fn()}
          onRemoveStyleImage={vi.fn()}
          showInitialStatusSelector={false}
        />
      </MemoryRouter>,
    );

    expect(screen.queryByLabelText("登録時ステータスを選択")).not.toBeInTheDocument();
  });

  it("#233: showInitialStatusSelector=true のとき登録時ステータス選択を表示し、選択で initialStatus を更新する", async () => {
    const user = userEvent.setup();
    const onFormChange = vi.fn();

    render(
      <MemoryRouter>
        <TrimmingLeftColumn
          formData={baseFormData}
          courses={[]}
          options={[]}
          styleImagePreview={null}
          onFormChange={onFormChange}
          onCourseModalOpen={vi.fn()}
          onStyleImageChange={vi.fn()}
          onRemoveStyleImage={vi.fn()}
          showInitialStatusSelector
        />
      </MemoryRouter>,
    );

    const trigger = screen.getByLabelText("登録時ステータスを選択");
    expect(trigger).toHaveTextContent("進行中");

    await user.click(trigger);
    await user.click(screen.getByRole("option", { name: "予約" }));

    expect(onFormChange).toHaveBeenCalledWith({ initialStatus: "pending" });
  });
});

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
