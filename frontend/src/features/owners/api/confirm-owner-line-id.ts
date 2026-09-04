import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

async function confirmOwnerLineId(clinicId: string, ownerId: string): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line-id-confirm`,
  );
  return transformOwner(data);
}

export function useConfirmOwnerLineId(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return confirmOwnerLineId(clinicId, ownerId);
    },
    onSuccess: (owner) => {
      queryClient.setQueryData(queryKeys.owners.detail(ownerId), owner);
      queryClient.invalidateQueries({ queryKey: queryKeys.owners.detail(ownerId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success("LINE ID確認を記録しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE ID確認");
    },
  });
}
