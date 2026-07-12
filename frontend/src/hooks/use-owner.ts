import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformOwner, type OwnerApiResponse } from "@/lib/transforms/owner";

/**
 * Shared hook for fetching a single owner by ID.
 * Uses the same query key as features/owners to share React Query cache.
 */
export function useGetOwner(id: string) {
  return useQuery({
    queryKey: queryKeys.owners.detail(id),
    queryFn: async () => {
      const { data } = await axios.get<OwnerApiResponse>(`/v1/owners/${id}`);
      return transformOwner(data);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!id,
  });
}
