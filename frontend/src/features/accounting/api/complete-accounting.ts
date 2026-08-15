import { axios } from "@/lib/axios";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type {
  BackendAccounting,
  CompleteAccountingRequest,
} from "./types";

/**
 * BUG-018: 会計確定を単一 aggregate command で送信する。
 * Idempotency-Key は mutation 単位で1回生成し、retry で再利用する。
 */
export const completeAccounting = async (
  req: CompleteAccountingRequest,
  idempotencyKey: string,
): Promise<Accounting> => {
  const { data } = await axios.post<BackendAccounting>(
    "/v1/accountings/complete",
    req,
    {
      headers: {
        "Idempotency-Key": idempotencyKey,
      },
    },
  );
  return transformToAccounting(data);
};

/** mutation 単位の Idempotency-Key。画面再読込では新しい UUID。 */
export function createAccountingCompletionIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  // Fallback for non-secure contexts in tests
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
