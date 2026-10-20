import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import type { ShiftMutationReceipt } from "./contract-types";

/** DELETE /v1/shifts/{id} の要求変数。期待版と operation_key は query で送る */
export interface DeleteShiftVariables {
  shiftId: string;
  expectedVersion: number;
  operationKey: string;
}

/**
 * DELETE /v1/shifts/{id}?expected_version=&operation_key=（EMR-514 固定API契約）。
 * 削除の成功契約は receipt.shift = NULL / resulting_version = NULL
 * （削除後の架空版を作らない）。一覧一致だけで成功判定しない。
 */
export async function deleteShift({
  shiftId,
  expectedVersion,
  operationKey,
}: DeleteShiftVariables): Promise<ShiftMutationReceipt> {
  const { data } = await axios.delete<ShiftMutationReceipt>(`/v1/shifts/${shiftId}`, {
    params: { expected_version: expectedVersion, operation_key: operationKey },
  });
  return data;
}

export function useDeleteShift() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteShift,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
    },
  });
}
