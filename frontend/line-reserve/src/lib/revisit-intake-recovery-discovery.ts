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
import { LIFF_REVISIT_INTAKE_RECOVERY_FENCE_AVAILABLE } from "./revisit-intake-recovery-fence";
import type {
  RevisitIntakeRecoveryItem,
  RevisitIntakeRecoveryPage,
  RevisitIntakeRecoveryQueryRequest,
  RevisitIntakeRecoveryTargetRef,
} from "../api/revisit-intake-recovery-schemas";

/** 1 page は server 固定 50 件。1000 件を超える履歴は先を読まず truncated（保留）にする。 */
export const RECOVERY_MAX_PAGES = 20;

/** 発見全体の上限。各 request の timeout と合わせ、paging が checking 状態を延々保持しないようにする。 */
export const RECOVERY_DISCOVERY_DEADLINE_MS = 60_000;

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
  | "fence_unavailable" // Slice A: fence が無いため、発見の結果によらず新規送信を許可しない
  | "unresolved_operation" // 結果未確定の送信がある
  | "result_unavailable" // 結果を読めない（discovery_state=unavailable）
  | "active_operation" // 世代の active key が残る（解除は後続 Slice の fence が担当）
  | "scan_incomplete"; // 走査が完結していない（complete=false または打ち切り）

export interface RevisitIntakeRecoverySummary {
  newSubmissionBlocked: boolean;
  /** fence_unavailable 以外の保留理由がある（表示・再確認の対象がある）。 */
  hasOperationHold: boolean;
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

const TIMESTAMP = /^(\d{4})-(\d{2})-(\d{2})[Tt](\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?(?:[Zz]|([+-])(\d{2}):(\d{2}))$/;

/**
 * RFC3339 を UTC 整数秒 + 小数部文字列へ正規化して全精度で比較する（-1 / 0 / 1）。
 * 解釈できない値・閏秒（:60）・極端な年は null（呼び出し側が契約違反として扱う）。
 */
export function compareRfc3339(a: string, b: string): -1 | 0 | 1 | null {
  const norm = (value: string): { sec: number; frac: string } | null => {
    const m = TIMESTAMP.exec(value);
    if (m === null) return null;
    const [year, month, day, hour, minute, second] = m.slice(1, 7).map(Number);
    if (second > 59 || year < 1000) return null;
    const offset = m[8] === undefined ? 0 : (m[8] === "-" ? -1 : 1) * (Number(m[9]) * 60 + Number(m[10]));
    const sec = Date.UTC(year, month - 1, day, hour, minute, second) / 1000 - offset * 60;
    return Number.isFinite(sec) ? { sec, frac: m[7] ?? "" } : null;
  };
  const x = norm(a);
  const y = norm(b);
  if (x === null || y === null) return null;
  if (x.sec !== y.sec) return x.sec > y.sec ? 1 : -1;
  const width = Math.max(x.frac.length, y.frac.length);
  const fx = x.frac.padEnd(width, "0");
  const fy = y.frac.padEnd(width, "0");
  if (fx === fy) return 0;
  return fx > fy ? 1 : -1;
}

/** claimed_at DESC → operation_key DESC を全精度で検査する。同一 instant（表記違いを含む）は key で比較する。 */
function isDescending(prev: RevisitIntakeRecoveryItem, next: RevisitIntakeRecoveryItem): boolean {
  const cmp = compareRfc3339(prev.claimed_at, next.claimed_at);
  if (cmp === null) return false;
  if (cmp === 0) return prev.operation_key > next.operation_key;
  return cmp > 0;
}

/** page 取得を残り時間で打ち切る。signal を結合し、signal を無視する fetcher でも race で止める。 */
async function fetchWithinDeadline(
  fetchPage: RecoveryPageFetcher,
  request: RevisitIntakeRecoveryQueryRequest,
  deadline: number,
  parent?: AbortSignal,
): Promise<RevisitIntakeRecoveryPage> {
  const remaining = deadline - Date.now();
  const expired = () =>
    new RevisitIntakeRecoveryQueryError("network", "recovery discovery deadline exceeded");
  if (remaining <= 0) throw expired();
  const controller = new AbortController();
  const onParentAbort = () => controller.abort();
  parent?.addEventListener("abort", onParentAbort, { once: true });
  let timer: ReturnType<typeof setTimeout> | undefined;
  let timedOut = false;
  const deadlinePromise = new Promise<never>((_, reject) => {
    timer = setTimeout(() => {
      timedOut = true;
      controller.abort();
      reject(expired());
    }, remaining);
  });
  try {
    const page = await Promise.race([fetchPage(request, controller.signal), deadlinePromise]);
    // 期限後に完了した page は採用しない。
    if (timedOut || Date.now() > deadline) throw expired();
    return page;
  } catch (err: unknown) {
    throw timedOut ? expired() : err;
  } finally {
    if (timer !== undefined) clearTimeout(timer);
    parent?.removeEventListener("abort", onParentAbort);
  }
}

export async function discoverRevisitIntakeRecovery(
  fetchPage: RecoveryPageFetcher,
  scope: RevisitIntakeRecoveryScope,
  options: { maxPages?: number; signal?: AbortSignal; deadlineMs?: number } = {},
): Promise<RevisitIntakeRecoveryDiscovery> {
  const maxPages = options.maxPages ?? RECOVERY_MAX_PAGES;
  const deadline = Date.now() + (options.deadlineMs ?? RECOVERY_DISCOVERY_DEADLINE_MS);
  const items: RevisitIntakeRecoveryItem[] = [];
  const seenKeys = new Set<string>();
  const seenCursors = new Set<string>();
  let cursor: string | null = null;
  let first: RevisitIntakeRecoveryPage | null = null;
  let last: RevisitIntakeRecoveryPage | null = null;
  let truncated = false;

  for (let pageNo = 0; ; pageNo += 1) {
    if (options.signal?.aborted) throw new RevisitIntakeRecoveryQueryError("aborted", "aborted");
    if (Date.now() > deadline) {
      // 未読の page を残したまま完結扱いにしない。失敗として保留側へ倒す。
      throw new RevisitIntakeRecoveryQueryError("network", "recovery discovery deadline exceeded");
    }
    if (pageNo >= maxPages) {
      truncated = true;
      break;
    }
    const page = await fetchWithinDeadline(fetchPage, requestFor(scope, cursor), deadline, options.signal);
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
  const hasOperationHold = reasons.length > 0;
  // fence が無い間は、0 件・complete・理由なしでも解除しない（再読込直後と区別できず、旧 receipt は target query に出ない）。
  if (!LIFF_REVISIT_INTAKE_RECOVERY_FENCE_AVAILABLE) reasons.push("fence_unavailable");
  return { newSubmissionBlocked: reasons.length > 0, hasOperationHold, reasons };
}

/** 既定の fetcher（実 adapter）。mock へのフォールバックは持たない。 */
export function createRecoveryPageFetcher(clinicId: string, idToken: string): RecoveryPageFetcher {
  return (request, signal) =>
    queryRevisitIntakeOperationRecoveryPage({ clinicId, idToken, request, signal });
}
