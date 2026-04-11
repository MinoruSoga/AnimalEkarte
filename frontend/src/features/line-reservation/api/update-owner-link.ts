import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { ReservationCustomer } from "./types";

export async function updateOwnerLink(
  clinicId: string,
  customerId: number,
  ownerID: number | null,
): Promise<ReservationCustomer> {
  const { data } = await axios.patch<ReservationCustomer>(
    `/v1/clinics/${clinicId}/line-customers/${customerId}/link-owner`,
    { owner_id: ownerID },
  );
  return data;
}

export function useUpdateOwnerLink(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      customerId,
      ownerID,
    }: {
      customerId: number;
      ownerID: number | null;
    }) => updateOwnerLink(clinicId!, customerId, ownerID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["line-customers", clinicId] });
    },
    onError: (error) => {
      handleApiError(error, "オーナー紐付け");
    },
  });
}
