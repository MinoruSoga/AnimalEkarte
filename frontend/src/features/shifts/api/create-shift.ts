import { axios } from "@/lib/axios";
import type { ShiftMutationReceipt } from "./contract-types";
import type { CreateShiftRequest } from "../types";

/**
 * POST /v1/shifts（EMR-514 固定API契約）。
 * 全入力を明示した CreateShiftRequest を送り、確定 receipt を返す。
 * 成功判定は一覧の再取得ではなく receipt で行う。
 */
export async function createShift(request: CreateShiftRequest): Promise<ShiftMutationReceipt> {
  const { data } = await axios.post<ShiftMutationReceipt>("/v1/shifts", request);
  return data;
}
