import { useActionState, useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { TrimmingFormData } from "@/types/trimming";
import {
  TrimmingLeftColumn,
  TrimmingRightColumn,
  type TrimmingHistoryItem,
} from "../lib/trimming-form-columns";

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

interface LeftColumnActionState {
  success: boolean;
  staffError: string | null;
}

function LeftColumnActionHarness() {
  const [formData, setFormData] = useState<TrimmingFormData>({
    ...baseFormData,
    courseId: "10",
    styleRequest: "短め",
  });
  const [formState, formAction] = useActionState(
    async (): Promise<LeftColumnActionState> => {
      if (!formData.staffId) {
        return { success: false, staffError: "担当者を選択してください" };
      }
      return { success: true, staffError: null };
    },
    { success: false, staffError: null },
  );

  return (
    <MemoryRouter>
      <form action={formAction}>
        <TrimmingLeftColumn
          formData={formData}
          courses={[{ id: "10", name: "フルコース", price: 5000 }]}
          options={[
            { id: "20", name: "爪切り", price: 500 },
            { id: "21", name: "耳掃除", price: 400 },
          ]}
          styleImagePreview={null}
          onFormChange={(updates) => setFormData((prev) => ({ ...prev, ...updates }))}
          onCourseModalOpen={vi.fn()}
          onStyleImageChange={vi.fn()}
          onRemoveStyleImage={vi.fn()}
          showInitialStatusSelector={false}
        />
        {formState.staffError ? <p>{formState.staffError}</p> : null}
        <button type="submit">保存</button>
      </form>
    </MemoryRouter>
  );
}

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

  it("バリデーションエラー後もオプション選択が保持される", async () => {
    const user = userEvent.setup();
    render(<LeftColumnActionHarness />);

    await user.click(screen.getByRole("checkbox", { name: "爪切り" }));
    await user.click(screen.getByRole("checkbox", { name: "耳掃除" }));

    expect(screen.getByRole("checkbox", { name: "爪切り" })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: "耳掃除" })).toBeChecked();
    expect(screen.getByText("フルコース")).toBeInTheDocument();
    expect(screen.getByLabelText("スタイルの希望")).toHaveValue("短め");

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("担当者を選択してください")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole("checkbox", { name: "爪切り" })).toBeChecked();
      expect(screen.getByRole("checkbox", { name: "耳掃除" })).toBeChecked();
    });
    expect(screen.getByText("フルコース")).toBeInTheDocument();
    expect(screen.getByLabelText("スタイルの希望")).toHaveValue("短め");
  });

  it("ユーザー操作ではオプションのチェックを外せる", async () => {
    const user = userEvent.setup();
    render(<LeftColumnActionHarness />);

    await user.click(screen.getByRole("checkbox", { name: "爪切り" }));
    expect(screen.getByRole("checkbox", { name: "爪切り" })).toBeChecked();

    await user.click(screen.getByRole("checkbox", { name: "爪切り" }));
    expect(screen.getByRole("checkbox", { name: "爪切り" })).not.toBeChecked();
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
