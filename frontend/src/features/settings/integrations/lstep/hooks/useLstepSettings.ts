import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { getClinicId } from "./get-clinic-id";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface LstepSettingsResponse {
  lstep_api_key_masked: string | null;
  lstep_base_url: string | null;
  line_channel_access_token_masked: string | null;
  line_channel_secret_masked: string | null;
  liff_id: string | null;
  line_account_name: string | null;
  is_configured: boolean;
  last_updated_at: string | null;
  is_sync_enabled: boolean;
  sync_enabled_at: string | null;
  cpm_version: string;
  dormant_prevention_180_days: number;
  dormant_prevention_210_days: number;
  dormant_prevention_240_days: number;
  dormant_prevention_365_days: number;
  // CPM V2 来院数閾値 (33ca50b2)
  cpm_v2_coming_threshold: number;
  cpm_v2_good_threshold: number;
  cpm_v2_family_threshold: number;
  cpm_v2_noah_threshold: number;
  // CPM V1 判定閾値 (8ca5181b)
  cpm_v1_dormant_days: number;
  cpm_v1_noah_days: number;
  cpm_v1_noah_annual_visits: number;
  cpm_v1_noah_ltv: number;
  cpm_v1_core_days: number;
  cpm_v1_core_annual_visits: number;
  cpm_v1_core_ltv: number;
  cpm_v1_spot_min_amount: number;
  cpm_v1_spot_inactive_days: number;
  cpm_v1_growing_max_days: number;
  cpm_v1_growing_min_visits: number;
  cpm_v1_growing_max_visits: number;
  cpm_v1_ltv_break_low: number;
  // 健診・予防タグ判定閾値
  health_prevention_lookback_days: number;
  vaccine_deadline_days: number;
}

export interface LstepSettingsRequest {
  lstep_api_key?: string;
  lstep_base_url?: string;
  line_channel_access_token?: string;
  line_channel_secret?: string;
  liff_id?: string;
  line_account_name?: string;
  is_sync_enabled?: boolean;
  cpm_version?: string;
  dormant_prevention_180_days?: number;
  dormant_prevention_210_days?: number;
  dormant_prevention_240_days?: number;
  dormant_prevention_365_days?: number;
  cpm_v2_coming_threshold?: number;
  cpm_v2_good_threshold?: number;
  cpm_v2_family_threshold?: number;
  cpm_v2_noah_threshold?: number;
  cpm_v1_dormant_days?: number;
  cpm_v1_noah_days?: number;
  cpm_v1_noah_annual_visits?: number;
  cpm_v1_noah_ltv?: number;
  cpm_v1_core_days?: number;
  cpm_v1_core_annual_visits?: number;
  cpm_v1_core_ltv?: number;
  cpm_v1_spot_min_amount?: number;
  cpm_v1_spot_inactive_days?: number;
  cpm_v1_growing_max_days?: number;
  cpm_v1_growing_min_visits?: number;
  cpm_v1_growing_max_visits?: number;
  cpm_v1_ltv_break_low?: number;
  health_prevention_lookback_days?: number;
  vaccine_deadline_days?: number;
}

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const LSTEP_QUERY_KEY = (clinicId: string) =>
  ["lstep-settings", clinicId] as const;

// ─────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────

async function fetchLstepSettings(): Promise<LstepSettingsResponse> {
  const clinicId = getClinicId();
  const { data } = await axios.get<LstepSettingsResponse>(
    `/v1/clinics/${clinicId}/lstep-settings`,
  );
  return data;
}

async function patchLstepSettings(
  req: LstepSettingsRequest,
): Promise<LstepSettingsResponse> {
  const clinicId = getClinicId();
  const { data } = await axios.patch<LstepSettingsResponse>(
    `/v1/clinics/${clinicId}/lstep-settings`,
    req,
  );
  return data;
}

async function postLstepTest(): Promise<void> {
  const clinicId = getClinicId();
  await axios.post(
    `/v1/clinics/${clinicId}/lstep-settings/test-connection`,
  );
}

async function deleteLstepSettings(): Promise<void> {
  const clinicId = getClinicId();
  await axios.delete(
    `/v1/clinics/${clinicId}/lstep-settings`,
  );
}

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetLstepSettings() {
  const clinicId = getClinicId();
  return useQuery({
    queryKey: LSTEP_QUERY_KEY(clinicId),
    queryFn: fetchLstepSettings,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function useUpdateLstepSettings() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: patchLstepSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: LSTEP_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "Lステップ設定の更新"),
  });
}

export function useTestLstepConnection() {
  return useMutation({
    mutationFn: postLstepTest,
    onError: (error) => handleApiError(error, "接続テスト"),
  });
}

// BE has a single test-connection endpoint; this alias keeps the form unchanged.
export function useTestLineMessagingConnection() {
  return useMutation({
    mutationFn: postLstepTest,
    onError: (error) => handleApiError(error, "LINE Messaging API接続テスト"),
  });
}

export function useDeleteLstepSettings() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: deleteLstepSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: LSTEP_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "Lステップ設定の削除"),
  });
}
