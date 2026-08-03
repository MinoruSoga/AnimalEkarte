import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Staff as ModelStaff } from "@/types/generated/models";

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
 * Returns minimal staff data needed for selection UIs.
 * Uses a distinct query key from master CRUD staff list (full Staff shape).
 */
export function useGetStaffs() {
  return useQuery({
    queryKey: queryKeys.masters.staffSelectorList(),
    queryFn: async (): Promise<StaffItem[]> => {
      const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
      return data.map(transformStaffSelectorItem);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
