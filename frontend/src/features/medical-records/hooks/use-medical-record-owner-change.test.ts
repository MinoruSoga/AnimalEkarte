import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { MedicalRecord } from "../api/transforms";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

const existingRecord = { version: 4 } as MedicalRecord;

function renderOwnerChange(canEdit: boolean) {
  const mutateAsync = vi.fn().mockResolvedValue(undefined);
  const hook = renderHook(() =>
    useMedicalRecordOwnerChange({
      owner: { discountRate: 0, membershipType: "regular" },
      recordId: "record-1",
      existingRecord,
      updateMutation: { mutateAsync },
      startSaveTransition: (callback) => callback(),
      canEdit,
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
