import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export const deleteExamination = async (id: string): Promise<void> => {
  await axios.delete(`/v1/examinations/${id}`);
};

export const useDeleteExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteExamination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
  });
};
