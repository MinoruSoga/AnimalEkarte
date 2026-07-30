import { axios } from "@/lib/axios";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";

/**
 * POST body for nested hospitalization treatment plans.
 * Matches backend createTreatmentPlanRequest (subtotal is ignored server-side).
 */
export interface CreateTreatmentPlanRequest {
  treatment_content: string;
  memo?: string;
  is_insurance?: boolean;
  unit_price?: number;
  quantity: number;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
}

/**
 * POST /v1/hospitalizations/:id/treatment-plans
 * Permission: hospitalization:create (BE nested route).
 */
export async function createTreatmentPlanForHospitalization(
  hospitalizationId: string,
  body: CreateTreatmentPlanRequest,
): Promise<TreatmentPlanResponse> {
  const { data } = await axios.post<TreatmentPlanResponse>(
    `/v1/hospitalizations/${hospitalizationId}/treatment-plans`,
    body,
  );
  return data;
}
