import type { EstimateStatus } from "../types";

/**
 * Pure offer/reason policy for estimate successor draft (S07 TASK-012).
 * No React / no axios — unit-testable without app path aliases.
 */

export function shouldOfferEstimateSuccessor(params: {
  canCreate: boolean;
  status: EstimateStatus | string;
}): boolean {
  if (!params.canCreate) {
    return false;
  }
  return params.status === "approved" || params.status === "rejected";
}

/**
 * Mirrors BE reason binding: trimmed non-empty string, Unicode length 1..500.
 */
export function isValidSuccessorReason(reason: unknown): boolean {
  if (typeof reason !== "string") {
    return false;
  }
  const trimmed = reason.trim();
  const len = Array.from(trimmed).length;
  return len >= 1 && len <= 500;
}
