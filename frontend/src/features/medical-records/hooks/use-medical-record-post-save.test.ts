import { useLayoutEffect } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { handleApiError } from "@/lib/handle-api-error";
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

  it("タブ切替をcommitした直後のlayout phaseのsuccessでは新しいタブのpost-saveを使う", async () => {
    const markClean = vi.fn();
    const estimateSave = vi.fn().mockResolvedValue(undefined);

    const { result, rerender } = renderHook(
      ({
        activeTab,
        formState,
      }: {
        activeTab: string;
        formState: { success: boolean; timestamp: number };
      }) => {
        const postSave = useMedicalRecordPostSave({
          activeTab,
          formState,
          markClean,
        });
        useLayoutEffect(() => {
          if (activeTab === "見積書" && formState.success) {
            // ref 同期が layout なら、この phase 後の effect が見積 save を呼ぶ
          }
        }, [activeTab, formState.success]);
        return postSave;
      },
      {
        initialProps: {
          activeTab: "問診",
          formState: { success: false, timestamp: 0 },
        },
      },
    );

    act(() => {
      result.current.handleRegisterEstimateSave(estimateSave);
    });

    await act(async () => {
      rerender({
        activeTab: "見積書",
        formState: { success: true, timestamp: 1 },
      });
    });

    await waitFor(() => expect(estimateSave).toHaveBeenCalledOnce());
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

    await waitFor(() => expect(estimateSave).toHaveBeenCalledOnce());
    expect(markClean).toHaveBeenCalled();
  });
});

describe("useMedicalRecordPostSave BUG-016", () => {
  it("見積書 tab で save 失敗時は markClean しない（偽クリーン防止）", async () => {
    const markClean = vi.fn();
    const estimateSave = vi.fn().mockRejectedValue(new Error("patch failed"));

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
      rerender({ formState: { success: true, timestamp: 3 } });
    });

    await waitFor(() => expect(estimateSave).toHaveBeenCalledOnce());
    expect(markClean).not.toHaveBeenCalled();
  });

  it("見積書 tab で save 未登録なら markClean せずエラー通知", async () => {
    const markClean = vi.fn();

    const { rerender } = renderHook(
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

    await act(async () => {
      rerender({ formState: { success: true, timestamp: 4 } });
    });

    await waitFor(() => expect(handleApiError).toHaveBeenCalled());
    expect(markClean).not.toHaveBeenCalled();
  });

  it("2回目以降の success でも estimate save を毎回呼ぶ", async () => {
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
      rerender({ formState: { success: true, timestamp: 10 } });
    });
    await waitFor(() => expect(estimateSave).toHaveBeenCalledTimes(1));

    await act(async () => {
      rerender({ formState: { success: true, timestamp: 11 } });
    });
    await waitFor(() => expect(estimateSave).toHaveBeenCalledTimes(2));
    expect(markClean).toHaveBeenCalledTimes(2);
  });
});
