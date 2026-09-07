import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Staff as ModelStaff } from "@/types/generated/models";

/** Raw `/v1/masters/staffs` cache. Selector and master CRUD share this key. */
export const STAFFS_RAW_QUERY_KEY = queryKeys.masters.category("staffs");

export async function fetchStaffsRaw(): Promise<ModelStaff[]> {
  const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
  return data;
}

// ─────────────────────────────────────────────────
// Type
// ─────────────────────────────────────────────────

export interface StaffItem {
  id: string;
  name: string;
  isActive: boolean;
  occupationName: string | null;
  /**
   * API staff_type as-is. Missing/unknown values stay empty — never fail-open to "doctor".
   * Doctor-selectable UIs must require staffType === "doctor" && isActive.
   */
  staffType: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

/** Exported for unit tests — untrusted API staff_type must not default to doctor. */
export function transformStaffSelectorItem(data: ModelStaff): StaffItem {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    isActive: data.is_active ?? true,
    occupationName: data.occupation?.name ?? null,
    staffType: typeof data.staff_type === "string" ? data.staff_type : "",
  };
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

/**
 * Read-only staff list hook for cross-feature consumption.
 * Shares the raw `/v1/masters/staffs` cache with master CRUD; select keeps the thin shape.
 */
export function useGetStaffs() {
  return useQuery({
    queryKey: STAFFS_RAW_QUERY_KEY,
    queryFn: fetchStaffsRaw,
    select: (rows) => rows.map(transformStaffSelectorItem),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
