import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

interface UpdateOwnerLstepOptOutBody {
  opt_out: boolean;
  reason?: string;
}

async function updateOwnerLstepOptOut(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerLstepOptOutBody
): Promise<void> {
  await axios.patch(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/opt-out`,
    body
  );
}

export function useUpdateOwnerLstepOptOut(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useMutation({
    mutationFn: (body: UpdateOwnerLstepOptOutBody) =>
      updateOwnerLstepOptOut(clinicId, ownerId, body),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success(variables.opt_out ? "配信を停止しました" : "配信を再開しました");
    },
    onError: (error) => {
      handleApiError(error, "配信設定変更");
    },
  });
}
