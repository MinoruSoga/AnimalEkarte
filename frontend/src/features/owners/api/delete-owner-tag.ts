import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

async function deleteOwnerTag(
  clinicId: string,
  ownerId: string,
  tagName: string
): Promise<void> {
  await axios.delete(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags/${encodeURIComponent(tagName)}`
  );
}

export function useDeleteOwnerTag(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? null;

  return useMutation({
    mutationFn: (tagName: string) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return deleteOwnerTag(clinicId, ownerId, tagName);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success("タグを解除しました");
    },
    onError: (error) => {
      handleApiError(error, "タグ解除");
    },
  });
}
