import { startTransition, useLayoutEffect, useRef } from "react";
import { QueryClient } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";

import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";
import { DEFAULT_CHIEF_COMPLAINT, DEFAULT_TREATMENT_POLICY } from "./use-medical-record-form-model";
import type { ClinicalPlanDiagnosisRow } from "../api/clinical-plan";

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
    diagnosis1CategoryId: null as number | null,
    diagnosis1NameId: null as number | null,
    diagnosis2CategoryId: null as number | null,
    diagnosis2NameId: null as number | null,
    physicalExam: "",
    plan: "",
    assessment: "",
    chiefComplaint: "",
    chiefComplaintDefault: "",
    chiefComplaintBaseline: "",
    chiefComplaintTypeId: null as number | null,
    treatmentPolicy: "",
    treatmentPolicyDefault: "",
    treatmentPolicyBaseline: "",
    existingRecordVersion: 1,
    existingClinicalPlanVersion: 1,
    setManualErrors: vi.fn(),
    queryClient,
    updateInquiryMutation: { mutateAsync: vi.fn() },
    updateTreatmentPlanMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
    // EMR-308: updateMutation（カルテ PATCH）は廃止。次回予定は tentative appointment 経路。
    // overrides で updateMutation を渡すテストは「保存がカルテ PATCH を発行しない」ことを検証する。
    ...overrides,
  };
}

// clinical-plan 応答の診断行（clinical_plan_response.go の closed shape）。
function loadedDiagnosisRow(
  slot: number,
  typeId: string | null,
  nameId: string | null,
  overrides: Partial<ClinicalPlanDiagnosisRow> = {},
): ClinicalPlanDiagnosisRow {
  return {
    id: `d-${slot}-${nameId ?? "none"}`,
    version: 1,
    diagnosis_slot: slot,
    diagnosis_type_id: typeId,
    diagnosis_name_id: nameId,
    diagnosis_name_snapshot: null,
    category_name_snapshot: null,
    sort_order: 0,
    ...overrides,
  };
}

// 読込済みの診断行: 診断1 は 2 行（スナップショット・sort_order 付き）、診断2 は 1 行。
const LOADED_DIAGNOSES: ClinicalPlanDiagnosisRow[] = [
  loadedDiagnosisRow(1, "3", "7", {
    diagnosis_name_snapshot: "胃腸炎",
    category_name_snapshot: "消化器",
    sort_order: 0,
  }),
  loadedDiagnosisRow(1, "3", "8", {
    diagnosis_name_snapshot: "胃炎",
    category_name_snapshot: "消化器",
    sort_order: 1,
  }),
  loadedDiagnosisRow(2, "4", "9", {
    diagnosis_name_snapshot: "皮膚炎",
    category_name_snapshot: "皮膚",
    sort_order: 0,
  }),
];

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
      // EMR-495: 診断を触っておらず読込済みの診断行も無いので diagnoses は送らない（= 変更なし）
      diagnoses: undefined,
      version: 1,
    });
    // EMR-308: 次回予定は tentative appointment 経路に移行したためカルテ PATCH は発行しない
    expect(updateRecord).not.toHaveBeenCalled();
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
      diagnoses: LOADED_DIAGNOSES,
      created_at: "2026-09-05T00:00:00Z",
      updated_at: "2026-09-05T00:00:00Z",
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

