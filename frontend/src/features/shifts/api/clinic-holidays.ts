import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ClinicHoliday } from "@/types/generated/models";

export type { ClinicHoliday };

export interface SetClinicHolidayInput {
  date: string;   // YYYY-MM-DD
  reason?: string;
}

const clinicHolidayKeys = {
  all: ["clinic-holidays"] as const,
  byMonth: (yearMonth: string) => ["clinic-holidays", yearMonth] as const,
};

// GET /v1/clinic-holidays?year_month=YYYY-MM
export const getClinicHolidays = async (yearMonth?: string): Promise<ClinicHoliday[]> => {
  const params: Record<string, string> = {};
  if (yearMonth) params.year_month = yearMonth;
  const { data } = await axios.get<ClinicHoliday[]>("/v1/clinic-holidays", { params });
  return data;
};

export function useGetClinicHolidays(yearMonth: string) {
  return useQuery({
    queryKey: clinicHolidayKeys.byMonth(yearMonth),
    queryFn: () => getClinicHolidays(yearMonth),
    enabled: !!yearMonth,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

// POST /v1/clinic-holidays
export function useCreateClinicHoliday() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: SetClinicHolidayInput) =>
      axios.post<ClinicHoliday>("/v1/clinic-holidays", { date: input.date, reason: input.reason ?? "" }).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: clinicHolidayKeys.all });
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
      queryClient.invalidateQueries({ queryKey: clinicHolidayKeys.all });
    },
    onError: (error) => handleApiError(error, "休診日の解除"),
  });
}
