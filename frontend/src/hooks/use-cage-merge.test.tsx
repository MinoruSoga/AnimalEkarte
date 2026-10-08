import type { ReactNode } from "react";
import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import { useMergeCages, useUnmergeCages } from "./use-cage-merge";

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

// EMR-306: ケージ一時連結 mutation。一覧キー "cages"（master CRUD）と
// "cage"（use-master-items 経由の入院 UI）は別配列要素のため両方 invalidate する。
describe("useMergeCages / useUnmergeCages (EMR-306)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("merge: POST /v1/masters/cages/:id/merge に member_cage_ids を送り、cages と cage の両カテゴリを invalidate する", async () => {
    let capturedId: string | undefined;
    let capturedBody: unknown;
    server.use(
      http.post("*/v1/masters/cages/:id/merge", async ({ request, params }) => {
        capturedId = String(params.id);
        capturedBody = await request.json();
        return HttpResponse.json({ id: Number(params.id) });
      }),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useMergeCages(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      result.current.mutate({ representativeId: "1", memberCageIds: [2, 3] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(capturedId).toBe("1");
    expect(capturedBody).toEqual({ member_cage_ids: [2, 3] });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("cages"),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("cage"),
    });
  });

  it("unmerge: POST /v1/masters/cages/:id/unmerge を呼び、cages と cage の両カテゴリを invalidate する", async () => {
    let capturedId: string | undefined;
    server.use(
      http.post("*/v1/masters/cages/:id/unmerge", ({ params }) => {
        capturedId = String(params.id);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useUnmergeCages(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      result.current.mutate("5");
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(capturedId).toBe("5");
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("cages"),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("cage"),
    });
  });
});
