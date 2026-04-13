import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export async function deleteShiftTemplate(id: string): Promise<void> {
  await axios.delete(`/v1/shift-templates/${id}`);
}

export function useDeleteShiftTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteShiftTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shift-templates"] });
    },
    onError: (error) => handleApiError(error, "シフトテンプレートの削除"),
  });
}
