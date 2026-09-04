import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

export interface OwnerLineStatus {
  line_user_id: string | null;
  is_linked: boolean;
  lstep_opt_out: boolean;
  tags: string[];
  fetched_at: string;
}

async function getOwnerLineTags(clinicId: string, ownerId: string): Promise<OwnerLineStatus> {
  const { data } = await axios.get<OwnerLineStatus>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags`,
  );
  return data;
}

/**
 * Shared hook for fetching owner LINE/LSTEP tag status.
 * Uses the same query key as features/owners to share React Query cache.
 */
export function useGetOwnerLineTags(ownerId: string) {
  const clinicId = getStoredClinicId();
  return useQuery({
    queryKey: queryKeys.ownerLineTags(ownerId),
    queryFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return getOwnerLineTags(clinicId, ownerId);
    },
    enabled: !!ownerId && clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
