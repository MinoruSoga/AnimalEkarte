import { startTransition } from "react";
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { UsePermissionResult } from "@/hooks/use-permission";

import type { PetChronicCondition } from "../api/pet-chronic-conditions";
import { usePetChronicConditionsSection } from "./use-pet-chronic-conditions-section";

const mocks = vi.hoisted(() => ({
  listData: [] as PetChronicCondition[],
  createMutateAsync: vi.fn(),
  updateMutateAsync: vi.fn(),
  deleteMutateAsync: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), warning: vi.fn(), error: vi.fn() },
}));

vi.mock("../api/pet-chronic-conditions", () => ({
  useGetPetChronicConditions: () => ({
    data: mocks.listData,
    isLoading: false,
    error: null,
  }),
  useCreatePetChronicCondition: () => ({
    mutateAsync: mocks.createMutateAsync,
    isPending: false,
  }),
  useUpdatePetChronicCondition: () => ({
    mutateAsync: mocks.updateMutateAsync,
    isPending: false,
  }),
  useDeletePetChronicCondition: () => ({
    mutateAsync: mocks.deleteMutateAsync,
    isPending: false,
  }),
}));

const ALL_PERMISSIONS: UsePermissionResult = {
  canView: true,
  canCreate: true,
  canEdit: true,
  canDelete: true,
};

