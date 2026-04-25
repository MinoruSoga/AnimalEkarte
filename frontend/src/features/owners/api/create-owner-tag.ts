import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

interface CreateOwnerTagBody {
  tag_name: string;
  reason?: string;
}

async function createOwnerTag(
  clinicId: string,
  ownerId: string,
  body: CreateOwnerTagBody
): Promise<void> {
  await axios.post(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/tags`,
    body
  );
}

export function useCreateOwnerTag(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useMutation({
    mutationFn: (body: CreateOwnerTagBody) =>
      createOwnerTag(clinicId, ownerId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success("タグを付与しました");
    },
    onError: (error) => {
      handleApiError(error, "タグ付与");
    },
  });
}
