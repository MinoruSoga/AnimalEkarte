// EMR-474 operation-recovery Slice A: 発見 query を page 単位でたどり、保留判定の材料へ畳む。
// 契約: docs/work/emr/emr-474-487-operation-recovery-contract-20261021.md §4 / §11。
//
// 不変条件:
// - 0 件・complete=true・経過時間は「新しい操作 key を出してよい」根拠ではない。この module は
//   許可を返さない。返すのは「保留すべきか」(newSubmissionBlocked) だけで、false は
//   「この発見では保留理由が見つからなかった」ことしか意味しない。
// - 想定外（cursor の再利用・世代の変動・重複 key・順序違反・page 数超過）は、結果を推測せず
//   契約違反または truncated として保留側へ倒す。
import {
  RevisitIntakeRecoveryQueryError,
  queryRevisitIntakeOperationRecoveryPage,
} from "../api/revisit-intake-recovery";
import type {
  RevisitIntakeRecoveryItem,
  RevisitIntakeRecoveryPage,
  RevisitIntakeRecoveryQueryRequest,
  RevisitIntakeRecoveryTargetRef,
} from "../api/revisit-intake-recovery-schemas";

/** 1 page は server 固定 50 件。1000 件を超える履歴は先を読まず truncated（保留）にする。 */
export const RECOVERY_MAX_PAGES = 20;

export type RevisitIntakeRecoveryScope =
  | { scope: "target"; targetRef: RevisitIntakeRecoveryTargetRef }
  | { scope: "self" };

export type RecoveryPageFetcher = (
  request: RevisitIntakeRecoveryQueryRequest,
  signal?: AbortSignal,
) => Promise<RevisitIntakeRecoveryPage>;

export interface RevisitIntakeRecoveryDiscovery {
  scope: "target" | "self";
  items: RevisitIntakeRecoveryItem[];
  /** target scope のみ非 null（self は常に null）。 */
  recoveryGeneration: string | null;
  activeOperationKey: string | null;
  /** 最終 page の complete。上限の走査が終わったことだけを表す。 */
  scanComplete: boolean;
  /** RECOVERY_MAX_PAGES で打ち切った（未読の page が残る）。 */
  truncated: boolean;
}

type RecoveryHoldReason =
  | "unresolved_operation" // 結果未確定の送信がある
  | "result_unavailable" // 結果を読めない（discovery_state=unavailable）
  | "active_operation" // 世代の active key が残る（解除は後続 Slice の fence が担当）
  | "scan_incomplete"; // 走査が完結していない（complete=false または打ち切り）

export interface RevisitIntakeRecoverySummary {
  newSubmissionBlocked: boolean;
  reasons: RecoveryHoldReason[];
}

function requestFor(scope: RevisitIntakeRecoveryScope, cursor: string | null): RevisitIntakeRecoveryQueryRequest {
  return scope.scope === "target"
    ? { scope: "target", target_ref: scope.targetRef, cursor }
    : { scope: "self", cursor };
}

function violation(message: string): RevisitIntakeRecoveryQueryError {
  return new RevisitIntakeRecoveryQueryError("contract_violation", message);
}

/** claimed_at DESC → operation_key DESC の順序を ms 精度で検査する（同一 ms 内は key で比較）。 */
function isDescending(prev: RevisitIntakeRecoveryItem, next: RevisitIntakeRecoveryItem): boolean {
  const a = Date.parse(prev.claimed_at);
  const b = Date.parse(next.claimed_at);
  if (Number.isNaN(a) || Number.isNaN(b)) return false;
  if (a !== b) return a > b;
  return prev.operation_key >= next.operation_key;
}

export async function discoverRevisitIntakeRecovery(
  fetchPage: RecoveryPageFetcher,
  scope: RevisitIntakeRecoveryScope,
  options: { maxPages?: number; signal?: AbortSignal } = {},
): Promise<RevisitIntakeRecoveryDiscovery> {
  const maxPages = options.maxPages ?? RECOVERY_MAX_PAGES;
  const items: RevisitIntakeRecoveryItem[] = [];
  const seenKeys = new Set<string>();
  const seenCursors = new Set<string>();
  let cursor: string | null = null;
  let first: RevisitIntakeRecoveryPage | null = null;
  let last: RevisitIntakeRecoveryPage | null = null;
  let truncated = false;

  for (let pageNo = 0; ; pageNo += 1) {
    if (options.signal?.aborted) throw new RevisitIntakeRecoveryQueryError("aborted", "aborted");
    if (pageNo >= maxPages) {
      truncated = true;
      break;
    }
    const page: RevisitIntakeRecoveryPage = await fetchPage(requestFor(scope, cursor), options.signal);
    if (first === null) first = page;
    // 世代・active key が page 間で変わったら、集合が一貫した 1 時点のものでないため信用しない。
    if (
      page.recovery_generation !== first.recovery_generation ||
      page.active_operation_key !== first.active_operation_key
    ) {
      throw violation("recovery generation or active key changed between pages");
    }
    for (const item of page.items) {
      if (seenKeys.has(item.operation_key)) throw violation("duplicate operation_key across pages");
      const prev = items[items.length - 1];
      if (prev !== undefined && !isDescending(prev, item)) throw violation("items are not ordered");
      seenKeys.add(item.operation_key);
      items.push(item);
    }
    last = page;
    if (page.next_cursor === null) break;
    if (seenCursors.has(page.next_cursor) || page.next_cursor === cursor) {
      throw violation("cursor did not advance");
    }
    seenCursors.add(page.next_cursor);
    cursor = page.next_cursor;
  }

  if (first === null || last === null) {
    // maxPages=0 など。1 page も読めていない発見は完結扱いにしない。
    throw violation("no page was read");
  }
  return {
    scope: scope.scope,
    items,
    recoveryGeneration: first.recovery_generation,
    activeOperationKey: first.active_operation_key,
    scanComplete: last.complete && !truncated,
    truncated,
  };
}

/** 保留理由の抽出。false は「許可」ではなく「この発見では理由が見つからない」にすぎない。 */
export function summarizeRevisitIntakeRecovery(
  discovery: RevisitIntakeRecoveryDiscovery,
): RevisitIntakeRecoverySummary {
  const reasons: RecoveryHoldReason[] = [];
  if (discovery.items.some((i) => i.discovery_state === "unavailable")) {
    reasons.push("result_unavailable");
  }
  if (
    discovery.items.some((i) => i.discovery_state === "result" && i.execution_state === "unresolved")
  ) {
    reasons.push("unresolved_operation");
  }
  if (discovery.activeOperationKey !== null) reasons.push("active_operation");
  if (!discovery.scanComplete) reasons.push("scan_incomplete");
  return { newSubmissionBlocked: reasons.length > 0, reasons };
}

/** 既定の fetcher（実 adapter）。mock へのフォールバックは持たない。 */
export function createRecoveryPageFetcher(clinicId: string, idToken: string): RecoveryPageFetcher {
  return (request, signal) =>
    queryRevisitIntakeOperationRecoveryPage({ clinicId, idToken, request, signal });
}
