import { isAxiosError } from "axios";
import { axios } from "@/lib/axios";
import type {
  ShiftContractError,
  ShiftMutationIntent,
  ShiftMutationReceipt,
  ShiftOperationResult,
} from "./contract-types";
import { extractShiftContractError } from "./contract-types";

/**
 * UUID v4 の operation_key を発行する。
 * mutation 意図ごとに 1 回発行し、結果照会が確定するまで同じ key を保持する
 * （unknown 中の新 key 再送は契約で禁止）。
 * crypto.randomUUID はセキュアコンテキスト専用のため、非対応環境では
 * getRandomValues による v4 生成へフォールバックする。
 */
export function newOperationKey(): string {
  const c = globalThis.crypto;
  if (typeof c?.randomUUID === "function") {
    return c.randomUUID();
  }
  const bytes = c.getRandomValues(new Uint8Array(16));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

/**
 * GET /v1/shifts/operations/{operation_key}?intent=（EMR-514 固定API契約）。
 * unknown は未実行・失敗・不存在の証明ではない。結果store障害は500、
 * 権限失効は403で返り、unknown や空receiptの成功へ変換されない。
 */
export async function getShiftOperationResult(
  operationKey: string,
  intent: ShiftMutationIntent,
): Promise<ShiftOperationResult> {
  const { data } = await axios.get<ShiftOperationResult>(`/v1/shifts/operations/${operationKey}`, {
    params: { intent },
  });
  return data;
}

/** mutation失敗の分類。code分岐は契約どおり error 文字列ではなく code で行う */
export type ShiftMutationError =
  | { kind: "contract"; status: 409; payload: ShiftContractError }
  | { kind: "http"; status: number }
  | { kind: "network" };

/**
 * axios例外を契約の失敗種別へ分類する。
 * - 409 + ShiftContractError shape → contract（code分岐対象）
 * - その他のHTTPエラー → http（400/401/403/404/500 等）
 * - 応答なし（通信断・タイムアウト）→ network（unknown照会の対象）
 */
export function classifyMutationError(err: unknown): ShiftMutationError {
  if (isAxiosError(err)) {
    const status = err.response?.status;
    if (status === 409) {
      const payload = extractShiftContractError(err);
      if (payload !== null) {
        return { kind: "contract", status: 409, payload };
      }
    }
    if (typeof status === "number") {
      return { kind: "http", status };
    }
    return { kind: "network" };
  }
  return { kind: "network" };
}

/**
 * 結果応答の喪失・5xx を unknown 扱いにするかを判定する。
 * 契約: 「結果応答喪失や5xxはunknownとして同じkeyの結果照会を行う」。
 */
export function isUnknownOutcome(error: ShiftMutationError): boolean {
  return error.kind === "network" || (error.kind === "http" && error.status >= 500);
}

/**
 * receipt の intent 別成功契約を検証する（B8/m3）。
 * - delete: shift=NULL・resulting_version=NULL（削除後の架空版を作らない）
 * - create/update: shift と resulting_version が非NULLで一致
 * - 照合: resulting_version === shift.version かつ shift_id === shift.id（server/clientの意味検証）
 */
export function isReceiptCoherent(receipt: ShiftMutationReceipt): boolean {
  if (receipt.intent === "delete") {
    return receipt.shift === null && receipt.resulting_version === null;
  }
  const shift = receipt.shift;
  if (shift === null || receipt.resulting_version === null) return false;
  return receipt.shift_id === shift.id && receipt.resulting_version === shift.version;
}
