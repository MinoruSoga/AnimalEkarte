import { useQuery } from "@tanstack/react-query";
import type { MasterItem } from "@/types";
import { getMasterItemsByEndpoint, MASTER_CATEGORY_ENDPOINT } from "./get-master-items";

export const getMasterItem = async (_id: string): Promise<MasterItem | null> => {
  // Cannot look up a single item without knowing the category.
  // Return null as fallback.
  return null;
};

export const useGetMasterItem = (id: string) => {
  return useQuery({
    queryKey: ["masterItem", id],
    queryFn: () => getMasterItem(id),
    enabled: !!id,
  });
};

export const getMasterItemsByCategory = async (category: string): Promise<MasterItem[]> => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  if (!endpoint) return [];
  return getMasterItemsByEndpoint(endpoint);
};

export const useGetMasterItemsByCategory = (category: string) => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  return useQuery({
    queryKey: ["masterItems", "category", category],
    queryFn: () => getMasterItemsByCategory(category),
    enabled: !!endpoint,
  });
};

// status filtering is now handled server-side per endpoint — stub for compat
export const getMasterItemsByStatus = async (_status: string): Promise<MasterItem[]> => [];
export const useGetMasterItemsByStatus = (status: string) =>
  useQuery({
    queryKey: ["masterItems", "status", status],
    queryFn: () => getMasterItemsByStatus(status),
    enabled: false,
  });
