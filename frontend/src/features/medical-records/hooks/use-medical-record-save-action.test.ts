import { startTransition, useLayoutEffect, useRef } from "react";
import { QueryClient } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

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
        const saveAction = useMedicalRecordSaveAction({
          recordId: "medical-record-1",
          activeTab: "問診",
          canEdit,
          isFinalized: false,
          isNextVisitDateValid: true,
          diagnosis1CategoryId: null,
          diagnosis1NameId: null,
          diagnosis2CategoryId: null,
          diagnosis2NameId: null,
          plan: "",
          assessment: "",
          chiefComplaint: "食欲低下",
          chiefComplaintDefault: "",
          chiefComplaintTypeId: null,
          treatmentPolicy: "",
          treatmentPolicyDefault: "",
          nextVisitDate: "",
          existingRecordVersion: 1,
          existingClinicalPlanVersion: 1,
          setManualErrors: vi.fn(),
          queryClient,
          updateInquiryMutation: { mutateAsync: updateInquiry },
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        });
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
        const saveAction = useMedicalRecordSaveAction({
          recordId: "medical-record-1",
          activeTab: "問診",
          canEdit: true,
          isSelectedPetDeceased,
          isFinalized: false,
          isNextVisitDateValid: true,
          diagnosis1CategoryId: null,
          diagnosis1NameId: null,
          diagnosis2CategoryId: null,
          diagnosis2NameId: null,
          plan: "",
          assessment: "",
          chiefComplaint: "食欲低下",
          chiefComplaintDefault: "",
          chiefComplaintTypeId: null,
          treatmentPolicy: "",
          treatmentPolicyDefault: "",
          nextVisitDate: "",
          existingRecordVersion: 1,
          existingClinicalPlanVersion: 1,
          setManualErrors: vi.fn(),
          queryClient,
          updateInquiryMutation: { mutateAsync: updateInquiry },
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        });
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
        useMedicalRecordSaveAction({
          recordId: "medical-record-1",
          activeTab: "診察/治療プラン",
          canEdit: true,
          isSelectedPetDeceased,
          isFinalized: false,
          isNextVisitDateValid: true,
          diagnosis1CategoryId: null,
          diagnosis1NameId: null,
          diagnosis2CategoryId: null,
          diagnosis2NameId: null,
          plan: "経過観察",
          assessment: "",
          chiefComplaint: "",
          chiefComplaintDefault: "",
          chiefComplaintTypeId: null,
          treatmentPolicy: "",
          treatmentPolicyDefault: "",
          nextVisitDate: "",
          existingRecordVersion: 1,
          existingClinicalPlanVersion: 1,
          setManualErrors: vi.fn(),
          queryClient,
          updateInquiryMutation: { mutateAsync: vi.fn() },
          updateTreatmentPlanMutation: { mutateAsync: updateTreatmentPlan },
          updateMutation: { mutateAsync: updateRecord },
        }),
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
  });
});
