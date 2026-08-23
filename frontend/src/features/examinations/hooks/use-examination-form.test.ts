import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { useExaminationForm } from "./use-examination-form";
import { useSearchParams } from "react-router";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useDeleteExamination } from "../api/delete-examination";
import { useGetExamination } from "../api/get-examination";
import { useCreateExamination } from "../api/create-examination";
import { useUpdateExamination } from "../api/update-examination";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";

// Mock dependencies
const mockNavigate = vi.fn();

vi.mock("react-router", () => ({
  useSearchParams: vi.fn(),
  useNavigate: vi.fn(() => mockNavigate),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
}));

vi.mock("@/hooks/use-pet-selection", () => ({
  usePetSelection: vi.fn(() => ({
    selectedPets: [],
    setSelectedPets: vi.fn(),
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({
    data: null,
    isLoading: false,
  })),
}));

vi.mock("../api/get-examination", () => ({
  useGetExamination: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
}));

vi.mock("../api/create-examination", () => ({
  useCreateExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
  })),
}));

vi.mock("../api/update-examination", () => ({
  useUpdateExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
  })),
}));

vi.mock("../api/delete-examination", () => ({
  useDeleteExamination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock("../api/unconfirm-examination", () => ({
  useUnconfirmExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
    isPending: false,
  })),
}));

vi.mock("../api/get-examination-items", () => ({
  useGetExaminationItems: vi.fn(() => ({
    data: undefined,
    isSuccess: false,
    isError: false,
  })),
}));

vi.mock("../api/update-examination-items", () => ({
  useUpdateExaminationItems: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue([]),
    isPending: false,
  })),
}));

vi.mock("../api/get-exam-type-fields", () => ({
  useGetExamTypeFields: vi.fn(() => ({ data: undefined })),
}));

const ALLOWED_MUTATION_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
  canUnconfirm: true,
} as const;

function selectedPet(status: "生存" | "死亡" | "不明") {
  return {
    id: "42",
    name: "ポチ",
    ownerName: "田中",
    ownerId: "5",
    species: "犬",
    breed: "",
    birthday: "",
    gender: "男",
    weight: null,
    imageUrl: null,
    status,
    microchipNumber: null,
    insuranceNumber: null,
    insuranceExpiry: null,
    memo: null,
  };
}

function renderExaminationForm(id?: string) {
  return renderHook(() =>
    useExaminationForm(id, undefined, ALLOWED_MUTATION_PERMISSIONS),
  );
}

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — 新規作成モード", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
  });

  it("isEdit は false（id なし）", () => {
    const { result } = renderExaminationForm();
    expect(result.current.isEdit).toBe(false);
  });

  it("初期 isSaving は false", () => {
    const { result } = renderExaminationForm();
    expect(result.current.isSaving).toBe(false);
  });

  it("初期 isDeleting は false", () => {
    const { result } = renderExaminationForm();
    expect(result.current.isDeleting).toBe(false);
  });

  it('status の初期値は "依頼中"', () => {
    const { result } = renderExaminationForm();
    expect(result.current.formData.status).toBe("依頼中");
  });

  it("doctorId なしの場合、formData.doctorId は undefined", () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(""),
      vi.fn(),
    ]);
    const { result } = renderExaminationForm();
    expect(result.current.formData.doctorId).toBeUndefined();
  });

  it("doctorId あり → formData.doctorId に反映される", () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("doctorId=789"),
      vi.fn(),
    ]);
    const { result } = renderExaminationForm();
    expect(result.current.formData.doctorId).toBe("789");
  });

  it("複数クエリパラメータがある場合、doctorId を正確に抽出", () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("petId=456&doctorId=789&medicalRecordId=101"),
      vi.fn(),
    ]);
    const { result } = renderExaminationForm();
    expect(result.current.formData.doctorId).toBe("789");
  });

  it("petId がない & ローディングでもない場合、ペット選択ページへリダイレクト", async () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(""),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    await act(async () => {
      renderExaminationForm();
      await Promise.resolve();
    });

    expect(mockNavigate).toHaveBeenCalledWith(
      expect.stringContaining("select-pet"),
    );
  });
});

