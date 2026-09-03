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
 * revokePetDeathMutate を一切呼ばないこと、および BUG-002 の専用 lifecycle 同期
 * (handlePetLifecycleChange) が pets / editingPet を不変更新することを回帰ガードする。
 *
 * FE-RC-045: 元 use-pet-form-list-state.test.ts (>800行) から死亡/生存ライフサイクル系のみ分離。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
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

const ALL_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
} as const;

describe("usePetFormListState.handleSavePet — 汎用 Save は生死ステータスに関与しない（BUG-415）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("初期 status=死亡 のペットを生存で保存しても revokePetDeathMutate は呼ばれない（旧 isPetRevival の回帰ガード）", () => {
    const deceasedPet = makePet({ status: "死亡" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [deceasedPet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    // 編集開始（editingPet に初期 status=死亡 がロードされる）
    act(() => result.current.handleEditPet(deceasedPet));
    // 生存で保存（旧実装ではここで revokePetDeathMutate が補完発火していた）
    act(() => result.current.handleSavePet(makePet({ status: "生存" })));

    expect(updatePetMutate).not.toHaveBeenCalled();
    expect(revokePetDeathMutate).not.toHaveBeenCalled();
  });

  it("明示的な死亡ペットの通常編集はupdate mutationを発行しない", () => {
    const deceasedPet = makePet({ status: "死亡" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [deceasedPet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(deceasedPet));
    act(() => result.current.handleSavePet(makePet({ status: "死亡", remarks: "通常編集" })));

    expect(updatePetMutate).not.toHaveBeenCalled();
    expect(revokePetDeathMutate).not.toHaveBeenCalled();
  });

  it("初期 status=生存 のペットを生存のまま保存 → revokePetDeathMutate は呼ばれない（回帰ガード）", () => {
    const livingPet = makePet({ status: "生存" });
    const { mutations, updatePetMutate, revokePetDeathMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [livingPet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(livingPet));
    // status を触らず生存のまま保存（一般的なペット編集）
    act(() => result.current.handleSavePet(makePet({ status: "生存", remarks: "メモ更新" })));

    expect(updatePetMutate).toHaveBeenCalledTimes(1);
    expect(revokePetDeathMutate).not.toHaveBeenCalled();
  });

  it("更新リクエストのペイロードに status キーを含めない（transformUpdatePetRequest の BUG-415 修正と整合）", () => {
    const livingPet = makePet({ status: "生存" });
    const { mutations, updatePetMutate } = makePetMutations();

    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [livingPet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(livingPet));
    act(() => result.current.handleSavePet(makePet({ status: "生存" })));

    const [{ req }] = updatePetMutate.mock.calls[0] as [{ req: Record<string, unknown> }];
    expect(req).not.toHaveProperty("status");
  });
});

// BUG-002: 死亡登録/解除の成功結果を外側 pets 一覧へ不変同期する。
// モーダル内 form だけ更新しても OwnerPetsSection の生死列は pets ローカル state を見るため、
// handlePetLifecycleChange が ID 一致ペットだけを status/deceasedAt の組で差し替える必要がある。
describe("usePetFormListState.handlePetLifecycleChange (BUG-002 outer list sync)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    {
      name: "死亡成功: 対象 pet-synth-1 のみ 死亡+提出日へ更新し sibling 参照を保つ",
      change: {
        petId: "pet-synth-1",
        status: "死亡" as const,
        deceasedAt: "2026-07-15",
      },
      expectTarget: { status: "死亡", deceasedAt: "2026-07-15" },
      expectSiblingUnchanged: true,
    },
    {
      name: "解除成功: 対象 pet-synth-1 のみ 生存+deceasedAt null へ戻す",
      change: {
        petId: "pet-synth-1",
        status: "生存" as const,
        deceasedAt: null,
      },
      expectTarget: { status: "生存", deceasedAt: null },
      expectSiblingUnchanged: true,
      initialTarget: {
        status: "死亡" as const,
        deceasedAt: "2026-07-15",
      },
    },
    {
      name: "存在しない pet-foreign では一覧を変更しない",
      change: {
        petId: "pet-foreign",
        status: "死亡" as const,
        deceasedAt: "2026-07-15",
      },
      expectTarget: { status: "生存", deceasedAt: undefined },
      expectSiblingUnchanged: true,
      expectNoChange: true,
    },
  ])("BUG-002 $name", ({ change, expectTarget, expectSiblingUnchanged, initialTarget, expectNoChange }) => {
    const petA = makePet({
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: initialTarget?.status ?? "生存",
      deceasedAt: initialTarget?.deceasedAt,
    });
    const petB = makePet({
      id: "pet-synth-2",
      petName: "合成ペット乙",
      status: "生存",
    });
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [petA, petB],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );
    const siblingBefore = result.current.pets[1];

    act(() => {
      result.current.handlePetLifecycleChange(change);
    });

    const [afterA, afterB] = result.current.pets;
    expect(afterA.id).toBe("pet-synth-1");
    expect(afterA.status).toBe(expectTarget.status);
    expect(afterA.deceasedAt).toBe(expectTarget.deceasedAt);
    if (expectSiblingUnchanged) {
      expect(afterB).toBe(siblingBefore);
      expect(afterB.status).toBe("生存");
    }
    if (expectNoChange) {
      expect(afterA.status).toBe("生存");
      expect(result.current.pets).toHaveLength(2);
    }
    // 汎用 Save 経路を経由しない（lifecycle は専用 mutation の結果同期のみ）
    expect(mutations.updatePetMutate).not.toHaveBeenCalled();
    expect(mutations.revokePetDeathMutate).not.toHaveBeenCalled();
  });
});

