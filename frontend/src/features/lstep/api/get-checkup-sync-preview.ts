import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export type CheckupType =
  | "annual"
  | "dental"
  | "blood"
  | "skin"
  | "cancer"
  | "other";

// ISSUE-009: CPM ステージ値域は backend service.CPMStage* と一致させる。
export type CPMStage =
  | "cpm_encounter"
  | "cpm_growing"
  | "cpm_core"
  | "cpm_spot"
  | "cpm_noah"
  | "cpm_dormant";

export interface CheckupSyncParams {
  checkup_type: CheckupType;
  species?: string;
  last_visit_before?: string;
  last_visit_after?: string;

  // ISSUE-009: 追加フィルタ（すべてオプショナル）
  min_age_years?: number;
  max_age_years?: number;
  has_chronic_condition?: boolean;
  cpm_stage?: CPMStage;
  min_total_amount?: number;
  min_annual_visit_count?: number;
  last_checkup_before?: string;
  last_checkup_after?: string;
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

  // ISSUE-009: 追加表示フィールド（additive）
  min_pet_age_years: number | null;
  max_pet_age_years: number | null;
  has_chronic_condition: boolean;
  cpm_stage: CPMStage | "";
  total_amount: number;
  annual_visit_count: number;
  last_checkup_date: string | null;
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
  const clinicId = localStorage.getItem("auth_current_clinic:v1");
  if (!clinicId) {
    throw new Error("クリニックが選択されていません。ページをリロードしてください。");
  }

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
