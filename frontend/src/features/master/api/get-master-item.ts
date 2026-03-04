import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { transformMasterItem } from "./transforms";
import type { BackendMasterItem } from "./types";

export const getMasterItem = async (id: string): Promise<MasterItem> => {
  const { data } = await axios.get<BackendMasterItem>(`/v1/master-items/${id}`);
  return transformMasterItem(data);
};

export const useGetMasterItem = (id: string) => {
  return useQuery({
    queryKey: ["masterItem", id],
    queryFn: () => getMasterItem(id),
    enabled: !!id,
  });
};

// Fetch master items by category
export const getMasterItemsByCategory = async (
  category: string
): Promise<MasterItem[]> => {
  const { data } = await axios.get<BackendMasterItem[]>(
    `/v1/master-items/category/${category}`
  );
  return data.map(transformMasterItem);
};

export const useGetMasterItemsByCategory = (category: string) => {
  return useQuery({
    queryKey: ["masterItems", "category", category],
    queryFn: () => getMasterItemsByCategory(category),
    enabled: !!category,
  });
};

// Fetch master items by status
export const getMasterItemsByStatus = async (
  status: string
): Promise<MasterItem[]> => {
  const { data } = await axios.get<BackendMasterItem[]>(
    `/v1/master-items/status/${status}`
  );
  return data.map(transformMasterItem);
};

export const useGetMasterItemsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["masterItems", "status", status],
    queryFn: () => getMasterItemsByStatus(status),
    enabled: !!status,
  });
};