// BUG-002 follow-up: outer pets 同期後も editingPet が古いと OwnerForm の
// editingPetRef.status === "死亡" 飼主変更ガードが stale になる。
// death 後は editingPet も 死亡+提出日、revocation 後は 生存+null に揃える。
describe("usePetFormListState.handlePetLifecycleChange (BUG-002 follow-up editingPet)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    {
      name: "death: pets と editingPet が 死亡+提出日に揃う",
      change: {
        petId: "pet-synth-1",
        status: "死亡" as const,
        deceasedAt: "2026-07-15",
      },
      initialTarget: { status: "生存" as const, deceasedAt: null as string | null },
      expectTarget: { status: "死亡", deceasedAt: "2026-07-15" },
    },
    {
      name: "revocation: pets と editingPet が 生存+null に揃う",
      change: {
        petId: "pet-synth-1",
        status: "生存" as const,
        deceasedAt: null,
      },
      initialTarget: {
        status: "死亡" as const,
        deceasedAt: "2026-07-15" as string | null,
      },
      expectTarget: { status: "生存", deceasedAt: null },
    },
  ])("BUG-002 follow-up $name", ({ change, initialTarget, expectTarget }) => {
    const petA = makePet({
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: initialTarget.status,
      deceasedAt: initialTarget.deceasedAt,
    });
    const petB = makePet({
      id: "pet-synth-2",
      petName: "合成ペット乙",
      status: "生存",
    });
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [petA, petB],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(petA));
    const siblingBefore = result.current.pets[1];
    const editingBefore = result.current.editingPet;

    act(() => {
      result.current.handlePetLifecycleChange(change);
    });

    const [afterA, afterB] = result.current.pets;
    const editingAfter = result.current.editingPet;
    expect(afterA.status).toBe(expectTarget.status);
    expect(afterA.deceasedAt).toBe(expectTarget.deceasedAt);
    expect(editingAfter).not.toBeNull();
    expect(editingAfter).not.toBe(editingBefore);
    expect(editingAfter?.id).toBe("pet-synth-1");
    expect(editingAfter?.status).toBe(expectTarget.status);
    expect(editingAfter?.deceasedAt).toBe(expectTarget.deceasedAt);
    expect(afterB).toBe(siblingBefore);
  });

  it("BUG-002 follow-up foreign ID では pets と editingPet を変更しない", () => {
    const petA = makePet({
      id: "pet-synth-1",
      petName: "合成ペット甲",
      status: "生存",
      deceasedAt: null,
    });
    const petB = makePet({
      id: "pet-synth-2",
      petName: "合成ペット乙",
      status: "生存",
    });
    const { mutations } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [petA, petB],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(petA));
    const petsBefore = result.current.pets;
    const editingBefore = result.current.editingPet;
    const siblingBefore = result.current.pets[1];

    act(() => {
      result.current.handlePetLifecycleChange({
        petId: "pet-foreign",
        status: "死亡",
        deceasedAt: "2026-07-15",
      });
    });

    expect(result.current.pets[0]?.status).toBe("生存");
    expect(result.current.editingPet).toBe(editingBefore);
    expect(result.current.editingPet?.status).toBe("生存");
    expect(result.current.pets[1]).toBe(siblingBefore);
    // foreign ID は完全 no-op: 配列参照ごと変更しない
    expect(result.current.pets).toBe(petsBefore);
  });

  it("BUG-002 follow-up handlePetLifecycleChange 参照は同等入力の再レンダーで安定", () => {
    const pet = makePet({ id: "pet-synth-1", petName: "合成ペット甲" });
    const { mutations } = makePetMutations();
    const { result, rerender } = renderHook(
      ({ permissions }: { permissions: typeof ALL_PERMISSIONS }) =>
        usePetFormListState({
          id: "owner-1",
          initialPets: [pet],
          petMutations: mutations,
          permissions,
        }),
      { initialProps: { permissions: ALL_PERMISSIONS } },
    );

    const first = result.current.handlePetLifecycleChange;
    rerender({ permissions: ALL_PERMISSIONS });
    expect(result.current.handlePetLifecycleChange).toBe(first);
  });
});
