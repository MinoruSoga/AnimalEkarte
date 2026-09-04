import { startTransition, useLayoutEffect, useRef } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import { createTestWrapper } from "@/testing/TestUtils";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { Owner } from "@/types/owner";
import type { PetMutations } from "@/types/pet";
import { ownerLoader } from "../loaders";
import { useOwnerForm } from "./use-owner-form";

const CREATE_PERMISSIONS = {
  canCreate: true,
  canEdit: false,
  canDelete: false,
} as const;

const EDIT_PERMISSIONS = {
  canCreate: false,
  canEdit: true,
  canDelete: false,
} as const;

const { mockAxiosGet, mockCreateOwner, mockGetOwner, mockUpdateOwner } = vi.hoisted(() => ({
  mockAxiosGet: vi.fn(),
  mockCreateOwner: vi.fn(),
  mockGetOwner: vi.fn(),
  mockUpdateOwner: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: mockAxiosGet,
  },
}));

vi.mock("../api/create-owner", () => ({
  createOwner: mockCreateOwner,
}));

vi.mock("../api/get-owner", () => ({
  getOwner: mockGetOwner,
}));

vi.mock("../api/update-owner", () => ({
  updateOwner: mockUpdateOwner,
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

function makeOwner(overrides: Partial<Owner> = {}): Owner {
  return {
    id: "123",
    ownerName: "山田太郎",
    ownerNameKana: "ヤマダタロウ",
    company: "",
    postalCode: "",
    address1: "",
    address2: "",
    homePostalCode: "",
    homeAddress1: "",
    homeAddress2: "",
    birthDate: "1980-01-02",
    phone: "090-1234-5678",
    companyPhone: "",
    email: "",
    remarks: "",
    isDangerous: false,
    discountRate: 0,
    membershipType: "non_member",
    deliveryExcluded: false,
    deliveryCaution: false,
    isTransferred: false,
    lstepOptOut: false,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    pets: [],
    ...overrides,
  };
}

async function submitForm(formAction: ReturnType<typeof useOwnerForm>["formAction"]) {
  await act(async () => {
    startTransition(() => formAction(new FormData()));
  });
}

describe("useOwnerForm birth_date payload (BUG-432)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
    mockUpdateOwner.mockResolvedValue(makeOwner());
  });

  it("新規登録時に入力した birth_date を create payload へ送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
        birthDate: "1990-04-01",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateOwner).toHaveBeenCalledWith(
        expect.objectContaining({ birth_date: "1990-04-01" }),
      );
    });
  });

  it("編集時に変更した birth_date を update payload へ送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner(), undefined, EDIT_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        birthDate: "1991-05-02",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockUpdateOwner).toHaveBeenCalledWith(
        "123",
        expect.objectContaining({ birth_date: "1991-05-02" }),
      );
    });
  });

  it("編集時に空へ戻した birth_date は null として送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner(), undefined, EDIT_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        birthDate: "",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockUpdateOwner).toHaveBeenCalledWith(
        "123",
        expect.objectContaining({ birth_date: null }),
      );
    });
  });
});

