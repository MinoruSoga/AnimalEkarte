import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

const deleteReservation = async (id: string): Promise<void> => {
  await axios.delete(`/v1/reservations/${id}`);
};

export const useDeleteReservation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteReservation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reservations"] });
    },
    onError: (error) => handleApiError(error, "予約削除"),
  });
};
