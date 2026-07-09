import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { getClinicId } from "./get-clinic-id";
import { fetchTriggerPriorities, patchTriggerPriorities } from "../api/trigger-priorities";
import type { UpdateTriggerPrioritiesRequest } from "../api/trigger-priorities";

export type {
  TriggerPriorityItem,
  TriggerPriorityListResponse,
  UpdateTriggerPrioritiesRequest,
} from "../api/trigger-priorities";

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const TRIGGER_PRIORITY_QUERY_KEY = (clinicId: string | null) =>
  ["trigger-priorities", clinicId] as const;

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetTriggerPriorities() {
  const clinicId = getClinicId();
  return useQuery({
    queryKey: TRIGGER_PRIORITY_QUERY_KEY(clinicId),
    queryFn: () => {
      // refetch時も都度最新値を読む（挙動保存。useLstepSettings.ts の同種修正を参照）。
      const currentClinicId = getClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return fetchTriggerPriorities(currentClinicId);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function useUpdateTriggerPriorities() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: (req: UpdateTriggerPrioritiesRequest) => {
      const currentClinicId = getClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return patchTriggerPriorities(currentClinicId, req);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: TRIGGER_PRIORITY_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "配信優先順位の更新"),
  });
}
