import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { getStoredClinicId } from "@/lib/current-clinic";
import {
  fetchLstepSettings,
  patchLstepSettings,
  postLstepTest,
  deleteLstepSettings,
} from "../api/lstep-settings";
import type { LstepSettingsRequest } from "../api/lstep-settings";

export type { LstepSettingsResponse, LstepSettingsRequest } from "../api/lstep-settings";

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetLstepSettings() {
  const clinicId = getStoredClinicId();
  return useQuery({
    queryKey: queryKeys.lstepSettings(clinicId),
    queryFn: () => {
      // refetch（window focus 等）時も都度最新値を読む — 元の fetchLstepSettings() は
      // 呼び出しごとに getStoredClinicId() していたため、render 時点の clinicId をクロージャで
      // 固定しない（挙動保存）。
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return fetchLstepSettings(currentClinicId);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function useUpdateLstepSettings() {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();
  return useMutation({
    mutationFn: (req: LstepSettingsRequest) => {
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return patchLstepSettings(currentClinicId, req);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.lstepSettings(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "Lステップ設定の更新"),
  });
}

export function useTestLstepConnection() {
  return useMutation({
    mutationFn: () => {
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return postLstepTest(currentClinicId);
    },
    onError: (error) => handleApiError(error, "接続テスト"),
  });
}

// BE has a single test-connection endpoint; this alias keeps the form unchanged.
export function useTestLineMessagingConnection() {
  return useMutation({
    mutationFn: () => {
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return postLstepTest(currentClinicId);
    },
    onError: (error) => handleApiError(error, "LINE Messaging API接続テスト"),
  });
}

export function useDeleteLstepSettings() {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();
  return useMutation({
    mutationFn: () => {
      const currentClinicId = getStoredClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return deleteLstepSettings(currentClinicId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.lstepSettings(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "Lステップ設定の削除"),
  });
}
