import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingCourseType as ModelTrimmingCourseType } from "@/types/generated/models";

export function transformTrimmingCourseType(data: ModelTrimmingCourseType) {
  return {
    id: String(data.id),
    name: data.name,
    isActive: data.is_active,
    sortOrder: data.sort_order,
  };
}

export type TrimmingCourseType = ReturnType<typeof transformTrimmingCourseType>;

async function getTrimmingCourseTypes(): Promise<TrimmingCourseType[]> {
  const { data } = await axios.get<ModelTrimmingCourseType[]>("/v1/masters/trimming-course-types");
  return data.map(transformTrimmingCourseType);
}

/**
 * Shared hook for fetching trimming course types.
 * Uses the same query key as features/master to share React Query cache.
 */
export function useGetTrimmingCourseTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("trimming-course-types"),
    queryFn: getTrimmingCourseTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
