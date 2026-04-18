import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Staff as ModelStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Type
// ─────────────────────────────────────────────────

export interface StaffItem {
  id: string;
  name: string;
  isActive: boolean;
  occupationName: string | null;
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
    queryKey: ["masters", "staffs"] as const,
    queryFn: async (): Promise<StaffItem[]> => {
      const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
      return data.map(transformStaff);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
