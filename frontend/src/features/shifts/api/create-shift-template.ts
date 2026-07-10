import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { CreateShiftTemplateInput, ShiftTemplate } from "../types";

async function createShiftTemplate(input: CreateShiftTemplateInput): Promise<ShiftTemplate> {
  const { data } = await axios.post<ShiftTemplate>("/v1/shift-templates", input);
  return data;
}

export function useCreateShiftTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createShiftTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shift-templates"] });
    },
    onError: (error) => handleApiError(error, "シフトテンプレートの作成"),
  });
}
