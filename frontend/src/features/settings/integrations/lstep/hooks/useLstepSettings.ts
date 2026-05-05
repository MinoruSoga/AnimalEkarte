import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

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
  fire_hour_jst: number;
}

export interface LstepSettingsRequest {
  lstep_api_key?: string;
  lstep_base_url?: string;
  line_channel_access_token?: string;
  line_channel_secret?: string;
  liff_id?: string;
  line_account_name?: string;
  is_sync_enabled?: boolean;
  fire_hour_jst?: number;
}

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const LSTEP_QUERY_KEY = (clinicId: string) =>
  ["lstep-settings", clinicId] as const;

// ─────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────

function getClinicId(): string {
  return localStorage.getItem("auth_current_clinic:v1") ?? "";
}

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
