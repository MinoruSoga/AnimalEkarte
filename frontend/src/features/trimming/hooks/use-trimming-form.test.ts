import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { useTrimmingForm } from "./use-trimming-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useGetTrimmings } from "../api/get-trimmings";

const {
  mockNavigate,
  mockToast,
  mockSelectedPets,
  mockSetSelectedPets,
  mockCreateMutateAsync,
  mockUpdateMutateAsync,
  mockDeleteMutate,
  mockLocationStateHolder,
  mockSearchParamsHolder,
  mockExistingTrimmingHolder,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
  mockSelectedPets: [] as Array<{ id: string; ownerId: string; name: string; status?: string }>,
  mockSetSelectedPets: vi.fn(),
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  mockDeleteMutate: vi.fn(),
  mockLocationStateHolder: { value: null as Record<string, unknown> | null },
  mockSearchParamsHolder: { value: new URLSearchParams() },
  mockExistingTrimmingHolder: { value: undefined as Record<string, unknown> | undefined },
}));

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParamsHolder.value, vi.fn()],
  useLocation: () => ({ state: mockLocationStateHolder.value }),
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

vi.mock("@/hooks/use-reservation-types", () => ({
  useGetReservationTypesGrouped: vi.fn(() => ({
    data: [
      {
        label: "トリミング",
        types: [
          {
            id: 9,
            name: "シャンプーコース",
            category: "trimming",
            is_internal: false,
            sort_order: 1,
          },
        ],
      },
    ],
  })),
}));

vi.mock("../api/get-trimming", () => ({
  useGetTrimming: vi.fn((id: string) => ({
    data: id ? mockExistingTrimmingHolder.value : undefined,
    isLoading: false,
  })),
}));

