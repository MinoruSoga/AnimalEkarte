import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { StaffSchedule, UpsertScheduleRequest } from "./types";

export async function upsertStaffSchedule(
  clinicId: string,
  staffId: number,
  date: string,
  payload: UpsertScheduleRequest
): Promise<StaffSchedule> {
  const { data } = await axios.put<StaffSchedule>(
    `/v1/clinics/${clinicId}/reservation-staffs/${staffId}/schedules/${date}`,
    payload
  );
  return data;
}

export function useUpsertStaffSchedule(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      staffId,
      date,
      payload,
    }: {
      staffId: number;
      date: string;
      payload: UpsertScheduleRequest;
    }) => upsertStaffSchedule(clinicId!, staffId, date, payload),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: ["staff-schedules", clinicId, variables.staffId],
      });
    },
  });
}
