import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { BackendInventoryItem, CreateInventoryItemRequest, UpdateInventoryItemRequest } from "./types";

interface GetInventoryItemsParams {
  category?: string;
  status?: string;
}

interface InventoryListResponse {
  data: BackendInventoryItem[];
  total: number;
  page: number;
  limit: number;
}

export function transformInventoryItem(raw: BackendInventoryItem) {
  return {
    id: String(raw.id ?? 0),
    clinicId: String(raw.clinic_id ?? 0),
    name: raw.name,
    category: raw.category,
    quantity: raw.quantity,
    unit: raw.unit,
    minStockLevel: raw.min_stock_level,
    location: raw.location || undefined,
    expiryDate: raw.expiry_date ?? undefined,
    supplier: raw.supplier || undefined,
    lastRestocked: raw.last_restocked ?? undefined,
    status: raw.status,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

export type InventoryItem = ReturnType<typeof transformInventoryItem>;

export const getInventoryItems = async (params?: GetInventoryItemsParams): Promise<InventoryItem[]> => {
  const { data } = await axios.get<InventoryListResponse>("/v1/inventory", { params });
  return data.data.map(transformInventoryItem);
};

export const useGetInventoryItems = (params?: GetInventoryItemsParams) => {
  return useQuery({
    queryKey: ["inventoryItems", params],
    queryFn: () => getInventoryItems(params),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export const getInventoryItem = async (id: string): Promise<InventoryItem> => {
  const { data } = await axios.get<BackendInventoryItem>(`/v1/inventory/${id}`);
  return transformInventoryItem(data);
};

export const useGetInventoryItem = (id: string) => {
  return useQuery({
    queryKey: ["inventoryItem", id],
    queryFn: () => getInventoryItem(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export const createInventoryItem = async (req: CreateInventoryItemRequest): Promise<InventoryItem> => {
  const { data } = await axios.post<BackendInventoryItem>("/v1/inventory", req);
  return transformInventoryItem(data);
};

export const useCreateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInventoryItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventoryItems"] });
    },
    onError: (error) => handleApiError(error, "在庫作成"),
  });
};

export const updateInventoryItem = async (id: string, req: UpdateInventoryItemRequest): Promise<InventoryItem> => {
  const { data } = await axios.patch<BackendInventoryItem>(`/v1/inventory/${id}`, req);
  return transformInventoryItem(data);
};

export const useUpdateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInventoryItemRequest }) => updateInventoryItem(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventoryItems"] });
    },
    onError: (error) => handleApiError(error, "在庫更新"),
  });
};

export const deleteInventoryItem = async (id: string): Promise<void> => {
  await axios.delete(`/v1/inventory/${id}`);
};