vi.mock("../api/get-trimmings", () => ({
  useGetTrimmings: vi.fn(() => ({ data: [], isLoading: false })),
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

const ALLOW_ALL_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
} as const;

const DECEASED_SAVE_MESSAGE = "死亡したペットのトリミング記録は保存できません";
const DECEASED_DELETE_MESSAGE = "死亡したペットのトリミング記録は削除できません";
const PET_LOADING_MESSAGE = "ペット情報の読み込みが完了してから保存してください";
const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const LIVING_EDIT_PET = {
  id: "10",
  ownerId: "20",
  name: "ポチ",
  status: "生存",
};

const DECEASED_EDIT_PET = {
  id: "10",
  ownerId: "20",
  name: "ポチ",
  status: "死亡",
};

function mockLivingEditPet(petId = "10") {
  mockExistingTrimmingHolder.value = {
    ...(mockExistingTrimmingHolder.value ?? {}),
    id: "15",
    petId,
  };
  vi.mocked(useGetPet).mockImplementation(
    (requestedPetId) =>
      ({
        data: requestedPetId === petId ? LIVING_EDIT_PET : undefined,
        isLoading: false,
      }) as ReturnType<typeof useGetPet>,
  );
}

function renderTrimmingForm(id?: string) {
  return renderHook(() => useTrimmingForm(id, ALLOW_ALL_PERMISSIONS));
}

describe("useTrimmingForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocationStateHolder.value = null;
    mockSearchParamsHolder.value = new URLSearchParams();
    mockExistingTrimmingHolder.value = undefined;
    mockSelectedPets.length = 0;
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
    vi.mocked(useGetTrimmings).mockReturnValue({ data: [], isLoading: false } as ReturnType<
      typeof useGetTrimmings
    >);
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "トリミング",
          types: [
            {
              id: 9,
              name: "シャンプーコース",
              category: "trimming",
              is_internal: false,
              sort_order: 1,
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
  });

  it("新規保存時に画面入力項目を create payload に含める", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

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
        reservation_type_id: 9,
        staff_id: 3,
        course_id: 4,
        status: "in_consultation",
        reservation_route: "record_shortcut",
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

  it("#233: カルテ直接新規作成で initialStatus=pending を選択した場合は status=pending を送る", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
        initialStatus: "pending",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        status: "pending",
        reservation_route: "record_shortcut",
      }),
    );
  });

  it("#233: hasExistingAppointment はカルテ直接新規作成で false を返す", () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    expect(result.current.hasExistingAppointment).toBe(false);
  });

  it("#233: 受付から遷移した新規作成では initialStatus を無視し status/reservation_route を送らない", async () => {
    mockLocationStateHolder.value = { appointmentId: "77" };
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    expect(result.current.hasExistingAppointment).toBe(true);

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
        initialStatus: "pending",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.not.objectContaining({
        status: expect.anything(),
        reservation_route: expect.anything(),
      }),
    );
  });

  it("visitDate 指定の一覧新規作成では appointment も同じ日付で作成する", async () => {
    mockSearchParamsHolder.value = new URLSearchParams({ visitDate: "2026-06-01" });
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        appointment_id: undefined,
        pet_id: 10,
        reservation_type_id: 9,
        staff_id: 3,
        course_id: 4,
        status: "in_consultation",
        reservation_route: "record_shortcut",
      }),
    );
    const payload = mockCreateMutateAsync.mock.calls[0]?.[0] as {
      start_time: string;
      end_time: string;
    };
    expect(payload.start_time).toMatch(/^2026-06-01T\d{2}:\d{2}:\d{2}\.\d{3}\+09:00$/);
    expect(payload.end_time).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}\+09:00$/);
  });

  it("一覧新規作成では同日同ペットの未完了 trimming appointment に詳細を作成する", async () => {
    mockSearchParamsHolder.value = new URLSearchParams({ petId: "10", visitDate: "2026-06-01" });
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });
    vi.mocked(useGetTrimmings).mockReturnValue({
      data: [
        {
          id: "77",
          hasDetail: false,
          status: "予約",
          reservationTypeId: "9",
        },
      ],
      isLoading: false,
    } as ReturnType<typeof useGetTrimmings>);

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        appointment_id: 77,
        start_time: undefined,
        end_time: undefined,
      }),
    );
  });

  it("一覧新規作成で同日同ペットの trimming detail がある場合は update payload を送る", async () => {
    mockSearchParamsHolder.value = new URLSearchParams({ petId: "10", visitDate: "2026-06-01" });
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });
    vi.mocked(useGetTrimmings).mockReturnValue({
      data: [
        {
          id: "77",
          hasDetail: true,
          status: "予約",
          reservationTypeId: "9",
        },
      ],
      isLoading: false,
    } as ReturnType<typeof useGetTrimmings>);

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: "77",
      req: expect.objectContaining({
        staff_id: 3,
        course_id: 4,
      }),
    });
  });

  it("編集保存時にコース・担当者・オプションを update payload に含める", async () => {
    mockLivingEditPet();
    const { result } = renderTrimmingForm("15");

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

    const { result } = renderTrimmingForm();

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.fieldErrors.staffId).toBe("担当者を選択してください");
    expect(result.current.fieldErrors.courseId).toBe("コースを選択してください");
  });

  it("active な trimming 予約区分がない新規保存では API を呼ばない", async () => {
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "一般診療",
          types: [
            {
              id: 1,
              name: "一般診察",
              category: "general",
              is_internal: false,
              sort_order: 1,
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.fieldErrors.reservationTypeId).toBe(
      "トリミング予約区分が設定されていません",
    );
  });

  it("受付から遷移した新規保存では既存 appointment_id を使い日時を上書きしない", async () => {
    mockLocationStateHolder.value = { appointmentId: "77" };
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({
        staffId: "3",
        courseId: "4",
      });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        appointment_id: 77,
        pet_id: 10,
        reservation_type_id: 9,
        staff_id: 3,
        course_id: 4,
        start_time: undefined,
        end_time: undefined,
      }),
    );
  });

  it("受付から遷移した新規フォームでは既存予約の担当者と予約種別を初期表示する", async () => {
    mockLocationStateHolder.value = { appointmentId: "77" };
    mockExistingTrimmingHolder.value = {
      hasDetail: false,
      reservationTypeId: "9",
      staffId: "33",
      staff: "さくら（デモ）",
    };

    const { result } = renderTrimmingForm();

    await waitFor(() => {
      expect(result.current.formData.staffId).toBe("33");
    });
    expect(result.current.formData.staffName).toBe("さくら（デモ）");
    expect(result.current.formData.reservationTypeId).toBe("9");
    expect(result.current.formData.optionIds).toEqual([]);
  });

  it("既存予約に詳細がない場合でも optionIds 未定義で保存処理が落ちない", async () => {
    mockLocationStateHolder.value = { appointmentId: "77" };
    mockExistingTrimmingHolder.value = {
      hasDetail: false,
      reservationTypeId: "9",
      staffId: "33",
      staff: "さくら（デモ）",
    };
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ" });

    const { result } = renderTrimmingForm();

    await waitFor(() => {
      expect(result.current.formData.staffId).toBe("33");
    });

    act(() => {
      result.current.setFormData({ courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        appointment_id: 77,
        staff_id: 33,
        course_id: 4,
        option_ids: [],
      }),
    );
  });

  it("受付から既存トリミング記録を開いた保存では update payload を送る", async () => {
    mockLocationStateHolder.value = { appointmentId: "77" };
    mockExistingTrimmingHolder.value = {
      hasDetail: true,
      reservationTypeId: "9",
      staffId: "33",
      staff: "さくら（デモ）",
      courseId: "4",
      optionIds: ["7"],
      styleRequest: "短め",
      bw: "",
      bt: "",
      usedShampoo: "薬用A",
      usedRibbon: "赤",
      remarks: "皮膚注意",
    };

    const { result } = renderTrimmingForm();

    await waitFor(() => {
      expect(result.current.formData.courseId).toBe("4");
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: "77",
      req: expect.objectContaining({
        staff_id: 33,
        course_id: 4,
        style_request: "短め",
        used_shampoo: "薬用A",
        used_ribbon: "赤",
        remarks: "皮膚注意",
        option_ids: [7],
      }),
    });
  });
});

