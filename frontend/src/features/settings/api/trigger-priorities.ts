import { axios } from "@/lib/axios";

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
// API functions
// ─────────────────────────────────────────────────

export async function fetchTriggerPriorities(
  clinicId: string,
): Promise<TriggerPriorityListResponse> {
  const { data } = await axios.get<TriggerPriorityListResponse>(
    `/v1/clinics/${clinicId}/lstep/trigger-priorities`,
  );
  return data;
}

export async function patchTriggerPriorities(
  clinicId: string,
  req: UpdateTriggerPrioritiesRequest,
): Promise<void> {
  await axios.patch(`/v1/clinics/${clinicId}/lstep/trigger-priorities`, req);
}
