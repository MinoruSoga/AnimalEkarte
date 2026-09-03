import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

interface UpdateOwnerDeliveryExclusionBody {
  excluded: boolean;
  reason?: string | null;
}

async function updateOwnerDeliveryExclusion(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerDeliveryExclusionBody,
): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/delivery-exclusion`,
    {
      excluded: body.excluded,
      reason: body.reason,
    },
  );
  return transformOwner(data);
}

export function useUpdateOwnerDeliveryExclusion(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: UpdateOwnerDeliveryExclusionBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return updateOwnerDeliveryExclusion(clinicId, ownerId, body);
    },
    onSuccess: (owner, variables) => {
      queryClient.setQueryData(queryKeys.owners.detail(ownerId), owner);
      queryClient.invalidateQueries({ queryKey: queryKeys.owners.detail(ownerId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerLineTags(ownerId) });
      toast.success(variables.excluded ? "配信を除外しました" : "配信除外を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "配信除外設定");
    },
  });
}
