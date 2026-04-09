import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export async function deleteCourse(clinicId: string, id: number): Promise<void> {
  await axios.delete(`/v1/clinics/${clinicId}/reservation-courses/${id}`);
}

export function useDeleteCourse(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteCourse(clinicId!, id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-courses", clinicId] });
    },
  });
}
