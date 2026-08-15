import { useLayoutEffect, useRef } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { MedicalRecord } from "../api/transforms";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

const existingRecord = { version: 4 } as MedicalRecord;

function renderOwnerChange(canEdit: boolean, isSelectedPetDeceased = false) {
  const mutateAsync = vi.fn().mockResolvedValue(undefined);
  const hook = renderHook(() =>
    useMedicalRecordOwnerChange({
      owner: { discountRate: 0, membershipType: "regular" },
      recordId: "record-1",
      existingRecord,
      updateMutation: { mutateAsync },
      startSaveTransition: (callback) => callback(),
      canEdit,
      isSelectedPetDeceased,
    }),
  );

  return { ...hook, mutateAsync };
}

const sameOwnerValues = {
  id: "2",
  name: "新しい飼主",
  discountRate: 0,
  membershipType: "regular",
};

describe("useMedicalRecordOwnerChange — mutation permission boundary", () => {
  it("権限なしではowner変更mutationを発行しない", () => {
    const { result, mutateAsync } = renderOwnerChange(false);

    act(() => result.current.requestOwnerChange(sameOwnerValues));

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("同一commitで権限を失ったlayout phaseでは取得済みowner変更callbackがmutationを発行しない", () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined);
    const { rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const ownerChange = useMedicalRecordOwnerChange({
          owner: { discountRate: 0, membershipType: "regular" },
          recordId: "record-1",
          existingRecord,
          updateMutation: { mutateAsync },
          startSaveTransition: (callback) => callback(),
          canEdit,
          isSelectedPetDeceased: false,
        });
        const capturedRequestRef = useRef(ownerChange.requestOwnerChange);

        useLayoutEffect(() => {
          if (!canEdit) {
            capturedRequestRef.current(sameOwnerValues);
          }
        }, [canEdit]);

        return ownerChange;
      },
      { initialProps: { canEdit: true } },
    );

    act(() => rerender({ canEdit: false }));

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("同一commitで選択ペットが死亡したlayout phaseでは取得済みowner変更callbackがmutationを発行しない", () => {
    const mutateAsync = vi.fn().mockResolvedValue(undefined);
    const { rerender } = renderHook(
      ({ isSelectedPetDeceased }: { isSelectedPetDeceased: boolean }) => {
        const ownerChange = useMedicalRecordOwnerChange({
          owner: { discountRate: 0, membershipType: "regular" },
          recordId: "record-1",
          existingRecord,
          updateMutation: { mutateAsync },
          startSaveTransition: (callback) => callback(),
          canEdit: true,
          isSelectedPetDeceased,
        });
        const capturedRequestRef = useRef(ownerChange.requestOwnerChange);

        useLayoutEffect(() => {
          if (isSelectedPetDeceased) {
            capturedRequestRef.current(sameOwnerValues);
          }
        }, [isSelectedPetDeceased]);

        return ownerChange;
      },
      { initialProps: { isSelectedPetDeceased: false } },
    );

    act(() => rerender({ isSelectedPetDeceased: true }));

    expect(mutateAsync).not.toHaveBeenCalled();
  });

  it("権限ありではowner変更mutationが従来どおり実行される", async () => {
    const { result, mutateAsync } = renderOwnerChange(true);

    act(() => result.current.requestOwnerChange(sameOwnerValues));

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledOnce());
    expect(mutateAsync).toHaveBeenCalledWith({
      id: "record-1",
      req: { owner_id: 2, version: 4 },
    });
  });
});
