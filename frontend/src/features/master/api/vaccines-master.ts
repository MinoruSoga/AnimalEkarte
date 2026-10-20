import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformVaccine } from "@/lib/transforms/treatment";
import type { VaccineItem } from "@/lib/transforms/treatment";
import type { Vaccine } from "@/types/generated/models";
import type {
  CreateVaccineRequest,
  UpdateVaccineRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { VaccineItem };

// EMR-455: 予防接種の品目別保険設定。共有の生成型・transform・request 型は EMR-454 所有のため、
// vaccine 専有の wire 拡張をこのファイルに閉じる。wire の null（未保存/旧欠落）は undefined として保持し、
// false 等へ補完しない（共有 TreatmentItem.isNonInsurance?: boolean と同じ表現）。
type VaccineMasterResponse = Vaccine & { version: number; is_non_insurance?: boolean | null };

export type VaccineMasterItem = VaccineItem & { version: number; isNonInsurance?: boolean };

// 作成時に未送信の場合はサーバーが新規初期値（対象外=true）を保存する。
export type CreateVaccineMasterRequest = CreateVaccineRequest & { is_non_insurance?: boolean };

// 更新時は未送信=変更なし。null はサーバーが 400 で拒否するため型でも許さない。
export type UpdateVaccineMasterRequest = UpdateVaccineRequest & {
  version: number;
  is_non_insurance?: boolean;
};

export const transformVaccineMaster = (data: VaccineMasterResponse): VaccineMasterItem => ({
  ...transformVaccine(data),
  version: data.version,
  isNonInsurance: data.is_non_insurance ?? undefined,
});

export const getAllVaccinesMaster = async (): Promise<VaccineMasterItem[]> => {
  const { data } = await axios.get<VaccineMasterResponse[]>("/v1/masters/vaccines");
  return data.map(transformVaccineMaster);
};

export const createVaccineMaster = async (
  req: CreateVaccineMasterRequest,
): Promise<VaccineMasterItem> => {
  const { data } = await axios.post<VaccineMasterResponse>("/v1/masters/vaccines", req);
  return transformVaccineMaster(data);
};

export const updateVaccineMaster = async (
  id: string,
  req: UpdateVaccineMasterRequest,
): Promise<VaccineMasterItem> => {
  if (!Number.isSafeInteger(req.version) || req.version <= 0) {
    throw new Error("編集開始時のバージョンがありません。再読み込みしてください。");
  }
  const { data } = await axios.patch<VaccineMasterResponse>(`/v1/masters/vaccines/${id}`, req);
  return transformVaccineMaster(data);
};

export const useGetAllVaccinesMaster = () =>
  useQuery({
    queryKey: queryKeys.masters.category("vaccines"),
    queryFn: getAllVaccinesMaster,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createVaccineMaster,
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateVaccineMasterRequest }) =>
      updateVaccineMaster(id, req),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/vaccines/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const useReorderVaccinesMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) => axios.patch("/v1/masters/vaccines/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
