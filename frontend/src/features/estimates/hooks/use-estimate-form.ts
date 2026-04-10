import { useState, useCallback, useActionState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
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

interface FormState {
  success: boolean;
  timestamp: number;
}

export function useEstimateForm(estimate?: Estimate) {
  const navigate = useNavigate();
  const isEdit = !!estimate;

  const [form, setForm] = useState<EstimateFormState>(() => buildInitialState(estimate));

  // Sync with estimate data if it loads later
  // rerender-dependencies: estimate（オブジェクト）の代わりに estimate?.id（primitive）を deps に使用
  // estimate は id 変更時点で必ず最新参照のため、exhaustive-deps 警告を抑制
  useEffect(() => {
    if (estimate) {
      setForm(buildInitialState(estimate));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- estimate?.id で変更を検知。id が変わる時 estimate は必ず最新
  }, [estimate?.id]);

  const { mutateAsync: createEstimate } = useCreateEstimate();
  const { mutateAsync: updateEstimate } = useUpdateEstimate();

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      if (!form.title.trim()) {
        toast.error("タイトルを入力してください");
        return { success: false, timestamp: Date.now() };
      }

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
          toast.success("見積書を更新しました");
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
          await createEstimate(req);
          toast.success("見積書を作成しました");
          // Navigation is handled in the component via useEffect
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, isEdit ? "更新" : "作成");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  const handleChange = useCallback(<K extends keyof EstimateFormState>(
    field: K,
    value: EstimateFormState[K],
  ) => {
    setForm(prev => ({ ...prev, [field]: value }));
  }, []);

  const handleCancel = useCallback(() => {
    if (isEdit && estimate) {
      navigate(paths.estimates.detail.getHref(estimate.id));
    } else {
      navigate(paths.estimates.getHref());
    }
  }, [isEdit, estimate, navigate]);

  return { form, handleChange, formAction, formState, handleCancel, isPending };
}
