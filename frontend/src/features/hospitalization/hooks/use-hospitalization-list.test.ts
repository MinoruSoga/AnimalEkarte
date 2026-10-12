import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import type { Hospitalization, MasterItem } from "@/types";

import { updateHospitalization } from "../api/update-hospitalization";
import { swapHospitalizationCages } from "../api/swap-hospitalization-cages";
import { useHospitalizationList } from "./use-hospitalization-list";

const mocks = vi.hoisted(() => ({
  cages: [] as MasterItem[],
}));

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: () => ({ data: mocks.cages }),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("../api/update-hospitalization", () => ({
  updateHospitalization: vi.fn().mockResolvedValue({}),
}));

vi.mock("../api/swap-hospitalization-cages", () => ({
  swapHospitalizationCages: vi.fn().mockResolvedValue([]),
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function makeHospitalization(overrides: Partial<Hospitalization> = {}): Hospitalization {
  return {
    id: "hospitalization-1",
    hospitalizationNo: "H-001",
    ownerName: "飼主",
    ownerId: undefined,
    petName: "ポチ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-26",
    endDate: "",
    status: "入院中",
    cageId: "cage-1",
    petId: undefined,
    doctorId: undefined,
    doctorName: undefined,
    petIsDeceased: false,
    petBirthDate: undefined,
    memo: undefined,
    ownerRequest: undefined,
    staffNotes: undefined,
    ...overrides,
  };
}

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(
      QueryClientProvider,
      { client: queryClient },
      createElement(MemoryRouter, null, children),
    );
  };
}

describe("useHospitalizationList — cage move mutation boundary", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.cages = [];
  });

  it("callback取得後に編集権限が失効した場合は移動mutationを実行しない", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [makeHospitalization()],
      total: 1,
      page: 1,
      limit: 20,
    });
    const { result, rerender } = renderHook(({ canEdit }) => useHospitalizationList(canEdit), {
      initialProps: { canEdit: true },
      wrapper: createWrapper(queryClient),
    });
    const capturedMovePet = result.current.movePet;

    rerender({ canEdit: false });
    await act(async () => {
      await capturedMovePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).not.toHaveBeenCalled();
  });

  it("source petが死亡している場合は編集権限があっても移動mutationを実行しない", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [makeHospitalization({ petIsDeceased: true })],
      total: 1,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).not.toHaveBeenCalled();
  });

  it("swap先のpetが死亡している場合は両方の移動mutationを実行しない", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization(),
        makeHospitalization({
          id: "hospitalization-2",
          cageId: "cage-2",
          petIsDeceased: true,
        }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).not.toHaveBeenCalled();
    expect(swapHospitalizationCages).not.toHaveBeenCalled();
  });

  it("生存petかつ編集権限ありなら既存の移動payloadを維持する", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [makeHospitalization()],
      total: 1,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).toHaveBeenCalledWith("hospitalization-1", {
      cage_id: "cage-2",
      expected_cage_id: "cage-1",
    });
    expect(swapHospitalizationCages).not.toHaveBeenCalled();
  });

  it("交換元の表示時ケージが未割当なら expected_cage_is_null: true を送る", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [makeHospitalization({ cageId: undefined })],
      total: 1,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).toHaveBeenCalledWith("hospitalization-1", {
      cage_id: "cage-2",
      expected_cage_is_null: true,
    });
  });

  it("交換元の表示時ケージが未割当ならスワップを送信しない", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization({ cageId: undefined }),
        makeHospitalization({ id: "hospitalization-2", cageId: "cage-2" }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(swapHospitalizationCages).not.toHaveBeenCalled();
    expect(updateHospitalization).not.toHaveBeenCalled();
  });

  it("スワップの409を通知し個別PATCHへフォールバックしない", async () => {
    const error = { response: { status: 409 } };
    vi.mocked(swapHospitalizationCages).mockRejectedValueOnce(error);
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization(),
        makeHospitalization({ id: "hospitalization-2", cageId: "cage-2" }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(handleApiError).toHaveBeenCalledWith(error, "移動");
    expect(swapHospitalizationCages).toHaveBeenCalledTimes(1);
    expect(updateHospitalization).not.toHaveBeenCalled();
  });

  // EMR-306 (Grok review HIGH-1): 連結グループが絡むスワップは 2 本の独立 PATCH が
  // 片方だけ成功し得るため、送信前に丸ごと拒否する回帰テスト。
  it("連結代表ケージが絡むスワップは PATCH を1本も送らず toast で拒否する", async () => {
    // cage-2 が代表（cage-3 が mergedIntoCageId で指す）
    mocks.cages = [
      { id: "cage-1", name: "A", price: 0, status: "active" },
      { id: "cage-2", name: "B", price: 0, status: "active" },
      { id: "cage-3", name: "C", price: 0, status: "active", mergedIntoCageId: "cage-2" },
    ];
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization(),
        makeHospitalization({ id: "hospitalization-2", cageId: "cage-2" }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(updateHospitalization).not.toHaveBeenCalled();
    expect(swapHospitalizationCages).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("連結中のケージを含む入れ替えはできません");
  });

  it("構成メンバーケージが絡むスワップも PATCH を1本も送らない", async () => {
    // source の cage-1 が構成メンバー（代表は cage-2）
    mocks.cages = [
      { id: "cage-1", name: "A", price: 0, status: "active", mergedIntoCageId: "cage-2" },
      { id: "cage-2", name: "B", price: 0, status: "active" },
      { id: "cage-4", name: "D", price: 0, status: "active" },
    ];
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization(),
        makeHospitalization({ id: "hospitalization-4", cageId: "cage-4" }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-4");
    });

    expect(updateHospitalization).not.toHaveBeenCalled();
    expect(swapHospitalizationCages).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("連結中のケージを含む入れ替えはできません");
  });

  // EMR-306 (Codex review P1): 通常スワップは 2 本の独立 PATCH ではなく
  // POST /v1/hospitalizations/cage-swap の単一トランザクション API を呼ぶ。
  it("連結に関わらない通常スワップは cage-swap API を1回だけ呼ぶ", async () => {
    mocks.cages = [
      { id: "cage-1", name: "A", price: 0, status: "active" },
      { id: "cage-2", name: "B", price: 0, status: "active" },
    ];
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.list({ statusFilter: "active" }), {
      data: [
        makeHospitalization(),
        makeHospitalization({ id: "hospitalization-2", cageId: "cage-2" }),
      ],
      total: 2,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => useHospitalizationList(true), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.movePet("hospitalization-1", "cage-2");
    });

    expect(swapHospitalizationCages).toHaveBeenCalledTimes(1);
    expect(swapHospitalizationCages).toHaveBeenCalledWith(
      "hospitalization-1",
      "hospitalization-2",
      ["cage-1", "cage-2"],
    );
    expect(updateHospitalization).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });
});
