import { useMemo } from "react";
import type { InventoryItem } from "@/types";
import { useGetInventoryItems } from "../api/inventory";
import { normalizeKana } from "@/lib/normalize-kana";
type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";

interface UseInventoryParams {
  searchTerm: string;
  category?: CategoryFilter;
  statusFilter?: StatusFilter;
}

export function useInventoryList({
  searchTerm,
  category = "all",
  statusFilter = "all",
}: UseInventoryParams) {
  const serverParams = {
    category: category !== "all" ? category : undefined,
    status: statusFilter !== "all" ? statusFilter : undefined,
  };
  const { data: items = [], isLoading, isError } = useGetInventoryItems(serverParams);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return items;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return items.filter(
      (item) =>
        normalizeKana(item.name).toLowerCase().includes(normalizedTerm) ||
        (item.location ? normalizeKana(item.location).toLowerCase().includes(normalizedTerm) : false) ||
        (item.supplier ? normalizeKana(item.supplier).toLowerCase().includes(normalizedTerm) : false)
    );
  }, [items, searchTerm]);

  const summary = useMemo(() => {
    const total = items.length;
    const lowStock = items.filter((i) => i.status === "low").length;
    const outOfStock = items.filter((i) => i.status === "out_of_stock").length;
    return { total, lowStock, outOfStock };
  }, [items]);

  return { data: filteredItems, summary, isLoading, isError };
}
