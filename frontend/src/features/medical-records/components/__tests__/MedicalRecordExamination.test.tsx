import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { MedicalRecordExamination } from "../MedicalRecordExamination";
import { useGetRecordExaminations } from "../../api/get-record-examinations";
import type { ExamGroup } from "../../api/get-record-examinations";
import type { ExamResult } from "@/features/examinations";

vi.mock("../../api/get-record-examinations");

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: false, canDelete: false }),
}));

// ExaminationImportDialog は open=false のためマウントされないが、
// 依存モジュールのエラーを防ぐために最小限の stub を提供する。
vi.mock("../ExaminationImportDialog", () => ({
  ExaminationImportDialog: () => null,
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
    items: [
      makeItem({ id: "1", name: "グルコース" }),
      makeItem({ id: "2", name: "クレアチニン" }),
    ],
  },
  {
    id: 2,
    date: "2026-01-02 10:00",
    machine: "DRI-CHEM",
    items: [makeItem({ id: "3", name: "ビリルビン" })],
  },
];

beforeEach(() => {
  vi.mocked(useGetRecordExaminations).mockReturnValue({
    data: EXAM_GROUPS,
    isLoading: false,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useGetRecordExaminations>);
});

describe("MedicalRecordExamination — カナ混同検索", () => {
  it("空の検索語は全グループを表示する", () => {
    render(<MedicalRecordExamination petId="1" />);
    expect(screen.getByText("グルコース")).toBeInTheDocument();
    expect(screen.getByText("ビリルビン")).toBeInTheDocument();
  });

  it("ひらがな「ぐるこーす」でカタカナ検査名「グルコース」にヒットする", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordExamination petId="1" />);

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "ぐるこーす");

    await waitFor(() => {
      expect(screen.getByText("グルコース")).toBeInTheDocument();
      expect(screen.queryByText("ビリルビン")).not.toBeInTheDocument();
    });
  });

  it("カタカナ「グルコース」でカタカナ検査名にヒットする (かな統一検索)", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordExamination petId="1" />);

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "グルコース");

    await waitFor(() => {
      expect(screen.getByText("グルコース")).toBeInTheDocument();
      expect(screen.queryByText("ビリルビン")).not.toBeInTheDocument();
    });
  });

  it("ひらがな「びりるびん」でカタカナ検査名「ビリルビン」にヒットする", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordExamination petId="1" />);

    await user.type(screen.getByPlaceholderText("WBC, Cre, etc..."), "びりるびん");

    await waitFor(() => {
      expect(screen.getByText("ビリルビン")).toBeInTheDocument();
      expect(screen.queryByText("グルコース")).not.toBeInTheDocument();
    });
  });
});
