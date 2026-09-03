import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

interface UpdateOwnerTransferStatusBody {
  is_transferred: boolean;
}

async function updateOwnerTransferStatus(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerTransferStatusBody,
): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/transfer-status`,
    {
      is_transferred: body.is_transferred,
    },
  );
  return transformOwner(data);
}

export function useUpdateOwnerTransferStatus(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: UpdateOwnerTransferStatusBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return updateOwnerTransferStatus(clinicId, ownerId, body);
    },
    onSuccess: (owner, variables) => {
      queryClient.setQueryData(queryKeys.owners.detail(ownerId), owner);
      queryClient.invalidateQueries({ queryKey: queryKeys.owners.detail(ownerId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success(variables.is_transferred ? "転院済みに設定しました" : "転院設定を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "転院状態更新");
    },
  });
}
