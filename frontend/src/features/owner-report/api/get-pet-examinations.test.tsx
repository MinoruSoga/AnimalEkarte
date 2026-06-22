import type { ReactNode } from "react";
import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { useGetPetExaminations } from "./get-pet-examinations";

function makeBackendExam(id: number, status: string, typeName: string) {
  return {
    id,
    date: "2026-05-10T00:00:00+09:00",
    status,
    exam_type: { id: 1, name: typeName },
    items: [],
  };
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("useGetPetExaminations (#158 下書き除外)", () => {
  it("依頼中/検査中を除外し、結果入力済み/完了/確定のみ日付順で返す", async () => {
    server.use(
      http.get("/api/v1/examinations", () =>
        HttpResponse.json({
          data: [
            makeBackendExam(1, "pending", "A"),
            makeBackendExam(2, "in_progress", "B"),
            makeBackendExam(3, "result_entered", "C"),
            makeBackendExam(4, "completed", "D"),
            makeBackendExam(5, "confirmed", "E"),
          ],
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetExaminations("7"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const statuses = (result.current.data ?? []).map((e) => e.status);
    expect(statuses).toEqual(["結果入力済み", "完了", "確定"]);
    expect(statuses).not.toContain("依頼中");
    expect(statuses).not.toContain("検査中");
  });

  it("petId 未指定ならクエリは無効でフェッチしない", () => {
    const { result } = renderHook(() => useGetPetExaminations(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe("idle");
  });
});
