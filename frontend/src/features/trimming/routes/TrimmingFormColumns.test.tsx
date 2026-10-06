import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { TrimmingFormData } from "@/types/trimming";
import { TrimmingFormColumns, type TrimmingPatient } from "./TrimmingFormColumns";

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
  staffName: "鈴木 一郎",
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

const basePatient: TrimmingPatient = {
  ownerName: "山田 太郎",
  name: "ポチ",
};

function renderColumns() {
  return render(
    <MemoryRouter>
      <TrimmingFormColumns
        selectedPet={basePatient}
        formData={baseFormData}
        fieldErrors={{}}
        courses={[]}
        options={[]}
        styleImagePreview={null}
        completedImagePreview={null}
        sortedHistory={[]}
        isHistoryLoading={false}
        historySearchTerm=""
        historySortOrder="desc"
        historyDateRange={{ from: "", to: "" }}
        showInitialStatusSelector={false}
        onFormChange={vi.fn()}
        onOpenCourseModal={vi.fn()}
        onOpenStaffModal={vi.fn()}
        onStyleImageChange={vi.fn()}
        onCompletedImageChange={vi.fn()}
        onRemoveStyleImage={vi.fn()}
        onRemoveCompletedImage={vi.fn()}
        onHistorySearchTermChange={vi.fn()}
        onHistorySortOrderChange={vi.fn()}
        onHistoryClear={vi.fn()}
        onHistoryStartDateChange={vi.fn()}
        onHistoryEndDateChange={vi.fn()}
        onHistoryClick={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("TrimmingFormColumns", () => {
  it("EMR-272: 担当スタッフのラベルを「担当医」ではなく「担当」と表示する", () => {
    renderColumns();

    const staffButton = screen.getByRole("button", { name: "担当鈴木 一郎" });
    expect(staffButton).toBeInTheDocument();
    expect(staffButton).not.toHaveAccessibleName(/担当医/);
  });
});
