import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface TagCodeMappingItem {
  id: number;
  clinic_id: number;
  tag_name: string;
  code_type: string;
  codes: string[];
  species_scope?: string;
  age_min?: number;
}

export interface PutMappingEntry {
  code_type: string;
  codes: string[];
  species_scope?: string;
  age_min?: number;
}

export interface PutTagCodeMappingsRequest {
  entries: PutMappingEntry[];
}

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const MAPPINGS_QUERY_KEY = (clinicId: string) =>
  ["lstep-tag-code-mappings", clinicId] as const;

// ─────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────

function getClinicId(): string {
  return localStorage.getItem("auth_current_clinic:v1") ?? "";
}

async function fetchTagCodeMappings(): Promise<TagCodeMappingItem[]> {
  const clinicId = getClinicId();
  const { data } = await axios.get<TagCodeMappingItem[]>(
    `/v1/clinics/${clinicId}/lstep-tag-code-mappings`,
  );
  return data;
}

async function putTagCodeMappingsForTag(params: {
  tagName: string;
  req: PutTagCodeMappingsRequest;
}): Promise<TagCodeMappingItem[]> {
  const clinicId = getClinicId();
  const { data } = await axios.put<TagCodeMappingItem[]>(
    `/v1/clinics/${clinicId}/lstep-tag-code-mappings/${encodeURIComponent(params.tagName)}`,
    params.req,
  );
  return data;
}

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetTagCodeMappings() {
  const clinicId = getClinicId();
  return useQuery({
    queryKey: MAPPINGS_QUERY_KEY(clinicId),
    queryFn: fetchTagCodeMappings,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function usePutTagCodeMappingsForTag() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: putTagCodeMappingsForTag,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: MAPPINGS_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "タグコードマッピングの更新"),
  });
}
