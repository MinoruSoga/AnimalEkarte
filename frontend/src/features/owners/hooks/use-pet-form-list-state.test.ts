/**
 * P2-2 Bug #2: usePetFormListState.handleSavePet — 死亡→生存 遷移時のみ死亡記録解除
 *
 * 死亡記録の解除（revokePetDeathMutate）は「編集開始時の初期 status が死亡」かつ
 * 「送信 status が生存」の実際の遷移時のみ発火する。送信後が生存というだけでは発火しない
 * （大半のペット編集は status を触らず生存のままであり、無条件発火は不要な Lstep API 呼び出しと
 * 誤った "pet_revival" 監査ログを毎回生む）。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { usePetFormListState } from "./use-pet-form-list-state";
import type { PetMutations } from "@/types/pet";
import type { PetFormData } from "../types";

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
  // 更新成功をシミュレートするため updatePetMutate は onSuccess を即時呼ぶ。
  // revoke は onSuccess 内で発火するため、これを呼ばないと revival ケースが検出できない。
  const updatePetMutate = vi.fn(
    (_args: unknown, callbacks: { onSuccess: () => void }) => callbacks.onSuccess(),
  );
  const revokePetDeathMutate = vi.fn();
  const mutations: PetMutations = {
    updatePetMutate,
    revokePetDeathMutate,
    createPetMutate: vi.fn(),
    deletePetMutate: vi.fn(),
    createPetFn: vi.fn() as never,
  };
  return { mutations, updatePetMutate, revokePetDeathMutate };
}

describe("usePetFormListState.handleSavePet — 死亡記録解除の遷移ガード", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("初期 status=死亡 のペットを生存で保存 → revokePetDeathMutate が1回呼ばれる", () => {
    const deceasedPet = makePet({ status: "死亡" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({ id: "owner-1", initialPets: [deceasedPet], petMutations: mutations }),
    );

    // 編集開始（editingPet に初期 status=死亡 がロードされる）
    act(() => result.current.handleEditPet(deceasedPet));
    // 生存で保存（死亡→生存の遷移）
    act(() => result.current.handleSavePet(makePet({ status: "生存" })));

    expect(updatePetMutate).toHaveBeenCalledTimes(1);
    expect(revokePetDeathMutate).toHaveBeenCalledTimes(1);
    expect(revokePetDeathMutate).toHaveBeenCalledWith("pet-1");
  });

  it("初期 status=生存 のペットを生存のまま保存 → revokePetDeathMutate は呼ばれない（回帰ガード）", () => {
    const livingPet = makePet({ status: "生存" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({ id: "owner-1", initialPets: [livingPet], petMutations: mutations }),
    );

    act(() => result.current.handleEditPet(livingPet));
    // status を触らず生存のまま保存（一般的なペット編集）
    act(() => result.current.handleSavePet(makePet({ status: "生存", remarks: "メモ更新" })));

    expect(updatePetMutate).toHaveBeenCalledTimes(1);
    expect(revokePetDeathMutate).not.toHaveBeenCalled();
  });
});
