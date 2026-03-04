import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BackendInventoryItem, CreateInventoryItemRequest, UpdateInventoryItemRequest } from "./types";

export const getInventoryItems = async (): Promise<BackendInventoryItem[]> => {
  const { data } = await axios.get<BackendInventoryItem[]>("/v1/inventory-items");
  return data;
};

export const useGetInventoryItems = () => {
  return useQuery({
    queryKey: ["inventoryItems"],
    queryFn: getInventoryItems,
  });
};

export const getInventoryItem = async (id: string): Promise<BackendInventoryItem> => {
  const { data } = await axios.get<BackendInventoryItem>(`/v1/inventory-items/${id}`);
  return data;
};

export const useGetInventoryItem = (id: string) => {
  return useQuery({
    queryKey: ["inventoryItem", id],
    queryFn: () => getInventoryItem(id),
    enabled: !!id,
  });
};

export const getInventoryItemsByCategory = async (category: string): Promise<BackendInventoryItem[]> => {
  const { data } = await axios.get<BackendInventoryItem[]>(`/v1/inventory-items/category/${category}`);
  return data;
};

export const useGetInventoryItemsByCategory = (category: string) => {
  return useQuery({
    queryKey: ["inventoryItems", "category", category],
    queryFn: () => getInventoryItemsByCategory(category),
    enabled: !!category,
  });
};

export const getInventoryItemsByStatus = async (status: string): Promise<BackendInventoryItem[]> => {
  const { data } = await axios.get<BackendInventoryItem[]>(`/v1/inventory-items/status/${status}`);
  return data;
};

export const useGetInventoryItemsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["inventoryItems", "status", status],
    queryFn: () => getInventoryItemsByStatus(status),
    enabled: !!status,
  });
};

export const createInventoryItem = async (req: CreateInventoryItemRequest): Promise<BackendInventoryItem> => {
  const { data } = await axios.post<BackendInventoryItem>("/v1/inventory-items", req);
  return data;
};

export const useCreateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInventoryItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventoryItems"] });
    },
  });
};

export const updateInventoryItem = async (id: string, req: UpdateInventoryItemRequest): Promise<BackendInventoryItem> => {
  const { data } = await axios.put<BackendInventoryItem>(`/v1/inventory-items/${id}`, req);
  return data;
};

export const useUpdateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInventoryItemRequest }) => updateInventoryItem(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventoryItems"] });
    },
  });
};

export const deleteInventoryItem = async (id: string): Promise<void> => {
  await axios.delete(`/v1/inventory-items/${id}`);
};

export const useDeleteInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteInventoryItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventoryItems"] });
    },
  });
};
