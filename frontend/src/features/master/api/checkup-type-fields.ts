import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { CheckupFieldType } from "@/types/checkup";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

// EMR-225: 定期健診パッケージのフィールド定義 write 側 API。
// GET（一覧）は @/hooks/use-checkup-fields の useGetCheckupTypeFields が正本（カルテ画面の
// 動的フォーム構築と query key を共有し、mutation 後の invalidate が双方へ効く）。
// options の永続化形状は manifest import と同じ { value, label } 配列。

export interface CheckupFieldOption {
  value: string;
  label: string;
}

// CheckupTypeFieldRow の id を string 化したエディタ用行型
// （useSortableList / DnD の item 同一性は string id 規約）。
export type CheckupTypeFieldEntry = Omit<CheckupTypeFieldRow, "id" | "checkupTypeId"> & {
  id: string;
  checkupTypeId: string;
};

export function toCheckupTypeFieldEntry(row: CheckupTypeFieldRow): CheckupTypeFieldEntry {
  return { ...row, id: String(row.id), checkupTypeId: String(row.checkupTypeId) };
}

export interface CreateCheckupTypeFieldRequest {
  name: string;
  field_type: CheckupFieldType;
  unit?: string;
  min_value?: number;
  max_value?: number;
  options?: CheckupFieldOption[];
  sort_order?: number;
}

// PATCH 部分更新: min/max を NULL に戻すには clear_min_value / clear_max_value を使う
// （min_value: null 送信は省略と同義になるため、明示クリアフラグが必須 — clear_parent_id 先例）。
export interface UpdateCheckupTypeFieldRequest {
  name?: string;
  field_type?: CheckupFieldType;
  unit?: string;
  min_value?: number;
  max_value?: number;
  options?: CheckupFieldOption[];
  sort_order?: number;
  clear_min_value?: boolean;
  clear_max_value?: boolean;
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

// use-checkup-fields.ts の transform と同一写像（共有 fetcher 側は export されていないため局所実装）。
function transformCheckupTypeField(f: CheckupTypeFieldApi): CheckupTypeFieldRow {
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

export async function createCheckupTypeField(
  checkupTypeId: string,
  req: CreateCheckupTypeFieldRequest,
): Promise<CheckupTypeFieldRow> {
  const { data } = await axios.post<CheckupTypeFieldApi>(
    `/v1/masters/checkup-types/${checkupTypeId}/fields`,
    req,
  );
  return transformCheckupTypeField(data);
}

export async function updateCheckupTypeField(
  checkupTypeId: string,
  fieldId: string,
  req: UpdateCheckupTypeFieldRequest,
): Promise<CheckupTypeFieldRow> {
  const { data } = await axios.patch<CheckupTypeFieldApi>(
    `/v1/masters/checkup-types/${checkupTypeId}/fields/${fieldId}`,
    req,
  );
  return transformCheckupTypeField(data);
}

export async function deleteCheckupTypeField(
  checkupTypeId: string,
  fieldId: string,
): Promise<void> {
  await axios.delete(`/v1/masters/checkup-types/${checkupTypeId}/fields/${fieldId}`);
}

export async function reorderCheckupTypeFields(
  checkupTypeId: string,
  ids: number[],
): Promise<void> {
  await axios.patch(`/v1/masters/checkup-types/${checkupTypeId}/fields/reorder`, { ids });
}

export async function invalidateCheckupTypeFieldQueries(
  queryClient: QueryClient,
  checkupTypeId: string,
): Promise<void> {
  // checkup_types 一覧レスポンスは fields を埋め込まないため typeFields key のみで十分。
  await queryClient.invalidateQueries({
    queryKey: queryKeys.checkups.typeFields(checkupTypeId),
  });
}

function useInvalidateCheckupTypeFields() {
  const queryClient = useQueryClient();
  return (checkupTypeId: string) => invalidateCheckupTypeFieldQueries(queryClient, checkupTypeId);
}

export function useCreateCheckupTypeField() {
  const invalidate = useInvalidateCheckupTypeFields();
  return useMutation({
    mutationFn: ({
      checkupTypeId,
      req,
    }: {
      checkupTypeId: string;
      req: CreateCheckupTypeFieldRequest;
    }) => createCheckupTypeField(checkupTypeId, req),
    onSuccess: (_data, variables) => invalidate(variables.checkupTypeId),
    onError: (error) => handleApiError(error, "健診項目の作成"),
  });
}

export function useUpdateCheckupTypeField() {
  const invalidate = useInvalidateCheckupTypeFields();
  return useMutation({
    mutationFn: ({
      checkupTypeId,
      fieldId,
      req,
    }: {
      checkupTypeId: string;
      fieldId: string;
      req: UpdateCheckupTypeFieldRequest;
    }) => updateCheckupTypeField(checkupTypeId, fieldId, req),
    onSuccess: (_data, variables) => invalidate(variables.checkupTypeId),
    onError: (error) => handleApiError(error, "健診項目の更新"),
  });
}

export function useDeleteCheckupTypeField() {
  const invalidate = useInvalidateCheckupTypeFields();
  return useMutation({
    mutationFn: ({ checkupTypeId, fieldId }: { checkupTypeId: string; fieldId: string }) =>
      deleteCheckupTypeField(checkupTypeId, fieldId),
    onSuccess: (_data, variables) => invalidate(variables.checkupTypeId),
    onError: (error) => handleApiError(error, "健診項目の削除"),
  });
}

export function useReorderCheckupTypeFields() {
  const invalidate = useInvalidateCheckupTypeFields();
  return useMutation({
    mutationFn: ({ checkupTypeId, ids }: { checkupTypeId: string; ids: number[] }) =>
      reorderCheckupTypeFields(checkupTypeId, ids),
    onSuccess: (_data, variables) => invalidate(variables.checkupTypeId),
    onError: (error) => handleApiError(error, "健診項目の並び替え"),
  });
}