describe("useMedicalRecordSaveAction EMR-495 diagnoses payload", () => {
  // 読込済みの行を PATCH 形に直したもの（ID は数値、スナップショットと sort_order は読込値のまま）。
  const slot1Rows = [
    {
      diagnosis_slot: 1,
      diagnosis_type_id: 3,
      diagnosis_name_id: 7,
      diagnosis_name_snapshot: "胃腸炎",
      category_name_snapshot: "消化器",
      sort_order: 0,
    },
    {
      diagnosis_slot: 1,
      diagnosis_type_id: 3,
      diagnosis_name_id: 8,
      diagnosis_name_snapshot: "胃炎",
      category_name_snapshot: "消化器",
      sort_order: 1,
    },
  ];
  const slot2Row = {
    diagnosis_slot: 2,
    diagnosis_type_id: 4,
    diagnosis_name_id: 9,
    diagnosis_name_snapshot: "皮膚炎",
    category_name_snapshot: "皮膚",
    sort_order: 0,
  };
  // UI が読込済みの代表行（診断1 = 3/7、診断2 = 4/9）と一致している選択。
  const loadedSelection = {
    diagnosis1CategoryId: 3,
    diagnosis1NameId: 7,
    diagnosis2CategoryId: 4,
    diagnosis2NameId: 9,
  };

  async function savedPayload(overrides: Record<string, unknown>) {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          existingClinicalPlanDiagnoses: LOADED_DIAGNOSES,
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          ...overrides,
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateTreatmentPlan).toHaveBeenCalledOnce();
    return updateTreatmentPlan.mock.calls[0]?.[0] as { diagnoses?: unknown };
  }

  it("診断を触っていなければ diagnoses を送らない（読込済みの複数行も変更なし）", async () => {
    const payload = await savedPayload(loadedSelection);

    expect(payload.diagnoses).toBeUndefined();
  });

  it("変更していない診断1 の読込済み 2 行（スナップショット・sort_order 付き）はそのまま送り、変更した診断2 だけ新しい 1 行へ置換する", async () => {
    const payload = await savedPayload({ ...loadedSelection, diagnosis2NameId: 10 });

    expect(payload.diagnoses).toEqual([
      ...slot1Rows,
      {
        diagnosis_slot: 2,
        diagnosis_type_id: 4,
        diagnosis_name_id: 10,
        diagnosis_name_snapshot: null,
        category_name_snapshot: null,
        sort_order: 0,
      },
    ]);
  });

  it("診断2 だけをクリアすると診断2 の行を送らず、診断1 の読込済み 2 行は保持する", async () => {
    const payload = await savedPayload({
      ...loadedSelection,
      diagnosis2CategoryId: null,
      diagnosis2NameId: null,
    });

    expect(payload.diagnoses).toEqual(slot1Rows);
  });

  it("診断1 だけをクリアすると診断1 の 2 行を送らず、診断2 の読込済み行は保持する", async () => {
    const payload = await savedPayload({
      ...loadedSelection,
      diagnosis1CategoryId: null,
      diagnosis1NameId: null,
    });

    expect(payload.diagnoses).toEqual([slot2Row]);
  });

  it("両 slot をユーザーがクリアしたときだけ空配列で全行クリアを送る", async () => {
    const payload = await savedPayload({
      diagnosis1CategoryId: null,
      diagnosis1NameId: null,
      diagnosis2CategoryId: null,
      diagnosis2NameId: null,
    });

    expect(payload.diagnoses).toEqual([]);
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

  // EMR-308: 旧仕様は「clinical-plan 保存後に次回来院推奨日をカルテ PATCH する」第2書き込みが
  // あったが撤廃。診察/治療プラン保存は clinical-plan PATCH のみでカルテ PATCH を発行しない。
  it("診察/治療プラン保存は clinical-plan PATCH のみ発行しカルテ PATCH は行わない", async () => {
    const updateTreatmentPlan = vi.fn().mockResolvedValue(undefined);
    const updateRecord = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          plan: "経過観察",
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
    expect(updateRecord).not.toHaveBeenCalled();
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

describe("useMedicalRecordSaveAction chief_complaint_type unset/clear", () => {
  it("N unset new save: 問診 PATCH sends chief_complaint_type_id JSON null and keeps body text", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          chiefComplaintTypeId: null,
          chiefComplaint: "食欲低下",
          chiefComplaintDefault: "",
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith({
      chief_complaint: "食欲低下",
      chief_complaint_type_id: null,
      notes: undefined,
    });
  });

  it("C intentional clear: null type after prior selection still sends chief_complaint_type_id null", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result, rerender } = renderHook(
      ({ chiefComplaintTypeId }: { chiefComplaintTypeId: number | null }) =>
        useMedicalRecordSaveAction(
          buildSaveArgs({
            activeTab: "問診",
            chiefComplaintTypeId,
            chiefComplaint: "食欲低下",
            chiefComplaintDefault: "",
            updateInquiryMutation: { mutateAsync: updateInquiry },
          }),
        ),
      { initialProps: { chiefComplaintTypeId: 5 as number | null } },
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith(
      expect.objectContaining({ chief_complaint_type_id: 5 }),
    );

    updateInquiry.mockClear();
    rerender({ chiefComplaintTypeId: null });

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateInquiry).toHaveBeenCalled());
    expect(updateInquiry).toHaveBeenCalledWith(
      expect.objectContaining({
        chief_complaint: "食欲低下",
        chief_complaint_type_id: null,
      }),
    );
  });
});

describe("useMedicalRecordSaveAction EMR-215 inquiry template revert", () => {
  it("保存値が既定文と異なる場合、主訴を既定文と完全一致に戻しても既定文を送信する", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          chiefComplaint: DEFAULT_CHIEF_COMPLAINT,
          chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
          chiefComplaintBaseline: "保存済みの主訴テキスト",
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith(
      expect.objectContaining({ chief_complaint: DEFAULT_CHIEF_COMPLAINT }),
    );
  });

  it("保存値が既定文と異なる場合、問診メモを既定文と完全一致に戻しても既定文を送信する", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          treatmentPolicy: DEFAULT_TREATMENT_POLICY,
          treatmentPolicyDefault: DEFAULT_TREATMENT_POLICY,
          treatmentPolicyBaseline: "保存済みの治療方針メモ",
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith(
      expect.objectContaining({ notes: DEFAULT_TREATMENT_POLICY }),
    );
  });

  it("保存値も既定文（未保存・未取得）の場合、未編集の既定文は chief_complaint と notes の両方を undefined で送る", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          chiefComplaint: DEFAULT_CHIEF_COMPLAINT,
          chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
          chiefComplaintBaseline: DEFAULT_CHIEF_COMPLAINT,
          treatmentPolicy: DEFAULT_TREATMENT_POLICY,
          treatmentPolicyDefault: DEFAULT_TREATMENT_POLICY,
          treatmentPolicyBaseline: DEFAULT_TREATMENT_POLICY,
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith(
      expect.objectContaining({ chief_complaint: undefined, notes: undefined }),
    );
  });

  it("主訴の空文字は既定文との差分として明示クリア送信する", async () => {
    const updateInquiry = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useMedicalRecordSaveAction(
        buildSaveArgs({
          activeTab: "問診",
          chiefComplaint: "",
          chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
          chiefComplaintBaseline: "保存済みの主訴テキスト",
          updateInquiryMutation: { mutateAsync: updateInquiry },
        }),
      ),
    );

    act(() => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.formState.success).toBe(true));
    expect(updateInquiry).toHaveBeenCalledWith(expect.objectContaining({ chief_complaint: "" }));
  });
});
