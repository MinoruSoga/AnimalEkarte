import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { CheckupType } from "./get-checkup-sync-preview";

export interface CheckupSyncRequest {
  checkup_type: CheckupType;
  owner_ids: string[];
  tag_name: string;
}

export interface CheckupSyncResult {
  success_count: number;
  failed_count: number;
  skip_count: number;
  failed_owner_ids: string[];
  skipped_owner_ids: string[];
}

async function createCheckupSync(
  clinicId: string,
  body: CheckupSyncRequest
): Promise<CheckupSyncResult> {
  const { data } = await axios.post<CheckupSyncResult>(
    `/v1/clinics/${clinicId}/lstep/checkup-sync`,
    body
  );
  return data;
}

// POST /v1/clinics/:clinicId/lstep/checkup-sync
export function useCreateCheckupSync() {
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useMutation({
    mutationFn: (body: CheckupSyncRequest) =>
      createCheckupSync(clinicId, body),
    onSuccess: () => {
      toast.success("Lステップタグの一括付与が完了しました");
    },
    onError: (error) => {
      handleApiError(error, "Lステップタグ一括付与");
    },
  });
}
