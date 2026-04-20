import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ClosingSpecialPeriod } from "@/types/generated/models";

export interface CreateSpecialPeriodRequest {
  start_date: string;
  end_date: string;
  am_pm_boundary: string;
  pm_end: string;
  note?: string;
}

export interface UpdateSpecialPeriodRequest {
  start_date?: string;
  end_date?: string;
  am_pm_boundary?: string;
  pm_end?: string;
  note?: string;
}

export const createSpecialPeriod = async (
  data: CreateSpecialPeriodRequest,
): Promise<ClosingSpecialPeriod> => {
  const { data: res } = await axios.post<ClosingSpecialPeriod>(
    "/v1/closing-settings/special-periods",
    data,
  );
  return res;
};

export const updateSpecialPeriod = async (
  id: number,
  data: UpdateSpecialPeriodRequest,
): Promise<ClosingSpecialPeriod> => {
  const { data: res } = await axios.patch<ClosingSpecialPeriod>(
    `/v1/closing-settings/special-periods/${id}`,
    data,
  );
  return res;
};

export const deleteSpecialPeriod = async (id: number): Promise<void> => {
  await axios.delete(`/v1/closing-settings/special-periods/${id}`);
};

export const useCreateSpecialPeriod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createSpecialPeriod,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
  });
};

export const useUpdateSpecialPeriod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateSpecialPeriodRequest }) =>
      updateSpecialPeriod(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
  });
};

export const useDeleteSpecialPeriod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteSpecialPeriod(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
  });
};
