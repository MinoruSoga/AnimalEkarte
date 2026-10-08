import type { ReactNode } from "react";
import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { useCageOccupancy } from "./use-cage-occupancy";

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

const occupancyState = (queryClient: QueryClient, status: "admitted" | "reserved") =>
  queryClient.getQueryState(queryKeys.hospitalizations.cageOccupancy(status));

describe("useCageOccupancy (EMR-306)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("admitted/reserved 両ステータスの cage_id を和集合で返す", async () => {
    server.use(
      http.get("*/v1/hospitalizations", ({ request }) => {
        const status = new URL(request.url).searchParams.get("status");
        const rows =
          status === "admitted"
            ? [{ cage_id: 1 }, { cage_id: 2 }]
            : [{ cage_id: 2 }, { cage_id: 7 }];
        return HttpResponse.json({ data: rows, total: rows.length });
      }),
    );

    const { result } = renderOccupancy();

    await waitFor(() => expect(result.current.occupiedCageIds).toBeDefined());
    expect([...(result.current.occupiedCageIds ?? [])].sort()).toEqual(["1", "2", "7"]);
  });

  it("クエリがエラーのときは undefined を返す（fail-closed）", async () => {
    server.use(http.get("*/v1/hospitalizations", () => new HttpResponse(null, { status: 500 })));

    const { result, queryClient } = renderOccupancy();

    await waitFor(() => expect(occupancyState(queryClient, "admitted")?.status).toBe("error"));
    await waitFor(() => expect(occupancyState(queryClient, "reserved")?.status).toBe("error"));
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("total が先頭ページ取得数を超える応答は打ち切りと見なし undefined を返す", async () => {
    server.use(
      http.get("*/v1/hospitalizations", ({ request }) => {
        const status = new URL(request.url).searchParams.get("status");
        if (status === "admitted") {
          return HttpResponse.json({ data: [{ cage_id: 1 }], total: 101 });
        }
        return HttpResponse.json({ data: [], total: 0 });
      }),
    );

    const { result, queryClient } = renderOccupancy();

    await waitFor(() => expect(occupancyState(queryClient, "admitted")?.status).toBe("success"));
    await waitFor(() => expect(occupancyState(queryClient, "reserved")?.status).toBe("success"));
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("total を欠く応答は信頼できないため undefined を返す", async () => {
    server.use(
      http.get("*/v1/hospitalizations", () => HttpResponse.json({ data: [{ cage_id: 1 }] })),
    );

    const { result, queryClient } = renderOccupancy();

    await waitFor(() => expect(occupancyState(queryClient, "admitted")?.status).toBe("success"));
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("enabled: false のときはクエリを発行せず undefined を返す", async () => {
    let calls = 0;
    server.use(
      http.get("*/v1/hospitalizations", () => {
        calls += 1;
        return HttpResponse.json({ data: [], total: 0 });
      }),
    );

    const { result } = renderOccupancy({ enabled: false });

    expect(result.current.occupiedCageIds).toBeUndefined();
    // 遅延発行もないことを確認（非同期でクエリが走るなら呼ばれる）
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(calls).toBe(0);
  });
});
