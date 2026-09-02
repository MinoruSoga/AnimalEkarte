import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CheckupFieldType } from "@/types/checkup";

/** 飼い主レポート用: 親 checkup の日付・パッケージ名を付与した健診結果値。 */
export interface PetCheckupResult {
  id: number;
  checkupId: number;
  date: string;
  checkupTypeName: string;
  fieldName: string;
  fieldType: CheckupFieldType;
  unit: string;
  valueNumber?: number;
  valueText: string;
  valueBool?: boolean;
  valueList: string[];
  isAbnormal: boolean;
  status: string;
}

interface PetCheckupResultApi {
  id: number;
  checkup_id: number;
  date: string;
  checkup_type_name: string;
  field_name: string;
  field_type: CheckupFieldType;
  unit: string;
  value_number?: number;
  value_text: string;
  value_bool?: boolean;
  value_list: string[] | null;
  is_abnormal: boolean;
  status: string;
}

function transformPetCheckupResult(r: PetCheckupResultApi): PetCheckupResult {
  return {
    id: r.id,
    checkupId: r.checkup_id,
    date: r.date ?? "",
    checkupTypeName: r.checkup_type_name ?? "",
    fieldName: r.field_name ?? "",
    fieldType: r.field_type,
    unit: r.unit ?? "",
    valueNumber: r.value_number ?? undefined,
    valueText: r.value_text ?? "",
    valueBool: r.value_bool ?? undefined,
    valueList: r.value_list ?? [],
    isAbnormal: r.is_abnormal ?? false,
    status: r.status ?? "normal",
  };
}

// GET /v1/checkups/field-results?pet_id=X
const getPetCheckupResults = async (petId: string): Promise<PetCheckupResult[]> => {
  const { data } = await axios.get<PetCheckupResultApi[]>("/v1/checkups/field-results", {
    params: { pet_id: petId },
  });
  return (data ?? []).map(transformPetCheckupResult);
};

/**
 * Shared hook for fetching a pet's checkup (package) results.
 * Uses the same query key as features/checkups to share React Query cache.
 */
export const useGetPetCheckupResults = (petId?: string) => {
  return useQuery({
    queryKey: queryKeys.petCheckupResultsReport(petId!),
    queryFn: () => getPetCheckupResults(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
