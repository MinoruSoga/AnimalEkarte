import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ItemCategory, TaxType } from "@/types";

// ─── HTTP DTO（非永続・EMR-527） ─────────────────────────────

/**
 * 物販品マスタ応答。保存 model（生成 MerchandiseItem）ではなく handler の応答列を写す。
 * accounting の get-merchandise-items.ts も同じ struct を写す（feature 間 import は禁止）。
 * @see backend/internal/inventory/merchandise_item_response.go merchandiseItemResponse
 */
interface LegacyMerchandiseItemResponse {
  id: number;
  clinic_id: number;
  name: string;
  category: ItemCategory;
  unit_price: number;
  tax_type: TaxType;
  tax_rate: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

/** 一覧は配列で返る。旧 { data } 包みも従来どおり受ける。 */
type MerchandiseItemsPayload =
  | LegacyMerchandiseItemResponse[]
  | { data: LegacyMerchandiseItemResponse[] };

// ─── Transform ────────────────────────────────────────────

function transformMerchandiseItem(item: LegacyMerchandiseItemResponse) {
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

/**
 * 物販品作成リクエスト。送信側（buildMerchandiseCreateRequest）が送るキーだけを列挙する。
 * @see backend/internal/inventory/merchandise_item_request.go createMerchandiseItemRequest
 */
export interface CreateMerchandiseItemRequest {
  name: string;
  category: ItemCategory;
  unit_price: number;
  tax_type: TaxType;
  tax_rate?: number;
  is_active: boolean;
}

/**
 * 物販品更新リクエスト（PATCH・全フィールド optional）。
 * @see backend/internal/inventory/merchandise_item_request.go updateMerchandiseItemRequest
 */
export interface UpdateMerchandiseItemRequest {
  name?: string;
  category?: ItemCategory;
  unit_price?: number;
  tax_type?: TaxType;
  tax_rate?: number;
  is_active?: boolean;
}

interface ReorderMerchandiseItemsRequest {
  ids: number[];
}

export type { ReorderMerchandiseItemsRequest };

// ─── Queries ──────────────────────────────────────────────

const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItemsPayload>("/v1/masters/merchandise-items");
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
      const { data } = await axios.post<LegacyMerchandiseItemResponse>(
        "/v1/masters/merchandise-items",
        req,
      );
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
      const { data } = await axios.patch<LegacyMerchandiseItemResponse>(
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
