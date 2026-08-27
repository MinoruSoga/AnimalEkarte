import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ExaminationRecord } from "@/lib/transforms/examination";
import { ExaminationImportDialog } from "./ExaminationImportDialog";
import { useGetExaminations } from "@/hooks/use-examinations";
import { useUpdateExamination } from "@/hooks/use-update-examination";

vi.mock("@/hooks/use-examinations", () => ({
  useGetExaminations: vi.fn(),
}));

vi.mock("@/hooks/use-update-examination", () => ({
  useUpdateExamination: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function makeExam(overrides: Partial<ExaminationRecord> & { id: string }): ExaminationRecord {
  return {
    date: "2026-08-01",
    ownerName: "飼主",
    petName: "ポチ",
    petId: "pet-1",
    medicalRecordId: undefined,
    testType: "血液検査",
    testTypeId: "1",
    doctor: "山田",
    doctorId: "2",
    status: "依頼中",
    currentRevisionVersion: undefined,
    resultSummary: undefined,
    machine: undefined,
    items: undefined,
    ...overrides,
  };
}

const mutateAsync = vi.fn();

beforeEach(() => {
  mutateAsync.mockReset();
  mutateAsync.mockResolvedValue({});
  vi.mocked(useUpdateExamination).mockReturnValue({
    mutateAsync,
  } as unknown as ReturnType<typeof useUpdateExamination>);
  vi.mocked(useGetExaminations).mockReturnValue({
    data: [],
    isLoading: false,
  } as unknown as ReturnType<typeof useGetExaminations>);
});

function renderDialog() {
  return render(
    <ExaminationImportDialog
      open
      onOpenChange={vi.fn()}
      petId="pet-1"
      medicalRecordId="10"
      onImported={vi.fn()}
    />,
  );
}

describe("ExaminationImportDialog (BUG-014)", () => {
  it("does not offer confirmed exams as active import choices", async () => {
    vi.mocked(useGetExaminations).mockReturnValue({
      data: [
        makeExam({ id: "1", testType: "確定済み血液", status: "確定" }),
        makeExam({ id: "2", testType: "未確定尿検査", status: "依頼中" }),
      ],
      isLoading: false,
    } as unknown as ReturnType<typeof useGetExaminations>);

    renderDialog();

    expect(await screen.findByText("未確定尿検査")).toBeInTheDocument();
    expect(screen.queryByText("確定済み血液")).not.toBeInTheDocument();
  });

  it("keeps importable unconfirmed exams selectable", async () => {
    vi.mocked(useGetExaminations).mockReturnValue({
      data: [makeExam({ id: "3", testType: "取込可エコー", status: "検査中" })],
      isLoading: false,
    } as unknown as ReturnType<typeof useGetExaminations>);

    const user = userEvent.setup();
    renderDialog();

    const row = await screen.findByRole("button", { name: /取込可エコー/ });
    expect(row).toBeEnabled();
    await user.click(row);
    expect(screen.getByRole("button", { name: /1件取り込む/ })).toBeEnabled();
  });

  it("import mutation succeeds for an allowed exam", async () => {
    vi.mocked(useGetExaminations).mockReturnValue({
      data: [makeExam({ id: "42", testType: "取込成功対象", status: "結果入力済み" })],
      isLoading: false,
    } as unknown as ReturnType<typeof useGetExaminations>);

    const onImported = vi.fn();
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    render(
      <ExaminationImportDialog
        open
        onOpenChange={onOpenChange}
        petId="pet-1"
        medicalRecordId="10"
        onImported={onImported}
      />,
    );

    await user.click(await screen.findByRole("button", { name: /取込成功対象/ }));
    await user.click(screen.getByRole("button", { name: /1件取り込む/ }));

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        id: "42",
        req: { medical_record_id: 10 },
      });
    });
    await waitFor(() => {
      expect(onImported).toHaveBeenCalled();
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it("shows empty state when only non-importable exams exist", async () => {
    vi.mocked(useGetExaminations).mockReturnValue({
      data: [
        makeExam({ id: "c", testType: "確定のみ", status: "確定" }),
        makeExam({
          id: "r",
          testType: "リビジョンあり",
          status: "依頼中",
          currentRevisionVersion: 1,
        }),
      ],
      isLoading: false,
    } as unknown as ReturnType<typeof useGetExaminations>);

    renderDialog();

    expect(
      await screen.findByText("取り込める検査記録がありません"),
    ).toBeInTheDocument();
    expect(screen.queryByText("確定のみ")).not.toBeInTheDocument();
  });
});
