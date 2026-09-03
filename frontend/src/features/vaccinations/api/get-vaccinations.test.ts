import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { useGetVaccinations } from "./get-vaccinations";

describe("useGetVaccinations (BUG-007)", () => {
  it("sends pet_id + page/limit so pet history is not page-window filtered client-side", async () => {
    let seenUrl = "";
    server.use(
      http.get("/api/v1/vaccinations", ({ request }) => {
        seenUrl = request.url;
        const url = new URL(request.url);
        expect(url.searchParams.get("pet_id")).toBe("1000002");
        expect(url.searchParams.get("page")).toBe("1");
        expect(url.searchParams.get("limit")).toBe(String(HISTORY_FETCH_LIMIT));
        return HttpResponse.json({
          data: [
            {
              id: 1091849,
              pet_id: 1000002,
              vaccine_id: 1,
              date: "2026-07-31T00:00:00+09:00",
              next_date: "2026-09-15T00:00:00+09:00",
              vaccine: { name: "バンガードL4" },
              pet: { name: "豆助", owner: { name: "伊藤 史安" } },
            },
          ],
          total: 1,
          page: 1,
          limit: HISTORY_FETCH_LIMIT,
        });
      }),
    );

    const { result } = renderHook(
      () => useGetVaccinations({ petId: "1000002" }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.petId).toBe("1000002");
    expect(result.current.data?.[0]?.date).toBe("2026-07-31");
    expect(seenUrl).toContain("pet_id=1000002");
    expect(seenUrl).not.toContain("petId=");
  });

  it("list without petId still requests HISTORY_FETCH_LIMIT (not silent default 20)", async () => {
    server.use(
      http.get("/api/v1/vaccinations", ({ request }) => {
        const url = new URL(request.url);
        expect(url.searchParams.get("limit")).toBe(String(HISTORY_FETCH_LIMIT));
        expect(url.searchParams.get("pet_id")).toBeNull();
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: HISTORY_FETCH_LIMIT });
      }),
    );

    const { result } = renderHook(() => useGetVaccinations(), {
      wrapper: createTestWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it("BUG-502: list search + end_date are sent server-side so PACO is not client-windowed", async () => {
    let seenUrl = "";
    server.use(
      http.get("/api/v1/vaccinations", ({ request }) => {
        seenUrl = request.url;
        const url = new URL(request.url);
        expect(url.searchParams.get("search")).toBe("PACO");
        expect(url.searchParams.get("end_date")).toBe("2026-08-29");
        expect(url.searchParams.get("limit")).toBe(String(HISTORY_FETCH_LIMIT));
        return HttpResponse.json({
          data: [
            {
              id: 1000000000,
              pet_id: 42,
              vaccine_id: 1,
              date: "2026-08-29T00:00:00+09:00",
              next_date: "2026-09-10T00:00:00+09:00",
              vaccine: { name: "混合" },
              pet: { name: "PACO", owner: { name: "S03" } },
            },
          ],
          total: 1,
          page: 1,
          limit: HISTORY_FETCH_LIMIT,
        });
      }),
    );

    const { result } = renderHook(
      () =>
        useGetVaccinations({
          search: "PACO",
          endDate: "2026-08-29",
          page: 1,
          limit: HISTORY_FETCH_LIMIT,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.id).toBe("1000000000");
    expect(result.current.data?.[0]?.petName).toBe("PACO");
    expect(result.current.data?.[0]?.nextDate).toBe("2026-09-10");
    expect(seenUrl).toContain("search=PACO");
    expect(seenUrl).toContain("end_date=2026-08-29");
  });
});
