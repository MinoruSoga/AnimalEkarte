export type EstimateStatus = 'draft' | 'sent' | 'approved' | 'rejected';

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
