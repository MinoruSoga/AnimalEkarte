import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { useUpdateReservation } from "./use-update-reservation";

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

// FE4-6 fix: 更新成功後、detail クエリキー ["reservation", id]（単数形）が
// invalidate されることを固定する回帰テスト。旧実装は ["reservations"]（複数形）と
// ["reception"] のみ invalidate しており、詳細画面が stale のまま残るバグがあった
// （先例: update-reservation-route.ts は既に両方を invalidate している）。
describe("useUpdateReservation (FE4-6)", () => {
  it("成功後に list prefix / reception と detail キーの両方を invalidate する", async () => {
    server.use(
      http.patch("/api/v1/reservations/:id", () =>
        HttpResponse.json({
          id: 99,
          start_time: "2026-07-11T09:00:00+09:00",
          end_time: "2026-07-11T09:30:00+09:00",
        }),
      ),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useUpdateReservation(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      result.current.mutate({ id: "99", req: {} });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["reservations"] });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["reception"] });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["reservation", "99"] });
  });
});
