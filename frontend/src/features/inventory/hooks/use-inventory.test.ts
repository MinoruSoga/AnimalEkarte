import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useInventoryList } from "./use-inventory";
import type { InventoryItem } from "@/types";

const { mockUseGetInventoryItems } = vi.hoisted(() => ({
  mockUseGetInventoryItems: vi.fn(),
}));

vi.mock("../api/inventory", () => ({
  useGetInventoryItems: mockUseGetInventoryItems,
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

describe("useInventoryList", () => {
  beforeEach(() => {
    mockUseGetInventoryItems.mockReset();
    mockUseGetInventoryItems.mockReturnValue({ data: [], isLoading: false, isError: false });
  });

  describe("summary 集計", () => {
    it("status='low' の件数を lowStock として集計する", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [
          makeItem({ id: "1", status: "low" }),
          makeItem({ id: "2", status: "low" }),
          makeItem({ id: "3", status: "sufficient" }),
        ],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));

      expect(result.current.summary).toEqual({ total: 3, lowStock: 2, outOfStock: 0 });
    });

    it("status='out_of_stock' の件数を outOfStock として集計する", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [
          makeItem({ id: "1", status: "out_of_stock" }),
          makeItem({ id: "2", status: "sufficient" }),
        ],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));

      expect(result.current.summary).toEqual({ total: 2, lowStock: 0, outOfStock: 1 });
    });

    it("回帰防止: 未知の status 値は lowStock にも outOfStock にも計上されない（total のみ増える）", () => {
      // status は string 型のため、バックエンドが新しい状態値（例: 'expiring_soon'）を
      // 追加してもフロントの filter 条件（=== 'low' / === 'out_of_stock'）は追従しない。
      // このテストはその現状の挙動を固定する回帰テストであり、新ステータス追加時は
      // useInventoryList 側の summary 集計ロジックの追従漏れに注意が必要。
      mockUseGetInventoryItems.mockReturnValue({
        data: [
          makeItem({ id: "1", status: "expiring_soon" as InventoryItem["status"] }),
          makeItem({ id: "2", status: "low" }),
        ],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));

      expect(result.current.summary).toEqual({ total: 2, lowStock: 1, outOfStock: 0 });
    });

    it("空配列のとき全て 0 を返す", () => {
      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));
      expect(result.current.summary).toEqual({ total: 0, lowStock: 0, outOfStock: 0 });
    });
  });

  describe("filteredItems 検索", () => {
    it("searchTerm が空文字なら全件返す（フィルタしない）", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [makeItem({ id: "1" }), makeItem({ id: "2", name: "猫用フード" })],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));

      expect(result.current.data).toHaveLength(2);
    });

    it("name の部分一致でフィルタする", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [makeItem({ id: "1", name: "犬用フード" }), makeItem({ id: "2", name: "猫用フード" })],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "犬" }));

      expect(result.current.data).toHaveLength(1);
      expect(result.current.data[0].id).toBe("1");
    });

    it("location の部分一致でフィルタする", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [
          makeItem({ id: "1", name: "薬品A", location: "倉庫2階" }),
          makeItem({ id: "2", name: "薬品B", location: "倉庫1階" }),
        ],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "2階" }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });

    it("supplier の部分一致でフィルタする", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [
          makeItem({ id: "1", name: "薬品A", supplier: "アニマル商事" }),
          makeItem({ id: "2", name: "薬品B", supplier: "ペット卸" }),
        ],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "アニマル" }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });

    it("location/supplier が undefined のアイテムでも例外にならず単に非マッチとして扱う", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [makeItem({ id: "1", name: "在庫A", location: undefined, supplier: undefined })],
        isLoading: false,
        isError: false,
      });

      expect(() =>
        renderHook(() => useInventoryList({ searchTerm: "倉庫" })),
      ).not.toThrow();
    });

    it("かな正規化（全角/半角・カタカナ/ひらがな）を考慮して検索する", () => {
      mockUseGetInventoryItems.mockReturnValue({
        data: [makeItem({ id: "1", name: "ワクチン" })],
        isLoading: false,
        isError: false,
      });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "わくちん" }));

      expect(result.current.data.map((i) => i.id)).toEqual(["1"]);
    });
  });

  describe("サーバーパラメータ", () => {
    it("category / statusFilter が 'all' 以外のとき useGetInventoryItems に渡す", () => {
      renderHook(() =>
        useInventoryList({ searchTerm: "", category: "medicine", statusFilter: "low" }),
      );

      expect(mockUseGetInventoryItems).toHaveBeenCalledWith({
        category: "medicine",
        status: "low",
      });
    });

    it("category / statusFilter が 'all' のとき undefined を渡す（サーバー側で絞り込まない）", () => {
      renderHook(() => useInventoryList({ searchTerm: "", category: "all", statusFilter: "all" }));

      expect(mockUseGetInventoryItems).toHaveBeenCalledWith({
        category: undefined,
        status: undefined,
      });
    });
  });

  describe("isLoading / isError の透過", () => {
    it("useGetInventoryItems の isLoading / isError をそのまま返す", () => {
      mockUseGetInventoryItems.mockReturnValue({ data: [], isLoading: true, isError: true });

      const { result } = renderHook(() => useInventoryList({ searchTerm: "" }));

      expect(result.current.isLoading).toBe(true);
      expect(result.current.isError).toBe(true);
    });
  });
});
