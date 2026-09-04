import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { UpdateShiftTemplateInput, ShiftTemplate } from "../types";

async function updateShiftTemplate(
  id: string,
  input: UpdateShiftTemplateInput,
): Promise<ShiftTemplate> {
  const { data } = await axios.patch<ShiftTemplate>(`/v1/shift-templates/${id}`, input);
  return data;
}

export function useUpdateShiftTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateShiftTemplateInput }) =>
      updateShiftTemplate(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.shiftTemplates.all() });
    },
    onError: (error) => handleApiError(error, "シフトテンプレートの更新"),
  });
}