describe("useTrimmingForm permissions (FE-RC-101 fail-closed)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocationStateHolder.value = null;
    mockSearchParamsHolder.value = new URLSearchParams();
    mockExistingTrimmingHolder.value = undefined;
    mockSelectedPets.length = 0;
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
    vi.mocked(useGetTrimmings).mockReturnValue({ data: [], isLoading: false } as ReturnType<
      typeof useGetTrimmings
    >);
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "トリミング",
          types: [
            {
              id: 9,
              name: "シャンプーコース",
              category: "trimming",
              is_internal: false,
              sort_order: 1,
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
  });

  it("canCreate=false の新規保存では create mutation を呼ばず toast.error する", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result } = renderHook(() =>
      useTrimmingForm(undefined, { canCreate: false, canEdit: true, canDelete: true }),
    );

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false の編集保存では update mutation を呼ばず toast.error する", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result } = renderHook(() =>
      useTrimmingForm("15", { canCreate: true, canEdit: false, canDelete: true }),
    );

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("permissions 未指定（既定 deny）では create/update いずれも呼ばない", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result } = renderHook(() => useTrimmingForm());

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
  });

  it("canDelete=false では delete mutation を呼ばず toast.error する", () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result } = renderHook(() =>
      useTrimmingForm("15", { canCreate: true, canEdit: true, canDelete: false }),
    );

    act(() => {
      result.current.handleDelete();
    });

    expect(mockDeleteMutate).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("作成権限が剥奪された後は取得済み formAction でも create mutation を発行しない", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) =>
        useTrimmingForm(undefined, { canCreate, canEdit: true, canDelete: true }),
      { initialProps: { canCreate: true } },
    );

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });
    const capturedAction = result.current.formAction;

    rerender({ canCreate: false });
    await submitFormAction(capturedAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
  });

  it("取得済み formAction は後続 commit の formDataRef を送る", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "生存" });

    const { result } = renderTrimmingForm();
    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });
    const capturedAction = result.current.formAction;

    act(() => {
      result.current.setFormData({ staffId: "8", courseId: "4" });
    });
    await submitFormAction(capturedAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        staff_id: 8,
        course_id: 4,
      }),
    );
  });

  it("編集で petFromEdit 未着なら update mutation を発行しない", async () => {
    const { result } = renderTrimmingForm("15");

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });
    await submitFormAction(result.current.formAction);

    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PET_LOADING_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PET_LOADING_MESSAGE);
  });

  it("編集で petFromEdit 未着なら delete mutation を発行しない", () => {
    const { result } = renderTrimmingForm("15");

    act(() => {
      result.current.handleDelete();
    });

    expect(mockDeleteMutate).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(
      "ペット情報の読み込みが完了してから削除してください",
    );
  });
});

describe("useTrimmingForm deceased pet (FE-RC-102 dual-gate)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocationStateHolder.value = null;
    mockSearchParamsHolder.value = new URLSearchParams();
    mockExistingTrimmingHolder.value = undefined;
    mockSelectedPets.length = 0;
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
    vi.mocked(useGetTrimmings).mockReturnValue({ data: [], isLoading: false } as ReturnType<
      typeof useGetTrimmings
    >);
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "トリミング",
          types: [
            {
              id: 9,
              name: "シャンプーコース",
              category: "trimming",
              is_internal: false,
              sort_order: 1,
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
  });

  it("死亡ペットの新規保存では create mutation を呼ばず toast.error する", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "死亡" });

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(DECEASED_SAVE_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(DECEASED_SAVE_MESSAGE);
  });

  it("死亡ペットの編集保存では update mutation を呼ばず toast.error する", async () => {
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "死亡" });

    const { result } = renderTrimmingForm("15");

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(DECEASED_SAVE_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(DECEASED_SAVE_MESSAGE);
  });

  it("死亡ペットの同日詳細更新でも update mutation を呼ばない", async () => {
    mockSearchParamsHolder.value = new URLSearchParams({ petId: "10", visitDate: "2026-06-01" });
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "死亡" });
    vi.mocked(useGetTrimmings).mockReturnValue({
      data: [
        {
          id: "77",
          hasDetail: true,
          status: "予約",
          reservationTypeId: "9",
        },
      ],
      isLoading: false,
    } as ReturnType<typeof useGetTrimmings>);

    const { result } = renderTrimmingForm();

    act(() => {
      result.current.setFormData({ staffId: "3", courseId: "4" });
    });

    await submitFormAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(DECEASED_SAVE_MESSAGE);
  });

  it("死亡ペットでは delete mutation を呼ばず toast.error する", () => {
    mockExistingTrimmingHolder.value = { id: "15", petId: "10" };
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "10" ? DECEASED_EDIT_PET : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    mockSelectedPets.push({ id: "10", ownerId: "20", name: "ポチ", status: "死亡" });

    const { result } = renderTrimmingForm("15");

    act(() => {
      result.current.handleDelete();
    });

    expect(mockDeleteMutate).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(DECEASED_DELETE_MESSAGE);
  });
});
