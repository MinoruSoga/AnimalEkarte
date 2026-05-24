import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

interface UpdateOwnerTransferStatusBody {
  is_transferred: boolean;
}

async function updateOwnerTransferStatus(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerTransferStatusBody
): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(`/v1/clinics/${clinicId}/owners/${ownerId}/transfer-status`, {
    is_transferred: body.is_transferred,
  });
  return transformOwner(data);
}

export function useUpdateOwnerTransferStatus(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";

  return useMutation({
    mutationFn: (body: UpdateOwnerTransferStatusBody) =>
      updateOwnerTransferStatus(clinicId, ownerId, body),
    onSuccess: (owner, variables) => {
      queryClient.setQueryData(["owners", ownerId], owner);
      queryClient.invalidateQueries({ queryKey: ["owners", ownerId] });
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success(variables.is_transferred ? "転院済みに設定しました" : "転院設定を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "転院状態更新");
    },
  });
}
