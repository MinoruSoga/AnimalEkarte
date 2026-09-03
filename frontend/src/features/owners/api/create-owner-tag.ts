import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

interface CreateOwnerTagBody {
  tag_name: string;
  reason?: string;
}

async function createOwnerTag(
  clinicId: string,
  ownerId: string,
  body: CreateOwnerTagBody,
): Promise<void> {
  await axios.post(`/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags`, body);
}

export function useCreateOwnerTag(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: CreateOwnerTagBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return createOwnerTag(clinicId, ownerId, body);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success("タグを付与しました");
    },
    onError: (error) => {
      handleApiError(error, "タグ付与");
    },
  });
}
