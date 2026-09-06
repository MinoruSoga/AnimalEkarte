import { startTransition, useLayoutEffect, useRef } from "react";
import { QueryClient } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";

import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

beforeEach(() => {
  vi.mocked(toast.success).mockClear();
  vi.mocked(toast.error).mockClear();
  vi.mocked(handleApiError).mockClear();
});

function buildSaveArgs(overrides: Record<string, unknown> = {}) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return {
    recordId: "medical-record-1",
    activeTab: "診察/治療プラン",
    canEdit: true,
    isSelectedPetDeceased: false,
    isFinalized: false,
    isNextVisitDateValid: true,
    diagnosis1CategoryId: null as number | null,
    diagnosis1NameId: null as number | null,
    diagnosis2CategoryId: null as number | null,
    diagnosis2NameId: null as number | null,
    physicalExam: "",
    plan: "",
    assessment: "",
    chiefComplaint: "",
    chiefComplaintDefault: "",
    chiefComplaintTypeId: null as number | null,
    treatmentPolicy: "",
    treatmentPolicyDefault: "",
    nextVisitDate: "",
    existingRecordVersion: 1,
    existingClinicalPlanVersion: 1,
    setManualErrors: vi.fn(),
    queryClient,
    updateInquiryMutation: { mutateAsync: vi.fn() },
    updateTreatmentPlanMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
    updateMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
    ...overrides,
  };
}

describe("useMedicalRecordSaveAction BUG-010 clinical plan payload", () => {
  it("診察/治療プラン保存時に physical_exam・diagnosis_details・treatment_policy の入力値をそのまま1回の PATCH に送る（DEFAULT 固定文字列で置換しない）", async () => {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          physicalExam: "CLEAN-TEST 身体検査所見の内容ABC",
          assessment: "CLEAN-TEST 診断詳細の内容XYZ",
          plan: "CLEAN-TEST 治療方針の内容123",
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));

    expect(updateTreatmentPlan).toHaveBeenCalledOnce();
    expect(updateTreatmentPlan).toHaveBeenCalledWith({
      physical_exam: "CLEAN-TEST 身体検査所見の内容ABC",
      diagnosis_details: "CLEAN-TEST 診断詳細の内容XYZ",
      treatment_policy: "CLEAN-TEST 治療方針の内容123",
      diagnosis_type_id: undefined,
      diagnosis_name_id: undefined,
      diagnosis_2_type_id: null,
      diagnosis_2_name_id: null,
      version: 1,
    });
    // next-visit は別責務のため clinical-plan PATCH 回数に含めない
    expect(updateRecord).toHaveBeenCalledOnce();
  });

  it("空文字の明示クリアを physical_exam / diagnosis_details / treatment_policy に含めて送信する", async () => {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          physicalExam: "",
          assessment: "",
          plan: "",
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateTreatmentPlan).toHaveBeenCalledWith(
      expect.objectContaining({
        physical_exam: "",
        diagnosis_details: "",
        treatment_policy: "",
      }),
    );
  });

  it("clinical-plan PATCH が 4xx/競合で失敗した場合は成功 toast を出さない", async () => {
    const conflictError = { response: { status: 409 } };
    const updateTreatmentPlan = vi.fn().mockRejectedValue(conflictError);
    const onClinicalPlanSaved = vi.fn();
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          physicalExam: "所見",
          assessment: "診断",
          plan: "方針",
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          onClinicalPlanSaved,
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(false));
    expect(toast.success).not.toHaveBeenCalled();
    expect(handleApiError).toHaveBeenCalledWith(conflictError, "保存");
    expect(onClinicalPlanSaved).not.toHaveBeenCalled();
  });

  it("clinical-plan PATCH の成功応答だけを次の baseline snapshot として通知する", async () => {
    const savedClinicalPlan = {
      id: "cp-1",
      medical_record_id: "medical-record-1",
      version: 2,
      physical_exam: "保存済み所見",
      treatment_policy: "保存済み治療方針",
      diagnosis_details: "保存済み診断詳細",
      diagnosis_type_id: "3",
      diagnosis_name_id: "7",
      diagnosis_2_type_id: "4",
      diagnosis_2_name_id: "9",
      created_at: "2026-09-05T00:00:00Z",
      updated_at: "2026-09-05T00:00:00Z",
      diagnosis_type: { id: "3", name: "診断種別" },
      diagnosis_name: { id: "7", name: "診断名" },
      diagnosis_2_type: { id: "4", name: "診断種別2" },
      diagnosis_2_name: { id: "9", name: "診断名2" },
    };
    const updateTreatmentPlan = vi.fn().mockResolvedValue(savedClinicalPlan);
    const onClinicalPlanSaved = vi.fn();
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          onClinicalPlanSaved,
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(onClinicalPlanSaved).toHaveBeenCalledWith(
      savedClinicalPlan,
      expect.objectContaining({ diagnosis1CategoryId: null, diagnosis2NameId: null }),
    );
  });

  it("clinical-plan version 未取得（hydrate 前）は PATCH せず既存所見を空クリアしない", async () => {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          physicalExam: "",
          assessment: "",
          plan: "",
          existingClinicalPlanVersion: undefined,
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(updateTreatmentPlan).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });
});

