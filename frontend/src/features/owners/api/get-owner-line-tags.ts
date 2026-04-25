import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

export interface LstepTag {
  tag_name: string;
  category: "auto" | "manual";
  created_at: string;
}

export interface OwnerLineStatus {
  line_user_id: string | null;
  is_linked: boolean;
  tags: LstepTag[];
  lstep_opt_out: boolean;
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
