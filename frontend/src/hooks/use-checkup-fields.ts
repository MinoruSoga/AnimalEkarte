import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CheckupFieldType } from "@/types/checkup";

interface CheckupFieldOption {
  value: string;
  label: string;
}

/** FE 動的フォーム構築に使う健診パッケージのフィールド定義行。 */
export interface CheckupTypeFieldRow {
  id: number;
  checkupTypeId: number;
  name: string;
  fieldType: CheckupFieldType;
  unit: string;
  minValue?: number;
  maxValue?: number;
  options: CheckupFieldOption[];
  isProvisional: boolean;
  sortOrder: number;
}

interface CheckupTypeFieldApi {
  id: number;
  checkup_type_id: number;
  name: string;
  field_type: CheckupFieldType;
  unit: string;
  min_value?: number;
  max_value?: number;
  options: CheckupFieldOption[] | null;
  is_provisional: boolean;
  sort_order: number;
}

function transform(f: CheckupTypeFieldApi): CheckupTypeFieldRow {
  return {
    id: f.id,
    checkupTypeId: f.checkup_type_id,
    name: f.name,
    fieldType: f.field_type,
    unit: f.unit ?? "",
    minValue: f.min_value ?? undefined,
    maxValue: f.max_value ?? undefined,
    options: f.options ?? [],
    isProvisional: f.is_provisional,
    sortOrder: f.sort_order,
  };
}

// GET /v1/masters/checkup-types/:id/fields
const getCheckupTypeFields = async (checkupTypeId: string): Promise<CheckupTypeFieldRow[]> => {
  const { data } = await axios.get<CheckupTypeFieldApi[]>(
    `/v1/masters/checkup-types/${checkupTypeId}/fields`,
  );
  return (data ?? []).map(transform);
};

export const useGetCheckupTypeFields = (checkupTypeId: string) => {
  return useQuery({
    queryKey: queryKeys.checkups.typeFields(checkupTypeId),
    queryFn: () => getCheckupTypeFields(checkupTypeId),
    enabled: !!checkupTypeId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

/** PUT 一括置換に送る健診結果値 1 件分。status / is_abnormal はサーバ側で導出する。 */
export interface CheckupFieldResultInput {
  checkup_type_field_id: number;
  value_number?: number | null;
  value_text?: string;
  value_bool?: boolean | null;
  value_list?: string[];
}

// PUT /v1/medical-records/:medicalRecordId/checkups/:checkupId/field-results
// 既存全削除→一括登録の PUT セマンティクス。
export const replaceCheckupFieldResults = async (
  medicalRecordId: string | number,
  checkupId: string | number,
  results: CheckupFieldResultInput[],
): Promise<void> => {
  await axios.put(`/v1/medical-records/${medicalRecordId}/checkups/${checkupId}/field-results`, {
    results,
  });
};
