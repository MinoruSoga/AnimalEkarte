import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { describe, expect, it, vi } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import { server } from "@/testing/mocks/node";
import { useUnconfirmExamination } from "./unconfirm-examination";

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe("useUnconfirmExamination", () => {
  it("理由を専用endpointへPOSTし、関連する検査cacheを無効化する", async () => {
    let requestBody: unknown;
    server.use(
      http.post("/api/v1/examinations/:id/unconfirm", async ({ request }) => {
        requestBody = await request.json();
        return HttpResponse.json({
          id: 7,
          clinic_id: 1,
          exam_type_id: 2,
          date: "2026-08-03T00:00:00Z",
          result_summary: "",
          machine: "",
          status: "completed",
          current_revision_version: 2,
          created_at: "2026-08-03T00:00:00Z",
          updated_at: "2026-08-03T00:00:00Z",
        });
      }),
    );

    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUnconfirmExamination(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ id: "7", reason: "再確認のため" });
    });

    expect(requestBody).toEqual({ reason: "再確認のため" });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.examinations.all(),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.examinations.detail("7"),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.examinations.items("7"),
    });
  });

  it("関連cacheの無効化が完了するまで mutation を完了扱いにしない", async () => {
    server.use(
      http.post("/api/v1/examinations/:id/unconfirm", () =>
        HttpResponse.json({
          id: 7,
          clinic_id: 1,
          exam_type_id: 2,
          date: "2026-08-03T00:00:00Z",
          result_summary: "",
          machine: "",
          status: "completed",
          created_at: "2026-08-03T00:00:00Z",
          updated_at: "2026-08-03T00:00:00Z",
        }),
      ),
    );

    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    let releaseInvalidations: (() => void) | undefined;
    const invalidationsPending = new Promise<void>((resolve) => {
      releaseInvalidations = resolve;
    });
    const invalidateSpy = vi
      .spyOn(queryClient, "invalidateQueries")
      .mockReturnValue(invalidationsPending);
    const { result } = renderHook(() => useUnconfirmExamination(), {
      wrapper: createWrapper(queryClient),
    });

    let settled = false;
    const mutation = result.current.mutateAsync({ id: "7", reason: "再確認のため" }).then(() => {
      settled = true;
    });

    await waitFor(() => expect(invalidateSpy).toHaveBeenCalledTimes(3));
    await Promise.resolve();
    expect(settled).toBe(false);

    releaseInvalidations?.();
    await act(async () => mutation);
    expect(settled).toBe(true);
  });
});
