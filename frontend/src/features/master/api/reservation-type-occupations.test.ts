import { describe, expect, it, afterEach, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { queryKeys } from "@/lib/query-keys";
import {
  useGetReservationTypeOccupations,
  useLinkOccupation,
} from "./reservation-type-occupations";

afterEach(() => {
  server.resetHandlers();
});

const ENDPOINT = "/api/v1/masters/reservation-types/5/occupations";

function makeRawLinked(id: number, occupationName: string) {
  return {
    id,
    clinic_id: 1,
    reservation_type_id: 5,
    occupation_id: id + 100,
    occupation: {
      id: id + 100,
      clinic_id: 1,
      name: occupationName,
      description: "",
      sort_order: 1,
      is_active: true,
      created_at: "2026-09-23T00:00:00Z",
      updated_at: "2026-09-23T00:00:00Z",
    },
    created_at: "2026-09-23T00:00:00Z",
  };
}

describe("useGetReservationTypeOccupations", () => {
  it("BE の裸配列レスポンスを変換して返す（EMR-209）", async () => {
    server.use(http.get(ENDPOINT, () => HttpResponse.json([makeRawLinked(10, "V04職種")])));

    const { result } = renderHook(() => useGetReservationTypeOccupations("1", "5"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]).toMatchObject({
      id: 10,
      reservationTypeId: 5,
      occupationId: 110,
      occupation: { id: 110, name: "V04職種" },
    });
  });

  it("{data: [...]} エンベロープ形式も許容する", async () => {
    server.use(
      http.get(ENDPOINT, () => HttpResponse.json({ data: [makeRawLinked(11, "トリマー")] })),
    );

    const { result } = renderHook(() => useGetReservationTypeOccupations("1", "5"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].occupation?.name).toBe("トリマー");
  });
});

function createQueryWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

describe("useLinkOccupation", () => {
  it("201 の {data: <link>} エンベロープを解釈し、一覧クエリを無効化する（EMR-209）", async () => {
    server.use(
      http.post(ENDPOINT, () =>
        HttpResponse.json({ data: makeRawLinked(20, "V04職種") }, { status: 201 }),
      ),
    );
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useLinkOccupation("1", "5"), {
      wrapper: createQueryWrapper(queryClient),
    });

    let linked: Awaited<ReturnType<typeof result.current.mutateAsync>> | undefined;
    await act(async () => {
      linked = await result.current.mutateAsync(3);
    });

    expect(linked).toMatchObject({
      id: 20,
      reservationTypeId: 5,
      occupationId: 120,
      occupation: { id: 120, name: "V04職種" },
    });
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.reservationTypeSubResource("1", "5", "occupations"),
    });
  });

  it("409 競合時も職種一覧クエリを無効化してバッジ表示を同期する（EMR-209）", async () => {
    server.use(
      http.post(ENDPOINT, () =>
        HttpResponse.json(
          {
            error: "この職種はすでに紐付けられています",
            code: "ALREADY_EXISTS",
            data: makeRawLinked(21, "V04職種"),
          },
          { status: 409 },
        ),
      ),
    );
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useLinkOccupation("1", "5"), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await expect(result.current.mutateAsync(3)).rejects.toThrow();
    });

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.reservationTypeSubResource("1", "5", "occupations"),
    });
  });
});
