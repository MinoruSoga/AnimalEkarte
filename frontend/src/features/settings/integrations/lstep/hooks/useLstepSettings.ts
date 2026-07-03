import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { getClinicId } from "./get-clinic-id";
import {
  fetchLstepSettings,
  patchLstepSettings,
  postLstepTest,
  deleteLstepSettings,
} from "../api/lstep-settings";
import type { LstepSettingsRequest } from "../api/lstep-settings";

export type { LstepSettingsResponse, LstepSettingsRequest } from "../api/lstep-settings";

// ─────────────────────────────────────────────────
// Query key
// ─────────────────────────────────────────────────

const LSTEP_QUERY_KEY = (clinicId: string | null) =>
  ["lstep-settings", clinicId] as const;

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

export function useGetLstepSettings() {
  const clinicId = getClinicId();
  return useQuery({
    queryKey: LSTEP_QUERY_KEY(clinicId),
    queryFn: () => {
      if (clinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return fetchLstepSettings(clinicId);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!clinicId,
  });
}

export function useUpdateLstepSettings() {
  const queryClient = useQueryClient();
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: (req: LstepSettingsRequest) => {
      const currentClinicId = getClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return patchLstepSettings(currentClinicId, req);
    },
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
    mutationFn: () => {
      const currentClinicId = getClinicId();
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
      const currentClinicId = getClinicId();
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
  const clinicId = getClinicId();
  return useMutation({
    mutationFn: () => {
      const currentClinicId = getClinicId();
      if (currentClinicId === null) {
        throw new Error("clinic_id is not selected");
      }
      return deleteLstepSettings(currentClinicId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: LSTEP_QUERY_KEY(clinicId),
      });
    },
    onError: (error) => handleApiError(error, "Lステップ設定の削除"),
  });
}
