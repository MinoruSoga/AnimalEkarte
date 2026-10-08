import { useLayoutEffect, useRef } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { QueryClient } from "@tanstack/react-query";
import { useMedicalRecordQuickPatchActions } from "./use-medical-record-quick-patch-actions";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

function makeQueryClient() {
  return {
    invalidateQueries: vi.fn().mockResolvedValue(undefined),
  } as unknown as QueryClient;
}

function renderQuickPatchActions(
  canEdit: boolean,
  recordId: string | undefined = "record-1",
  isSelectedPetDeceased = false,
) {
  const mutateAsync = vi.fn().mockResolvedValue(undefined);
  const queryClient = makeQueryClient();
  const setVisitType = vi.fn();
  const hook = renderHook(() =>
    useMedicalRecordQuickPatchActions({
      recordId,
      existingRecordVersion: 3,
      visitType: "再診",
      setVisitType,
      queryClient,
      updateMutation: { mutateAsync },
      canEdit,
      isSelectedPetDeceased,
    }),
  );

  return { ...hook, mutateAsync, queryClient, setVisitType };
}

describe("useMedicalRecordQuickPatchActions — mutation permission boundary", () => {
  it("権限なしでは4種のquick patch mutationを発行しない", () => {
    const { result, mutateAsync } = renderQuickPatchActions(false);

    act(() => {
      result.current.handleChangeDoctor("7", "担当医");
      result.current.handleVisitTypeChange("初診");
      result.current.handleChangeDate("2026-08-15");
      result.current.handleFinalize();
    });

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("同一commitで権限を失ったlayout phaseでは取得済み4種callbackがmutationを発行しない", () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined);
    const queryClient = makeQueryClient();
    const { rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const actions = useMedicalRecordQuickPatchActions({
          recordId: "record-1",
          existingRecordVersion: 3,
          visitType: "再診",
          setVisitType: vi.fn(),
          queryClient,
          updateMutation: { mutateAsync },
          canEdit,
          isSelectedPetDeceased: false,
        });
        const capturedActionsRef = useRef(actions);

        useLayoutEffect(() => {
          if (!canEdit) {
            capturedActionsRef.current.handleChangeDoctor("7", "担当医");
            capturedActionsRef.current.handleVisitTypeChange("初診");
            capturedActionsRef.current.handleChangeDate("2026-08-15");
            capturedActionsRef.current.handleFinalize();
          }
        }, [canEdit]);

        return actions;
      },
      { initialProps: { canEdit: true } },
    );

    act(() => rerender({ canEdit: false }));

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("同一commitで選択ペットが死亡したlayout phaseでは取得済み4種callbackがmutationを発行しない", () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined);
    const queryClient = makeQueryClient();
    const { rerender } = renderHook(
      ({ isSelectedPetDeceased }: { isSelectedPetDeceased: boolean }) => {
        const actions = useMedicalRecordQuickPatchActions({
          recordId: "record-1",
          existingRecordVersion: 3,
          visitType: "再診",
          setVisitType: vi.fn(),
          queryClient,
          updateMutation: { mutateAsync },
          canEdit: true,
          isSelectedPetDeceased,
        });
        const capturedActionsRef = useRef(actions);

        useLayoutEffect(() => {
          if (isSelectedPetDeceased) {
            capturedActionsRef.current.handleChangeDoctor("7", "担当医");
            capturedActionsRef.current.handleVisitTypeChange("初診");
            capturedActionsRef.current.handleChangeDate("2026-08-15");
            capturedActionsRef.current.handleFinalize();
          }
        }, [isSelectedPetDeceased]);

        return actions;
      },
      { initialProps: { isSelectedPetDeceased: false } },
    );

    act(() => rerender({ isSelectedPetDeceased: true }));

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  // 来院種別は新規作成時(recordIdなし)にローカルstateのみ更新する既存契約を持つ。
  // guard追加でこの経路を殺さないことを固定する(Mode 3照合で検出したregressionの再発防止)。
  it("新規作成時(recordIdなし)は権限ありでもローカルstateを更新しmutationを発行しない", () => {
    const { result, mutateAsync, setVisitType } = renderQuickPatchActions(true, "");

    act(() => {
      result.current.handleVisitTypeChange("初診");
    });

    expect(setVisitType).toHaveBeenCalledWith("初診");
    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("権限ありでは4種のquick patchが従来どおりのpayloadで実行される", async () => {
    const { result, mutateAsync, queryClient } = renderQuickPatchActions(true);

    act(() => {
      result.current.handleChangeDoctor("7", "担当医");
      result.current.handleVisitTypeChange("初診");
      result.current.handleChangeDate("2026-08-15");
      result.current.handleFinalize();
    });

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(4));
    expect(mutateAsync).toHaveBeenNthCalledWith(1, {
      id: "record-1",
      req: { doctor_id: 7, version: 3 },
    });
    expect(mutateAsync).toHaveBeenNthCalledWith(2, {
      id: "record-1",
      req: { visit_type: "first", version: 3 },
    });
    expect(mutateAsync).toHaveBeenNthCalledWith(3, {
      id: "record-1",
      req: { date: "2026-08-15T00:00:00+09:00", version: 3 },
    });
    expect(mutateAsync).toHaveBeenNthCalledWith(4, {
      id: "record-1",
      req: { status: "finalized", version: 3 },
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledTimes(3);
  });
});
