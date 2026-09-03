import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useExaminationForm } from "./use-examination-form";
import { useSearchParams } from "react-router";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";

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

// FE-RC-045: use-examination-form.test.ts (2662行) をトピック別に分割した1ファイル。
// 各ファイルは vi.mock 済みモジュールを共有できないため、ヘッダ（import/mock/helper）を複製している。

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

  it("mutateAsync が失敗した場合、caller は再通知せず失敗を返す（FE-RC-005）", async () => {
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

    // FE-RC-005: 通知は api/create-examination onError → handleApiError。caller は二重 toast しない。
    expect(toast.error).not.toHaveBeenCalled();
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
