import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

async function confirmOwnerLineId(clinicId: string, ownerId: string): Promise<Owner> {
  const { data } = await axios.patch<OwnerApiResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line-id-confirm`
  );
  return transformOwner(data);
}

export function useConfirmOwnerLineId(ownerId: string) {
  const queryClient = useQueryClient();
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? null;

  return useMutation({
    mutationFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return confirmOwnerLineId(clinicId, ownerId);
    },
    onSuccess: (owner) => {
      queryClient.setQueryData(["owners", ownerId], owner);
      queryClient.invalidateQueries({ queryKey: ["owners", ownerId] });
      queryClient.invalidateQueries({ queryKey: ["owner-line-tags", ownerId] });
      toast.success("LINE ID確認を記録しました");
    },
    onError: (error) => {
      handleApiError(error, "LINE ID確認");
    },
  });
}