describe("useMedicalRecordSaveAction permission boundary", () => {
  it("編集権限剥奪をcommitした直後のlayout phaseで取得済みformActionが発火してもmutationを発行しない", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const saveAction = useMedicalRecordSaveAction(
          buildSaveArgs({
            activeTab: "問診",
            canEdit,
            chiefComplaint: "食欲低下",
            queryClient,
            updateInquiryMutation: { mutateAsync: updateInquiry },
            updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
            updateMutation: { mutateAsync: updateRecord },
          }),
        );
        const capturedActionRef = useRef(saveAction.formAction);

        useLayoutEffect(() => {
          if (!canEdit) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canEdit]);

        return saveAction;
      },
      { initialProps: { canEdit: true } },
    );
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canEdit: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(updateTreatmentPlan).not.toHaveBeenCalled();
    expect(updateRecord).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe("この操作を行う権限がありません");
  });

  it("確定済みカルテの保存は権限エラーを返す", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          isFinalized: true,
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe("この操作を行う権限がありません");
  });

  it("タブ切替をcommitした直後のlayout phaseで取得済みformActionは新しいタブを参照する", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({ activeTab }: { activeTab: string }) => {
        const saveAction = useMedicalRecordSaveAction(
          buildSaveArgs({
            activeTab,
            physicalExam: "所見",
            plan: "方針",
            assessment: "診断",
            updateInquiryMutation: { mutateAsync: updateInquiry },
            updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
            updateMutation: { mutateAsync: updateRecord },
          }),
        );
        const capturedActionRef = useRef(saveAction.formAction);

        useLayoutEffect(() => {
          if (activeTab === "診察/治療プラン") {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [activeTab]);

        return saveAction;
      },
      { initialProps: { activeTab: "問診" } },
    );

    await act(async () => {
      rerender({ activeTab: "診察/治療プラン" });
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateTreatmentPlan).toHaveBeenCalledOnce();
    expect(updateInquiry).not.toHaveBeenCalled();
  });

  it("選択ペットの死亡をcommitした直後のlayout phaseで取得済みformActionが発火してもmutationを発行しない", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    const { result, rerender } = renderHook(
      ({ isSelectedPetDeceased }: { isSelectedPetDeceased: boolean }) => {
        const saveAction = useMedicalRecordSaveAction(
          buildSaveArgs({
            activeTab: "問診",
            isSelectedPetDeceased,
            chiefComplaint: "食欲低下",
            queryClient,
            updateInquiryMutation: { mutateAsync: updateInquiry },
            updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
            updateMutation: { mutateAsync: updateRecord },
          }),
        );
        const capturedActionRef = useRef(saveAction.formAction);

        useLayoutEffect(() => {
          if (isSelectedPetDeceased) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [isSelectedPetDeceased]);

        return saveAction;
      },
      { initialProps: { isSelectedPetDeceased: false } },
    );
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ isSelectedPetDeceased: true });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(updateTreatmentPlan).not.toHaveBeenCalled();
    expect(updateRecord).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe("死亡したペットのカルテは保存できません");
  });

  it("治療プラン更新待機中に選択ペットが死亡へ変わると後続のカルテ更新を発行しない", async () => {
    let resolveTreatmentPlan!: (value: undefined) => void;
    const updateTreatmentPlan = vi.fn().mockReturnValue(
      new Promise<undefined>((resolve) => {
        resolveTreatmentPlan = resolve;
      }),
    );
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    const { result, rerender } = renderHook(
      ({ isSelectedPetDeceased }: { isSelectedPetDeceased: boolean }) =>
        useMedicalRecordSaveAction(
          buildSaveArgs({
            isSelectedPetDeceased,
            plan: "経過観察",
            queryClient,
            updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
            updateMutation: { mutateAsync: updateRecord },
          }),
        ),
      { initialProps: { isSelectedPetDeceased: false } },
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateTreatmentPlan).toHaveBeenCalledOnce());

    rerender({ isSelectedPetDeceased: true });
    await act(async () => {
      resolveTreatmentPlan(undefined);
      await Promise.resolve();
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(updateRecord).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe("死亡したペットのカルテは保存できません");
  });
});

