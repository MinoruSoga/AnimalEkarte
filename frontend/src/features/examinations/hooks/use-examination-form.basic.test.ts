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
