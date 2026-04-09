import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export async function deleteReservationStaff(clinicId: string, id: number): Promise<void> {
  await axios.delete(`/v1/clinics/${clinicId}/reservation-staffs/${id}`);
}

export function useDeleteReservationStaff(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteReservationStaff(clinicId!, id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-staffs", clinicId] });
    },
  });
}
