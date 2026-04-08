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
    id: String(data.id ?? 0),
    name: data.name ?? "",
    category: (data.category ?? "other") as InventoryItem["category"],
    quantity: data.quantity ?? 0,
    unit: data.unit ?? "",
    minStockLevel: data.min_stock_level ?? 0,
    location: data.location,
    expiryDate: data.expiry_date ?? undefined,
    supplier: data.supplier,
    lastRestocked: data.last_restocked ?? undefined,
    status: (data.status ?? "sufficient") as InventoryItem["status"],
  };
}

export function useInventory({
  searchTerm,
  category = "all",
  statusFilter = "all",
}: UseInventoryParams) {
  const serverParams = {
    category: category !== "all" ? category : undefined,
    status: statusFilter !== "all" ? statusFilter : undefined,
  };
  const { data: backendItems = [], isLoading, isError } = useGetInventoryItems(serverParams);

  const items = useMemo(
    () => backendItems.map(transformInventoryItem),
    [backendItems]
  );

  const filteredItems = useMemo(() => {
    if (!searchTerm) return items;
    const lowerTerm = searchTerm.toLowerCase();
    return items.filter(
      (item) =>
        item.name.toLowerCase().includes(lowerTerm) ||
        (item.location?.toLowerCase().includes(lowerTerm) ?? false) ||
        (item.supplier?.toLowerCase().includes(lowerTerm) ?? false)
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
