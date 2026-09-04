import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

interface UpdateOwnerLineBody {
  line_user_id: string;
}

async function updateOwnerLine(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerLineBody,
): Promise<void> {
  await axios.patch(`/v1/clinics/${clinicId}/owners/${ownerId}/line-user-id`, body);
}

async function unlinkOwnerLine(clinicId: string, ownerId: string): Promise<void> {
  await axios.patch(`/v1/clinics/${clinicId}/owners/${ownerId}/line-user-id`, {
    line_user_id: null,
  });
}

export function useUpdateOwnerLine(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: UpdateOwnerLineBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return updateOwnerLine(clinicId, ownerId, body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success("LINE IDを設定しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE ID設定");
    },
  });
}

export function useDeleteOwnerLine(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return unlinkOwnerLine(clinicId, ownerId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success("LINE連携を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE連携解除");
    },
  });
}
