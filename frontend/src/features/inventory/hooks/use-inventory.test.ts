import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useInventoryList } from "./use-inventory";
import type { InventoryItem } from "@/types";

const { mockUseGetInventoryItemsPage } = vi.hoisted(() => ({
  mockUseGetInventoryItemsPage: vi.fn(),
}));

vi.mock("../api/inventory", () => ({
  useGetInventoryItemsPage: mockUseGetInventoryItemsPage,
}));

function makeItem(overrides: Partial<InventoryItem> = {}): InventoryItem {
  return {
    id: "1",
    clinicId: "1",
    name: "犬用フード",
    category: "food",
    quantity: 10,
    unit: "袋",
    minStockLevel: 2,
    location: undefined,
    expiryDate: undefined,
    supplier: undefined,
    lastRestocked: undefined,
    status: "sufficient",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  } as InventoryItem;
}

function mockPage(items: InventoryItem[], total?: number, page = 1, limit = 20) {
  return { data: items, total: total ?? items.length, page, limit };
}

describe("useInventoryList", () => {
  beforeEach(() => {
    mockUseGetInventoryItemsPage.mockReset();
    mockUseGetInventoryItemsPage.mockReturnValue({ data: mockPage([]), isLoading: false, isError: false });
  });

  describe("summary 集計（BUG-412: ページ非依存でなければならない）", () => {
    it("status='low'/'out_of_stock' クエリの実 total を summary として使う", () => {
      mockUseGetInventoryItemsPage.mockImplementation((params: { status?: string }) => {
        if (params.status === "low") return { data: mockPage([], 2), isLoading: false, isError: false };
        if (params.status === "out_of_stock") return { data: mockPage([], 1), isLoading: false, isError: false };
        return {
          data: mockPage([makeItem({ id: "1" }), makeItem({ id: "2" }), makeItem({ id: "3" })], 3),
          isLoading: false,
          isError: false,
        };
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));

      expect(result.current.summary).toEqual({ total: 3, lowStock: 2, outOfStock: 1, isError: false });
    });

    it("回帰防止: 現在ページの items を集計しない（ページ送りで件数が変動しない）", () => {
      mockUseGetInventoryItemsPage.mockImplementation((params: { status?: string; page?: number }) => {
        if (params.status === "low") return { data: mockPage([], 2), isLoading: false, isError: false };
        if (params.status === "out_of_stock") return { data: mockPage([], 1), isLoading: false, isError: false };
        // page=2 のメインクエリは1件しか返さないが、summary はこれを流用してはならない
        return { data: mockPage([makeItem({ id: "21" })], 3, 2, 20), isLoading: false, isError: false };
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 2, limit: 20 }));

      expect(result.current.summary).toEqual({ total: 3, lowStock: 2, outOfStock: 1, isError: false });
    });

    it("未ロード時は全て 0 を返す", () => {
      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));
      expect(result.current.summary).toEqual({ total: 0, lowStock: 0, outOfStock: 0, isError: false });
    });

    it("code-reviewer指摘(HIGH)回帰防止: low/out_of_stockクエリが失敗した場合、summary.isErrorがtrueになり0件へサイレントに丸め込まない", () => {
      mockUseGetInventoryItemsPage.mockImplementation((params: { status?: string }) => {
        if (params.status === "low") return { data: undefined, isLoading: false, isError: true };
        if (params.status === "out_of_stock") return { data: mockPage([], 0), isLoading: false, isError: false };
        return { data: mockPage([makeItem()], 1), isLoading: false, isError: false };
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));

      expect(result.current.summary.isError).toBe(true);
    });
  });

  describe("filteredItems 検索（現在ページ内のみ・クライアントサイド）", () => {
    it("searchTerm が空文字なら全件返す（フィルタしない）", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage([makeItem({ id: "1" }), makeItem({ id: "2", name: "猫用フード" })], 2),
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));

      expect(result.current.data).toHaveLength(2);
    });

    it("name の部分一致でフィルタする", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage([makeItem({ id: "1", name: "犬用フード" }), makeItem({ id: "2", name: "猫用フード" })], 2),
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "犬", page: 1, limit: 20 }));

      expect(result.current.data).toHaveLength(1);
      expect(result.current.data[0].id).toBe("1");
    });

    it("location の部分一致でフィルタする", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage(
          [
            makeItem({ id: "1", name: "薬品A", location: "倉庫2階" }),
            makeItem({ id: "2", name: "薬品B", location: "倉庫1階" }),
          ],
          2
        ),
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "2階", page: 1, limit: 20 }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });

    it("supplier の部分一致でフィルタする", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage(
          [
            makeItem({ id: "1", name: "薬品A", supplier: "アニマル商事" }),
            makeItem({ id: "2", name: "薬品B", supplier: "ペット卸" }),
          ],
          2
        ),
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "アニマル", page: 1, limit: 20 }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });

    it("location/supplier が undefined のアイテムでも例外にならず単に非マッチとして扱う", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage([makeItem({ id: "1", name: "在庫A", location: undefined, supplier: undefined })], 1),
        isLoading: false,
        isError: false,
      });

      expect(() =>
        renderHook(() => useInventoryList({ searchTerm: "倉庫", page: 1, limit: 20 })),
      ).not.toThrow();
    });

    it("かな正規化（全角/半角・カタカナ/ひらがな）を考慮して検索する", () => {
      mockUseGetInventoryItemsPage.mockReturnValue({
        data: mockPage([makeItem({ id: "1", name: "ワクチン" })], 1),
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "わくちん", page: 1, limit: 20 }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });
  });

  describe("サーバーパラメータ（BUG-412: page/limit を必ず送る）", () => {
    it("category / statusFilter / page / limit を useGetInventoryItemsPage に渡す", () => {
      renderHook(() =>
        useInventoryList({ searchTerm: "", category: "medicine", statusFilter: "low", page: 2, limit: 20 }),
      );

      expect(mockUseGetInventoryItemsPage).toHaveBeenCalledWith({
        category: "medicine",
        status: "low",
        page: 2,
        limit: 20,
      });
    });

    it("category / statusFilter が 'all' のとき undefined を渡す（サーバー側で絞り込まない）", () => {
      renderHook(() =>
        useInventoryList({ searchTerm: "", category: "all", statusFilter: "all", page: 1, limit: 20 }),
      );

      expect(mockUseGetInventoryItemsPage).toHaveBeenCalledWith({
        category: undefined,
        status: undefined,
        page: 1,
        limit: 20,
      });
    });

    it("回帰防止: 20件超のtotalでも実totalをそのまま返す（旧実装はtotal自体を捨てていた）", () => {
      mockUseGetInventoryItemsPage.mockImplementation((params: { status?: string }) => {
        if (params.status === "low" || params.status === "out_of_stock") {
          return { data: mockPage([], 0), isLoading: false, isError: false };
        }
        return { data: mockPage([makeItem()], 30, 1, 20), isLoading: false, isError: false };
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));

      expect(result.current.total).toBe(30);
      expect(result.current.page).toBe(1);
      expect(result.current.limit).toBe(20);
    });
  });

  describe("isLoading / isError の透過", () => {
    it("メインクエリの isLoading / isError をそのまま返す", () => {
      mockUseGetInventoryItemsPage.mockImplementation((params: { status?: string }) => {
        if (params.status === "low" || params.status === "out_of_stock") {
          return { data: mockPage([], 0), isLoading: false, isError: false };
        }
        return { data: undefined, isLoading: true, isError: true };
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "", page: 1, limit: 20 }));

      expect(result.current.isLoading).toBe(true);
      expect(result.current.isError).toBe(true);
    });
  });
});
