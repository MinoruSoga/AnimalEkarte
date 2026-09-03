import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { useUpdateVaccination } from "./update-vaccination";

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

// FE4-6 fix: 更新成功後、detail クエリキー ["vaccination", id]（単数形）が
// invalidate されることを固定する回帰テスト。旧実装は ["vaccinations"]（複数形の
// list prefix）のみ invalidate しており、詳細画面が stale のまま残るバグがあった。
describe("useUpdateVaccination (FE4-6)", () => {
  it("成功後に list prefix と detail キーの両方を invalidate する", async () => {
    server.use(http.patch("/api/v1/vaccinations/:id", () => HttpResponse.json({ id: 42 })));

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useUpdateVaccination(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      result.current.mutate({ id: "42", req: {} });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: queryKeys.vaccinations.all() });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: queryKeys.vaccinations.detail("42") });
  });
});
