import type { ReactNode } from "react";
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { CURRENT_CLINIC_STORAGE_KEY, setStoredClinicId } from "@/lib/current-clinic";
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
  queryClient.getQueryState(queryKeys.hospitalizations.cageOccupancy("1"));

describe("useCageOccupancy (EMR-306)", () => {
  beforeEach(() => {
    setStoredClinicId("1");
  });

  afterEach(() => {
    server.resetHandlers();
    localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
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

  it("query key の医院を header に固定し、通信開始前の選択変更に影響されない", async () => {
    const clinicHeaders: Array<string | null> = [];
    server.use(
      http.get(OCCUPANCY_URL, ({ request }) => {
        const clinicId = request.headers.get("X-Clinic-ID");
        clinicHeaders.push(clinicId);
        return HttpResponse.json({ occupied_cage_ids: clinicId === "1" ? [11] : [22] });
      }),
    );

    const { queryClient } = renderOccupancy();
    // Axios interceptor runs asynchronously after the hook captured clinic 1.
    setStoredClinicId("2");

    await waitFor(() => expect(occupancyState(queryClient)?.status).toBe("success"));
    expect(queryClient.getQueryData(queryKeys.hospitalizations.cageOccupancy("1"))).toEqual(
      new Set(["11"]),
    );
    expect(clinicHeaders[0]).toBe("1");
  });

  it("医院切替後は旧医院の fresh cache を表示せず、未取得の新医院を不明にする", async () => {
    let finishNewClinic = () => {};
    const responseReady = new Promise<void>((resolve) => {
      finishNewClinic = resolve;
    });
    server.use(
      http.get(OCCUPANCY_URL, async () => {
        await responseReady;
        return HttpResponse.json({ occupied_cage_ids: [22] });
      }),
    );
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(queryKeys.hospitalizations.cageOccupancy("1"), new Set(["11"]));
    const { result, rerender } = renderHook(() => useCageOccupancy(), {
      wrapper: createWrapper(queryClient),
    });
    expect(result.current.occupiedCageIds).toEqual(new Set(["11"]));

    setStoredClinicId("2");
    rerender();
    try {
      expect(result.current.occupiedCageIds).toBeUndefined();
    } finally {
      finishNewClinic();
    }
    await waitFor(() => expect(result.current.occupiedCageIds).toEqual(new Set(["22"])));
  });

  it.each([false, true])(
    "医院切替前の遅延応答（error=%s）が新医院の結果を上書きしない",
    async (oldClinicError) => {
      let finishOldClinic = () => {};
      const responseReady = new Promise<void>((resolve) => {
        finishOldClinic = resolve;
      });
      const clinicHeaders: Array<string | null> = [];
      server.use(
        http.get(OCCUPANCY_URL, async ({ request }) => {
          const clinicId = request.headers.get("X-Clinic-ID");
          clinicHeaders.push(clinicId);
          if (clinicId === "1") {
            await responseReady;
            return oldClinicError
              ? new HttpResponse(null, { status: 500 })
              : HttpResponse.json({ occupied_cage_ids: [11] });
          }
          return HttpResponse.json({ occupied_cage_ids: [22] });
        }),
      );
      const { result, queryClient, rerender } = renderOccupancy();
      await waitFor(() => expect(clinicHeaders).toEqual(["1"]));
      setStoredClinicId("2");
      rerender();
      try {
        await waitFor(() => expect(result.current.occupiedCageIds).toEqual(new Set(["22"])));
      } finally {
        finishOldClinic();
      }
      await waitFor(() =>
        expect(occupancyState(queryClient)?.status).toBe(oldClinicError ? "error" : "success"),
      );
      expect(result.current.occupiedCageIds).toEqual(new Set(["22"]));
      expect(clinicHeaders).toEqual(["1", "2"]);
      expect(queryClient.getQueryData(queryKeys.hospitalizations.cageOccupancy("2"))).toEqual(
        new Set(["22"]),
      );
    },
  );

  it("切替先の取得失敗でも旧医院の cache に戻らず不明を返す", async () => {
    server.use(http.get(OCCUPANCY_URL, () => new HttpResponse(null, { status: 500 })));
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(queryKeys.hospitalizations.cageOccupancy("1"), new Set(["11"]));
    const { result, rerender } = renderHook(() => useCageOccupancy(), {
      wrapper: createWrapper(queryClient),
    });
    setStoredClinicId("2");
    rerender();
    await waitFor(() =>
      expect(queryClient.getQueryState(queryKeys.hospitalizations.cageOccupancy("2"))?.status).toBe(
        "error",
      ),
    );
    expect(result.current.occupiedCageIds).toBeUndefined();
    expect(queryClient.getQueryData(queryKeys.hospitalizations.cageOccupancy("1"))).toEqual(
      new Set(["11"]),
    );
  });

  it("医院未選択では通信せず、占有状況を不明にする", async () => {
    localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
    let calls = 0;
    server.use(
      http.get(OCCUPANCY_URL, () => {
        calls += 1;
        return HttpResponse.json({ occupied_cage_ids: [] });
      }),
    );
    const { result } = renderOccupancy();
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(calls).toBe(0);
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("disabled は取得済み cache があっても占有状況を返さない", () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.hospitalizations.cageOccupancy("1"), new Set<string>());
    const { result } = renderHook(() => useCageOccupancy({ enabled: false }), {
      wrapper: createWrapper(queryClient),
    });
    expect(result.current.occupiedCageIds).toBeUndefined();
  });

  it("入退院 mutation の共通 prefix invalidation が医院付きキーも再取得する", async () => {
    let calls = 0;
    server.use(
      http.get(OCCUPANCY_URL, () => {
        calls += 1;
        return HttpResponse.json({ occupied_cage_ids: calls === 1 ? [11] : [12] });
      }),
    );
    const { result, queryClient } = renderOccupancy();
    await waitFor(() => expect(result.current.occupiedCageIds).toEqual(new Set(["11"])));
    await queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
    await waitFor(() => expect(result.current.occupiedCageIds).toEqual(new Set(["12"])));
    expect(calls).toBe(2);
  });
});
