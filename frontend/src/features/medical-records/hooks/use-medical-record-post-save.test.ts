import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useMedicalRecordPostSave } from "./use-medical-record-post-save";

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

describe("useMedicalRecordPostSave BUG-010", () => {
  it("診察/治療プラン tab の success 後に clinical-plan の再保存を呼ばない（単一 writer）", async () => {
    const markClean = vi.fn();
    const clinicalPlanSave = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({ formState }) =>
        useMedicalRecordPostSave({
          activeTab: "診察/治療プラン",
          formState,
          markClean,
        }),
      {
        initialProps: {
          formState: { success: false, timestamp: 0 },
        },
      },
    );

    // 旧 API が残っていても clinical plan は登録経路を持たない
    expect(result.current).not.toHaveProperty("handleRegisterClinicalPlanSave");
    expect(result.current.handleRegisterEstimateSave).toEqual(expect.any(Function));

    // 万一旧実装が clinical save ref を残していても post-save は clinical を呼ばない
    void clinicalPlanSave;

    await act(async () => {
      rerender({ formState: { success: true, timestamp: 1 } });
    });

    expect(clinicalPlanSave).not.toHaveBeenCalled();
    expect(markClean).toHaveBeenCalled();
  });

  it("見積書 tab の success 後は estimate save を1回呼ぶ", async () => {
    const markClean = vi.fn();
    const estimateSave = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({ formState }) =>
        useMedicalRecordPostSave({
          activeTab: "見積書",
          formState,
          markClean,
        }),
      {
        initialProps: {
          formState: { success: false, timestamp: 0 },
        },
      },
    );

    act(() => {
      result.current.handleRegisterEstimateSave(estimateSave);
    });

    await act(async () => {
      rerender({ formState: { success: true, timestamp: 2 } });
    });

    expect(estimateSave).toHaveBeenCalledOnce();
    expect(markClean).toHaveBeenCalled();
  });
});
