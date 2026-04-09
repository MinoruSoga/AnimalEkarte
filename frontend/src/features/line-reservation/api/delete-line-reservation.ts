import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export async function deleteLineReservation(clinicId: string, id: number): Promise<void> {
  await axios.delete(`/v1/clinics/${clinicId}/reservations/${id}`);
}

export function useDeleteLineReservation(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteLineReservation(clinicId!, id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["line-reservations", clinicId] });
    },
  });
}
