/**
 * FE-RC-045: 元 use-pet-form-list-state.test.ts (>800行) から権限境界 CRUD 系のみ分離。
 * 対象: 作成/編集/削除 mutation の権限ガード（FE12-02 C6a）、pending ペットのローカル
 * のみ更新、mutation onSuccess/onError の一覧反映（BUG-002）。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { Pet } from "@/types";
import type { PetMutations } from "@/types/pet";
import type { PetFormData } from "../types";
import { usePetFormListState } from "./use-pet-form-list-state";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), warning: vi.fn(), error: vi.fn() },
}));

function makePet(overrides: Partial<PetFormData> = {}): PetFormData {
  return {
    id: "pet-1",
    petNumber: "P001",
    petName: "ポチ",
    petNameKana: "",
    status: "生存",
    species: "犬",
    animalSpeciesId: "10",
    gender: "オス",
    birthDate: "",
    color: "",
    weight: "",
    environment: "",
    remarks: "",
    ...overrides,
  };
}

function makePetMutations() {
  const updatePetMutate = vi.fn((_args: unknown, callbacks: { onSuccess: () => void }) =>
    callbacks.onSuccess(),
  );
  const mutations: PetMutations = {
    updatePetMutate,
    revokePetDeathMutate: vi.fn(),
    createPetMutate: vi.fn(),
    deletePetMutate: vi.fn(),
    createPetFn: vi.fn() as never,
  };
  return { mutations, updatePetMutate };
}

const ALL_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
} as const;

describe("usePetFormListState mutation permission boundary (FE12-02 C6a)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("取得済みcreate callbackは最新の作成権限がfalseならcreate mutationを発行しない", () => {
    const { mutations } = makePetMutations();
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) =>
        usePetFormListState({
          id: "owner-1",
          initialPets: [],
          petMutations: mutations,
          permissions: { canCreate, canEdit: true, canDelete: true },
        }),
      { initialProps: { canCreate: true } },
    );
    const capturedSave = result.current.handleSavePet;

    rerender({ canCreate: false });
    act(() => capturedSave(makePet()));

    expect(mutations.createPetMutate).not.toHaveBeenCalled();
  });

  it("取得済みupdate callbackは最新の編集権限がfalseならupdate mutationを発行しない", () => {
    const pet = makePet();
    const { mutations, updatePetMutate } = makePetMutations();
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) =>
        usePetFormListState({
          id: "owner-1",
          initialPets: [pet],
          petMutations: mutations,
          permissions: { canCreate: true, canEdit, canDelete: true },
        }),
      { initialProps: { canEdit: true } },
    );
    act(() => result.current.handleEditPet(pet));
    const capturedSave = result.current.handleSavePet;

    rerender({ canEdit: false });
    act(() => capturedSave(makePet({ remarks: "更新" })));

    expect(updatePetMutate).not.toHaveBeenCalled();
  });

  it("取得済みdelete callbackは最新の削除権限がfalseならdelete mutationを発行しない", () => {
    const pet = makePet();
    const { mutations } = makePetMutations();
    const { result, rerender } = renderHook(
      ({ canDelete }: { canDelete: boolean }) =>
        usePetFormListState({
          id: "owner-1",
          initialPets: [pet],
          petMutations: mutations,
          permissions: { canCreate: true, canEdit: true, canDelete },
        }),
      { initialProps: { canDelete: true } },
    );
    const capturedDelete = result.current.handleDeletePet;

    rerender({ canDelete: false });
    act(() => capturedDelete(pet.id));

    expect(mutations.deletePetMutate).not.toHaveBeenCalled();
  });

  it("取得済みdelete callbackは対象が明示的な死亡へ変わったらdelete mutationを発行しない", () => {
    const pet = makePet();
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );
    const capturedDelete = result.current.handleDeletePet;

    act(() => {
      result.current.setPets([{ ...pet, status: "死亡" }]);
    });
    act(() => capturedDelete(pet.id));

    expect(mutations.deletePetMutate).not.toHaveBeenCalled();
  });
});

// BUG-002 coverage: 同一 touched file の未通過分岐を最小テストで閉じ、
// フォーカス coverage 80% floor を満たす（lifecycle 同期以外の副作用は変更しない）。
describe("usePetFormListState focused coverage (BUG-002)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("BUG-002 handleAddPet はモーダルを新規モードで開く", () => {
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleAddPet());
    expect(result.current.petModalOpen).toBe(true);
    expect(result.current.editingPet).toBeNull();
  });

  it("BUG-002 pending ペット削除は mutation なしでローカルから除去する", () => {
    const pending = makePet({ id: "temp-1", isPending: true });
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: undefined,
        initialPets: [pending],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleDeletePet("temp-1"));
    expect(result.current.pets).toEqual([]);
    expect(mutations.deletePetMutate).not.toHaveBeenCalled();
  });

  it("BUG-002 既存ペット削除成功で一覧から除去する", () => {
    const pet = makePet({ id: "pet-synth-1", petName: "合成ペット甲" });
    const deletePetMutate = vi.fn((_id: string, callbacks: { onSuccess: () => void }) =>
      callbacks.onSuccess(),
    );
    const mutations: PetMutations = {
      ...makePetMutations().mutations,
      deletePetMutate,
    };
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleDeletePet("pet-synth-1"));
    expect(result.current.pets).toEqual([]);
    expect(deletePetMutate).toHaveBeenCalledTimes(1);
  });

  it("BUG-002 pending ペット編集保存はローカル map 更新のみ", () => {
    const pending = makePet({ id: "temp-1", isPending: true, petName: "合成ペット甲" });
    const { mutations, updatePetMutate } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: undefined,
        initialPets: [pending],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(pending));
    act(() =>
      result.current.handleSavePet({
        ...pending,
        petName: "合成ペット甲改",
      }),
    );

    expect(updatePetMutate).not.toHaveBeenCalled();
    expect(result.current.pets[0]?.petName).toBe("合成ペット甲改");
    expect(result.current.pets[0]?.isPending).toBe(true);
  });

  it("BUG-002 飼主未保存時の新規ペットは isPending でローカル追加する", () => {
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: undefined,
        initialPets: [],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() =>
      result.current.handleSavePet(
        makePet({ id: "", petName: "合成ペット乙", animalSpeciesId: "10" }),
      ),
    );

    expect(mutations.createPetMutate).not.toHaveBeenCalled();
    expect(result.current.pets).toHaveLength(1);
    expect(result.current.pets[0]?.isPending).toBe(true);
    expect(result.current.pets[0]?.petName).toBe("合成ペット乙");
  });

  it("BUG-002 既存飼主への作成 onSuccess で一覧へ追加する", () => {
    const createPetMutate = vi.fn((_req: unknown, callbacks: { onSuccess: (pet: Pet) => void }) => {
      callbacks.onSuccess({
        id: "pet-created-1",
        petNumber: "P-NEW",
        name: "合成ペット丙",
        petNameKana: "",
        status: "生存",
        species: "犬",
        animalSpeciesId: "10",
        gender: "オス",
        birthDate: "",
        color: "",
        weight: "",
        environment: "",
        remarks: "",
      } as Pet);
    });
    const mutations: PetMutations = {
      ...makePetMutations().mutations,
      createPetMutate,
    };
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() =>
      result.current.handleSavePet(
        makePet({ id: "", petName: "合成ペット丙", animalSpeciesId: "10" }),
      ),
    );

    expect(createPetMutate).toHaveBeenCalledTimes(1);
    expect(result.current.pets).toHaveLength(1);
    expect(result.current.pets[0]?.id).toBe("pet-created-1");
    expect(result.current.pets[0]?.petName).toBe("合成ペット丙");
  });

  it("BUG-002 更新 onError / 削除 onError / 作成 onError でも一覧 status を変えない", () => {
    const updatePetMutate = vi.fn(
      (_args: unknown, callbacks: { onSuccess: () => void; onError: (e: unknown) => void }) =>
        callbacks.onError(new Error("update failed")),
    );
    const deletePetMutate = vi.fn(
      (_id: string, callbacks: { onSuccess: () => void; onError: (e: unknown) => void }) =>
        callbacks.onError(new Error("delete failed")),
    );
    const createPetMutate = vi.fn(
      (
        _req: unknown,
        callbacks: {
          onSuccess: (pet: Pet) => void;
          onError: (e: unknown) => void;
        },
      ) => callbacks.onError(new Error("create failed")),
    );
    const pet = makePet({ id: "pet-synth-1", petName: "合成ペット甲" });
    const mutations: PetMutations = {
      ...makePetMutations().mutations,
      updatePetMutate,
      deletePetMutate,
      createPetMutate,
    };
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(pet));
    act(() => result.current.handleSavePet(makePet({ remarks: "x" })));
    act(() => result.current.handleDeletePet("pet-synth-1"));
    expect(updatePetMutate).toHaveBeenCalled();
    expect(deletePetMutate).toHaveBeenCalled();
    expect(result.current.pets[0]?.status).toBe("生存");

    // 作成経路は editingPet 無しで別 hook を使う
    const { result: createResult } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );
    act(() =>
      createResult.current.handleSavePet(
        makePet({ id: "", petName: "合成ペット丁", animalSpeciesId: "10" }),
      ),
    );
    expect(createPetMutate).toHaveBeenCalled();
    expect(createResult.current.pets).toHaveLength(0);
  });
});
