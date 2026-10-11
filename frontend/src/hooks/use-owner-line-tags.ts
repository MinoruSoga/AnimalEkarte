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
  // Codex PR606 再々指摘: path だけでなく X-Clinic-ID も対象行の医院に揃える
  // （未指定だと interceptor が選択中医院を補完し、backend が header で scope する）。
  const { data } = await axios.get<OwnerLineStatus>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags`,
    { headers: { "X-Clinic-ID": clinicId } },
  );
  return data;
}

/**
 * Shared hook for fetching owner LINE/LSTEP tag status.
 * Uses the same query key as features/owners to share React Query cache.
 */
export function useGetOwnerLineTags(ownerId: string, clinicId?: string) {
  // Codex PR606 再指摘: 拠点横断フォームは行の clinicId で URL path を組み立てる
  const effectiveClinicId = clinicId ?? getStoredClinicId();
  return useQuery({
    queryKey: clinicId
      ? [...queryKeys.ownerLineTags(ownerId), clinicId]
      : queryKeys.ownerLineTags(ownerId),
    queryFn: () => {
      if (effectiveClinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return getOwnerLineTags(effectiveClinicId, ownerId);
    },
    enabled: !!ownerId && effectiveClinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
