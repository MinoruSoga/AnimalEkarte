import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationSetting, UpdateLineReservationSettingRequest } from "./types";

export async function updateLineReservationSetting(
  clinicId: string,
  payload: UpdateLineReservationSettingRequest
): Promise<ReservationSetting> {
  const { data } = await axios.put<ReservationSetting>(
    `/v1/clinics/${clinicId}/line-reservation-settings`,
    payload
  );
  return data;
}

export function useUpdateLineReservationSetting(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: UpdateLineReservationSettingRequest) =>
      updateLineReservationSetting(clinicId!, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["reservation-settings", clinicId] });
    },
  });
}
