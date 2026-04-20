import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ClosingHoliday } from "@/types/generated/models";

export interface CreateHolidayRequest {
  date: string;
  note?: string;
}

export const createHoliday = async (data: CreateHolidayRequest): Promise<ClosingHoliday> => {
  const { data: res } = await axios.post<ClosingHoliday>(
    "/v1/closing-settings/holidays",
    data,
  );
  return res;
};

export const deleteHoliday = async (id: number): Promise<void> => {
  await axios.delete(`/v1/closing-settings/holidays/${id}`);
};

export const useCreateHoliday = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createHoliday,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
  });
};

export const useDeleteHoliday = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteHoliday(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
  });
};
