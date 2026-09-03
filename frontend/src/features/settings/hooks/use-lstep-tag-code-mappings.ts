import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { getStoredClinicId } from "@/lib/current-clinic";
import { fetchTagCodeMappings, putTagCodeMappingsForTag } from "../api/lstep-tag-code-mappings";
import type { PutTagCodeMappingsRequest } from "../api/lstep-tag-code-mappings";

export type {
  TagCodeMappingItem,
  PutMappingEntry,
  PutTagCodeMappingsRequest,
} from "../api/lstep-tag-code-mappings";

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetTagCodeMappings() {
  const clinicId = getStoredClinicId();
  return useQuery({
    queryKey: queryKeys.lstepTagCodeMappings(clinicId),
    queryFn: () => {
      // refetch時も都度最新値を読む（挙動保存。useLstepSettings.ts の同種修正を参照）。
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return fetchTagCodeMappings(currentClinicId);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function usePutTagCodeMappingsForTag() {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();
  return useMutation({
    mutationFn: (params: { tagName: string; req: PutTagCodeMappingsRequest }) => {
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return putTagCodeMappingsForTag(currentClinicId, params);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.lstepTagCodeMappings(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "タグコードマッピングの更新"),
  });
}
