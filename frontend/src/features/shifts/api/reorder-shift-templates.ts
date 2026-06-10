import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export async function reorderShiftTemplates(ids: number[]): Promise<void> {
  await axios.patch("/v1/shift-templates/reorder", { ids });
}

export function useReorderShiftTemplates() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (ids: number[]) => reorderShiftTemplates(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shift-templates"] });
    },
    onError: (error) => handleApiError(error, "シフトテンプレートの並べ替え"),
  });
}
