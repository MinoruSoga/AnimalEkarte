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

// FE-RC-045: use-examination-form.test.ts (2662行) をトピック別に分割した1ファイル。
// 各ファイルは vi.mock 済みモジュールを共有できないため、ヘッダ（import/mock/helper）を複製している。

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
