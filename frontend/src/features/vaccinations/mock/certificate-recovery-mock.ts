import { CertificateApiError } from "../api/certificate-errors";
import {
  certificateRecoveryQueryRequestSchema,
  RECOVERY_PAGE_LIMIT,
  type CertificateRecoveryItem,
  type CertificateRecoveryQueryRequest,
  type CertificateRecoveryQueryResponse,
} from "../api/certificate-recovery-schemas";
import type { CertificateOperationIntent } from "../api/certificate-api";
import type { CertificateResult } from "../api/certificate-schemas";
import { mockHash } from "./certificate-fixtures";

/**
 * EMR-487 Slice A: 回復 discovery の契約モック（テスト専用・実 API の代替ではない）。
 * 契約どおりの形・順序（claimed_at DESC, key DESC）・1 ページ 50 件・opaque cursor を模す。
 * 世代は論理世代 "1" 固定、fence / 新 key 再開は持たない（Slice B）。
 */

export type MockRecoveryExecution = "succeeded" | "unresolved" | "closed_without_effect";

export interface MockRecoveryEntry {
  key: string;
  intent: CertificateOperationIntent;
  editionId: number | null;
  claimedAt: string;
  execution: MockRecoveryExecution;
  discoveryUnavailable: boolean;
  result: CertificateResult | null;
}

const BASE_TIME = Date.parse("2026-10-20T14:00:00Z");

function encodeCursor(payload: object): string {
  return btoa(JSON.stringify(payload)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function decodeCursor(cursor: string): { upperClaimedAt: string; upperKey: string; lastClaimedAt: string; lastKey: string } {
  try {
    const padded = cursor.replace(/-/g, "+").replace(/_/g, "/");
    const parsed = JSON.parse(atob(padded + "=".repeat((4 - (padded.length % 4)) % 4))) as Record<string, unknown>;
    if (
      parsed.v !== 1 ||
      typeof parsed.upper_claimed_at !== "string" ||
      typeof parsed.upper_key !== "string" ||
      typeof parsed.last_claimed_at !== "string" ||
      typeof parsed.last_key !== "string"
    ) {
      throw new Error("bad cursor");
    }
    return {
      upperClaimedAt: parsed.upper_claimed_at,
      upperKey: parsed.upper_key,
      lastClaimedAt: parsed.last_claimed_at,
      lastKey: parsed.last_key,
    };
  } catch {
    throw new CertificateApiError("validation", { httpStatus: 400, code: "VALIDATION" });
  }
}

/** (claimed_at DESC, key DESC) の比較。a が b より前（新しい側）なら負。 */
function compareDesc(a: { claimedAt: string; key: string }, b: { claimedAt: string; key: string }): number {
  if (a.claimedAt !== b.claimedAt) return a.claimedAt < b.claimedAt ? 1 : -1;
  if (a.key !== b.key) return a.key < b.key ? 1 : -1;
  return 0;
}

export function createMockRecoveryLog() {
  const entries = new Map<string, MockRecoveryEntry>();
  let seq = 0;

  function nextClaimedAt(): string {
    seq += 1;
    return new Date(BASE_TIME + seq * 1000).toISOString();
  }

  function toItem(entry: MockRecoveryEntry): CertificateRecoveryItem {
    const base = { operation_key: entry.key, intent: entry.intent, claimed_at: entry.claimedAt };
    if (entry.execution === "succeeded" && entry.result !== null && entry.result.status !== "not-succeeded") {
      return {
        ...base,
        discovery_state: "result",
        execution_state: "succeeded",
        receipt_digest: mockHash(`receipt:${entry.key}`),
        result: entry.result,
      };
    }
    if (entry.discoveryUnavailable) {
      return { ...base, discovery_state: "unavailable", execution_state: "unresolved", receipt_digest: null, result: null };
    }
    return {
      ...base,
      discovery_state: "result",
      execution_state: entry.execution === "closed_without_effect" ? "closed_without_effect" : "unresolved",
      receipt_digest: null,
      result: { status: "not-succeeded", operation_key: entry.key },
    };
  }

  /** claim 済みにする（既にあれば何もしない）。 */
  function claim(key: string, intent: CertificateOperationIntent, editionId: number | null): void {
    if (entries.has(key)) return;
    entries.set(key, {
      key,
      intent,
      editionId,
      claimedAt: nextClaimedAt(),
      execution: "unresolved",
      discoveryUnavailable: false,
      result: null,
    });
  }

  return {
    get: (key: string) => entries.get(key),
    claim,
    succeed(key: string, result: CertificateResult): void {
      const entry = entries.get(key);
      if (entry) entries.set(key, { ...entry, execution: "succeeded", result });
    },
    markUnavailable(key: string): void {
      const entry = entries.get(key);
      if (entry) entries.set(key, { ...entry, discoveryUnavailable: true });
    },
    closeWithoutEffect(key: string): void {
      const entry = entries.get(key);
      if (entry && entry.execution !== "succeeded") entries.set(key, { ...entry, execution: "closed_without_effect" });
    },
    /** 未確定の claim を count 件作る（ページ送りの確認用）。key は決定的。 */
    seedUnresolved(count: number, intent: CertificateOperationIntent, editionId: number | null): string[] {
      const keys: string[] = [];
      for (let i = 0; i < count; i += 1) {
        const key = `00000000-0000-4000-8000-${(entries.size + 1).toString(16).padStart(12, "0")}`;
        claim(key, intent, editionId);
        keys.push(key);
      }
      return keys;
    },
    query(request: CertificateRecoveryQueryRequest): CertificateRecoveryQueryResponse {
      const checked = certificateRecoveryQueryRequestSchema.safeParse(request);
      if (!checked.success) throw new CertificateApiError("validation", { httpStatus: 400, code: "VALIDATION" });
      const req = checked.data;
      const scoped = [...entries.values()]
        .filter((e) => e.intent === req.intent && (req.intent === "issue" || e.editionId === req.target_ref.edition_id))
        .sort(compareDesc);
      const cursor = req.cursor === null ? null : decodeCursor(req.cursor);
      const upper = cursor ? { claimedAt: cursor.upperClaimedAt, key: cursor.upperKey } : scoped[0] ?? null;
      const candidates = scoped.filter(
        (e) =>
          (upper === null || compareDesc(upper, e) <= 0) &&
          (cursor === null || compareDesc({ claimedAt: cursor.lastClaimedAt, key: cursor.lastKey }, e) < 0),
      );
      const page = candidates.slice(0, RECOVERY_PAGE_LIMIT);
      const last = page[page.length - 1];
      let nextCursor: string | null = null;
      if (candidates.length > RECOVERY_PAGE_LIMIT && upper !== null && last !== undefined) {
        nextCursor = encodeCursor({
          v: 1,
          upper_claimed_at: upper.claimedAt,
          upper_key: upper.key,
          last_claimed_at: last.claimedAt,
          last_key: last.key,
        });
      }
      return {
        items: page.map(toItem),
        next_cursor: nextCursor,
        complete: nextCursor === null,
        // control 不存在は論理世代 "1"・active=null。ここでは最新 claim を active とみなす。
        recovery_generation: "1",
        active_operation_key: scoped[0]?.key ?? null,
      };
    },
  };
}

export type MockRecoveryLog = ReturnType<typeof createMockRecoveryLog>;
