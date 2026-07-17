import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

async function deleteShift(id: string): Promise<void> {
  await axios.delete(`/v1/shifts/${id}`);
}

export function useDeleteShift() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteShift,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
    },
    onError: (error) => handleApiError(error, "シフトの削除"),
  });
}
