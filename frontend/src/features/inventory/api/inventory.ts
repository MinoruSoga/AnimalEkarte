import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { BackendInventoryItem, CreateInventoryItemRequest, UpdateInventoryItemRequest } from "./types";

// BUG-412: page/limit をサーバへ転送し total を消費するページネーション対応パラメータ。
// BUG-411（会計・検査, 93c54ede9）の getAccountingsPage と同型。
interface GetInventoryItemsPageParams {
  category?: string;
  status?: string;
  page: number;
  limit: number;
}

interface InventoryListResponse {
  data: BackendInventoryItem[];
  total: number;
  page: number;
  limit: number;
}

export interface InventoryItemsPage {
  data: InventoryItem[];
  total: number;
  page: number;
  limit: number;
}

function transformInventoryItem(raw: BackendInventoryItem) {
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

const getInventoryItemsPage = async (params: GetInventoryItemsPageParams): Promise<InventoryItemsPage> => {
  const { data } = await axios.get<InventoryListResponse>("/v1/inventory", { params });
  return {
    data: data.data.map(transformInventoryItem),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

// BUG-412: page/limit をサーバへ転送し、レスポンスの total/page/limit をそのまま消費する。
// placeholderData: keepPreviousData で、ページ遷移中に表が一瞬空にならないようにする(BUG-411踏襲)。
export const useGetInventoryItemsPage = (params: GetInventoryItemsPageParams) => {
  return useQuery({
    queryKey: queryKeys.inventoryItems.list(params),
    queryFn: () => getInventoryItemsPage(params),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
    placeholderData: keepPreviousData,
  });
};

const getInventoryItem = async (id: string): Promise<InventoryItem> => {
  const { data } = await axios.get<BackendInventoryItem>(`/v1/inventory/${id}`);
  return transformInventoryItem(data);
};

export const useGetInventoryItem = (id: string) => {
  return useQuery({
    queryKey: queryKeys.inventoryItems.detail(id),
    queryFn: () => getInventoryItem(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

const createInventoryItem = async (req: CreateInventoryItemRequest): Promise<InventoryItem> => {
  const { data } = await axios.post<BackendInventoryItem>("/v1/inventory", req);
  return transformInventoryItem(data);
};

export const useCreateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInventoryItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.inventoryItems.all() });
    },
    onError: (error) => handleApiError(error, "在庫作成"),
  });
};

const updateInventoryItem = async (id: string, req: UpdateInventoryItemRequest): Promise<InventoryItem> => {
  const { data } = await axios.patch<BackendInventoryItem>(`/v1/inventory/${id}`, req);
  return transformInventoryItem(data);
};

export const useUpdateInventoryItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInventoryItemRequest }) => updateInventoryItem(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.inventoryItems.all() });
    },
    onError: (error) => handleApiError(error, "在庫更新"),
  });
};
