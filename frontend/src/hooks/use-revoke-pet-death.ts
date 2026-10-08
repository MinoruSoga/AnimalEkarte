import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

export function useRevokePetDeath() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (petId: string) => {
      const clinicId = requireStoredClinicId();
      await axios.delete(`/v1/clinics/${clinicId}/pets/${petId}/death`);
    },
    onSuccess: (_, petId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.detail(petId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
      // use-record-pet-death と同じ sentinel 埋め込みキャッシュを更新する
      // （解除後も stale deceased 表示が残ると復活した個体を死亡扱いし続ける）。
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.detailAll() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.detailAll() });
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerSharedPets.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.ownerReportPetsAll() });
      toast.success("死亡記録を解除しました");
    },
    onError: (error) => {
      handleApiError(error, "死亡記録解除");
    },
  });
}
