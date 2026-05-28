import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition } from "react";
import { useTrimmingForm } from "./use-trimming-form";

const {
  mockNavigate,
  mockToast,
  mockSelectedPets,
  mockSetSelectedPets,
  mockCreateMutateAsync,
  mockUpdateMutateAsync,
  mockDeleteMutate,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn() },
  mockSelectedPets: [] as Array<{ id: string; ownerId: string; name: string }>,
  mockSetSelectedPets: vi.fn(),
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  mockDeleteMutate: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

vi.mock("@/hooks/use-pet-selection", () => ({
  usePetSelection: vi.fn(() => ({
    selectedPets: mockSelectedPets,
    setSelectedPets: mockSetSelectedPets,
    togglePetSelection: vi.fn(),
    isPetSelected: vi.fn(() => false),
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false })),
}));

vi.mock("../api/get-trimming", () => ({
  useGetTrimming: vi.fn(() => ({ data: undefined, isLoading: false })),
}));

vi.mock("../api/create-trimming", () => ({
  useCreateTrimming: vi.fn(() => ({ mutateAsync: mockCreateMutateAsync })),
}));

vi.mock("../api/update-trimming", () => ({
  useUpdateTrimming: vi.fn(() => ({ mutateAsync: mockUpdateMutateAsync })),
}));

vi.mock("../api/delete-trimming", () => ({
  useDeleteTrimming: vi.fn(() => ({
    mutate: mockDeleteMutate,
    isPending: false,
  })),
}));

vi.mock("@/config/paths", () => ({
  paths: {
    trimming: {
      selectPet: { getHref: () => "/trimming/select-pet" },
    },
  },
}));

async function submitFormAction(action: (payload: FormData) => void | Promise<unknown>) {
  await act(async () => {
    await new Promise<void>((resolve, reject) => {
      startTransition(() => {
        Promise.resolve(action(new FormData())).then(resolve).catch(reject);
      });
    });
  });
}

describe("useTrimmingForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSelectedPets.length = 0;
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("新規保存時に画面入力項目を create payload に含める", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderHook(() => useTrimmingForm());

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
        styleRequest: "短め",
        bw: "5.2",
        bwUnit: "Kg",
        bt: "38.5",
        usedShampoo: "薬用A",
        usedRibbon: "赤",
        remarks: "皮膚注意",
        optionIds: ["7", "8"],
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        pet_id: 10,
        staff_id: 3,
        course_id: 4,
        style_request: "短め",
        bw: 5.2,
        bw_unit: "Kg",
        bt: 38.5,
        used_shampoo: "薬用A",
        used_ribbon: "赤",
        remarks: "皮膚注意",
        option_ids: [7, 8],
      }),
    );
  });

  it("編集保存時にコース・担当者・オプションを update payload に含める", async () => {
    const { result } = renderHook(() => useTrimmingForm("15"));

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
        styleRequest: "そろえる",
        usedShampoo: "",
        usedRibbon: "",
        remarks: "",
        optionIds: [],
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: "15",
      req: expect.objectContaining({
        staff_id: 3,
        course_id: 4,
        style_request: "そろえる",
        used_shampoo: "",
        used_ribbon: "",
        remarks: "",
        option_ids: [],
      }),
    });
  });

  it("担当者とコースが未選択の新規保存では API を呼ばない", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderHook(() => useTrimmingForm());

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.fieldErrors.staffId).toBe("担当者を選択してください");
    expect(result.current.fieldErrors.courseId).toBe("コースを選択してください");
  });
});
