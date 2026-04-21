import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { BodySize, BillingUnit, HospitalizationPlan as ModelHospitalizationPlan, TaxType } from "@/types/generated/models";
import {
  BodySizeSmall,
  BodySizeMedium,
  BodySizeLarge,
  BillingUnitPerDay,
  BillingUnitPerNight,
} from "@/types/generated/models";

// Re-export enum constants for use in components
export {
  BodySizeSmall,
  BodySizeMedium,
  BodySizeLarge,
  BillingUnitPerDay,
  BillingUnitPerNight,
};

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface CreateHospitalizationPlanRequest {
  name: string;
  price?: number;
  description?: string;
  is_active?: boolean;
  body_size?: BodySize;
  billing_unit?: BillingUnit;
  tax_type?: TaxType;
  tax_rate?: number;
}

export interface UpdateHospitalizationPlanRequest {
  name?: string;
  price?: number;
  description?: string;
  is_active?: boolean;
  body_size?: BodySize | null;
  billing_unit?: BillingUnit | null;
  tax_type?: TaxType;
  tax_rate?: number;
}

export const BODY_SIZE_OPTIONS: { value: BodySize; label: string }[] = [
  { value: BodySizeSmall, label: "小型" },
  { value: BodySizeMedium, label: "中型" },
  { value: BodySizeLarge, label: "大型" },
];

export const BODY_SIZE_LABELS: Record<string, string> = {
  [BodySizeSmall]: "小型",
  [BodySizeMedium]: "中型",
  [BodySizeLarge]: "大型",
};

export const BILLING_UNIT_OPTIONS: { value: BillingUnit; label: string }[] = [
  { value: BillingUnitPerDay, label: "1日あたり" },
  { value: BillingUnitPerNight, label: "1泊あたり" },
];

export const BILLING_UNIT_LABELS: Record<string, string> = {
  [BillingUnitPerDay]: "1日あたり",
  [BillingUnitPerNight]: "1泊あたり",
};

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformHospitalizationPlan(
  data: ModelHospitalizationPlan,
) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    price: data.price ?? 0,
    isActive: data.is_active,
    description: data.description ?? "",
    bodySize: data.body_size ?? null,
    billingUnit: data.billing_unit ?? null,
    taxType: (data.tax_type ?? "excluded") as TaxType,
    taxRate: data.tax_rate ?? 0.1,
    sortOrder: data.sort_order,
  };
}

export type HospitalizationPlan = ReturnType<typeof transformHospitalizationPlan>;

// ─────────────────────────────────────────────────
// API Functions
// ─────────────────────────────────────────────────

const ENDPOINT = "/v1/masters/hospitalization-plans";

export const getAllHospitalizationPlans = async (): Promise<
  HospitalizationPlan[]
> => {
  const { data } = await axios.get<ModelHospitalizationPlan[]>(ENDPOINT);
  return data.map(transformHospitalizationPlan);
};

export const createHospitalizationPlan = async (
  req: CreateHospitalizationPlanRequest,
): Promise<HospitalizationPlan> => {
  const { data } = await axios.post<ModelHospitalizationPlan>(ENDPOINT, req);
  return transformHospitalizationPlan(data);
};

export const updateHospitalizationPlan = async (
  id: string,
  req: UpdateHospitalizationPlanRequest,
): Promise<HospitalizationPlan> => {
  const { data } = await axios.patch<ModelHospitalizationPlan>(
    `${ENDPOINT}/${id}`,
    req,
  );
  return transformHospitalizationPlan(data);
};

export const deleteHospitalizationPlan = async (id: string): Promise<void> => {
  await axios.delete(`${ENDPOINT}/${id}`);
};

// ─────────────────────────────────────────────────
// React Query Hooks
// ─────────────────────────────────────────────────

const HOSPITALIZATION_PLANS_QUERY_KEY = ["masters", "hospitalization-plans"] as const;

export const useGetAllHospitalizationPlans = () => {
  return useQuery({
    queryKey: HOSPITALIZATION_PLANS_QUERY_KEY,
    queryFn: getAllHospitalizationPlans,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const useCreateHospitalizationPlan = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createHospitalizationPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: HOSPITALIZATION_PLANS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateHospitalizationPlan = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateHospitalizationPlanRequest;
    }) => updateHospitalizationPlan(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: HOSPITALIZATION_PLANS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteHospitalizationPlan = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteHospitalizationPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: HOSPITALIZATION_PLANS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const reorderHospitalizationPlans = async (ids: number[]): Promise<void> => {
  await axios.patch(`${ENDPOINT}/reorder`, { ids });
};

export const useReorderHospitalizationPlans = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (ids: number[]) => reorderHospitalizationPlans(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: HOSPITALIZATION_PLANS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
