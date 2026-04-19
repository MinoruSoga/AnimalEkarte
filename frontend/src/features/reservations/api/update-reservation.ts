import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Reservation } from "@/types";
import { transformReservation } from "./transforms";
import type { Appointment as BackendReservation } from "@/types/generated/models";
import type { UpdateReservationRequest } from "./types";

export const updateReservation = async (
  id: string,
  req: UpdateReservationRequest
): Promise<Reservation> => {
  const { data } = await axios.patch<BackendReservation>(
    `/v1/reservations/${id}`,
    req
  );
  return transformReservation(data);
};

export const useUpdateReservation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationRequest }) =>
      updateReservation(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reservations"] });
    },
    onError: (error) => handleApiError(error, "予約更新"),
  });
};
