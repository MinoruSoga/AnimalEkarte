import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import type { CPMStage } from "@/lib/cpm-stage";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export type CheckupType =
  | "annual"
  | "dental"
  | "blood"
  | "skin"
  | "cancer"
  | "other";

// ISSUE-009/180: CPM ステージ値域は共有定義 @/lib/cpm-stage に集約（単一真実源）。
// ファイル内では上部 import の CPMStage を利用。外部からは @/lib/cpm-stage を直接 import すること。

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
  // clinicId 欠落時は throw せず enabled で無効化する (use-reservation-types.ts と同パターン)。
  // レンダー中の throw は ErrorBoundary 頼みになりページ全体を落とすため禁止。
  const clinicId = getStoredClinicId();

  return useQuery({
    queryKey: queryKeys.checkupSyncPreview(clinicId, params),
    queryFn: async () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }

      // BUG-032: client timeout so permanent pending does not leave the UI on 読み込み中.
      // BE also enforces a 15s context deadline + SQL LIMIT.
      const { data } = await axios.get<CheckupSyncPreviewResponse>(
        `/v1/clinics/${clinicId}/lstep/checkup-sync/preview`,
        { params: params ?? undefined, timeout: 20_000 },
      );
      return data;
    },
    enabled: params !== null && clinicId !== null,
    staleTime: QUERY_STALE_TIMES.NONE,
  });
}
