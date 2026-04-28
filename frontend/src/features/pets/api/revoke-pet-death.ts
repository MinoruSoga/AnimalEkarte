import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export function useRevokePetDeath() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (petId: string) => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      await axios.delete(`/v1/clinics/${clinicId}/pets/${petId}/death`);
    },
    onSuccess: (_, petId) => {
      queryClient.invalidateQueries({ queryKey: ["pet", petId] });
      queryClient.invalidateQueries({ queryKey: ["pets"] });
      toast.success("死亡記録を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "死亡記録解除");
    },
  });
}
