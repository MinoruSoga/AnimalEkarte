import { useQuery } from "@tanstack/react-query";
import type { MasterItem } from "@/types";
import { getMasterItemsByEndpoint, MASTER_CATEGORY_ENDPOINT } from "./get-master-items";

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
