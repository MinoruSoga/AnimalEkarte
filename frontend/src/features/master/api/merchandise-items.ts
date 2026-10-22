import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { MerchandiseItem } from "@/types/generated/models";

// EMR-530: tax_rate は 有限な非負 exact decimal 文字列（採択405・483の十進輸送）。
// 生成modelの number 型（同期前）を Omit し、wire 契約どおりの文字列へ差し替える。
// 生成modelの再同期（EMR-527）後も Omit 差し替えは文字列契約を維持する。
export type MerchandiseItemExact = Omit<MerchandiseItem, "tax_rate"> & { tax_rate: string };

// ─── Transform ────────────────────────────────────────────

function transformMerchandiseItem(item: MerchandiseItemExact) {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    taxType: item.tax_type ?? "excluded",
    // EMR-530: 入力字句と保存値の十進文字列をそのまま保持する（number 化しない）。
    taxRate: item.tax_rate,
    isActive: item.is_active,
    sortOrder: item.sort_order,
  };
}

export type FrontendMerchandiseItem = ReturnType<typeof transformMerchandiseItem>;

// ─── API Request Types ────────────────────────────────────

export type CreateMerchandiseItemRequest = Omit<
  MerchandiseItemExact,
  "id" | "clinic_id" | "sort_order" | "created_at" | "updated_at"
>;

export type UpdateMerchandiseItemRequest = Partial<CreateMerchandiseItemRequest>;

interface ReorderMerchandiseItemsRequest {
  ids: number[];
}

export type { ReorderMerchandiseItemsRequest };

// ─── Queries ──────────────────────────────────────────────

const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItemExact[] | { data: MerchandiseItemExact[] }>(
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
      const { data } = await axios.post<MerchandiseItemExact>("/v1/masters/merchandise-items", req);
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
      const { data } = await axios.patch<MerchandiseItemExact>(
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
