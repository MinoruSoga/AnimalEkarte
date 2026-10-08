import type { ReactNode } from "react";
import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { useCageOccupancy } from "./use-cage-occupancy";

const OCCUPANCY_URL = "*/v1/masters/cages/occupancy";

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

function renderOccupancy(options?: { enabled?: boolean }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const view = renderHook(() => useCageOccupancy(options), {
    wrapper: createWrapper(queryClient),
  });
  return { ...view, queryClient };
}

const occupancyState = (queryClient: QueryClient) =>
  queryClient.getQueryState(queryKeys.hospitalizations.cageOccupancy());

describe("useCageOccupancy (EMR-306)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("GET /v1/masters/cages/occupancy の occupied_cage_ids を Set で返す", async () => {
    server.use(http.get(OCCUPANCY_URL, () => HttpResponse.json({ occupied_cage_ids: [1, 2, 7] })));

    const { result } = renderOccupancy();

    await waitFor(() => expect(result.current.occupiedCageIds).toBeDefined());
    expect([...(result.current.occupiedCageIds ?? [])].sort()).toEqual(["1", "2", "7"]);
  });

  it("空の occupied_cage_ids は空 Set（未定義ではない）を返す", async () => {
    server.use(http.get(OCCUPANCY_URL, () => HttpResponse.json({ occupied_cage_ids: [] })));

    const { result } = renderOccupancy();

    await waitFor(() => expect(result.current.occupiedCageIds).toBeDefined());
    expect(result.current.occupiedCageIds?.size).toBe(0);
  });

  it("クエリがエラーのときは undefined を返す（fail-closed）", async () => {
    server.use(http.get(OCCUPANCY_URL, () => new HttpResponse(null, { status: 500 })));

    const { result, queryClient } = renderOccupancy();

    await waitFor(() => expect(occupancyState(queryClient)?.status).toBe("error"));
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("occupied_cage_ids を欠く応答は契約外として undefined を返す（fail-closed）", async () => {
    server.use(http.get(OCCUPANCY_URL, () => HttpResponse.json({})));

    const { result, queryClient } = renderOccupancy();

    // queryFn が throw → query は error になり occupiedCageIds は undefined
    await waitFor(() => expect(occupancyState(queryClient)?.status).toBe("error"));
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("enabled: false のときはクエリを発行せず undefined を返す", async () => {
    let calls = 0;
    server.use(
      http.get(OCCUPANCY_URL, () => {
        calls += 1;
        return HttpResponse.json({ occupied_cage_ids: [] });
      }),
    );

    const { result } = renderOccupancy({ enabled: false });

    expect(result.current.occupiedCageIds).toBeUndefined();
    // 遅延発行もないことを確認（非同期でクエリが走るなら呼ばれる）
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(calls).toBe(0);
  });
});
