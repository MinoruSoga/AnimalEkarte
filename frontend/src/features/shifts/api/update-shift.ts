import { axios } from "@/lib/axios";
import type { ShiftMutationReceipt } from "./contract-types";
import type { UpdateShiftRequest } from "../types";

/**
 * PATCH /v1/shifts/{id}（EMR-514 固定API契約）。
 * 未指定は既存値維持。breaks は配列全置換（[] は全解除）。
 * 期待版（expected_version）不一致は 409 version_conflict。
 * 成功判定は一覧の再取得ではなく receipt で行う。
 */
export async function updateShift(
  id: string,
  request: UpdateShiftRequest,
): Promise<ShiftMutationReceipt> {
  const { data } = await axios.patch<ShiftMutationReceipt>(`/v1/shifts/${id}`, request);
  return data;
}
