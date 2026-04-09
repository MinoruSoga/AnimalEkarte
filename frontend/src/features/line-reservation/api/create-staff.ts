import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Staff, CreateStaffRequest } from "./types";

export async function createReservationStaff(
  clinicId: string,
  payload: CreateStaffRequest
): Promise<Staff> {
  const { data } = await axios.post<Staff>(
    `/v1/clinics/${clinicId}/reservation-staffs`,
    payload
  );
  return data;
}

export function useCreateReservationStaff(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateStaffRequest) => createReservationStaff(clinicId!, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-staffs", clinicId] });
    },
  });
}