describe("useExaminationForm — 患者変更と確定解除", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: selectedPet("生存"),
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [{ ...selectedPet("生存"), id: "84" }],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "完了" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);
  });

  afterEach(async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUnconfirmExamination } =
      await import("../api/unconfirm-examination");
    vi.mocked(useGetExamination).mockReturnValue({ data: null } as ReturnType<
      typeof useGetExamination
    >);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: undefined,
      isSuccess: false,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUnconfirmExamination).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
      isPending: false,
    } as ReturnType<typeof useUnconfirmExamination>);
  });

  it("初回confirm前の患者変更をpet_idとしてPATCHする", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    expect(result.current.isPatientChangeLocked).toBe(false);

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({ pet_id: 84 }),
      }),
    );
  });

  it.each([{ status: "完了" as const, currentRevisionVersion: 2 }])(
    "履歴がある状態では患者変更をPATCHしない: $status/$currentRevisionVersion",
    async ({ status, currentRevisionVersion }) => {
      const { useGetExamination } = await import("../api/get-examination");
      const { useUpdateExamination } =
        await import("../api/update-examination");
      vi.mocked(usePetSelection).mockReturnValue({
        selectedPets: [selectedPet("生存")],
        setSelectedPets: vi.fn(),
      } as ReturnType<typeof usePetSelection>);
      vi.mocked(useGetExamination).mockReturnValue({
        data: {
          id: "exam-001",
          petId: "42",
          testTypeId: "5",
          doctorId: "3",
          status,
          currentRevisionVersion,
          ownerName: "",
          petName: "",
          date: "",
        },
      } as ReturnType<typeof useGetExamination>);
      const updateMutate = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateExamination).mockReturnValue({
        mutateAsync: updateMutate,
      } as ReturnType<typeof useUpdateExamination>);

      const { result } = renderExaminationForm("exam-001");
      expect(result.current.isPatientChangeLocked).toBe(true);

      await act(async () => {
        startTransition(() => result.current.formAction(new FormData()));
      });

      await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
      expect(updateMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          req: expect.not.objectContaining({ pet_id: expect.anything() }),
        }),
      );
    },
  );

  it("確定済みでは formAction が PATCH 自体を発行しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useUpdateExamination } = await import("../api/update-examination");
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [selectedPet("生存")],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "確定" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    expect(result.current.isPatientChangeLocked).toBe(true);
    expect(result.current.isPersistedConfirmed).toBe(true);

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() =>
      expect(result.current.formState.success).toBe(false),
    );
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("revision lock が到着済みの異なる患者候補を保存せず fail-closed にする", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useUpdateExamination } = await import("../api/update-examination");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "完了" as const,
        currentRevisionVersion: 2,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });

  it("状態不明の患者候補を pet_id として保存しない", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [{ ...selectedPet("不明"), id: "84" }],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });

  it("確定解除はtrim済み理由と最新の専用権限でだけ実行する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useUnconfirmExamination } =
      await import("../api/unconfirm-examination");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "確定" as const,
        currentRevisionVersion: 1,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    const unconfirmMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUnconfirmExamination).mockReturnValue({
      mutateAsync: unconfirmMutate,
      isPending: false,
    } as ReturnType<typeof useUnconfirmExamination>);

    const { result, rerender } = renderHook(
      ({ canUnconfirm }: { canUnconfirm: boolean }) =>
        useExaminationForm("exam-001", undefined, {
          ...ALLOWED_MUTATION_PERMISSIONS,
          canEdit: false,
          canUnconfirm,
        }),
      { initialProps: { canUnconfirm: true } },
    );

    await expect(result.current.handleUnconfirm("   ")).resolves.toBe(false);
    await expect(
      result.current.handleUnconfirm("あ".repeat(501)),
    ).resolves.toBe(false);
    expect(unconfirmMutate).not.toHaveBeenCalled();

    await expect(
      result.current.handleUnconfirm("  再確認のため  "),
    ).resolves.toBe(true);
    expect(unconfirmMutate).toHaveBeenCalledWith({
      id: "exam-001",
      reason: "再確認のため",
    });

    rerender({ canUnconfirm: false });
    await expect(result.current.handleUnconfirm("再確認のため")).resolves.toBe(
      false,
    );
    expect(unconfirmMutate).toHaveBeenCalledOnce();
  });
});

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — petFromQuery あり", () => {
  const mockPet = {
    id: "42",
    name: "ポチ",
    ownerName: "田中太郎",
    ownerId: "5",
    species: "犬",
    breed: "",
    birthday: "",
    gender: "男",
    weight: null,
    imageUrl: null,
    status: "生存" as const,
    microchipNumber: null,
    insuranceNumber: null,
    insuranceExpiry: null,
    memo: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("petId=42"),
      vi.fn(),
    ]);
  });

  it("petFromQuery が非null のとき setSelectedPets が呼ばれる（line 140）", async () => {
    const mockSetSelectedPets = vi.fn();
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: mockSetSelectedPets,
    } as ReturnType<typeof usePetSelection>);

    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    await act(async () => {
      renderExaminationForm();
      await Promise.resolve();
    });

    expect(mockSetSelectedPets).toHaveBeenCalledWith([mockPet]);
  });

  it("petFromQuery から ownerName / petName を formData に反映する", async () => {
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [mockPet],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    const { result } = renderExaminationForm();

    expect(result.current.formData.ownerName).toBe("田中太郎");
    expect(result.current.formData.petName).toBe("ポチ");
  });
});

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — 編集モード（id あり）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("id を渡すと isEdit = true になる", () => {
    const { result } = renderExaminationForm("exam-001");
    expect(result.current.isEdit).toBe(true);
  });

  it("isEdit = true のとき useEffect でリダイレクトしない", async () => {
    await act(async () => {
      renderExaminationForm("exam-001");
      await Promise.resolve();
    });
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it("handleDelete は isEdit = false のとき何もしない", () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderExaminationForm(); // no id → isEdit = false
    act(() => {
      result.current.handleDelete();
    });
    expect(mockMutate).not.toHaveBeenCalled();
  });

  it("handleDelete は isEdit = true のとき deleteMutation.mutate を呼ぶ（line 151-155）", async () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    expect(mockMutate).toHaveBeenCalledWith(
      "exam-001",
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("handleDelete の onSuccess コールバックが toast.success を呼ぶ", async () => {
    const { toast } = await import("sonner");
    let capturedOnSuccess: (() => void) | undefined;
    const mockMutate = vi.fn((_id: string, opts: { onSuccess: () => void }) => {
      capturedOnSuccess = opts.onSuccess;
    });
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    act(() => {
      capturedOnSuccess?.();
    });
    expect(toast.success).toHaveBeenCalledWith("検査記録を削除しました");
  });
});

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — formAction（useActionState コールバック）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("testTypeId と doctorId がない場合、バリデーションエラーを返す（line 92-97）", async () => {
    const { result } = renderExaminationForm();

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.fieldErrors).toMatchObject({
      testTypeId: expect.any(String),
      doctorId: expect.any(String),
    });
  });

  it("doctorId のみない場合、doctorId バリデーションエラーを返す", async () => {
    const { result } = renderExaminationForm();

    act(() => {
      result.current.setFormData({ testTypeId: "5" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.fieldErrors?.doctorId).toBeDefined();
    expect(result.current.formState.fieldErrors?.testTypeId).toBeUndefined();
  });

  it("BUG-017: 空送信後に testTypeId を直すと当該 error のみ消え doctorId error は残る", async () => {
    const { result } = renderExaminationForm();

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.fieldErrors).toMatchObject({
      testTypeId: "検査種別を選択してください",
      doctorId: "担当医を選択してください",
    });
    expect(result.current.fieldErrors).toMatchObject({
      testTypeId: "検査種別を選択してください",
      doctorId: "担当医を選択してください",
    });

    act(() => {
      result.current.setFormData({ testTypeId: "5" });
    });

    expect(result.current.fieldErrors?.testTypeId).toBeUndefined();
    expect(result.current.fieldErrors?.doctorId).toBe(
      "担当医を選択してください",
    );
  });

  it("バリデーション通過 & 新規 & selectedPets なし → success: false（line 115）", async () => {
    // selectedPets = [] なので pet がない → early return
    const { result } = renderExaminationForm();

    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.success).toBe(false);
  });

  it("バリデーション通過 & 新規 & selectedPets あり → createMutation.mutateAsync 呼ぶ（line 125）", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: mockMutateAsync,
    } as ReturnType<typeof useCreateExamination>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存",
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();

    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        pet_id: 42,
        exam_type_id: 5,
        doctor_id: 3,
      }),
    );
    expect(result.current.formState.success).toBe(true);
  });

  it("mutateAsync が失敗した場合、toast.error を呼ぶ（line 129-130）", async () => {
    const { toast } = await import("sonner");
    const { useCreateExamination } = await import("../api/create-examination");
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: vi.fn().mockRejectedValue(new Error("API error")),
    } as ReturnType<typeof useCreateExamination>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存",
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(toast.error).toHaveBeenCalledWith("API error");
    expect(result.current.formState.success).toBe(false);
  });

  it("編集モード & バリデーション通過 → updateMutation.mutateAsync 呼ぶ（line 112）", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: mockMutateAsync,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存",
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm("exam-001");

    act(() => {
      result.current.setFormData({
        testTypeId: "5",
        doctorId: "3",
        status: "完了",
        resultSummary: "正常",
      });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          status: "completed",
          result_summary: "正常",
        }),
      }),
    );
    expect(result.current.formState.success).toBe(true);
  });
});

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — setFormData（ローカルオーバーライド）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("setFormData でフィールドを更新できる", () => {
    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ resultSummary: "異常なし" });
    });
    expect(result.current.formData.resultSummary).toBe("異常なし");
  });

  it("setFormData を複数回呼んだ場合マージされる", () => {
    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ resultSummary: "異常なし" });
    });
    act(() => {
      result.current.setFormData({ machine: "MRI" });
    });
    expect(result.current.formData.resultSummary).toBe("異常なし");
    expect(result.current.formData.machine).toBe("MRI");
  });

  it("setFormData で doctorId を上書きできる", () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("doctorId=1"),
      vi.fn(),
    ]);
    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ doctorId: "99" });
    });
    expect(result.current.formData.doctorId).toBe("99");
  });
});

