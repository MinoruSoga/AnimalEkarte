import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export async function deleteStaffSchedule(
  clinicId: string,
  staffId: number,
  date: string
): Promise<void> {
  await axios.delete(
    `/v1/clinics/${clinicId}/reservation-staffs/${staffId}/schedules/${date}`
  );
}

export function useDeleteStaffSchedule(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ staffId, date }: { staffId: number; date: string }) =>
      deleteStaffSchedule(clinicId!, staffId, date),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: ["staff-schedules", clinicId, variables.staffId],
      });
    },
  });
}
