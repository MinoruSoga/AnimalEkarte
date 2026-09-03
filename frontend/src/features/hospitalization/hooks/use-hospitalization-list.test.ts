import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import type { Hospitalization } from "@/types";

import { updateHospitalization } from "../api/update-hospitalization";
import { useHospitalizationList } from "./use-hospitalization-list";

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: () => ({ data: [] }),
}));

vi.mock("../api/update-hospitalization", () => ({
  updateHospitalization: vi.fn().mockResolvedValue({}),
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function makeHospitalization(overrides: Partial<Hospitalization> = {}): Hospitalization {
  return {
    id: "hospitalization-1",
    hospitalizationNo: "H-001",
    ownerName: "飼主",
    petName: "ポチ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-26",
    endDate: "",
    status: "入院中",
    cageId: "cage-1",
    petIsDeceased: false,
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

    expect(updateHospitalization).toHaveBeenCalledWith("hospitalization-1", { cage_id: "cage-2" });
  });
});
