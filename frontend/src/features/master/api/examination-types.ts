import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface ExamTypeItemInput {
  name: string;
  inspection_value: string;
  normal_value: string;
}

export interface ExamTypeItemResponse {
  id: number;
  exam_type_id: number;
  name: string;
  inspection_value: string;
  normal_value: string;
  sort_order: number;
  created_at: string;
}

export interface ExamTypeResponse {
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  sort_order: number;
  items?: ExamTypeItemResponse[];
  created_at: string;
  updated_at: string;
}

export const replaceExamTypeItems = async (
  examTypeId: string,
  items: ExamTypeItemInput[]
): Promise<ExamTypeResponse> => {
  const { data } = await axios.put<ExamTypeResponse>(
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
