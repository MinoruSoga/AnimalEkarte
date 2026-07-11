import type { ReactNode } from "react";
import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { useUpdateExamination } from "./use-update-examination";

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

// FE4-6 fix: 更新成功後、detail クエリキー ["examination", id]（単数形）が
// invalidate されることを固定する回帰テスト。旧実装は ["examinations"]（複数形の
// list prefix）のみ invalidate しており、詳細画面が stale のまま残るバグがあった
// （先例: update-examination-items.ts は既に両方を invalidate している）。
describe("useUpdateExamination (FE4-6)", () => {
  it("成功後に list prefix と detail キーの両方を invalidate する", async () => {
    server.use(
      http.patch("/api/v1/examinations/:id", () => HttpResponse.json({ id: 7 })),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useUpdateExamination(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      result.current.mutate({ id: "7", req: {} });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["examinations"] });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["examination", "7"] });
  });
});
