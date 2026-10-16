import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import type { ComponentProps } from "react";

import type { ExamResult, ExaminationRecord } from "@/lib/transforms/examination";

import { MedicalRecordExamination } from "./MedicalRecordExamination";
import type { ExamGroup } from "../api/get-record-examinations";

// PR679-REVIEW-R5: ロック中・作成権限なしのカルテに検査取り込みの導線が残らないことを、
// 実親（MedicalRecordExamination）+ 実 ExaminationFilter + 実 ExaminationImportDialog で確認する。
const mocks = vi.hoisted(() => ({
  useGetRecordExaminations: vi.fn(),
  useGetExaminations: vi.fn(),
  useUpdateExamination: vi.fn(),
  updateExamination: vi.fn(),
  refetch: vi.fn(),
  permission: { canView: true, canCreate: true, canEdit: true, canDelete: false },
}));

// lazy ラッパーだけ実ダイアログへ差し替える（他の遅延モーダルは読み込まない）。
vi.mock("./MedicalRecordLazyModals", async () => {
  const actual = await vi.importActual<typeof import("./ExaminationImportDialog")>(
    "./ExaminationImportDialog",
  );
  return { ExaminationImportDialog: actual.ExaminationImportDialog };
});

vi.mock("../api/get-record-examinations", () => ({
  useGetRecordExaminations: mocks.useGetRecordExaminations,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permission,
}));

vi.mock("@/hooks/use-examinations", () => ({
  useGetExaminations: mocks.useGetExaminations,
}));

vi.mock("@/hooks/use-update-examination", () => ({
  useUpdateExamination: mocks.useUpdateExamination,
}));

