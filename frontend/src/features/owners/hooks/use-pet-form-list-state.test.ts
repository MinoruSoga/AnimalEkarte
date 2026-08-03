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
import {
  transformBackendPetToFrontend,
  transformCreatePetRequest,
  transformUpdatePetRequest,
} from "@/lib/transforms/pet";
import { ownersLoader } from "../loaders";
import { usePetFormListState } from "./use-pet-form-list-state";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { PetMutations } from "@/types/pet";
import type { PetFormData } from "../types";

const { mockAxiosGet } = vi.hoisted(() => ({
  mockAxiosGet: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: { get: mockAxiosGet },
}));

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

describe("dangerReason shared transform contract", () => {
  it("staff pet response の danger_reason を dangerReason へマッピングする", () => {
    const backendPet: PetResponse = {
      id: 7,
      version: 1,
      clinic_id: 1,
      owner_id: 42,
      animal_species_id: 1,
      pet_number: "42-1",
      name: "ポチ",
      pet_name_kana: "ぽち",
      gender: "male",
      status: "alive",
      breed: "",
      color: "",
      danger_level: "high",
      danger_reason: "保定時に噛む",
      food: "",
      environment: "",
      phone: "",
      remarks: "",
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
    };

    expect(transformBackendPetToFrontend(backendPet).dangerReason).toBe("保定時に噛む");
  });

  it("pets 一覧の再読込で danger_reason を dangerReason として保持する", async () => {
    mockAxiosGet.mockResolvedValue({
      data: {
        data: [{
          id: 7,
          clinic_id: 1,
          owner_id: 42,
          animal_species_id: 1,
          pet_number: "42-1",
          name: "ポチ",
          pet_name_kana: "ぽち",
          gender: "male",
          status: "alive",
          breed: "",
          color: "",
          danger_level: "high",
          danger_reason: "保定時に噛む",
          food: "",
          environment: "",
          remarks: "",
        }],
        total: 1,
        page: 1,
        limit: 20,
      },
    });

    const result = await ownersLoader({
      request: new Request("http://localhost/owners"),
    });

    expect(result.pets[0].dangerReason).toBe("保定時に噛む");
  });

  it("作成時は danger_reason の値を trim し、空なら省略する", () => {
    const withReason = transformCreatePetRequest({
      ownerId: "42",
      name: "ポチ",
      animalSpeciesId: "1",
      dangerReason: "  診察台で噛む  ",
    });
    const withoutReason = transformCreatePetRequest({
      ownerId: "42",
      name: "ポチ",
      animalSpeciesId: "1",
      dangerReason: " \t ",
    });

    expect(withReason.danger_reason).toBe("診察台で噛む");
    expect(withoutReason).not.toHaveProperty("danger_reason");
  });

  it.each([
    {
      name: "未指定なら省略",
      input: {},
      assertion: (request: Record<string, unknown>) =>
        expect(request).not.toHaveProperty("danger_reason"),
    },
    {
      name: "値なら trim して更新",
      input: { dangerReason: "  保定時に噛む  " },
      assertion: (request: Record<string, unknown>) =>
        expect(request.danger_reason).toBe("保定時に噛む"),
    },
    {
      name: "正規化後に同値なら省略",
      input: {
        dangerReason: "  保定時に噛む  ",
        originalDangerReason: "保定時に噛む",
      },
      assertion: (request: Record<string, unknown>) =>
        expect(request).not.toHaveProperty("danger_reason"),
    },
    {
      name: "空文字なら null でクリア",
      input: { dangerReason: "" },
      assertion: (request: Record<string, unknown>) =>
        expect(request.danger_reason).toBeNull(),
    },
    {
      name: "空白のみなら null でクリア",
      input: { dangerReason: " \n\t " },
      assertion: (request: Record<string, unknown>) =>
        expect(request.danger_reason).toBeNull(),
    },
  ])("更新時は danger_reason を $name", ({ input, assertion }) => {
    assertion(transformUpdatePetRequest(input) as Record<string, unknown>);
  });
});

describe("usePetFormListState dangerReason payload", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("変更していない危険理由は更新 payload から省略する", () => {
    const pet = makePet({ dangerReason: "保定時に噛む" });
    const { mutations, updatePetMutate } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(pet));
    act(() =>
      result.current.handleSavePet({
        ...pet,
        dangerReason: "  保定時に噛む  ",
      }),
    );

    const [{ req }] = updatePetMutate.mock.calls[0] as [{ req: Record<string, unknown> }];
    expect(req).not.toHaveProperty("danger_reason");
  });

  it("変更した危険理由は trim した値を更新 payload へ送る", () => {
    const pet = makePet({ dangerReason: "旧理由" });
    const { mutations, updatePetMutate } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(pet));
    act(() =>
      result.current.handleSavePet({
        ...pet,
        dangerReason: "  新しい理由  ",
      }),
    );

    const [{ req }] = updatePetMutate.mock.calls[0] as [{ req: Record<string, unknown> }];
    expect(req.danger_reason).toBe("新しい理由");
  });

  it("低へ変更して危険理由を空にしたら null を更新 payload へ送る", () => {
    const pet = makePet({ dangerLevel: "高", dangerReason: "旧理由" });
    const { mutations, updatePetMutate } = makePetMutations();
    const { result } = renderHook(() =>
      usePetFormListState({
        id: "owner-1",
        initialPets: [pet],
        petMutations: mutations,
        permissions: ALL_PERMISSIONS,
      }),
    );

    act(() => result.current.handleEditPet(pet));
    act(() =>
      result.current.handleSavePet({
        ...pet,
        dangerLevel: "低",
        dangerReason: " \t ",
      }),
    );

    const [{ req }] = updatePetMutate.mock.calls[0] as [{ req: Record<string, unknown> }];
    expect(req.danger_reason).toBeNull();
  });

  it("既存飼主への pet 作成 payload に選択した危険理由を送る", () => {
    const createPetMutate = vi.fn();
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
        makePet({ id: "", dangerLevel: "高", dangerReason: "  診察台で噛む  " }),
      ),
    );

    const [request] = createPetMutate.mock.calls[0] as [Record<string, unknown>];
    expect(request.danger_reason).toBe("診察台で噛む");
  });
});

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
    const deletePetMutate = vi.fn(
      (_id: string, callbacks: { onSuccess: () => void }) => callbacks.onSuccess(),
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
    const createPetMutate = vi.fn(
      (
        _req: unknown,
        callbacks: { onSuccess: (pet: Pet) => void },
      ) => {
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
      },
    );
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
      (
        _args: unknown,
        callbacks: { onSuccess: () => void; onError: (e: unknown) => void },
      ) => callbacks.onError(new Error("update failed")),
    );
    const deletePetMutate = vi.fn(
      (
        _id: string,
        callbacks: { onSuccess: () => void; onError: (e: unknown) => void },
      ) => callbacks.onError(new Error("delete failed")),
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
