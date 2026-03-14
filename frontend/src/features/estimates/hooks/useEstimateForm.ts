import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import type { Estimate, EstimateStatus } from '../types';
import { useCreateEstimate } from '../api/create-estimate';
import { useUpdateEstimate } from '../api/update-estimate';
import type { CreateEstimateRequest, UpdateEstimateRequest } from '../api/types';

interface EstimateFormState {
  title: string;
  status: EstimateStatus;
  ownerId: string;
  medicalRecordId: string;
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  insuranceAmount: number;
  discountAmount: number;
  validUntil: string;
  comment: string;
  notes: string;
}

function buildInitialState(estimate?: Estimate): EstimateFormState {
  return {
    title: estimate?.title ?? '',
    status: estimate?.status ?? 'draft',
    ownerId: estimate?.ownerId ?? '',
    medicalRecordId: estimate?.medicalRecordId ?? '',
    subtotal: estimate?.subtotal ?? 0,
    taxTotal: estimate?.taxTotal ?? 0,
    totalAmount: estimate?.totalAmount ?? 0,
    insuranceAmount: estimate?.insuranceAmount ?? 0,
    discountAmount: estimate?.discountAmount ?? 0,
    validUntil: estimate?.validUntil ?? '',
    comment: estimate?.comment ?? '',
    notes: estimate?.notes ?? '',
  };
}

export function useEstimateForm(estimate?: Estimate) {
  const navigate = useNavigate();
  const isEdit = !!estimate;

  const [form, setForm] = useState<EstimateFormState>(() => buildInitialState(estimate));
  const [isPending, setIsPending] = useState(false);

  const { mutateAsync: createEstimate } = useCreateEstimate();
  const { mutateAsync: updateEstimate } = useUpdateEstimate();

  const handleChange = useCallback(<K extends keyof EstimateFormState>(
    field: K,
    value: EstimateFormState[K],
  ) => {
    setForm(prev => ({ ...prev, [field]: value }));
  }, []);

  const handleSubmit = useCallback(async () => {
    if (!form.title.trim()) return;

    setIsPending(true);
    try {
      if (isEdit && estimate) {
        const req: UpdateEstimateRequest = {
          title: form.title,
          status: form.status,
          subtotal: form.subtotal,
          tax_total: form.taxTotal,
          total_amount: form.totalAmount,
          insurance_amount: form.insuranceAmount,
          discount_amount: form.discountAmount,
          valid_until: form.validUntil || null,
          comment: form.comment,
          notes: form.notes,
        };
        await updateEstimate({ id: estimate.id, data: req });
        navigate(`/estimates/${estimate.id}`);
      } else {
        const req: CreateEstimateRequest = {
          title: form.title,
          status: form.status,
          owner_id: form.ownerId ? Number(form.ownerId) : null,
          medical_record_id: form.medicalRecordId ? Number(form.medicalRecordId) : null,
          subtotal: form.subtotal,
          tax_total: form.taxTotal,
          total_amount: form.totalAmount,
          insurance_amount: form.insuranceAmount,
          discount_amount: form.discountAmount,
          valid_until: form.validUntil || null,
          comment: form.comment,
          notes: form.notes,
        };
        const created = await createEstimate(req);
        navigate(`/estimates/${created.id}`);
      }
    } finally {
      setIsPending(false);
    }
  }, [form, isEdit, estimate, createEstimate, updateEstimate, navigate]);

  const handleCancel = useCallback(() => {
    if (isEdit && estimate) {
      navigate(`/estimates/${estimate.id}`);
    } else {
      navigate('/estimates');
    }
  }, [isEdit, estimate, navigate]);

  return { form, handleChange, handleSubmit, handleCancel, isPending };
}
