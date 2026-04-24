import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformReservation } from "./transforms";
import type { Reservation } from "./transforms";
import type { Reservation as BackendReservation } from "@/types/generated/models";
import type { CreateReservationRequest } from "./types";

export const createReservation = async (
  req: CreateReservationRequest
): Promise<Reservation> => {
  const { data } = await axios.post<BackendReservation>(
    "/v1/reservations",
    req
  );
  return transformReservation(data);
};

export const useCreateReservation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createReservation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reservations"] });
    },
    onError: (error) => handleApiError(error, "予約作成"),
  });
};
