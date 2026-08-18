import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type {
  MedicineDoseBasis,
  MedicineDoseParam,
  MedicineDoseSpecies,
  MedicineRoundingMode,
} from "@/types/generated/models";

/**
 * PUT /v1/masters/medicines/:id/dose-params/:species リクエストボディ (#201)
 * species は URL パスに含まれるためボディには含めない。
 */
export interface UpsertMedicineDoseParamRequest {
  dose_basis?: MedicineDoseBasis;
  dose_per_kg: number;
  min_mg_per_kg?: number;
  max_mg_per_kg?: number;
  absolute_max_dose?: number;
  rounding_step?: number;
  rounding_mode?: MedicineRoundingMode;
  notes?: string;
}

const getMedicineDoseParams = async (medicineId: string): Promise<MedicineDoseParam[]> => {
  const { data } = await axios.get<MedicineDoseParam[]>(
    `/v1/masters/medicines/${medicineId}/dose-params`
  );
  return data ?? [];
};

export const useGetMedicineDoseParams = (medicineId: string) => {
  return useQuery({
    queryKey: queryKeys.masters.medicineDoseParams(medicineId),
    queryFn: () => getMedicineDoseParams(medicineId),
    enabled: !!medicineId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const upsertMedicineDoseParam = async (
  medicineId: string,
  species: MedicineDoseSpecies,
  input: UpsertMedicineDoseParamRequest
): Promise<MedicineDoseParam> => {
  const { data } = await axios.put<MedicineDoseParam>(
    `/v1/masters/medicines/${medicineId}/dose-params/${species}`,
    input
  );
  return data;
};

export const useUpsertMedicineDoseParam = (medicineId: string) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      species,
      input,
    }: {
      species: MedicineDoseSpecies;
      input: UpsertMedicineDoseParamRequest;
    }) => upsertMedicineDoseParam(medicineId, species, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.medicineDoseParams(medicineId) });
      toast.success("投与量パラメータを保存しました");
    },
    onError: (error) => handleApiError(error, "投与量パラメータの保存"),
  });
};

const deleteMedicineDoseParam = async (
  medicineId: string,
  species: MedicineDoseSpecies
): Promise<void> => {
  await axios.delete(`/v1/masters/medicines/${medicineId}/dose-params/${species}`);
};

export const useDeleteMedicineDoseParam = (medicineId: string) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (species: MedicineDoseSpecies) => deleteMedicineDoseParam(medicineId, species),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.medicineDoseParams(medicineId) });
    },
    onError: (error) => handleApiError(error, "投与量パラメータの削除"),
  });
};