describe("useOwnerForm dangerReason readback", () => {
  it("owner detail の再読込で staff pet detail の保存済み理由を pet form 初期値へ保持する", async () => {
    const backendPet: PetResponse = {
      id: 7,
      version: 1,
      clinic_id: 1,
      owner_id: 123,
      animal_species_id: 1,
      pet_number: "123-1",
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
    const ownerWithoutDangerReason = makeOwner({
      pets: [
        transformBackendPetToFrontend({
          ...backendPet,
          danger_reason: undefined,
        }),
      ],
    });
    mockGetOwner.mockResolvedValue(ownerWithoutDangerReason);
    mockAxiosGet.mockResolvedValue({ data: backendPet });

    const { owner } = await ownerLoader({ params: { id: "123" } });

    const { result } = renderHook(() => useOwnerForm("123", owner, undefined, EDIT_PERMISSIONS), {
      wrapper: createTestWrapper(),
    });

    expect(mockAxiosGet).toHaveBeenCalledWith("/v1/pets/7");
    expect(result.current.pets[0].dangerReason).toBe("保定時に噛む");
  });
});

describe("useOwnerForm death lifecycle readback (BUG-022)", () => {
  it("full reload相当のdetail再取得後も死亡statusと死亡日時をform stateへ保持する", async () => {
    const deceasedAt = "2026-07-10T12:00:00+09:00";
    const backendPet: PetResponse = {
      id: 7,
      version: 1,
      clinic_id: 1,
      owner_id: 123,
      animal_species_id: 1,
      pet_number: "P-SYNTH-7",
      name: "合成ペット",
      pet_name_kana: "ゴウセイペット",
      gender: "unknown",
      status: "deceased",
      breed: "",
      color: "",
      danger_level: "low",
      food: "",
      environment: "",
      phone: "",
      remarks: "",
      deceased_at: deceasedAt,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-07-10T12:00:00+09:00",
    };
    mockGetOwner.mockResolvedValue(
      makeOwner({ pets: [transformBackendPetToFrontend(backendPet)] }),
    );
    mockAxiosGet.mockResolvedValue({ data: backendPet });

    const { owner } = await ownerLoader({ params: { id: "123" } });
    const { result } = renderHook(() => useOwnerForm("123", owner, undefined, EDIT_PERMISSIONS), {
      wrapper: createTestWrapper(),
    });

    expect(mockAxiosGet).toHaveBeenCalledWith("/v1/pets/7");
    expect(result.current.pets[0]).toEqual(expect.objectContaining({ status: "死亡", deceasedAt }));
  });

  it("既存owner petの欠損statusは生存へ推測しない", () => {
    const pet = transformBackendPetToFrontend({
      id: 7,
      version: 1,
      clinic_id: 1,
      owner_id: 123,
      animal_species_id: 1,
      pet_number: "P-SYNTH-7",
      name: "合成ペット",
      pet_name_kana: "ゴウセイペット",
      gender: "unknown",
      status: "",
      breed: "",
      color: "",
      danger_level: "low",
      food: "",
      environment: "",
      phone: "",
      remarks: "",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    });
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner({ pets: [pet] }), undefined, EDIT_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    expect(result.current.pets[0].status).toBe("不明");
  });
});

describe("useOwnerForm atomic owner and pets creation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
  });

  it("pending petをnested petsとして1回のowner作成で送信し個別pet作成は呼ばない", async () => {
    const createPetFn = vi.fn();
    const petMutations: PetMutations = {
      createPetFn,
      createPetMutate: vi.fn(),
      updatePetMutate: vi.fn(),
      deletePetMutate: vi.fn(),
      revokePetDeathMutate: vi.fn(),
    };
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, petMutations, CREATE_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );

    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
      }));
      result.current.handleSavePet({
        id: "pet-fixture",
        petNumber: "P001",
        petName: "ポチ",
        petNameKana: "ポチ",
        status: "生存",
        species: "犬",
        animalSpeciesId: "10",
        gender: "雄",
        birthDate: "2020-01-02",
        breed: "柴犬",
        color: "茶",
        bloodType: "DEA 1.1+",
        microchipNumber: "392123456789012",
        weight: "12.5",
        neuteredDate: "2021-03-04",
        acquisitionType: "購入",
        dangerLevel: "高",
        dangerReason: "  診察台で噛む  ",
        food: "ドライ",
        environment: "室内",
        insuranceId: "20",
        remarks: "nested fixture",
      });
    });

    await waitFor(() => {
      expect(result.current.pets).toHaveLength(1);
      expect(result.current.pets[0]).toEqual(expect.objectContaining({ isPending: true }));
    });
    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateOwner).toHaveBeenCalledTimes(1);
    });
    expect(mockCreateOwner).toHaveBeenCalledWith(
      expect.objectContaining({
        pets: [
          {
            name: "ポチ",
            animal_species_id: 10,
            name_kana: "ポチ",
            breed: "柴犬",
            color: "茶",
            blood_type: "DEA 1.1+",
            microchip_number: "392123456789012",
            gender: "male",
            status: "alive",
            birth_date: "2020-01-02T00:00:00+09:00",
            weight: 12.5,
            neutered_date: "2021-03-04T00:00:00+09:00",
            acquisition_type: "purchased",
            danger_level: "high",
            danger_reason: "診察台で噛む",
            food: "ドライ",
            environment: "室内",
            insurance_id: 20,
            remarks: "nested fixture",
          },
        ],
      }),
    );
    const nestedPet = mockCreateOwner.mock.calls[0]?.[0]?.pets?.[0];
    expect(nestedPet).not.toHaveProperty("owner_id");
    expect(nestedPet).not.toHaveProperty("pet_number");
    expect(createPetFn).not.toHaveBeenCalled();
  });
});

