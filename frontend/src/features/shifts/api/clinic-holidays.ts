import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { ClinicHoliday } from "@/types/generated/models";

export type { ClinicHoliday };

export interface SetClinicHolidayInput {
  date: string;   // YYYY-MM-DD
  reason?: string;
}

export { useGetClinicHolidays } from "@/hooks/use-clinic-holidays";

// POST /v1/clinic-holidays
export function useCreateClinicHoliday() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: SetClinicHolidayInput) =>
      axios.post<ClinicHoliday>("/v1/clinic-holidays", { date: input.date, reason: input.reason ?? "" }).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clinicHolidays.all() });
    },
    onError: (error) => handleApiError(error, "休診日の設定"),
  });
}

// DELETE /v1/clinic-holidays/:date
export function useDeleteClinicHoliday() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (date: string) =>
      axios.delete(`/v1/clinic-holidays/${date}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clinicHolidays.all() });
    },
    onError: (error) => handleApiError(error, "休診日の解除"),
  });
}
