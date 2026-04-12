import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { ExaminationType, ExamTypeField } from "@/types/generated/models";

export type { ExaminationType, ExamTypeField };

export interface ExamTypeItemInput {
  name: string;
  inspection_value: string;
  normal_value: string;
}

export const replaceExamTypeItems = async (
  examTypeId: string,
  items: ExamTypeItemInput[]
): Promise<ExaminationType> => {
  const { data } = await axios.put<ExaminationType>(
    `/v1/masters/examination-types/${examTypeId}/items`,
    { items }
  );
  return data;
};

interface ReplaceExamTypeItemsVariables {
  examTypeId: string;
  items: ExamTypeItemInput[];
}

export const useReplaceExamTypeItems = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ examTypeId, items }: ReplaceExamTypeItemsVariables) =>
      replaceExamTypeItems(examTypeId, items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "examination-types"] });
    },
    onError: (error) => handleApiError(error, "操作"),
  });
};