function makeCondition(overrides: Partial<PetChronicCondition> = {}): PetChronicCondition {
  return {
    id: 1,
    clinic_id: 1,
    pet_id: 7,
    condition_code: "CKD",
    condition_name: "慢性腎臓病",
    diagnosed_at: "2024-03-01",
    notes: "定期検査",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

async function dispatchAction(action: (formData: FormData) => void) {
  await act(async () => {
    startTransition(() => action(new FormData()));
  });
}

function fillValidAddDraft(result: { current: ReturnType<typeof usePetChronicConditionsSection> }) {
  act(() => {
    result.current.setAddDraft({
      conditionCode: " CKD ",
      conditionName: " 慢性腎臓病 ",
      diagnosedAt: "2024-03-01",
      notes: "  定期検査  ",
      isActive: true,
    });
  });
}

describe("usePetChronicConditionsSection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.listData = [];
    mocks.createMutateAsync.mockResolvedValue(makeCondition());
    mocks.updateMutateAsync.mockResolvedValue(makeCondition());
    mocks.deleteMutateAsync.mockResolvedValue(undefined);
  });

  it("一覧はクエリ結果をそのまま返す", () => {
    const conditions = [makeCondition(), makeCondition({ id: 2, condition_name: "糖尿病" })];
    mocks.listData = conditions;
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));

    expect(result.current.conditions).toEqual(conditions);
  });

  it("追加は trim 済みの code/name/diagnosed_at/notes を POST する", async () => {
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));
    fillValidAddDraft(result);

    await dispatchAction(result.current.addAction);

    expect(mocks.createMutateAsync).toHaveBeenCalledWith({
      petId: "7",
      request: {
        condition_code: "CKD",
        condition_name: "慢性腎臓病",
        diagnosed_at: "2024-03-01",
        notes: "定期検査",
        is_active: true,
      },
    });
    expect(result.current.addState).toEqual({
      kind: "success",
      message: "慢性疾患を登録しました",
    });
    expect(result.current.addDraft.conditionName).toBe("");
  });

  it("空白のみの notes は null として送る", async () => {
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));
    act(() => {
      result.current.setAddDraft({
        conditionCode: "DM",
        conditionName: "糖尿病",
        diagnosedAt: "2023-11-10",
        notes: "   ",
        isActive: true,
      });
    });

    await dispatchAction(result.current.addAction);

    expect(mocks.createMutateAsync).toHaveBeenCalledWith({
      petId: "7",
      request: expect.objectContaining({ notes: null }),
    });
  });

  it("疾患名が空ならバリデーションエラーで送信しない", async () => {
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));
    act(() => {
      result.current.setAddDraft({
        conditionCode: "CKD",
        conditionName: "   ",
        diagnosedAt: "2024-03-01",
        notes: "",
        isActive: true,
      });
    });

    await dispatchAction(result.current.addAction);

    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
    expect(result.current.addState.kind).toBe("error");
    expect(result.current.addState.message).toContain("疾患名");
  });

  it("診断日が YYYY-MM-DD でないなら送信しない", async () => {
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));
    act(() => {
      result.current.setAddDraft({
        conditionCode: "CKD",
        conditionName: "慢性腎臓病",
        diagnosedAt: "2024/03/01",
        notes: "",
        isActive: true,
      });
    });

    await dispatchAction(result.current.addAction);

    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
    expect(result.current.addState.kind).toBe("error");
  });

  it("編集保存は全フィールドと明示的な is_active:false を PATCH する", async () => {
    const condition = makeCondition();
    mocks.listData = [condition];
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));
    act(() => {
      result.current.startEdit(condition);
    });
    act(() => {
      result.current.setEditDraft((prev) => ({
        ...prev,
        conditionName: " 慢性腎臓病（進行） ",
        notes: "   ",
        isActive: false,
      }));
    });

    await dispatchAction(result.current.editAction);

    expect(mocks.updateMutateAsync).toHaveBeenCalledWith({
      petId: "7",
      conditionId: 1,
      request: {
        condition_code: "CKD",
        condition_name: "慢性腎臓病（進行）",
        diagnosed_at: "2024-03-01",
        notes: null,
        is_active: false,
      },
    });
    expect(result.current.editingId).toBeNull();
    expect(result.current.editState).toEqual({
      kind: "success",
      message: "慢性疾患を更新しました",
    });
  });

  it("削除は canDelete=true のとき DELETE mutation を発行する", async () => {
    const { result } = renderHook(() => usePetChronicConditionsSection("7", true, ALL_PERMISSIONS));

    await act(async () => {
      result.current.handleDeleteCondition(1);
    });

    expect(mocks.deleteMutateAsync).toHaveBeenCalledWith({ petId: "7", conditionId: 1 });
  });

  it("canDelete=false のとき削除は発行しない", async () => {
    const { result } = renderHook(() =>
      usePetChronicConditionsSection("7", true, { ...ALL_PERMISSIONS, canDelete: false }),
    );

    await act(async () => {
      result.current.handleDeleteCondition(1);
    });

    expect(mocks.deleteMutateAsync).not.toHaveBeenCalled();
  });

  it("作成権限剥奪をcommitした後に取得済みaddActionが発火しても作成しない", async () => {
    const { result, rerender } = renderHook(
      ({ permissions }: { permissions: UsePermissionResult }) =>
        usePetChronicConditionsSection("7", true, permissions),
      { initialProps: { permissions: ALL_PERMISSIONS } },
    );
    fillValidAddDraft(result);
    const capturedAction = result.current.addAction;

    rerender({ permissions: { ...ALL_PERMISSIONS, canCreate: false } });
    await dispatchAction(capturedAction);

    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
    expect(result.current.addState.kind).toBe("error");
  });

  it("更新権限剥奪をcommitした後に取得済みeditActionが発火しても更新しない", async () => {
    const condition = makeCondition();
    mocks.listData = [condition];
    const { result, rerender } = renderHook(
      ({ permissions }: { permissions: UsePermissionResult }) =>
        usePetChronicConditionsSection("7", true, permissions),
      { initialProps: { permissions: ALL_PERMISSIONS } },
    );
    act(() => {
      result.current.startEdit(condition);
    });
    const capturedAction = result.current.editAction;

    rerender({ permissions: { ...ALL_PERMISSIONS, canEdit: false } });
    await dispatchAction(capturedAction);

    expect(mocks.updateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.editState.kind).toBe("error");
  });

  it("fieldset canEdit=false に倒れた後の取得済みaddActionは作成しない", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) =>
        usePetChronicConditionsSection("7", canEdit, ALL_PERMISSIONS),
      { initialProps: { canEdit: true } },
    );
    fillValidAddDraft(result);
    const capturedAction = result.current.addAction;

    rerender({ canEdit: false });
    await dispatchAction(capturedAction);

    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
  });
});
