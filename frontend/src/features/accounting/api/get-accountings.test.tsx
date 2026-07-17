import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { useGetAccountingsPage } from "./get-accountings";

const mockedGet = vi.mocked(axios.get);

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

// BUG-411: 会計一覧が backend デフォルト limit=20 で切られた配列に偽のクライアントページネーションを
// 行い、実データの大半（392,105件中20件のみ）が不可視化していた回帰の固定テスト。
// useGetAccountingsPage は page/limit を実際にリクエストへ転送し、サーバの total をそのまま返すこと。
describe("useGetAccountingsPage — BUG-411 サーバページネーション配線", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("page/limit をクエリパラメータとして送信する", async () => {
    mockedGet.mockResolvedValue({ data: { data: [], total: 0, page: 3, limit: 20 } });

    const { result } = renderHook(
      () => useGetAccountingsPage({ page: 3, limit: 20 }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockedGet).toHaveBeenCalledWith(
      "/v1/accountings",
      expect.objectContaining({ params: expect.objectContaining({ page: "3", limit: "20" }) }),
    );
  });

  it("サーバの total を data.length ではなく真の総件数として返す（偽ページネーション再発防止）", async () => {
    mockedGet.mockResolvedValue({
      data: {
        data: [{ id: 1, clinic_id: 1, owner_id: 1, pet_id: 1, status: "completed", scheduled_date: "2026-07-01", items: [] }],
        total: 392105,
        page: 1,
        limit: 20,
      },
    });

    const { result } = renderHook(
      () => useGetAccountingsPage({ page: 1, limit: 20 }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.total).toBe(392105);
    expect(result.current.data?.total).not.toBe(result.current.data?.data.length);
  });

  it("page=1 と page=2 で異なるクエリパラメータを送る（同一20件の使い回しではない）", async () => {
    mockedGet.mockResolvedValue({ data: { data: [], total: 100, page: 1, limit: 20 } });

    const { rerender } = renderHook(
      ({ page }: { page: number }) => useGetAccountingsPage({ page, limit: 20 }),
      { wrapper, initialProps: { page: 1 } },
    );
    await waitFor(() => expect(mockedGet).toHaveBeenCalledTimes(1));

    rerender({ page: 2 });
    await waitFor(() => expect(mockedGet).toHaveBeenCalledTimes(2));

    expect(mockedGet.mock.calls[0][1]).toMatchObject({ params: expect.objectContaining({ page: "1" }) });
    expect(mockedGet.mock.calls[1][1]).toMatchObject({ params: expect.objectContaining({ page: "2" }) });
  });
});
