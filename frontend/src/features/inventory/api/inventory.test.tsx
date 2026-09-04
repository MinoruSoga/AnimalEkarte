import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn(), post: vi.fn(), patch: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useGetInventoryItemsPage } from "./inventory";

const mockedGet = vi.mocked(axios.get);

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

// BUG-412 回帰テスト: 旧実装は page/limit を送らず、レスポンスの total/page/limit を捨てて
// data のみ返していたため、backend デフォルト limit=20 に切られた在庫が不可視化していた
// （seed 30件のうち10件が一覧・フィルタ・検索から到達不能）。
describe("useGetInventoryItemsPage（BUG-412）", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("page/limit をクエリパラメータとして /v1/inventory に送る", async () => {
    mockedGet.mockResolvedValue({ data: { data: [], total: 0, page: 2, limit: 20 } });

    const { result } = renderHook(
      () => useGetInventoryItemsPage({ category: "medicine", status: "low", page: 2, limit: 20 }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockedGet).toHaveBeenCalledWith("/v1/inventory", {
      params: { category: "medicine", status: "low", page: 2, limit: 20 },
    });
  });

  it("レスポンスの実 total（20 超）を捨てずにそのまま返す", async () => {
    mockedGet.mockResolvedValue({
      data: {
        data: [
          {
            id: 1,
            clinic_id: 1,
            name: "犬用フード",
            category: "food",
            quantity: 10,
            unit: "袋",
            min_stock_level: 2,
            location: null,
            expiry_date: null,
            supplier: null,
            last_restocked: null,
            status: "sufficient",
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        ],
        total: 30,
        page: 1,
        limit: 20,
      },
    });

    const { result } = renderHook(() => useGetInventoryItemsPage({ page: 1, limit: 20 }), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // 旧実装（getInventoryItems）は Promise<InventoryItem[]> のみを返し total/page/limit を破棄していた。
    expect(result.current.data?.total).toBe(30);
    expect(result.current.data?.page).toBe(1);
    expect(result.current.data?.limit).toBe(20);
    expect(result.current.data?.data).toHaveLength(1);
  });
});
