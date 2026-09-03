import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { useVaccinationForm } from "./use-vaccination-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { toast } from "sonner";
import { jstDateStartISOString } from "@/lib/jst-date";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false })),
}));
// BUG-401: 実マスタ参照化に伴い useGetAllVaccinesMaster をモックする。id="1"/"2" は既存の BUG-026
// 回帰テスト期待値（両方とも 1year）を保つ interval="1年" の固定値。BUG-401 固有の interval マッピング
// テストは個別に mockReturnValueOnce で上書きする。
vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: vi.fn(() => ({
    data: [
      { id: "1", name: "混合ワクチン", isActive: true, interval: "1年" },
      { id: "2", name: "狂犬病ワクチン", isActive: true, interval: "1年" },
    ],
  })),
}));
vi.mock("../api/get-vaccination", () => ({
  useGetVaccination: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
}));
vi.mock("../api/create-vaccination", () => ({
  useCreateVaccination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
    isPending: false,
  })),
}));
vi.mock("../api/update-vaccination", () => ({
  useUpdateVaccination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
    isPending: false,
  })),
}));
vi.mock("../api/delete-vaccination", () => ({
  useDeleteVaccination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

function runFormAction(action: (payload: FormData) => void) {
  act(() => {
    startTransition(() => {
      action(new FormData());
    });
  });
}

const ALLOWED_MUTATION_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
} as const;

const DECEASED_PET = {
  id: "5",
  ownerId: "1",
  name: "ポチ",
  status: "死亡",
} as NonNullable<ReturnType<typeof useGetPet>["data"]>;

const LIVING_PET = {
  id: "5",
  ownerId: "1",
  name: "ポチ",
  status: "生存",
} as NonNullable<ReturnType<typeof useGetPet>["data"]>;

function renderVaccinationForm(id?: string) {
  return renderHook(() => useVaccinationForm(id, ALLOWED_MUTATION_PERMISSIONS));
}
// FE-RC-045: use-vaccination-form.test.ts (1317行) をトピック別に分割した1ファイル。
// 元は describe("useVaccinationForm") 配下のネスト describe だったが、親の beforeEach/afterEach を
// 各ファイルの独立 describe に複製して分割している（振る舞いは維持）。

describe("useVaccinationForm — mutation permission boundary (FE12-02 U8)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    // Date のみ偽装する。setTimeout/setInterval を実タイマーに残さないと
    // waitFor の内部ポーリングが進まず全 async テストが 5000ms でタイムアウトする。
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-07-10T01:00:00.000Z")); // JST 2026-07-10 10:00
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false } as ReturnType<
      typeof useGetPet
    >);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);
    // BUG-401: renderHook 内の act() で複数回再レンダーが起きるため、mockReturnValueOnce だと
    // 2 回目以降の呼び出しでモック実装がデフォルトへ巻き戻ってしまう（1 回限りの upvalue が枯渇する）。
    // beforeEach でテストごとに明示的にデフォルトへ戻し、個別テストは mockReturnValue（永続）で上書きする。
    vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
      data: [
        { id: "1", name: "混合ワクチン", isActive: true, interval: "1年" },
        { id: "2", name: "狂犬病ワクチン", isActive: true, interval: "1年" },
      ],
    } as ReturnType<typeof useGetAllVaccinesMaster>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("作成権限なしでは有効な新規入力でも create mutation を発行しない", async () => {
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);

    const { result } = renderHook(() =>
      useVaccinationForm(undefined, {
        canCreate: false,
        canEdit: true,
        canDelete: true,
      }),
    );
    act(() => {
      result.current.petSelection.setSelectedPets([
        { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
          typeof result.current.petSelection.setSelectedPets
        >[0][number],
      ]);
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("編集権限なしでは有効な編集入力でも update mutation を発行しない", async () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);

    const { result } = renderHook(() =>
      useVaccinationForm("10", {
        canCreate: true,
        canEdit: false,
        canDelete: true,
      }),
    );

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("削除権限なしでは編集IDがあっても delete mutation を発行しない", () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteVaccination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteVaccination>);

    const { result } = renderHook(() =>
      useVaccinationForm("10", {
        canCreate: true,
        canEdit: true,
        canDelete: false,
      }),
    );

    act(() => {
      result.current.handleDelete();
    });

    expect(mockMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("作成権限が剥奪された後は取得済み formAction でも create mutation を発行しない", async () => {
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) =>
        useVaccinationForm(undefined, {
          canCreate,
          canEdit: true,
          canDelete: true,
        }),
      { initialProps: { canCreate: true } },
    );
    act(() => {
      result.current.petSelection.setSelectedPets([
        { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
          typeof result.current.petSelection.setSelectedPets
        >[0][number],
      ]);
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });
    const capturedAction = result.current.formAction;

    rerender({ canCreate: false });
    runFormAction(capturedAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("編集で petFromEdit 未着なら delete mutation を発行しない", () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    vi.mocked(useGetPet).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    const mockMutate = vi.fn();
    vi.mocked(useDeleteVaccination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteVaccination>);

    const { result } = renderVaccinationForm("10");
    act(() => {
      result.current.handleDelete();
    });

    expect(mockMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("ペット情報の読み込みが完了してから削除してください");
  });

  it("取得済み formAction は後続 commit の formDataRef を送る", async () => {
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    const { result } = renderVaccinationForm();
    act(() => {
      result.current.petSelection.setSelectedPets([LIVING_PET]);
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });
    const capturedAction = result.current.formAction;

    act(() => {
      result.current.form.setDate("2026-07-05");
    });
    runFormAction(capturedAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(true);
    });
    expect(mockMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        date: jstDateStartISOString("2026-07-05"),
      }),
    );
  });

  it("編集で petFromEdit 未着なら update mutation を発行しない", async () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    vi.mocked(useGetPet).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);

    const { result } = renderVaccinationForm("10");
    const initialTimestamp = result.current.formState.timestamp;
    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("ペット情報の読み込みが完了してから保存してください");
  });

  it("作成権限剥奪をcommitした直後のlayout phaseで取得済みformActionが発火してもmutationを発行しない", async () => {
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) => {
        const form = useVaccinationForm(undefined, {
          canCreate,
          canEdit: true,
          canDelete: true,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canCreate) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canCreate]);
        return form;
      },
      { initialProps: { canCreate: true } },
    );
    act(() => {
      result.current.petSelection.setSelectedPets([
        { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
          typeof result.current.petSelection.setSelectedPets
        >[0][number],
      ]);
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canCreate: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });
});

