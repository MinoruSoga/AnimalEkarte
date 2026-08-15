import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import type { SortOrder } from "@/types";

import type { ExaminationRecord } from "../api/transforms";
import { ExaminationHistoryPanel } from "./ExaminationHistoryPanel";

vi.mock("./ExamPivotTable", () => ({
  ExamPivotTable: ({ examinations }: { examinations: ExaminationRecord[] }) => (
    <div>ピボット:{examinations.length}</div>
  ),
}));

const EXAMINATION: ExaminationRecord = {
  id: "1",
  date: "2026-07-20",
  ownerName: "飼主",
  petName: "ポチ",
  petId: "7",
  medicalRecordId: undefined,
  testType: "血液検査",
  testTypeId: "10",
  doctor: "獣医師",
  doctorId: "3",
  status: "確定",
  resultSummary: "問題なし",
  machine: "DRI-CHEM",
  items: undefined,
};

function HistoryPanelHarness({
  initialView = "cards",
}: {
  initialView?: "cards" | "pivot";
}) {
  const [view, setView] = useState<"cards" | "pivot">(initialView);
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");

  return (
    <MemoryRouter>
      <ExaminationHistoryPanel
        filteredHistory={[EXAMINATION]}
        pivotHistory={[EXAMINATION]}
        currentPetId="7"
        historyStartDate=""
        historyEndDate=""
        historySearchTerm=""
        historySortOrder={sortOrder}
        historyView={view}
        onHistoryStartDateChange={vi.fn()}
        onHistoryEndDateChange={vi.fn()}
        onHistorySearchTermChange={vi.fn()}
        onHistorySortOrderChange={setSortOrder}
        onHistoryViewChange={setView}
        onHistoryClear={vi.fn()}
      />
    </MemoryRouter>
  );
}

describe("ExaminationHistoryPanel", () => {
  it("既存カード一覧をデフォルト表示し、時系列ピボットへ切り替えられる", async () => {
    const user = userEvent.setup();
    render(<HistoryPanelHarness />);

    expect(screen.getByText("血液検査")).toBeInTheDocument();
    expect(screen.queryByText("ピボット:1")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "時系列" }));

    expect(screen.getByText("ピボット:1")).toBeInTheDocument();
    expect(screen.queryByText("血液検査")).not.toBeInTheDocument();
  });

  it("deep-link指定時はピボットを初期表示する", () => {
    render(<HistoryPanelHarness initialView="pivot" />);

    expect(screen.getByText("ピボット:1")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "時系列" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
  });
});
