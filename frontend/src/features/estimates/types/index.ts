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

/** UI-facing estimate (camelCase, string IDs — post-transform) */
export interface Estimate {
  id: string;
  clinicId: string;
  estimateNo: string;
  medicalRecordId?: string | null;
  title: string;
  ownerId?: string | null;
  ownerName?: string;
  status: EstimateStatus;
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  insuranceAmount: number;
  discountAmount: number;
  validUntil?: string | null;
  comment?: string;
  notes?: string;
  createdBy?: string | null;
  items: EstimateLineItem[];
  createdAt: string;
  updatedAt: string;
}

/** @see {@link import("@/types/generated/models").EstimateItem} */
export interface EstimateLineItem {
  id: string;
  estimateId: string;
  name: string;
  category: string;
  unitPrice: number;
  quantity: number;
  taxRate: number;
  discountRate: number;
  discountAmount: number;
  isInsuranceApplicable: boolean;
  sortOrder: number;
  createdAt: string;
}
