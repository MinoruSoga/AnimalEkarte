import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Appointment } from "@/types";
import { transformReservation } from "./transforms";
import type { Appointment as BackendReservation } from "@/types/generated/models";
import type { CreateReservationRequest } from "./types";

export const createReservation = async (
  req: CreateReservationRequest
): Promise<Appointment> => {
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
  });
};
