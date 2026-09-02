import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { MedicalRecordExamination } from "./MedicalRecordExamination";
import { useGetRecordExaminations } from "../api/get-record-examinations";
import type { ExamGroup } from "../api/get-record-examinations";
import type { ExamResult } from "@/lib/transforms/examination";

vi.mock("../api/get-record-examinations");

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: false, canDelete: false }),
}));

// ExaminationImportDialog は open=false のためマウントされないが、
// 依存モジュールのエラーを防ぐために最小限の stub を提供する。
vi.mock("./ExaminationImportDialog", () => ({
  ExaminationImportDialog: () => null,
}));

vi.mock("@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner", () => ({
  LabDeviceUnlinkedBanner: () => null,
}));

const makeItem = (overrides: Partial<ExamResult> = {}): ExamResult => ({
  id: "1",
  examTypeFieldId: 1,
  name: "GLU",
  result: "95",
  inspectionValue: "95",
  normalValue: "70-110",
  unit: "mg/dL",
  referenceValue: "70-110",
  refMin: 70,
  refMax: 110,
  isAbnormal: false,
  status: "normal",
  sortOrder: 0,
  ...overrides,
});

const EXAM_GROUPS: ExamGroup[] = [
  {
    id: 1,
    date: "2026-01-01 10:00",
    machine: "DRI-CHEM",
    name: "血液検査",
    price: 4200,
    medicalRecordId: "9",
    items: [
      makeItem({ id: "1", name: "グルコース" }),
      makeItem({ id: "2", name: "クレアチニン" }),
    ],
  },
  {
    id: 2,
    date: "2026-01-02 10:00",
    machine: "DRI-CHEM",
    name: "血液検査",
    price: 4200,
    medicalRecordId: "9",
    items: [makeItem({ id: "3", name: "ビリルビン" })],
  },
];

beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn();
  vi.mocked(useGetRecordExaminations).mockReturnValue({
    data: { items: EXAM_GROUPS, isTruncated: false },
    isLoading: false,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useGetRecordExaminations>);
});

describe("MedicalRecordExamination — カナ混同検索", () => {
  function renderExaminations() {
    return render(
      <MemoryRouter>
        <MedicalRecordExamination petId="1" />
      </MemoryRouter>,
    );
  }

  it("空の検索語は全グループを表示する", () => {
    renderExaminations();
    expect(screen.getByText("グルコース")).toBeInTheDocument();
    expect(screen.getByText("ビリルビン")).toBeInTheDocument();
  });

  it("100件で打ち切られた場合は履歴の省略を表示する", () => {
    vi.mocked(useGetRecordExaminations).mockReturnValue({
      data: { items: EXAM_GROUPS, isTruncated: true },
      isLoading: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetRecordExaminations>);

    renderExaminations();

    expect(screen.getByText("直近100件を表示しています")).toBeInTheDocument();
  });

  it("ひらがな「ぐるこーす」でカタカナ検査名「グルコース」にヒットする", async () => {
    const user = userEvent.setup();
    renderExaminations();

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "ぐるこーす");

    await waitFor(() => {
      expect(screen.getByText("グルコース")).toBeInTheDocument();
      expect(screen.queryByText("ビリルビン")).not.toBeInTheDocument();
    });
  });

  it("カタカナ「グルコース」でカタカナ検査名にヒットする (かな統一検索)", async () => {
    const user = userEvent.setup();
    renderExaminations();

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "グルコース");

    await waitFor(() => {
      expect(screen.getByText("グルコース")).toBeInTheDocument();
      expect(screen.queryByText("ビリルビン")).not.toBeInTheDocument();
    });
  });

  it("ひらがな「びりるびん」でカタカナ検査名「ビリルビン」にヒットする", async () => {
    const user = userEvent.setup();
    renderExaminations();

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "びりるびん");

    await waitFor(() => {
      expect(screen.getByText("ビリルビン")).toBeInTheDocument();
      expect(screen.queryByText("グルコース")).not.toBeInTheDocument();
    });
  });

  it("examId があるとき対象検査を先頭にして強調する", () => {
    render(
      <MemoryRouter initialEntries={["/?examId=2"]}>
        <MedicalRecordExamination petId="1" medicalRecordId="9" />
      </MemoryRouter>,
    );

    expect(useGetRecordExaminations).toHaveBeenCalledWith("1", "9");
    const groups = document.querySelectorAll("[id^='exam-group-']");
    expect(groups[0]).toHaveAttribute("id", "exam-group-2");
    expect(groups[0]).toHaveAttribute("aria-current", "true");
    expect(groups[1]).toHaveAttribute("id", "exam-group-1");
  });

  it("各検査グループへ対象petIdを渡して時系列導線を構成する", () => {
    renderExaminations();

    const firstDetailLink = screen.getByRole("link", {
      name: "2026-01-01 10:00の検歴を表示",
    });
    const secondDetailLink = screen.getByRole("link", {
      name: "2026-01-02 10:00の検歴を表示",
    });
    expect(firstDetailLink).toHaveAttribute(
      "href",
      "/examinations/1?petId=1&historyView=pivot",
    );
    expect(secondDetailLink).toHaveAttribute(
      "href",
      "/examinations/2?petId=1&historyView=pivot",
    );
  });
});
