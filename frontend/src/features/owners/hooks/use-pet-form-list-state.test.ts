/**
 * BUG-415: usePetFormListState.handleSavePet — 死亡記録解除の補完発火(isPetRevival)を削除
 *
 * 旧ロジックは「編集開始時の初期 status が死亡」かつ「送信 status が生存」の遷移を検知し、
 * 汎用 Save の onSuccess から revokePetDeathMutate を追加発火していた（P2-2 Bug #2 対応）。
 * この補完発火は、汎用 PATCH (updatePetMutate) がもう status を送信しないこと、および
 * 生死の変更が PetCareSection → PetDeceasedRecordButton 経由の専用エンドポイント
 * (useRecordPetDeath/useRevokePetDeath) に一本化されたことにより不要となった
 * （専用ボタンが自身のクリックで即座にミューテーションを完結させるため、後続の汎用 Save で
 * 再度 revokePetDeathMutate を呼ぶと二重発火になる）。
 *
 * このテストは、汎用 Save (handleSavePet) がどのような status 遷移であっても
 * revokePetDeathMutate を一切呼ばないことを回帰ガードする。
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

describe("usePetFormListState.handleSavePet — 汎用 Save は生死ステータスに関与しない（BUG-415）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("初期 status=死亡 のペットを生存で保存しても revokePetDeathMutate は呼ばれない（旧 isPetRevival の回帰ガード）", () => {
    const deceasedPet = makePet({ status: "死亡" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({ id: "owner-1", initialPets: [deceasedPet], petMutations: mutations }),
    );

    // 編集開始（editingPet に初期 status=死亡 がロードされる）
    act(() => result.current.handleEditPet(deceasedPet));
    // 生存で保存（旧実装ではここで revokePetDeathMutate が補完発火していた）
    act(() => result.current.handleSavePet(makePet({ status: "生存" })));

    expect(updatePetMutate).toHaveBeenCalledTimes(1);
    expect(revokePetDeathMutate).not.toHaveBeenCalled();
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

  it("更新リクエストのペイロードに status キーを含めない（transformUpdatePetRequest の BUG-415 修正と整合）", () => {
    const deceasedPet = makePet({ status: "死亡" });
    const { mutations, updatePetMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({ id: "owner-1", initialPets: [deceasedPet], petMutations: mutations }),
    );

    act(() => result.current.handleEditPet(deceasedPet));
    act(() => result.current.handleSavePet(makePet({ status: "生存" })));

    const [{ req }] = updatePetMutate.mock.calls[0] as [{ req: Record<string, unknown> }];
    expect(req).not.toHaveProperty("status");
  });
});
