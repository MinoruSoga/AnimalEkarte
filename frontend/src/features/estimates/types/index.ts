/**
 * Estimates feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Estimate as BackendEstimate}, {@link EstimateItem as BackendEstimateItem} from models.ts
 */
import {
  EstimateStatusDraft,
  EstimateStatusSent,
  EstimateStatusApproved,
  EstimateStatusRejected,
} from "@/types/generated/models";

/** @see {@link import("@/types/generated/models").EstimateStatus} */
export type EstimateStatus =
  | typeof EstimateStatusDraft
  | typeof EstimateStatusSent
  | typeof EstimateStatusApproved
  | typeof EstimateStatusRejected;

export type { Estimate, EstimateLineItem } from "../api/transforms";
