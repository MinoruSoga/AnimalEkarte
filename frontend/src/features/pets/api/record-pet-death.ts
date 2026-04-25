import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

interface RecordPetDeathVariables {
  petId: string;
  deceasedAt: string;
  deceasedReason?: string;
}

export function useRecordPetDeath() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ petId, deceasedAt, deceasedReason }: RecordPetDeathVariables) => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.patch(
        `/v1/clinics/${clinicId}/pets/${petId}`,
        {
          deceased_at: deceasedAt,
          deceased_reason: deceasedReason,
        },
      );
      return data;
    },
    onSuccess: (_, { petId }) => {
      queryClient.invalidateQueries({ queryKey: ["pet", petId] });
      queryClient.invalidateQueries({ queryKey: ["pets"] });
      toast.success("死亡を記録しました");
    },
    onError: (error) => {
      handleApiError(error, "死亡記録");
    },
  });
}
