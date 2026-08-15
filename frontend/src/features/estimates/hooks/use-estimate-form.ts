import { useState, useCallback, useActionState } from 'react';
import { useNavigate } from 'react-router';
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
import type { Estimate, EstimateStatus } from '../types';
import { useCreateEstimate } from '../api/create-estimate';
import { useUpdateEstimate } from '../api/update-estimate';
import type { CreateEstimateRequest, UpdateEstimateRequest } from '../api/types';
import { CREATE_ALLOWED_STATUSES } from '../constants/estimate-status-options';
import {
  ESTIMATE_LOCKED_EDIT_MESSAGE,
  isEstimateLockedStatus,
} from '../lib/is-estimate-locked-status';

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
  fieldErrors?: Record<string, string>;
}

export interface UseEstimateFormArgs {
  /**
   * Route-derived mode. Must NOT be inferred from fetched object truthiness
   * (BUG-019: missing ID must not fall through to create-like blank edit).
   */
  mode: "create" | "edit";
  /** Present only when the edit entity was successfully loaded. */
  estimate?: Estimate;
}

export function useEstimateForm(args: UseEstimateFormArgs = { mode: "create" }) {
  const navigate = useNavigate();
  // BUG-019: mode is route-param driven, not `!!estimate`
  const isEdit = args.mode === "edit";
  const estimate = args.estimate;

  const [form, setForm] = useState<EstimateFormState>(() => buildInitialState(estimate));

  // Sync with estimate data if it loads later — previous-value pattern
  const [prevEstimateId, setPrevEstimateId] = useState(estimate?.id);
  if (prevEstimateId !== estimate?.id) {
    setPrevEstimateId(estimate?.id);
    if (estimate) {
      setForm(buildInitialState(estimate));
    }
  }

  const { mutateAsync: createEstimate } = useCreateEstimate();
  const { mutateAsync: updateEstimate } = useUpdateEstimate();

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      if (!form.title.trim()) {
        return { success: false, fieldErrors: { title: "タイトルを入力してください" }, timestamp: Date.now() };
      }

      try {
        if (isEdit) {
          // BUG-019: edit mode without a found estimate must not create/update
          if (!estimate) {
            return { success: false, timestamp: Date.now() };
          }
          if (isEstimateLockedStatus(estimate.status)) {
            toast.info(ESTIMATE_LOCKED_EDIT_MESSAGE);
            return { success: false, timestamp: Date.now() };
          }
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
          if (!CREATE_ALLOWED_STATUSES.includes(form.status)) {
            return {
              success: false,
              fieldErrors: { status: "作成時は下書きまたは送付済みのみ選択できます" },
              timestamp: Date.now(),
            };
          }
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

  return { form, handleChange, formAction, formState, handleCancel, isPending, isEdit };
}
