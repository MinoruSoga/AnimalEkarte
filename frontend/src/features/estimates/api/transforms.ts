import type { BackendEstimate, BackendEstimateItem } from './types';
import type { EstimateStatus } from '../types';

function transformEstimateItem(item: BackendEstimateItem) {
  return {
    id: String(item.id ?? 0),
    estimateId: String(item.estimate_id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price ?? 0,
    quantity: item.quantity ?? 1,
    taxType: item.tax_type ?? "excluded",
    taxRate: item.tax_rate ?? 0.1,
    discountRate: item.discount_rate ?? 0,
    discountAmount: item.discount_amount ?? 0,
    isInsuranceApplicable: item.is_insurance_applicable,
    sortOrder: item.sort_order ?? 0,
    createdAt: item.created_at,
  };
}

export function transformEstimate(data: BackendEstimate) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    estimateNo: data.estimate_no ?? '',
    medicalRecordId: data.medical_record_id != null ? String(data.medical_record_id) : null,
    title: data.title ?? '',
    ownerId: data.owner_id != null ? String(data.owner_id) : null,
    ownerName: data.owner?.name ?? undefined,
    petId: data.pet_id != null ? String(data.pet_id) : null,
    petName: data.pet?.name ?? undefined,
    status: (data.status ?? 'draft') as EstimateStatus,
    subtotal: data.subtotal ?? 0,
    taxTotal: data.tax_total ?? 0,
    totalAmount: data.total_amount ?? 0,
    insuranceAmount: data.insurance_amount ?? 0,
    discountAmount: data.discount_amount ?? 0,
    validUntil: data.valid_until ?? null,
    comment: data.comment ?? '',
    notes: data.notes ?? '',
    createdBy: data.created_by != null ? String(data.created_by) : null,
    items: (data.items ?? []).map(transformEstimateItem),
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type EstimateLineItem = ReturnType<typeof transformEstimateItem>;
export type Estimate = ReturnType<typeof transformEstimate>;