// ─────────────────────────────────────────────────────────────

describe("useExaminationForm — 検査項目テーブル（FE-EXAM-001）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("初期 formItems は空配列", () => {
    const { result } = renderExaminationForm();
    expect(result.current.formItems).toEqual([]);
  });

  it("編集モードで既存 items を formItems に反映する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "4.0-12.0",
          unit: "x10^3/μL",
          referenceValue: "4.0-12.0",
          refMin: 4,
          refMax: 12,
          isAssessed: true,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const { result } = renderExaminationForm("exam-001");
    expect(result.current.formItems).toHaveLength(1);
    expect(result.current.formItems[0].name).toBe("WBC");
    expect(result.current.formItems[0].inspectionValue).toBe("5.0");
    expect(result.current.formItems[0].status).toBe("normal");
    expect(result.current.formItems[0].isAssessed).toBe(true);
  });

  it("setInspectionValue で指定 row の inspectionValue を更新できる", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const { result } = renderExaminationForm("exam-001");
    const key = result.current.formItems[0].key;
    act(() => {
      result.current.setInspectionValue(key, "7.5");
    });
    expect(result.current.formItems[0].inspectionValue).toBe("7.5");
  });

  it("手動行をimmutableに追加・改名・削除できる", () => {
    const { result } = renderExaminationForm();
    const before = result.current.formItems;

    act(() => {
      result.current.addManualItem();
    });
    const afterAdd = result.current.formItems;

    expect(afterAdd).not.toBe(before);
    expect(before).toEqual([]);
    expect(afterAdd).toHaveLength(1);
    expect(afterAdd[0].examTypeFieldId).toBeUndefined();
    expect(afterAdd[0]).toMatchObject({ name: "", inspectionValue: "" });

    const key = afterAdd[0].key;
    act(() => {
      result.current.setItemName(key, "手動項目");
    });
    expect(result.current.formItems).not.toBe(afterAdd);
    expect(afterAdd[0].name).toBe("");
    expect(result.current.formItems[0].name).toBe("手動項目");

    const afterRename = result.current.formItems;
    act(() => {
      result.current.removeItem(key);
    });
    expect(result.current.formItems).not.toBe(afterRename);
    expect(afterRename).toHaveLength(1);
    expect(result.current.formItems).toEqual([]);
  });

  it("追加した手動行の名前と結果値をPATCH itemsへ送る", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "41",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    act(() => {
      result.current.addManualItem();
    });
    const key = result.current.formItems[0].key;
    act(() => {
      result.current.setItemName(key, "手動項目");
      result.current.setInspectionValue(key, "記録値");
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          items: expect.arrayContaining([
            expect.objectContaining({
              exam_type_field_id: null,
              name: "手動項目",
              inspection_value: "記録値",
            }),
          ]),
        }),
      }),
    );
  });

  it("結果値がある手動行の空名を拒否し、silent drop しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "41",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    act(() => result.current.addManualItem());
    const key = result.current.formItems[0].key;
    act(() => result.current.setInspectionValue(key, "記録値"));

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors).toEqual({
      examItems: "結果値を入力した手動項目には項目名が必要です",
    });
  });

  it("編集モード保存時に表示値を保持し判定境界を含まない items を PATCH する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "4.0-12.0",
          unit: "x10^3/μL",
          referenceValue: "4.0-12.0",
          refMin: 4,
          refMax: 12,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存" as const,
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          items: [
            {
              exam_type_field_id: 1,
              name: "WBC",
              inspection_value: "5.0",
              normal_value: "4.0-12.0",
              unit: "x10^3/μL",
              reference_value: "4.0-12.0",
              sort_order: 1,
            },
          ],
        }),
      }),
    );
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("編集モードで items が空でも空配列を PATCH に含める", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({ items: [] }),
      }),
    );
  });

  it.each([
    [
      "取得中",
      {
        data: undefined,
        isSuccess: false,
        isError: false,
        isPending: true,
      },
    ],
    [
      "取得失敗",
      {
        data: [],
        isSuccess: false,
        isError: true,
        isPending: false,
        error: new Error("items unavailable"),
      },
    ],
  ])(
    "編集モードで検査項目が%sなら保存 mutation を発行しない",
    async (_label, itemQueryResult) => {
      const { useGetExamination } = await import("../api/get-examination");
      const { useGetExaminationItems } =
        await import("../api/get-examination-items");
      const { useUpdateExamination } =
        await import("../api/update-examination");

      vi.mocked(useGetExamination).mockReturnValue({
        data: {
          id: "exam-001",
          testTypeId: "5",
          doctorId: "3",
          status: "検査中" as const,
          ownerName: "",
          petName: "",
          date: "",
        },
      } as ReturnType<typeof useGetExamination>);
      vi.mocked(useGetExaminationItems).mockReturnValue(
        itemQueryResult as ReturnType<typeof useGetExaminationItems>,
      );
      const updateMutate = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateExamination).mockReturnValue({
        mutateAsync: updateMutate,
      } as ReturnType<typeof useUpdateExamination>);

      const { result } = renderExaminationForm("exam-001");
      await waitFor(() =>
        expect(result.current.formData.testTypeId).toBe("5"),
      );

      await act(async () => {
        startTransition(() => result.current.formAction(new FormData()));
      });

      await waitFor(() =>
        expect(result.current.formState.fieldErrors).toEqual({
          examItems: "検査項目の読み込み完了後に保存してください",
        }),
      );
      expect(updateMutate).not.toHaveBeenCalled();
    },
  );

  it("検査ID変更時は旧itemsを破棄し、新IDの取得完了後だけ新itemsをPATCHする", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useGetExamTypeFields } =
      await import("../api/get-exam-type-fields");
    const { useUpdateExamination } = await import("../api/update-examination");

    const examinationFor = (examID: string) => ({
      id: examID,
      testTypeId: examID === "exam-a" ? "5" : "6",
      doctorId: examID === "exam-a" ? "3" : "4",
      status: "検査中" as const,
      ownerName: "",
      petName: "",
      date: examID === "exam-a" ? "2026-08-01" : "2026-08-02",
      resultSummary:
        examID === "exam-a" ? "server summary A" : "server summary B",
      machine: examID === "exam-a" ? "server machine A" : "server machine B",
    });
    const itemFor = (id: string, name: string, inspectionValue: string) => ({
      id,
      examTypeFieldId: 1,
      name,
      result: "",
      inspectionValue,
      normalValue: "",
      unit: "",
      referenceValue: "",
      refMin: undefined,
      refMax: undefined,
      isAbnormal: false,
      status: "normal" as const,
      sortOrder: 1,
    });
    let examinationBItemsReady = false;
    vi.mocked(useGetExamination).mockImplementation(
      (examID) =>
        ({ data: examinationFor(examID) }) as ReturnType<
          typeof useGetExamination
        >,
    );
    vi.mocked(useGetExaminationItems).mockImplementation((examID) => {
      if (examID === "exam-a") {
        return {
          data: [itemFor("item-a", "A-WBC", "1")],
          isSuccess: true,
          isError: false,
          isPending: false,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      if (!examinationBItemsReady) {
        return {
          data: undefined,
          isSuccess: false,
          isError: false,
          isPending: true,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      return {
        data: [itemFor("item-b", "B-WBC", "2")],
        isSuccess: true,
        isError: false,
        isPending: false,
      } as ReturnType<typeof useGetExaminationItems>;
    });
    vi.mocked(useGetExamTypeFields).mockImplementation(
      (examTypeID) =>
        ({
          data: [
            {
              id: Number(examTypeID) * 100,
              name: `TEMPLATE-${examTypeID}`,
              unit: "",
              normalValue: "",
              sortOrder: 1,
            },
          ],
        }) as ReturnType<typeof useGetExamTypeFields>,
    );
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result, rerender } = renderHook(
      ({ examID }: { examID: string }) =>
        useExaminationForm(
          examID,
          undefined,
          ALLOWED_MUTATION_PERMISSIONS,
        ),
      { initialProps: { examID: "exam-a" } },
    );
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("A-WBC"));

    act(() => {
      result.current.setFormData({
        testTypeId: "7",
        status: "確定",
        resultSummary: "unsaved summary A",
        machine: "unsaved machine A",
      });
    });
    await waitFor(() => expect(result.current.formData.testTypeId).toBe("7"));

    rerender({ examID: "exam-b" });
    expect(result.current.formData).toMatchObject({
      testTypeId: "6",
      doctorId: "4",
      status: "検査中",
      resultSummary: "server summary B",
      machine: "server machine B",
      date: "2026-08-02",
    });
    expect(result.current.formItems).toEqual([]);
    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() =>
      expect(result.current.formState.fieldErrors?.examItems).toBe(
        "検査項目の読み込み完了後に保存してください",
      ),
    );
    expect(updateMutate).not.toHaveBeenCalled();

    examinationBItemsReady = true;
    rerender({ examID: "exam-b" });
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("B-WBC"));
    expect(result.current.formItems[0]?.name).not.toBe("TEMPLATE-6");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-b",
        req: expect.objectContaining({
          status: "in_progress",
          result_summary: "server summary B",
          machine: "server machine B",
          items: [expect.objectContaining({ name: "B-WBC" })],
        }),
      }),
    );
  });

  it("A取得済み→B取得中→A再表示でもA itemsを復元してAだけをPATCHする", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    const itemA = {
      id: "item-a",
      examTypeFieldId: 1,
      name: "A-WBC",
      result: "",
      inspectionValue: "1",
      normalValue: "",
      unit: "",
      referenceValue: "",
      refMin: undefined,
      refMax: undefined,
      isAbnormal: false,
      status: "normal" as const,
      sortOrder: 1,
    };
    vi.mocked(useGetExamination).mockImplementation(
      (examID) =>
        ({
          data: {
            id: examID,
            testTypeId: "5",
            doctorId: "3",
            status: "検査中" as const,
            ownerName: "",
            petName: "",
            date: "2026-08-01",
          },
        }) as ReturnType<typeof useGetExamination>,
    );
    vi.mocked(useGetExaminationItems).mockImplementation((examID) => {
      if (examID === "exam-a") {
        return {
          data: [itemA],
          isSuccess: true,
          isError: false,
          isPending: false,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      return {
        data: undefined,
        isSuccess: false,
        isError: false,
        isPending: true,
      } as ReturnType<typeof useGetExaminationItems>;
    });
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result, rerender } = renderHook(
      ({ examID }: { examID: string }) =>
        useExaminationForm(
          examID,
          undefined,
          ALLOWED_MUTATION_PERMISSIONS,
        ),
      { initialProps: { examID: "exam-a" } },
    );
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("A-WBC"));

    rerender({ examID: "exam-b" });
    await waitFor(() => expect(result.current.formItems).toEqual([]));

    rerender({ examID: "exam-a" });
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("A-WBC"));

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-a",
        req: expect.objectContaining({
          items: [expect.objectContaining({ name: "A-WBC" })],
        }),
      }),
    );
  });

  it("同じ検査種別のA→B遷移後もBの種別変更で新テンプレへ再構築する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useGetExamTypeFields } =
      await import("../api/get-exam-type-fields");

    const templates = {
      "5": [
        {
          id: 500,
          name: "TEMPLATE-5",
          unit: "",
          normalValue: "",
          sortOrder: 1,
        },
      ],
      "6": [
        {
          id: 600,
          name: "TEMPLATE-6",
          unit: "",
          normalValue: "",
          sortOrder: 1,
        },
      ],
    };
    vi.mocked(useGetExamination).mockImplementation(
      (examID) =>
        ({
          data: {
            id: examID,
            testTypeId: "5",
            doctorId: "3",
            status: "検査中" as const,
            ownerName: "",
            petName: "",
            date: "2026-08-01",
          },
        }) as ReturnType<typeof useGetExamination>,
    );
    vi.mocked(useGetExaminationItems).mockImplementation(
      (examID) =>
        ({
          data: [
            {
              id: `item-${examID}`,
              examTypeFieldId: 1,
              name: `${examID}-WBC`,
              result: "",
              inspectionValue: "1",
              normalValue: "",
              unit: "",
              referenceValue: "",
              refMin: undefined,
              refMax: undefined,
              isAbnormal: false,
              status: "normal" as const,
              sortOrder: 1,
            },
          ],
          isSuccess: true,
          isError: false,
          isPending: false,
        }) as ReturnType<typeof useGetExaminationItems>,
    );
    vi.mocked(useGetExamTypeFields).mockImplementation(
      (examTypeID) =>
        ({ data: templates[examTypeID as keyof typeof templates] }) as ReturnType<
          typeof useGetExamTypeFields
        >,
    );

    const { result, rerender } = renderHook(
      ({ examID }: { examID: string }) =>
        useExaminationForm(
          examID,
          undefined,
          ALLOWED_MUTATION_PERMISSIONS,
        ),
      { initialProps: { examID: "exam-a" } },
    );
    await waitFor(() =>
      expect(result.current.formItems[0]?.name).toBe("exam-a-WBC"),
    );

    rerender({ examID: "exam-b" });
    await waitFor(() =>
      expect(result.current.formItems[0]?.name).toBe("exam-b-WBC"),
    );

    act(() => result.current.setFormData({ testTypeId: "6" }));
    await waitFor(() =>
      expect(result.current.formItems[0]?.name).toBe("TEMPLATE-6"),
    );
  });

  it("確定済み (status=確定) では PATCH を発行しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "確定" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() =>
      expect(result.current.isPersistedConfirmed).toBe(true),
    );
    expect(result.current.isPersistedResultsLocked).toBe(true);
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("BUG-033: 完了シールでは PATCH から items を省略し status 遷移を送る", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "完了" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await waitFor(() =>
      expect(result.current.isPersistedCompletedLocked).toBe(true),
    );
    expect(result.current.isPersistedResultsLocked).toBe(true);

    await act(async () => {
      result.current.setFormData({ status: "確定" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          status: "confirmed",
        }),
      }),
    );
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        req: expect.not.objectContaining({ items: expect.anything() }),
      }),
    );
  });

  it("未確定検査でステータスを確定に変えても items を PATCH に含め保存できる（A-S02-01）", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } =
      await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "結果入力済み" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await waitFor(() =>
      expect(result.current.isPersistedConfirmed).toBe(false),
    );

    await act(async () => {
      result.current.setFormData({ status: "確定" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          status: "confirmed",
          items: expect.arrayContaining([
            expect.objectContaining({ name: "WBC", inspection_value: "5.0" }),
          ]),
        }),
      }),
    );
    // ドラフト選択だけでは isPersistedConfirmed は true にならない
    expect(result.current.isPersistedConfirmed).toBe(false);
  });

  it("新規保存時に表示値を保持し判定境界を含まない items を POST する", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const { useGetExamTypeFields } =
      await import("../api/get-exam-type-fields");

    // テンプレ展開で formItems が初期化される
    vi.mocked(useGetExamTypeFields).mockReturnValue({
      data: [
        {
          id: 1,
          name: "WBC",
          unit: "x10^3/μL",
          normalValue: "4.0-12.0",
          sortOrder: 1,
        },
      ],
    } as ReturnType<typeof useGetExamTypeFields>);

    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存" as const,
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();

    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    // テンプレが反映されるのを待つ（useEffect で自動展開）
    await act(async () => {
      await Promise.resolve();
    });

    // 値を入力
    if (result.current.formItems.length > 0) {
      act(() => {
        result.current.setInspectionValue(
          result.current.formItems[0].key,
          "7.5",
        );
      });
    }

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(createMutate).toHaveBeenCalledOnce());
    expect(createMutate).toHaveBeenCalledWith({
      medical_record_id: null,
      pet_id: 42,
      exam_type_id: 5,
      doctor_id: 3,
      date: jstDateStartISOString(todayJSTISO()),
      result_summary: undefined,
      machine: undefined,
      items: [
        {
          exam_type_field_id: 1,
          name: "WBC",
          inspection_value: "7.5",
          normal_value: "4.0-12.0",
          unit: "x10^3/μL",
          reference_value: "4.0-12.0",
          sort_order: 1,
        },
      ],
    });
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("新規保存時に items が空でも空配列を POST に含める", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useGetExamTypeFields } =
      await import("../api/get-exam-type-fields");

    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useGetExamTypeFields).mockReturnValue({
      data: undefined,
    } as ReturnType<typeof useGetExamTypeFields>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [selectedPet("生存")],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(createMutate).toHaveBeenCalledOnce());
    expect(createMutate).toHaveBeenCalledWith(
      expect.objectContaining({ items: [] }),
    );
  });
});

