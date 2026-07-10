import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export interface ClosingHoliday {
  id: number;
  clinic_id: number;
  date: string;
  reason: string;
  created_at: string;
  updated_at: string;
}

interface CreateHolidayRequest {
  date: string;
  reason?: string;
}

const HOLIDAYS_QUERY_KEY = ["closing-settings", "holidays"] as const;

const getHolidays = async (): Promise<ClosingHoliday[]> => {
  const { data } = await axios.get<ClosingHoliday[]>("/v1/closing-settings/holidays");
  return data;
};

const createHoliday = async (req: CreateHolidayRequest): Promise<ClosingHoliday> => {
  const { data } = await axios.post<ClosingHoliday>("/v1/closing-settings/holidays", req);
  return data;
};

const deleteHoliday = async (date: string): Promise<void> => {
  await axios.delete(`/v1/closing-settings/holidays/${date}`);
};

export const useGetHolidays = () =>
  useQuery({
    queryKey: HOLIDAYS_QUERY_KEY,
    queryFn: getHolidays,
  });

export const useCreateHoliday = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createHoliday,
    onSuccess: () => qc.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY }),
    onError: (error) => handleApiError(error, "休診日の追加"),
  });
};

export const useDeleteHoliday = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (date: string) => deleteHoliday(date),
    onSuccess: () => qc.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY }),
    onError: (error) => handleApiError(error, "休診日の削除"),
  });
};
