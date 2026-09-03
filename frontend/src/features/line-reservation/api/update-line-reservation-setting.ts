import { axios } from "@/lib/axios";
import type { ReservationSetting, UpdateLineReservationSettingRequest } from "./types";

export async function updateLineReservationSetting(
  clinicId: string,
  payload: UpdateLineReservationSettingRequest,
): Promise<ReservationSetting> {
  const { data } = await axios.put<ReservationSetting>(
    `/v1/clinics/${clinicId}/line-reservation-settings`,
    payload,
  );
  return data;
}