describe("useExaminationForm — mutation permission boundary (FE12-02 U8)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存",
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("作成権限なしでは parent create と items replacement を発行しない", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    const { result } = renderHook(() =>
      useExaminationForm(undefined, undefined, {
        canCreate: false,
        canEdit: true,
        canDelete: true,
        canUnconfirm: false,
      }),
    );
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(createMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("作成権限があっても編集権限なしでは items を含む create を発行しない", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    const { result } = renderHook(() =>
      useExaminationForm(undefined, undefined, {
        canCreate: true,
        canEdit: false,
        canDelete: true,
        canUnconfirm: false,
      }),
    );
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(createMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("編集権限なしでは parent update と items replacement を発行しない", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    const { result } = renderHook(() =>
      useExaminationForm("exam-001", undefined, {
        canCreate: true,
        canEdit: false,
        canDelete: true,
        canUnconfirm: false,
      }),
    );
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("削除権限なしでは編集IDがあっても delete mutation を発行しない", async () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderHook(() =>
      useExaminationForm("exam-001", undefined, {
        canCreate: true,
        canEdit: true,
        canDelete: false,
        canUnconfirm: false,
      }),
    );

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    expect(mockMutate).not.toHaveBeenCalled();
  });

  it("編集権限が剥奪された後は取得済み formAction でも parent/items mutation を発行しない", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) =>
        useExaminationForm("exam-001", undefined, {
          canCreate: true,
          canEdit,
          canDelete: true,
          canUnconfirm: false,
        }),
      { initialProps: { canEdit: true } },
    );
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });
    const capturedAction = result.current.formAction;

    rerender({ canEdit: false });
    await act(async () => {
      await capturedAction(new FormData());
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("編集権限剥奪をcommitした直後のlayout phaseで取得済みformActionが発火してもmutationを発行しない", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const form = useExaminationForm("exam-001", undefined, {
          canCreate: true,
          canEdit,
          canDelete: true,
          canUnconfirm: false,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canEdit) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canEdit]);
        return form;
      },
      { initialProps: { canEdit: true } },
    );
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      rerender({ canEdit: false });
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("direct petIdのペットが死亡なら作成権限があってもcreate mutationを発行しない", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("petId=42"),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: {
        id: "42",
        name: "ポチ",
        ownerName: "田中",
        ownerId: "5",
        species: "犬",
        breed: "",
        gender: "雄",
        status: "死亡",
      },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(createMutate).not.toHaveBeenCalled();
  });

  it("direct petIdのペットが死亡なら編集権限があってもparent/items mutationを発行しない", async () => {
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("petId=42"),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: {
        id: "42",
        name: "ポチ",
        ownerName: "田中",
        ownerId: "5",
        species: "犬",
        breed: "",
        gender: "雄",
        status: "死亡",
      },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    const { result } = renderExaminationForm("exam-001");
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("direct petIdのペットが死亡なら削除権限があってもdelete mutationを発行しない", async () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams("petId=42"),
      vi.fn(),
    ]);
    vi.mocked(useGetPet).mockReturnValue({
      data: {
        id: "42",
        name: "ポチ",
        ownerName: "田中",
        ownerId: "5",
        species: "犬",
        breed: "",
        gender: "雄",
        status: "死亡",
      },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    expect(mockMutate).not.toHaveBeenCalled();
  });

  it("petIdなしの編集URLでもexistingExam.petIdの死亡ペットならupdate/delete/items mutationを発行しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } =
      await import("../api/update-examination-items");
    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    const deleteMutate = vi.fn();
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "42",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "ポチ",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "42" ? selectedPet("死亡") : null,
          isLoading: false,
          isError: false,
        }) as ReturnType<typeof useGetPet>,
    );
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: deleteMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      await result.current.formAction(new FormData());
      result.current.handleDelete();
      await Promise.resolve();
    });

    expect(useGetPet).toHaveBeenCalledWith("42");
    expect(updateMutate).not.toHaveBeenCalled();
    expect(updateItemsMutate).not.toHaveBeenCalled();
    expect(deleteMutate).not.toHaveBeenCalled();
  });
});


