/**
 * FE-RC-045: 元 use-pet-form-list-state.test.ts (>800行) から danger_reason 系のみ分離。
 * 対象: transformBackendPetToFrontend / ownersLoader / transformCreatePetRequest /
 * transformUpdatePetRequest の共有 transform 契約、および usePetFormListState 経由の
 * dangerReason payload 生成（trim・null クリア・未変更省略）。
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
  const updatePetMutate = vi.fn(
    (_args: unknown, callbacks: { onSuccess: () => void }) => callbacks.onSuccess(),
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