vi.mock("@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner", () => ({
  LabDeviceUnlinkedBanner: () => null,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn();
  mocks.useGetRecordExaminations.mockReturnValue({
    data: { items: [], isTruncated: false },
    isLoading: false,
    refetch: mocks.refetch,
  });
  mocks.useGetExaminations.mockReturnValue({ data: [], isLoading: false });
  mocks.useUpdateExamination.mockReturnValue({ mutateAsync: mocks.updateExamination });
  mocks.permission = { canView: true, canCreate: true, canEdit: true, canDelete: false };
  mocks.updateExamination.mockClear();
  mocks.refetch.mockClear();
});

function makeExam(overrides: Partial<ExaminationRecord> & { id: string }): ExaminationRecord {
  return {
    date: "2026-10-01",
    ownerName: "飼主",
    petName: "モカ",
    petId: "pet-1",
    medicalRecordId: undefined,
    testType: "血液検査",
    testTypeId: "1",
    doctor: "佐藤",
    doctorId: "2",
    // status は既定の「依頼中」（未確定 = 取り込み可）
    status: "依頼中",
    currentRevisionVersion: undefined,
    resultSummary: "",
    machine: "",
    items: undefined,
    ...overrides,
  };
}

function makeItem(overrides: Partial<ExamResult> = {}): ExamResult {
  return {
    id: "1",
    examTypeFieldId: 1,
    name: "WBC",
    result: "9.0",
    inspectionValue: "9.0",
    normalValue: "6.0-17.0",
    unit: "mg/dL",
    referenceValue: "6.0-17.0",
    refMin: 6,
    refMax: 17,
    isAssessed: false,
    isAbnormal: false,
    status: "normal",
    sortOrder: 0,
    ...overrides,
  };
}

function makeGroup(overrides: Partial<ExamGroup> = {}): ExamGroup {
  return {
    id: 1,
    date: "2026-10-02 10:00",
    machine: "DRI-CHEM",
    name: "血液検査",
    price: null,
    medicalRecordId: "99",
    items: [makeItem()],
    ...overrides,
  };
}

type ParentProps = ComponentProps<typeof MedicalRecordExamination>;

function parentTree(props: ParentProps) {
  return (
    <MemoryRouter>
      <MedicalRecordExamination petId="pet-1" {...props} />
    </MemoryRouter>
  );
}

function renderParent(props: ParentProps) {
  return render(parentTree(props));
}

describe("MedicalRecordExamination 検査取り込みのロック連動 (PR679-REVIEW-R5)", () => {
  it.each([
    ["ロック", { isLocked: true, isFinalized: false }],
    ["確定済みロック", { isLocked: true, isFinalized: true }],
  ] as const)("%s（作成権限あり）では取り込みを出さない", async (_label, lockedProps) => {
    const user = userEvent.setup();

    // 検査0件の空状態でも「記録を追加」は出さない
    const emptyView = renderParent({ medicalRecordId: "99", ...lockedProps });
    expect(screen.queryByRole("button", { name: "記録を追加" })).not.toBeInTheDocument();
    emptyView.unmount();

    // 検査がある場合もヘッダの「検査取り込み」は出さない
    mocks.useGetRecordExaminations.mockReturnValue({
      data: { items: [makeGroup()], isTruncated: false },
      isLoading: false,
      refetch: mocks.refetch,
    });
    renderParent({ medicalRecordId: "99", ...lockedProps });

    expect(screen.queryByRole("button", { name: "検査取り込み" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mocks.updateExamination).not.toHaveBeenCalled();

    // 絞り込みはロック中も使える（該当なしの語でグループが消える）
    expect(screen.getByText("WBC")).toBeInTheDocument();
    await user.type(screen.getByLabelText("検査項目検索"), "CRP");
    await waitFor(() => {
      expect(screen.queryByText("2026-10-02 10:00")).not.toBeInTheDocument();
    });
  });

  it("新規カルテでは取り込みの導線を出さない", () => {
    renderParent({ isNewRecord: true, medicalRecordId: undefined });

    expect(screen.getByText(/カルテ保存後に検査を追加できます/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "検査取り込み" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "記録を追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("作成権限なしでは取り込みも追加も出さない", () => {
    mocks.permission = { ...mocks.permission, canCreate: false };
    renderParent({ medicalRecordId: "99" });

    expect(screen.queryByRole("button", { name: "検査取り込み" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "記録を追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mocks.updateExamination).not.toHaveBeenCalled();
  });

  it("未ロックでは取り込みを開いて PATCH し再取得する", async () => {
    mocks.useGetExaminations.mockReturnValue({
      data: [makeExam({ id: "imp-1", testType: "取込可血液" })],
      isLoading: false,
    });

    const user = userEvent.setup();
    renderParent({ medicalRecordId: "99" });

    await user.click(screen.getByRole("button", { name: "検査取り込み" }));

    await user.click(await screen.findByRole("button", { name: /取込可血液/ }));
    await user.click(screen.getByRole("button", { name: /1件取り込む/ }));

    await waitFor(() => {
      expect(mocks.updateExamination).toHaveBeenCalledWith({
        id: "imp-1",
        req: { medical_record_id: 99 },
      });
    });
    await waitFor(() => {
      expect(mocks.refetch).toHaveBeenCalled();
    });
  });

  it("開いたままロックされると取り込みを閉じる", async () => {
    mocks.useGetExaminations.mockReturnValue({
      data: [makeExam({ id: "imp-1", testType: "取込可血液" })],
      isLoading: false,
    });

    const user = userEvent.setup();
    const view = renderParent({ medicalRecordId: "99" });

    await user.click(screen.getByRole("button", { name: "検査取り込み" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();

    view.rerender(parentTree({ medicalRecordId: "99", isLocked: true }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mocks.updateExamination).not.toHaveBeenCalled();
  });

  it("開いたまま作成権限を失うと取り込みを閉じる", async () => {
    mocks.useGetExaminations.mockReturnValue({
      data: [makeExam({ id: "imp-1", testType: "取込可血液" })],
      isLoading: false,
    });

    const user = userEvent.setup();
    const view = renderParent({ medicalRecordId: "99" });

    await user.click(screen.getByRole("button", { name: "検査取り込み" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();

    // mock 権限を再評価させる。省略時も false の isLocked を明示し、同じ患者・カルテを保つ。
    mocks.permission = { ...mocks.permission, canCreate: false };
    view.rerender(parentTree({ medicalRecordId: "99", isLocked: false }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mocks.updateExamination).not.toHaveBeenCalled();
  });
});
