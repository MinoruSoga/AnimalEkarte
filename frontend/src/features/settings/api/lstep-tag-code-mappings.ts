import { axios } from "@/lib/axios";

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
// API functions
// ─────────────────────────────────────────────────

export async function fetchTagCodeMappings(clinicId: string): Promise<TagCodeMappingItem[]> {
  const { data } = await axios.get<TagCodeMappingItem[]>(
    `/v1/clinics/${clinicId}/lstep-tag-code-mappings`,
  );
  return data;
}

export async function putTagCodeMappingsForTag(
  clinicId: string,
  params: {
    tagName: string;
    req: PutTagCodeMappingsRequest;
  },
): Promise<TagCodeMappingItem[]> {
  const { data } = await axios.put<TagCodeMappingItem[]>(
    `/v1/clinics/${clinicId}/lstep-tag-code-mappings/${encodeURIComponent(params.tagName)}`,
    params.req,
  );
  return data;
}
