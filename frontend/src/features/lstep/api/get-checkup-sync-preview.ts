import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export type CheckupType =
  | "annual"
  | "dental"
  | "blood"
  | "skin"
  | "cancer"
  | "other";

export interface CheckupSyncParams {
  checkup_type: CheckupType;
  species?: string;
  last_visit_before?: string;
  last_visit_after?: string;
}

export interface CheckupSyncPreviewOwner {
  owner_id: string;
  owner_name: string;
  pet_names: string[];
  last_visit_date: string | null;
  has_line: boolean;
  is_opt_out: boolean;
  has_living_pet: boolean;
  exclusion_reason: string | null;
  current_tags: string[];
}

export interface CheckupSyncPreviewResponse {
  owners: CheckupSyncPreviewOwner[];
  eligible_count: number;
  line_linked_count: number;
  opt_out_count: number;
  no_living_pet_count: number;
  total_count: number;
}

// GET /v1/clinics/:clinicId/lstep/checkup-sync/preview
export function useGetCheckupSyncPreview(params: CheckupSyncParams | null) {
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useQuery({
    queryKey: ["checkup-sync-preview", params],
    queryFn: async () => {
      const { data } = await axios.get<CheckupSyncPreviewResponse>(
        `/v1/clinics/${clinicId}/lstep/checkup-sync/preview`,
        { params: params ?? undefined }
      );
      return data;
    },
    enabled: params !== null,
    staleTime: 0,
  });
}
