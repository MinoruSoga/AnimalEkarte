import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

interface UpdateOwnerDeliveryCautionBody {
  caution: boolean;
  reason?: string | null;
}

async function updateOwnerDeliveryCaution(
  clinicId: string,
  ownerId: string,
  body: UpdateOwnerDeliveryCautionBody
): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(`/v1/clinics/${clinicId}/owners/${ownerId}/delivery-caution`, {
    caution: body.caution,
    reason: body.reason,
  });
  return transformOwner(data);
}

export function useUpdateOwnerDeliveryCaution(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = getStoredClinicId();

  return useMutation({
    mutationFn: (body: UpdateOwnerDeliveryCautionBody) => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return updateOwnerDeliveryCaution(clinicId, ownerId, body);
    },
    onSuccess: (owner, variables) => {
      queryClient.setQueryData(["owners", ownerId], owner);
      queryClient.invalidateQueries({ queryKey: ["owners", ownerId] });
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success(variables.caution ? "配信注意を設定しました" : "配信注意を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "配信注意設定");
    },
  });
}
