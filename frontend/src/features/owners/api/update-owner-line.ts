import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

interface UpdateOwnerLineBody {
  line_user_id: string;
}

async function updateOwnerLine(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerLineBody
): Promise<void> {
  await axios.patch(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line-user-id`,
    body
  );
}

async function unlinkOwnerLine(
  clinicId: string,
  ownerId: string
): Promise<void> {
  await axios.patch(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line-user-id`,
    { line_user_id: null }
  );
}

export function useUpdateOwnerLine(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? null;

  return useMutation({
    mutationFn: (body: UpdateOwnerLineBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return updateOwnerLine(clinicId, ownerId, body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success("LINE IDを設定しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE ID設定");
    },
  });
}

export function useDeleteOwnerLine(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? null;

  return useMutation({
    mutationFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return unlinkOwnerLine(clinicId, ownerId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success("LINE連携を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE連携解除");
    },
  });
}