describe("useVaccinationForm — deceased pet mutation boundary (FE12-02 C6a)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    // Date のみ偽装する。setTimeout/setInterval を実タイマーに残さないと
    // waitFor の内部ポーリングが進まず全 async テストが 5000ms でタイムアウトする。
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-07-10T01:00:00.000Z")); // JST 2026-07-10 10:00
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false } as ReturnType<
      typeof useGetPet
    >);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);
    // BUG-401: renderHook 内の act() で複数回再レンダーが起きるため、mockReturnValueOnce だと
    // 2 回目以降の呼び出しでモック実装がデフォルトへ巻き戻ってしまう（1 回限りの upvalue が枯渇する）。
    // beforeEach でテストごとに明示的にデフォルトへ戻し、個別テストは mockReturnValue（永続）で上書きする。
    vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
      data: [
        { id: "1", name: "混合ワクチン", isActive: true, interval: "1年" },
        { id: "2", name: "狂犬病ワクチン", isActive: true, interval: "1年" },
      ],
    } as ReturnType<typeof useGetAllVaccinesMaster>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("direct petIdのpetが死亡へ変わったcommit直後のlayout phaseでも取得済みformActionはcreate mutationを発行しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    const livingPet = { ...DECEASED_PET, status: "生存" as const };
    const petSnapshot = { current: livingPet };
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "5" ? petSnapshot.current : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    const { result, rerender } = renderHook(
      ({ status }: { status: "生存" | "死亡" }) => {
        petSnapshot.current = status === "死亡" ? DECEASED_PET : livingPet;
        const form = useVaccinationForm(undefined, ALLOWED_MUTATION_PERMISSIONS);
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (status === "死亡") {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [status]);
        return form;
      },
      { initialProps: { status: "生存" as const } },
    );
    act(() => {
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ status: "死亡" });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("死亡したペットの予防接種記録は保存できません");
  });

  it("編集petが死亡へ変わったcommit直後のlayout phaseでも取得済みformActionはupdate mutationを発行しない", async () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    const livingPet = { ...DECEASED_PET, status: "生存" as const };
    const petSnapshot = { current: livingPet };
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "5" ? petSnapshot.current : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    const { result, rerender } = renderHook(
      ({ status }: { status: "生存" | "死亡" }) => {
        petSnapshot.current = status === "死亡" ? DECEASED_PET : livingPet;
        const form = useVaccinationForm("10", ALLOWED_MUTATION_PERMISSIONS);
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (status === "死亡") {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [status]);
        return form;
      },
      { initialProps: { status: "生存" as const } },
    );
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ status: "死亡" });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("死亡したペットの予防接種記録は保存できません");
  });

  it("編集対象が明示的な死亡ペットならdelete mutationを発行しない", () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "5" ? DECEASED_PET : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    const mockMutate = vi.fn();
    vi.mocked(useDeleteVaccination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteVaccination>);

    const { result } = renderVaccinationForm("10");

    act(() => {
      result.current.handleDelete();
    });

    expect(mockMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("死亡したペットの予防接種記録は削除できません");
  });

  it("direct petIdから明示的な死亡ペットをhydrateしてもcreate mutationを発行しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "5" ? DECEASED_PET : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);

    const { result } = renderVaccinationForm();
    await waitFor(() => {
      expect(result.current.petSelection.selectedPets[0]?.status).toBe("死亡");
    });
    act(() => {
      result.current.form.setVaccineId("1");
      result.current.form.setDate("2026-07-01");
    });
    const initialTimestamp = result.current.formState.timestamp;

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(result.current.formState.success).toBe(false);
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("死亡したペットの予防接種記録は保存できません");
  });

  it("編集対象から明示的な死亡ペットをhydrateしてもupdate mutationを発行しない", async () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-07-01",
        nextDate: "",
        nextScheduleType: "1year",
      },
    } as ReturnType<typeof useGetVaccination>);
    vi.mocked(useGetPet).mockImplementation(
      (requestedPetId) =>
        ({
          data: requestedPetId === "5" ? DECEASED_PET : undefined,
          isLoading: false,
        }) as ReturnType<typeof useGetPet>,
    );
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);

    const { result } = renderVaccinationForm("10");
    await waitFor(() => {
      expect(result.current.petSelection.selectedPets[0]?.status).toBe("死亡");
    });
    const initialTimestamp = result.current.formState.timestamp;

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(result.current.formState.success).toBe(false);
    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("死亡したペットの予防接種記録は保存できません");
  });
});

