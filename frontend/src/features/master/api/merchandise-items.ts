import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MerchandiseItem } from "@/types/generated/models";

// ─── Transform ────────────────────────────────────────────

interface FrontendMerchandiseItem {
  id: string;
  name: string;
  category: string;
  unitPrice: number;
  taxRate: number;
  isActive: boolean;
  sortOrder: number;
}

function transformMerchandiseItem(item: MerchandiseItem): FrontendMerchandiseItem {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    taxRate: item.tax_rate,
    isActive: item.is_active,
    sortOrder: item.sort_order,
  };
}

export type { FrontendMerchandiseItem };

// ─── API Request Types ────────────────────────────────────

interface CreateMerchandiseItemRequest {
  name: string;
  category: string;
  unit_price: number;
  tax_rate: number;
  is_active?: boolean;
}

interface UpdateMerchandiseItemRequest {
  name?: string;
  category?: string;
  unit_price?: number;
  tax_rate?: number;
  is_active?: boolean;
}

interface ReorderMerchandiseItemsRequest {
  ids: number[];
}

export type { CreateMerchandiseItemRequest, UpdateMerchandiseItemRequest, ReorderMerchandiseItemsRequest };

// ─── Queries ──────────────────────────────────────────────

const QUERY_KEY = ["masters", "merchandise-items"] as const;

export const getAllMerchandiseItems = async (): Promise<FrontendMerchandiseItem[]> => {
  const { data } = await axios.get<MerchandiseItem[]>("/v1/masters/merchandise-items");
  return data.map(transformMerchandiseItem);
};

export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: [...QUERY_KEY],
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
      queryClient.invalidateQueries({ queryKey: [...QUERY_KEY] });
    },
  });
};

export const useUpdateMerchandiseItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, req }: { id: string; req: UpdateMerchandiseItemRequest }) => {
      const { data } = await axios.patch<MerchandiseItem>(`/v1/masters/merchandise-items/${id}`, req);
      return transformMerchandiseItem(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [...QUERY_KEY] });
    },
  });
};

export const useDeleteMerchandiseItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await axios.delete(`/v1/masters/merchandise-items/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [...QUERY_KEY] });
    },
  });
};

export const useReorderMerchandiseItems = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: ReorderMerchandiseItemsRequest) => {
      await axios.put("/v1/masters/merchandise-items/reorder", req);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [...QUERY_KEY] });
    },
  });
};
