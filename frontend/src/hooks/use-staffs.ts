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
   * Shared React Query key with `@/features/master` staff list.
   * Must include staffType so examination doctor filters are not empty-filtered
   * when this thinner transform populates the cache first.
   */
  staffType: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformStaff(data: ModelStaff): StaffItem {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    isActive: data.is_active ?? true,
    occupationName: data.occupation?.name ?? null,
    staffType: data.staff_type ?? "doctor",
  };
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

/**
 * Read-only staff list hook for cross-feature consumption.
 * Returns minimal staff data needed for selection UIs.
 * Query key matches features/master/api/staffs.ts to share the React Query cache.
 */
export function useGetStaffs() {
  return useQuery({
    queryKey: queryKeys.masters.category("staffs"),
    queryFn: async (): Promise<StaffItem[]> => {
      const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
      return data.map(transformStaff);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