describe("useOwnerForm format/range validation display (BUG-023)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
    mockUpdateOwner.mockResolvedValue(makeOwner());
  });

  async function fillRequiredAndSubmit(
    setOwnerData: ReturnType<typeof useOwnerForm>["setOwnerData"],
    formAction: ReturnType<typeof useOwnerForm>["formAction"],
    overrides: Partial<ReturnType<typeof useOwnerForm>["ownerData"]>,
  ) {
    act(() => {
      setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
        ...overrides,
      }));
    });
    await submitForm(formAction);
  }

  it("不正メールで create を送らず fieldErrors.email を返す", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    await fillRequiredAndSubmit(result.current.setOwnerData, result.current.formAction, {
      email: "abc",
    });

    await waitFor(() => {
      expect(result.current.fieldErrors.email).toBe("メールアドレスの形式が正しくありません");
    });
    expect(mockCreateOwner).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });

  it("不正電話で create を送らず fieldErrors.phone を返す", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    await fillRequiredAndSubmit(result.current.setOwnerData, result.current.formAction, {
      phone: "090-ABCD-4444",
    });

    await waitFor(() => {
      expect(result.current.fieldErrors.phone).toBe(
        "電話番号の形式が正しくありません（数字・ハイフンのみ）",
      );
    });
    expect(mockCreateOwner).not.toHaveBeenCalled();
  });

  it("不正郵便番号で create を送らず fieldErrors.postalCode を返す", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    await fillRequiredAndSubmit(result.current.setOwnerData, result.current.formAction, {
      postalCode: "12-3456",
    });

    await waitFor(() => {
      expect(result.current.fieldErrors.postalCode).toBe(
        "郵便番号の形式が正しくありません（例: 123-4567）",
      );
    });
    expect(mockCreateOwner).not.toHaveBeenCalled();
  });

  it("値引率101で update を送らず fieldErrors.discountRate を返す", async () => {
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner(), undefined, EDIT_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        discountRate: 101,
      }));
    });
    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(result.current.fieldErrors.discountRate).toBe(
        "値引率は0〜100の範囲で入力してください",
      );
    });
    expect(mockUpdateOwner).not.toHaveBeenCalled();
  });

  it("正常値では fieldErrors を空のまま create する", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );

    await fillRequiredAndSubmit(result.current.setOwnerData, result.current.formAction, {
      email: "taro@example.com",
      postalCode: "123-4567",
      discountRate: 10,
    });

    await waitFor(() => {
      expect(mockCreateOwner).toHaveBeenCalledTimes(1);
    });
    expect(result.current.fieldErrors).toEqual({});
    expect(result.current.formState.success).toBe(true);
  });
});

describe("useOwnerForm create success payload (BUG-010)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner", clinicId: "2" }));
  });

  it("登録先 clinicId を formState.data に載せる", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
        clinicId: "2",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(true);
    });
    expect(mockCreateOwner).toHaveBeenCalledWith(expect.objectContaining({ clinic_id: 2 }));
    expect(result.current.formState.data).toEqual({ id: "new-owner", clinicId: "2" });
  });

  it("登録先未指定でも API 応答の clinicId を data に載せる", async () => {
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner", clinicId: "1" }));
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      { wrapper: createTestWrapper() },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(true);
    });
    expect(result.current.formState.data).toEqual({ id: "new-owner", clinicId: "1" });
  });
});

describe("useOwnerForm mutation permission boundary (FE12-02 C6a)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
    mockUpdateOwner.mockResolvedValue(makeOwner());
  });

  it("作成権限剥奪をcommitしたlayout phaseで取得済みformActionが発火してもcreateOwnerを呼ばない", async () => {
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) => {
        const form = useOwnerForm(undefined, undefined, undefined, {
          canCreate,
          canEdit: false,
          canDelete: false,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canCreate) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canCreate]);
        return form;
      },
      {
        initialProps: { canCreate: true },
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
      }));
    });
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canCreate: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockCreateOwner).not.toHaveBeenCalled();
  });

  it("更新権限剥奪をcommitしたlayout phaseで取得済みformActionが発火してもupdateOwnerを呼ばない", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const form = useOwnerForm("123", makeOwner(), undefined, {
          canCreate: false,
          canEdit,
          canDelete: false,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canEdit) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canEdit]);
        return form;
      },
      {
        initialProps: { canEdit: true },
        wrapper: createTestWrapper(),
      },
    );
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canEdit: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockUpdateOwner).not.toHaveBeenCalled();
  });
});
