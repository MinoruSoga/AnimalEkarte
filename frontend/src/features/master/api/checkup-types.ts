import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformCheckupType } from "@/lib/transforms/treatment";
import type { CheckupTypeItem } from "@/lib/transforms/treatment";
import type { CheckupType } from "@/types/generated/models";
import type {
  CreateCheckupTypeRequest,
  UpdateCheckupTypeRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { CheckupTypeItem };

// EMR-456: 健診種別の保険設定（品目側の設定であり、患者の保険資格・補償の保証ではない）。
// null = 未保存（旧欠落）、true = 保険対象外、false = 保険対象。null を false/true に補完しない。
// 生成型 CheckupType は model の json:"-" により is_non_insurance / version を持たないため、
// domain-owned response（backend checkup_type_response.go）の wire 差分をここで定義する。
export type CheckupTypeInsuranceSetting = boolean | null;

export type CheckupTypeResponse = CheckupType & {
  is_non_insurance?: CheckupTypeInsuranceSetting;
  version?: number;
};

// 共有 transform（EMR-454 所有）の CheckupTypeItem に、保険設定と版を別名で添える。
// 共有 TreatmentItem.isNonInsurance（boolean・欠落は false 表示）と混同しないよう名前を分けている。
export type CheckupTypeMasterItem = CheckupTypeItem & {
  insuranceSetting: CheckupTypeInsuranceSetting;
  version: number | null;
};

export type CreateCheckupTypeInsuranceRequest = CreateCheckupTypeRequest & {
  // 未指定なら API が新規初期値（保険対象外）を明示保存する。明示 null は 400（送信しない）。
  is_non_insurance?: boolean;
};

export type UpdateCheckupTypeInsuranceRequest = UpdateCheckupTypeRequest & {
  // 未指定は変更なし。明示 null は 400（送信しない）。null（未保存）へ戻す送信は API が持たない。
  is_non_insurance?: boolean;
  // 楽観ロックの期待版。PATCH では必須（未指定は 400・チケット回答 2026-10-20）。不一致は 409 で何も保存されない。
  // 型は任意のままだが、送信側（454 の共通 builder）が reader→フォーム→builder へ version を通す。
  version?: number;
};

export function toCheckupTypeMasterItem(data: CheckupTypeResponse): CheckupTypeMasterItem {
  return {
    ...transformCheckupType(data),
    insuranceSetting: typeof data.is_non_insurance === "boolean" ? data.is_non_insurance : null,
    version: typeof data.version === "number" ? data.version : null,
  };
}

const getAllCheckupTypes = async (): Promise<CheckupTypeMasterItem[]> => {
  const { data } = await axios.get<CheckupTypeResponse[]>("/v1/masters/checkup-types");
  return data.map(toCheckupTypeMasterItem);
};

export const useGetAllCheckupTypes = () =>
  useQuery({
    queryKey: queryKeys.masters.category("checkup-types"),
    queryFn: getAllCheckupTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateCheckupTypeInsuranceRequest): Promise<CheckupTypeMasterItem> => {
      const { data } = await axios.post<CheckupTypeResponse>("/v1/masters/checkup-types", req);
      return toCheckupTypeMasterItem(data);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateCheckupTypeInsuranceRequest;
    }): Promise<CheckupTypeMasterItem> => {
      const { data } = await axios.patch<CheckupTypeResponse>(
        `/v1/masters/checkup-types/${id}`,
        req,
      );
      return toCheckupTypeMasterItem(data);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/checkup-types/${id}`),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderCheckupTypes = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/checkup-types/reorder", req),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};
