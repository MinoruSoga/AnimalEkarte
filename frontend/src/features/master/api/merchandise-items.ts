import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { MerchandiseItem } from "@/types/generated/models";

// ─── Transform ────────────────────────────────────────────

function transformMerchandiseItem(item: MerchandiseItem) {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    taxType: item.tax_type ?? "excluded",
    taxRate: item.tax_rate,
    isActive: item.is_active,
    sortOrder: item.sort_order,
  };
}

export type FrontendMerchandiseItem = ReturnType<typeof transformMerchandiseItem>;

// ─── API Request Types ────────────────────────────────────

export type CreateMerchandiseItemRequest = Omit<
  MerchandiseItem,
  "id" | "clinic_id" | "sort_order" | "created_at" | "updated_at"
>;

export type UpdateMerchandiseItemRequest = Partial<CreateMerchandiseItemRequest>;

interface ReorderMerchandiseItemsRequest {
  ids: number[];
}

export type { ReorderMerchandiseItemsRequest };

// ─── Queries ──────────────────────────────────────────────

const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItem[] | { data: MerchandiseItem[] }>(
    "/v1/masters/merchandise-items",
  );
  const items = Array.isArray(data) ? data : data.data;
  return items.map(transformMerchandiseItem);
};

export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("merchandise-items"),
    queryFn: getAllMerchandiseItems,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

// ─── Mutations ────────────────────────────────────────────

export const useCreateMerchandiseItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateMerchandiseItemRequest) => {
      const { data } = await axios.post<MerchandiseItem>("/v1/masters/merchandise-items", req);
      return transformMerchandiseItem(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("merchandise-items") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateMerchandiseItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, req }: { id: string; req: UpdateMerchandiseItemRequest }) => {
      const { data } = await axios.patch<MerchandiseItem>(
        `/v1/masters/merchandise-items/${id}`,
        req,
      );
      return transformMerchandiseItem(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("merchandise-items") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteMerchandiseItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await axios.delete(`/v1/masters/merchandise-items/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("merchandise-items") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const useReorderMerchandiseItems = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: ReorderMerchandiseItemsRequest) => {
      await axios.patch("/v1/masters/merchandise-items/reorder", req);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("merchandise-items") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
