import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

export interface OwnerLineStatus {
  line_user_id: string | null;
  is_linked: boolean;
  lstep_opt_out: boolean;
  tags: string[];
  fetched_at: string;
}

async function getOwnerLineTags(
  clinicId: string,
  ownerId: string
): Promise<OwnerLineStatus> {
  const { data } = await axios.get<OwnerLineStatus>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags`
  );
  return data;
}

export function useGetOwnerLineTags(ownerId: string) {
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";
  return useQuery({
    queryKey: ["owner-line-tags", ownerId],
    queryFn: () => getOwnerLineTags(clinicId, ownerId),
    enabled: !!ownerId && !!clinicId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
