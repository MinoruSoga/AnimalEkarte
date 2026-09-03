import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";

// ─────────────────────────────────────────────────
// Backend / Request types
// handler campaignResponse 形式（model.Campaign とは形が異なるため手動定義）
// ─────────────────────────────────────────────────

export type CampaignDiscountType = "rate" | "amount";

interface BackendCampaign {
  id: number;
  clinic_id: number;
  name: string;
  start_date: string; // YYYY-MM-DD
  end_date: string; // YYYY-MM-DD
  discount_type: CampaignDiscountType;
  discount_value: number;
  is_active: boolean;
  sort_order: number;
  target_categories: string[];
  target_item_ids: number[];
}

export interface CreateCampaignRequest {
  name: string;
  start_date: string;
  end_date: string;
  discount_type: CampaignDiscountType;
  discount_value: number;
  is_active?: boolean;
  sort_order?: number;
  target_categories?: string[];
  target_item_ids?: number[];
}

export type UpdateCampaignRequest = Partial<CreateCampaignRequest>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformCampaign(data: BackendCampaign) {
  return {
    id: String(data.id),
    name: data.name,
    startDate: data.start_date,
    endDate: data.end_date,
    discountType: data.discount_type,
    discountValue: data.discount_value,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    targetCategories: data.target_categories ?? [],
    targetItemIds: (data.target_item_ids ?? []).map(String),
  };
}

export type Campaign = ReturnType<typeof transformCampaign>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getCampaigns = async (): Promise<Campaign[]> => {
  const { data } = await axios.get<BackendCampaign[]>("/v1/masters/campaigns");
  return data.map(transformCampaign);
};

const createCampaign = async (req: CreateCampaignRequest): Promise<Campaign> => {
  const { data } = await axios.post<BackendCampaign>("/v1/masters/campaigns", req);
  return transformCampaign(data);
};

const updateCampaign = async (id: string, req: UpdateCampaignRequest): Promise<Campaign> => {
  const { data } = await axios.patch<BackendCampaign>(`/v1/masters/campaigns/${id}`, req);
  return transformCampaign(data);
};

const deleteCampaign = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/campaigns/${id}`);
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetCampaigns = () =>
  useQuery({
    queryKey: queryKeys.masters.category("campaigns"),
    queryFn: getCampaigns,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateCampaign = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createCampaign,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("campaigns") }),
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateCampaign = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateCampaignRequest }) =>
      updateCampaign(id, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("campaigns") }),
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteCampaign = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteCampaign,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("campaigns") }),
    onError: (error) => handleApiError(error, "削除"),
  });
};
