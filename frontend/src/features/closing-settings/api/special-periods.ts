import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { ClosingSpecialPeriod } from "@/types/generated/models";

interface CreateSpecialPeriodRequest {
  start_date: string;
  end_date: string;
  am_pm_boundary: string;
  pm_end: string;
  note?: string;
}

const createSpecialPeriod = async (
  data: CreateSpecialPeriodRequest,
): Promise<ClosingSpecialPeriod> => {
  const { data: res } = await axios.post<ClosingSpecialPeriod>(
    "/v1/closing-settings/special-periods",
    data,
  );
  return res;
};

const deleteSpecialPeriod = async (id: number): Promise<void> => {
  await axios.delete(`/v1/closing-settings/special-periods/${id}`);
};

export const useCreateSpecialPeriod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createSpecialPeriod,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
    onError: (error) => handleApiError(error, "特別期間の追加"),
  });
};

export const useDeleteSpecialPeriod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteSpecialPeriod(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["closing-settings"] }),
    onError: (error) => handleApiError(error, "特別期間の削除"),
  });
};
