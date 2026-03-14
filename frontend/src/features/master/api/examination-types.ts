import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ExaminationType, ExaminationTypeItem } from "@/types/generated/models";

export type { ExaminationType, ExaminationTypeItem };

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
      queryClient.invalidateQueries({ queryKey: ["masterItems", "examination"] });
    },
  });
};
