import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { recoveryPageMatchesRequest } from "../api/certificate-recovery-binding";
import {
  certificateRecoveryQueryResponseSchema,
  type CertificateRecoveryItem,
  type CertificateRecoveryQueryRequest,
} from "../api/certificate-recovery-schemas";

/**
 * EMR-487 Slice A: 回復 discovery の対象指定とページ送り。
 *
 * 守ること（契約 §4 / §11）:
 * - complete は「上限までの走査が終わった」だけを表す。0 件・complete・時間経過のどれも、
 *   新しい operation key を作ってよい証拠にしない（この関数は許可を返さず、見えた事実だけを返す）。
 * - 応答は adapter（http / mock）の実装に関わらず、ここで再度 strict に検査して要求へ束縛する。
 * - 壊れたページ・cursor の循環・重複 key・ページごとの世代/active のずれは例外（保留側）にする。
 */

export type CertificateRecoveryTarget =
  | { intent: "issue"; vaccinationId: number }
  | { intent: "reprint"; vaccinationId: number; editionId: number };

/** 1 ページ最大 50 件のため、20 ページ = 1000 操作で打ち切る（無制限に辿らない）。 */
export const MAX_RECOVERY_PAGES = 20;

export interface CertificateRecoveryDiscovery {
  readonly target: CertificateRecoveryTarget;
  /** 見えた順（claimed_at DESC, key DESC）。 */
  readonly items: readonly CertificateRecoveryItem[];
  /** 論理世代（10 進文字列）。Slice A では要求に載せず、表示・検証だけに使う。 */
  readonly recoveryGeneration: string;
  readonly activeOperationKey: string | null;
  /** 最後のページまで走査した。情報のみで、次操作の許可には使わない。 */
  readonly scanEnded: boolean;
  /** ページ上限で打ち切った。見えていない操作が残りうる。 */
  readonly truncated: boolean;
}

export function recoveryRequestFor(
  target: CertificateRecoveryTarget,
  cursor: string | null,
): CertificateRecoveryQueryRequest {
  return target.intent === "issue"
    ? { target_ref: { vaccination_id: target.vaccinationId, edition_id: null }, intent: "issue", cursor }
    : {
        target_ref: { vaccination_id: target.vaccinationId, edition_id: target.editionId },
        intent: "reprint",
        cursor,
      };
}

function malformed(message: string): CertificateApiError {
  return new CertificateApiError("malformed", { message });
}

function assertNotAborted(signal: AbortSignal | undefined): void {
  if (signal?.aborted) throw new CertificateApiError("network", { message: "回復照会を中止しました" });
}

export async function discoverCertificateOperations(
  api: CertificateApi,
  target: CertificateRecoveryTarget,
  signal?: AbortSignal,
): Promise<CertificateRecoveryDiscovery> {
  const seenCursors = new Set<string>();
  const seenKeys = new Set<string>();
  const items: CertificateRecoveryItem[] = [];
  let generation: string | null = null;
  let active: string | null = null;
  let cursor: string | null = null;

  for (let pageNo = 0; pageNo < MAX_RECOVERY_PAGES; pageNo += 1) {
    assertNotAborted(signal);
    const request = recoveryRequestFor(target, cursor);
    const raw = await api.queryOperationRecovery(request, signal);
    assertNotAborted(signal);

    const parsed = certificateRecoveryQueryResponseSchema.safeParse(raw);
    if (!parsed.success) throw malformed("回復照会の応答形式が不正です");
    const page = parsed.data;
    if (!recoveryPageMatchesRequest(page, request)) throw malformed("回復照会の応答が要求した対象と一致しません");

    // 世代/active がページ間で変わった = 走査中に状態が動いた。どちらの値も信用せず保留に倒す。
    if (generation !== null && (page.recovery_generation !== generation || page.active_operation_key !== active)) {
      throw malformed("回復照会のページ間で世代が一致しません");
    }
    generation = page.recovery_generation;
    active = page.active_operation_key;

    for (const item of page.items) {
      if (seenKeys.has(item.operation_key)) throw malformed("回復照会に同じ操作が重複しています");
      seenKeys.add(item.operation_key);
      items.push(item);
    }

    if (page.next_cursor === null) {
      return {
        target,
        items,
        recoveryGeneration: page.recovery_generation,
        activeOperationKey: page.active_operation_key,
        scanEnded: true,
        truncated: false,
      };
    }
    if (page.next_cursor === cursor || seenCursors.has(page.next_cursor)) {
      throw malformed("回復照会の cursor が循環しています");
    }
    seenCursors.add(page.next_cursor);
    cursor = page.next_cursor;
  }

  return {
    target,
    items,
    recoveryGeneration: generation ?? "1",
    activeOperationKey: active,
    scanEnded: false,
    truncated: true,
  };
}
