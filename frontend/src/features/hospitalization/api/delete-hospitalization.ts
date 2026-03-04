import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export const deleteHospitalization = async (id: string): Promise<void> => {
  await axios.delete(`/v1/hospitalizations/${id}`);
};

export const useDeleteHospitalization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteHospitalization,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["hospitalizations"] });
    },
  });
};
