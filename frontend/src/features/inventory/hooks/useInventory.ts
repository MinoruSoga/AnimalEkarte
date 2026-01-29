import { useMemo } from "react";
import { MOCK_INVENTORY_ITEMS } from "../api";
import type { InventoryItem } from "@/types";

type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";

interface UseInventoryParams {
  searchTerm: string;
  category?: CategoryFilter;
  statusFilter?: StatusFilter;
}

export function useInventory({
  searchTerm,
  category = "all",
  statusFilter = "all",
}: UseInventoryParams) {
  const items = MOCK_INVENTORY_ITEMS;

  const filteredItems = useMemo(() => {
    let result = items;

    // Category filter
    if (category !== "all") {
      result = result.filter((item) => item.category === category);
    }

    // Status filter
    if (statusFilter !== "all") {
      result = result.filter((item) => item.status === statusFilter);
    }

    // Search filter
    if (searchTerm) {
      const lowerTerm = searchTerm.toLowerCase();
      result = result.filter(
        (item) =>
          item.name.toLowerCase().includes(lowerTerm) ||
          (item.location?.toLowerCase().includes(lowerTerm) ?? false) ||
          (item.supplier?.toLowerCase().includes(lowerTerm) ?? false)
      );
    }

    return result;
  }, [items, searchTerm, category, statusFilter]);

  // Summary statistics
  const summary = useMemo(() => {
    const total = items.length;
    const lowStock = items.filter((i) => i.status === "low").length;
    const outOfStock = items.filter((i) => i.status === "out_of_stock").length;
    return { total, lowStock, outOfStock };
  }, [items]);

  return { data: filteredItems, summary };
}