describe("useVaccinationForm BUG-016 entity read", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-07-10T01:00:00.000Z"));
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false } as ReturnType<
      typeof useGetPet
    >);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
      config,
      data: { error: "not found" },
      headers: new AxiosHeaders(),
      status,
      statusText: "Error",
    });
  }

  it("404 → isReadNotFound、formAction で update/create が 0 回", async () => {
    const updateMutate = vi.fn().mockResolvedValue({});
    const createMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: createMutate,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(404),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("999999999");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.isReadError).toBe(false);
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

  it("403（別 clinic 相当）→ isReadNotFound と同一非開示、mutation 0 回", async () => {
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(403),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("42");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("forbiddenOrHidden");

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

  it("network error → isReadError（notFound と区別）かつ retry あり、mutation 0 回", async () => {
    const refetch = vi.fn();
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(undefined),
      refetch,
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("999999999");
    expect(result.current.isReadError).toBe(true);
    expect(result.current.isReadNotFound).toBe(false);
    expect(result.current.retryRead).toBeTypeOf("function");
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

  it("正常 edit: found レコードを form に反映する", () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-01-15T00:00:00+09:00",
        nextScheduleType: "1year",
        nextDate: "2027-01-15T00:00:00+09:00",
        doctor: "Dr.A",
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("10");
    expect(result.current.entityRead.status).toBe("found");
    expect(result.current.isEdit).toBe(true);
    expect(result.current.form.vaccineId).toBe("1");
    expect(result.current.form.date).toBe("2026-01-15");
  });

  it("create route (id なし): idle かつ default form", () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    const { result } = renderVaccinationForm();
    expect(result.current.entityRead.status).toBe("idle");
    expect(result.current.isEdit).toBe(false);
    expect(result.current.form.vaccineId).toBe("");
  });
});
