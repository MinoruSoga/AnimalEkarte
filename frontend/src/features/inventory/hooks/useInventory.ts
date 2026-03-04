import { useMemo } from "react";
import type { InventoryItem } from "@/types";
import { useGetInventoryItems } from "../api/inventory";
import type { BackendInventoryItem } from "../api/types";

type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";

interface UseInventoryParams {
  searchTerm: string;
  category?: CategoryFilter;
  statusFilter?: StatusFilter;
}

function transformInventoryItem(data: BackendInventoryItem): InventoryItem {
  return {
    id: data.id,
    name: data.name,
    category: data.category as InventoryItem["category"],
    quantity: data.quantity,
    unit: data.unit,
    minStockLevel: data.min_stock_level,
    location: data.location,
    expiryDate: data.expiry_date ?? undefined,
    supplier: data.supplier,
    lastRestocked: data.last_restocked ?? undefined,
    status: data.status,
  };
}

export function useInventory({
  searchTerm,
  category = "all",
  statusFilter = "all",
}: UseInventoryParams) {
  const { data: backendItems = [], isLoading } = useGetInventoryItems();

  const items = useMemo(
    () => backendItems.map(transformInventoryItem),
    [backendItems]
  );

  const filteredItems = useMemo(() => {
    let result = items;

    if (category !== "all") {
      result = result.filter((item) => item.category === category);
    }

    if (statusFilter !== "all") {
      result = result.filter((item) => item.status === statusFilter);
    }

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

  const summary = useMemo(() => {
    const total = items.length;
    const lowStock = items.filter((i) => i.status === "low").length;
    const outOfStock = items.filter((i) => i.status === "out_of_stock").length;
    return { total, lowStock, outOfStock };
  }, [items]);

  return { data: filteredItems, summary, isLoading };
}