// ──────────────────────────────────────────────────────────
// BUG-016: 不存在 ID / 別 clinic / network error を空 edit に潰さない
// ──────────────────────────────────────────────────────────

describe("useExaminationForm BUG-016 entity read", () => {
  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError(
      "request failed",
      AxiosError.ERR_BAD_RESPONSE,
      config,
      undefined,
      {
        config,
        data: { error: "not found" },
        headers: new AxiosHeaders(),
        status,
        statusText: "Error",
      },
    );
  }

  it("404 → isReadNotFound、formAction で mutation 0 回", async () => {
    const updateMutate = vi.fn().mockResolvedValue({});
    const createMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useGetExamination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(404),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetExamination>);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);

    const { result } = renderHook(() =>
      useExaminationForm("999999999", undefined, {
        canCreate: true,
        canEdit: true,
        canDelete: true,
        canUnconfirm: true,
      }),
    );
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("notFound");

    await act(async () => {
      startTransition(() => {
        result.current.formAction(new FormData());
      });
    });
    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(updateMutate).not.toHaveBeenCalled();
    expect(createMutate).not.toHaveBeenCalled();
  });

  it("403 → forbiddenOrHidden を isReadNotFound として非開示", async () => {
    vi.mocked(useGetExamination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(403),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetExamination>);

    const { result } = renderHook(() =>
      useExaminationForm("42", undefined, {
        canCreate: true,
        canEdit: true,
        canDelete: true,
        canUnconfirm: true,
      }),
    );
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("forbiddenOrHidden");
  });

  it("network error → isReadError と retry、mutation 0 回", async () => {
    const refetch = vi.fn();
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useGetExamination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(undefined),
      refetch,
    } as unknown as ReturnType<typeof useGetExamination>);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderHook(() =>
      useExaminationForm("999999999", undefined, {
        canCreate: true,
        canEdit: true,
        canDelete: true,
        canUnconfirm: true,
      }),
    );
    expect(result.current.isReadError).toBe(true);
    expect(result.current.isReadNotFound).toBe(false);
    result.current.retryRead?.();
    expect(refetch).toHaveBeenCalledTimes(1);

    await act(async () => {
      startTransition(() => {
        result.current.formAction(new FormData());
      });
    });
    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("create route: idle かつ isEdit false", () => {
    vi.mocked(useGetExamination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetExamination>);

    const { result } = renderHook(() =>
      useExaminationForm(undefined, undefined, {
        canCreate: true,
        canEdit: true,
        canDelete: true,
        canUnconfirm: true,
      }),
    );
    expect(result.current.entityRead.status).toBe("idle");
    expect(result.current.isEdit).toBe(false);
  });
});
