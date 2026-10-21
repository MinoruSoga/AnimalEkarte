import type {
  CertificateRecoveryItem,
  CertificateRecoveryQueryResponse,
} from "../api/certificate-recovery-schemas";
import type { CertificateSucceededResult } from "../api/certificate-schemas";
import { MOCK_VACCINATION_ID, mockHash } from "./certificate-fixtures";

/**
 * EMR-487 Slice A: 回復 discovery の合成 fixture（テスト専用）。
 * 契約の zod（strict）を満たす形だけを返す。値は合成で、実 key・実 hash ではない。
 */

export const RECOVERY_KEY_A = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
export const RECOVERY_KEY_B = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
export const RECOVERY_KEY_C = "cccccccc-cccc-4ccc-8ccc-cccccccccccc";

export function succeededResultFor(
  key: string,
  refs: { vaccinationId?: number; editionId?: number; artifactId?: number } = {},
): CertificateSucceededResult {
  return {
    status: "succeeded",
    operation_key: key,
    result_refs: {
      vaccination_id: String(refs.vaccinationId ?? MOCK_VACCINATION_ID),
      series_id: "201",
      edition_id: String(refs.editionId ?? 301),
      artifact_id: String(refs.artifactId ?? 401),
    },
    committed_versions: { edition: "1", source_version: "7" },
    recorded_at: "2026-10-20T14:30:00Z",
  };
}

export function succeededItem(
  key: string,
  options: { intent?: "issue" | "reprint"; claimedAt?: string; refs?: Parameters<typeof succeededResultFor>[1] } = {},
): CertificateRecoveryItem {
  return {
    operation_key: key,
    intent: options.intent ?? "issue",
    claimed_at: options.claimedAt ?? "2026-10-20T14:00:01Z",
    discovery_state: "result",
    execution_state: "succeeded",
    receipt_digest: mockHash(`receipt:${key}`),
    result: succeededResultFor(key, options.refs),
  };
}

export function unresolvedItem(
  key: string,
  options: { intent?: "issue" | "reprint"; claimedAt?: string; unavailable?: boolean } = {},
): CertificateRecoveryItem {
  const base = {
    operation_key: key,
    intent: options.intent ?? "issue",
    claimed_at: options.claimedAt ?? "2026-10-20T14:00:02Z",
    receipt_digest: null,
  } as const;
  return options.unavailable === true
    ? { ...base, discovery_state: "unavailable", execution_state: "unresolved", result: null }
    : {
        ...base,
        discovery_state: "result",
        execution_state: "unresolved",
        result: { status: "not-succeeded", operation_key: key },
      };
}

export function closedItem(
  key: string,
  options: { intent?: "issue" | "reprint"; claimedAt?: string } = {},
): CertificateRecoveryItem {
  return {
    operation_key: key,
    intent: options.intent ?? "issue",
    claimed_at: options.claimedAt ?? "2026-10-20T14:00:03Z",
    discovery_state: "result",
    execution_state: "closed_without_effect",
    receipt_digest: null,
    result: { status: "not-succeeded", operation_key: key },
  };
}

export function recoveryPage(
  items: CertificateRecoveryItem[],
  options: { nextCursor?: string | null; generation?: string; active?: string | null } = {},
): CertificateRecoveryQueryResponse {
  const nextCursor = options.nextCursor ?? null;
  return {
    items,
    next_cursor: nextCursor,
    complete: nextCursor === null,
    recovery_generation: options.generation ?? "1",
    active_operation_key: options.active ?? null,
  };
}
