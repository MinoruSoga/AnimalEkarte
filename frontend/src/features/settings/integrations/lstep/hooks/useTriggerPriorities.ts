import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { getClinicId } from "./get-clinic-id";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface TriggerPriorityItem {
  trigger_type: string;
  priority: number;
}

export interface TriggerPriorityListResponse {
  clinic_id: string;
  items: TriggerPriorityItem[];
}

export interface UpdateTriggerPrioritiesRequest {
  items: TriggerPriorityItem[];
}

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const TRIGGER_PRIORITY_QUERY_KEY = (clinicId: string) =>
  ["trigger-priorities", clinicId] as const;

// ─────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────

async function fetchTriggerPriorities(): Promise<TriggerPriorityListResponse> {
  const clinicId = getClinicId();
  const { data } = await axios.get<TriggerPriorityListResponse>(
    `/v1/clinics/${clinicId}/lstep/trigger-priorities`,
  );
  return data;
}

async function patchTriggerPriorities(
  req: UpdateTriggerPrioritiesRequest,
): Promise<void> {
  const clinicId = getClinicId();
  await axios.patch(`/v1/clinics/${clinicId}/lstep/trigger-priorities`, req);
}

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetTriggerPriorities() {
  const clinicId = getClinicId();
  return useQuery({
    queryKey: TRIGGER_PRIORITY_QUERY_KEY(clinicId),
    queryFn: fetchTriggerPriorities,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function useUpdateTriggerPriorities() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: patchTriggerPriorities,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: TRIGGER_PRIORITY_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "配信優先順位の更新"),
  });
}