describe("useMedicalRecordSaveAction BUG-016 estimate tab", () => {
  it("見積書タブ保存では親の汎用成功トーストを出さない（post-save が正本）", async () => {
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "見積書",
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("useMedicalRecordSaveAction BUG-001 vaccination tab", () => {
  it("予防接種タブ保存では inquiry/plan/record を呼ばず汎用成功トーストを出さない", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "予防接種",
          updateInquiryMutation: { mutateAsync: updateInquiry },
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(result.current.formState.success).toBe(false);
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(updateTreatmentPlan).not.toHaveBeenCalled();
    expect(updateRecord).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("useMedicalRecordSaveAction captured formAction snapshot", () => {
  it("取得済み formAction は後続 commit の physicalExam を送る", async () => {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({ physicalExam }: { physicalExam: string }) => {
        const saveAction = useMedicalRecordSaveAction(
          buildSaveArgs({
            physicalExam,
            plan: "方針",
            assessment: "診断",
            updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
            updateMutation: { mutateAsync: updateRecord },
          }),
        );
        const capturedActionRef = useRef(saveAction.formAction);

        useLayoutEffect(() => {
          if (physicalExam === "NEW-EXAM") {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [physicalExam]);

        return saveAction;
      },
      { initialProps: { physicalExam: "OLD-EXAM" } },
    );

    await act(async () => {
      rerender({ physicalExam: "NEW-EXAM" });
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateTreatmentPlan).toHaveBeenCalledWith(
      expect.objectContaining({ physical_exam: "NEW-EXAM" }),
    );
  });

  it("確定直後の取得済み formAction は isFinalized を見て mutation しない", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({ isFinalized }: { isFinalized: boolean }) => {
        const saveAction = useMedicalRecordSaveAction(
          buildSaveArgs({
            activeTab: "問診",
            isFinalized,
            chiefComplaint: "食欲低下",
            updateInquiryMutation: { mutateAsync: updateInquiry },
          }),
        );
        const capturedActionRef = useRef(saveAction.formAction);

        useLayoutEffect(() => {
          if (isFinalized) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [isFinalized]);

        return saveAction;
      },
      { initialProps: { isFinalized: false } },
    );

    await act(async () => {
      rerender({ isFinalized: true });
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe("この操作を行う権限がありません");
  });

  it("治療タブの親保存は未送信を成功にせず汎用トーストも出さない", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "治療",
          updateInquiryMutation: { mutateAsync: updateInquiry },
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(result.current.formState.success).toBe(false);
    expect(updateInquiry).not.toHaveBeenCalled();
    expect(updateTreatmentPlan).not.toHaveBeenCalled();
    expect(updateRecord).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
