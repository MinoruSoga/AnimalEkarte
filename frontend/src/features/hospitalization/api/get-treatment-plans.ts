import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";

/**
 * GET /v1/hospitalizations/:id/treatment-plans
 * Real wire for hospitalization treatment-plan rows (not embedded on detail).
 */
const getTreatmentPlans = async (
  hospitalizationId: string
): Promise<TreatmentPlanResponse[]> => {
  const { data } = await axios.get<TreatmentPlanResponse[]>(
    `/v1/hospitalizations/${hospitalizationId}/treatment-plans`
  );
  return data;
};

export const useGetTreatmentPlans = (hospitalizationId: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.hospitalizations.treatmentPlans(hospitalizationId ?? ""),
    queryFn: () => getTreatmentPlans(hospitalizationId!),
    enabled: !!hospitalizationId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export type { TreatmentPlanResponse };
