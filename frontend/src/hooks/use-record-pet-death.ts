import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

interface RecordPetDeathVariables {
  petId: string;
  deceasedAt: string;
  deceasedReason?: string;
}

export function useRecordPetDeath() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ petId, deceasedAt, deceasedReason }: RecordPetDeathVariables) => {
      const clinicId = requireStoredClinicId();
      await axios.patch(`/v1/clinics/${clinicId}/pets/${petId}/death`, {
        deceased_at: deceasedAt,
        reason: deceasedReason,
      });
    },
    onSuccess: (_, { petId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.detail(petId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
      toast.success("死亡を記録しました");
    },
    onError: (error) => {
      handleApiError(error, "死亡記録");
    },
  });
}
