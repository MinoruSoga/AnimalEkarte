import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { BodySize, BillingUnit, TaxType } from "@/types/generated/models";
import {
  BodySizeSmall,
  BodySizeMedium,
  BodySizeLarge,
  BillingUnitPerDay,
  BillingUnitPerNight,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// 移行 HTTP DTO（非永続・EMR-527）
// 保存 model（生成 HospitalizationPlan）から Omit / Pick せず、現行 handler の
// response struct と binding struct の列をそのまま写す。
// ─────────────────────────────────────────────────

/**
 * 入院プラン応答。body_size / billing_unit / tax_type は Go では string で直列化されるが、
 * 値域どおり enum union で受ける。
 * @see backend/internal/medicalrecord/hospitalization_plan_response.go hospitalizationPlanResponse
 */
interface LegacyHospitalizationPlanResponse {
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  body_size?: BodySize;
  billing_unit?: BillingUnit;
  tax_type: TaxType;
  tax_rate: number;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

/**
 * 入院プラン作成リクエスト。name のみ必須。送信側が送るキーだけを列挙する。
 * @see backend/internal/medicalrecord/hospitalization_plan_request.go createHospitalizationPlanRequest
 */
export interface CreateHospitalizationPlanRequest {
  name: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  body_size?: string;
  billing_unit?: string;
  tax_type?: TaxType;
  tax_rate?: number;
}

/**
 * 入院プラン更新リクエスト（PATCH）。body_size / billing_unit は null で未設定を表す。
 * @see backend/internal/medicalrecord/hospitalization_plan_request.go updateHospitalizationPlanRequest
 */
export interface UpdateHospitalizationPlanRequest {
  name?: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  body_size?: string | null;
  billing_unit?: string | null;
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

function transformHospitalizationPlan(data: LegacyHospitalizationPlanResponse) {
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

const getAllHospitalizationPlans = async (): Promise<HospitalizationPlan[]> => {
  const { data } = await axios.get<LegacyHospitalizationPlanResponse[]>(ENDPOINT);
  return data.map(transformHospitalizationPlan);
};

const createHospitalizationPlan = async (
  req: CreateHospitalizationPlanRequest,
): Promise<HospitalizationPlan> => {
  const { data } = await axios.post<LegacyHospitalizationPlanResponse>(ENDPOINT, req);
  return transformHospitalizationPlan(data);
};

const updateHospitalizationPlan = async (
  id: string,
  req: UpdateHospitalizationPlanRequest,
): Promise<HospitalizationPlan> => {
  const { data } = await axios.patch<LegacyHospitalizationPlanResponse>(`${ENDPOINT}/${id}`, req);
  return transformHospitalizationPlan(data);
};

const deleteHospitalizationPlan = async (id: string): Promise<void> => {
  await axios.delete(`${ENDPOINT}/${id}`);
};

// ─────────────────────────────────────────────────
// React Query Hooks
// ─────────────────────────────────────────────────

export const useGetAllHospitalizationPlans = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("hospitalization-plans"),
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
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("hospitalization-plans"),
      });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateHospitalizationPlan = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateHospitalizationPlanRequest }) =>
      updateHospitalizationPlan(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("hospitalization-plans"),
      });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteHospitalizationPlan = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteHospitalizationPlan,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.masters.category("hospitalization-plans"),
      });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};
