import { useState, useCallback, useActionState, useEffect, useLayoutEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { INITIAL_ACTION_STATE, type ActionState } from "@/types/form";
import type { Estimate, EstimateStatus } from "../types";
import { useCreateEstimate } from "../api/create-estimate";
import { useUpdateEstimate } from "../api/update-estimate";
import type { CreateEstimateRequest, UpdateEstimateRequest } from "../api/types";
import { CREATE_ALLOWED_STATUSES } from "../constants/estimate-status-options";
import {
  ESTIMATE_LOCKED_EDIT_MESSAGE,
  isEstimateLockedStatus,
} from "../lib/is-estimate-locked-status";

interface EstimateFormState {
  title: string;
  status: EstimateStatus;
  ownerId: string;
  petId: string;
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
    title: estimate?.title ?? "",
    status: estimate?.status ?? "draft",
    ownerId: estimate?.ownerId ?? "",
    petId: estimate?.petId ?? "",
    medicalRecordId: estimate?.medicalRecordId ?? "",
    subtotal: estimate?.subtotal ?? 0,
    taxTotal: estimate?.taxTotal ?? 0,
    totalAmount: estimate?.totalAmount ?? 0,
    insuranceAmount: estimate?.insuranceAmount ?? 0,
    discountAmount: estimate?.discountAmount ?? 0,
    validUntil: estimate?.validUntil ?? "",
    comment: estimate?.comment ?? "",
    notes: estimate?.notes ?? "",
  };
}

/** FE-RC-001: action 別の最新権限値。mutation 直前に isMutationAllowed() で再検査する。 */
export interface EstimateMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
}

const DENIED_ESTIMATE_MUTATION_PERMISSIONS: Readonly<EstimateMutationPermissions> = {
  canCreate: false,
  canEdit: false,
};

export interface UseEstimateFormArgs {
  /**
   * Route-derived mode. Must NOT be inferred from fetched object truthiness
   * (BUG-019: missing ID must not fall through to create-like blank edit).
   */
  mode: "create" | "edit";
  /** Present only when the edit entity was successfully loaded. */
  estimate?: Estimate;
  /** `/estimates/new?petId=` から解決した飼主・ペット（BUG-009）。 */
  initialOwnerId?: string;
  initialPetId?: string;
  /** FE-RC-001: action 別の最新権限値（mutation 直前の再検査に使用）。 */
  permissions?: Readonly<EstimateMutationPermissions>;
  /**
   * FE-RC-002/004: 死亡ペットへの新規見積書作成を fail-closed で拒否する理由。
   * render 側（fieldset disabled + banner）と同じ判定を callback 側でも再検査する（二重防壁）。
   */
  blockCreateReason?: string;
}

export function useEstimateForm(args: UseEstimateFormArgs = { mode: "create" }) {
  const navigate = useNavigate();
  // BUG-019: mode is route-param driven, not `!!estimate`
  const isEdit = args.mode === "edit";
  const estimate = args.estimate;

  const [form, setForm] = useState<EstimateFormState>(() => buildInitialState(estimate));
  const formRef = useRef(form);
  useLayoutEffect(() => {
    formRef.current = form;
  }, [form]);

  const permissions = args.permissions ?? DENIED_ESTIMATE_MUTATION_PERMISSIONS;
  const { canCreate, canEdit } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit };
  }, [canCreate, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof EstimateMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  const blockCreateReasonRef = useRef(args.blockCreateReason);
  useLayoutEffect(() => {
    blockCreateReasonRef.current = args.blockCreateReason;
  }, [args.blockCreateReason]);

  useEffect(() => {
    if (isEdit) return;
    const ownerId = args.initialOwnerId;
    const petId = args.initialPetId;
    if (!ownerId && !petId) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 新規作成時に URL の owner/pet を初期値へ同期
    setForm((prev) => ({
      ...prev,
      ownerId: prev.ownerId || ownerId || "",
      petId: prev.petId || petId || "",
    }));
  }, [args.initialOwnerId, args.initialPetId, isEdit]);

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
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const current = formRef.current;
      if (!current.title.trim()) {
        return {
          success: false,
          fieldErrors: { title: "タイトルを入力してください" },
          timestamp: Date.now(),
        };
      }

      try {
        if (isEdit) {
          // FE-RC-001: UI の disabled/非表示だけを最終防壁にせず、mutation 直前に最新権限を再検査する。
          if (!isMutationAllowed("canEdit")) {
            return { success: false, timestamp: Date.now() };
          }
          // BUG-019: edit mode without a found estimate must not create/update
          if (!estimate) {
            return { success: false, timestamp: Date.now() };
          }
          if (isEstimateLockedStatus(estimate.status)) {
            toast.info(ESTIMATE_LOCKED_EDIT_MESSAGE);
            return { success: false, timestamp: Date.now() };
          }
          const req: UpdateEstimateRequest = {
            title: current.title,
            status: current.status,
            subtotal: current.subtotal,
            tax_total: current.taxTotal,
            total_amount: current.totalAmount,
            insurance_amount: current.insuranceAmount,
            discount_amount: current.discountAmount,
            valid_until: current.validUntil || null,
            comment: current.comment,
            notes: current.notes,
          };
          // FE-RC-005: 成功 toast は useUpdateEstimate の onSuccess に一元化済み（ここでは通知しない）。
          await updateEstimate({ id: estimate.id, data: req });
        } else {
          // FE-RC-001: 新規作成も同様に mutation 直前で権限を再検査する。
          if (!isMutationAllowed("canCreate")) {
            return { success: false, timestamp: Date.now() };
          }
          // FE-RC-002/004: 死亡ペットの新規見積書作成は callback 側でも fail-closed に拒否する（二重防壁）。
          if (blockCreateReasonRef.current) {
            toast.error(blockCreateReasonRef.current);
            return { success: false, timestamp: Date.now() };
          }
          if (!CREATE_ALLOWED_STATUSES.includes(current.status)) {
            return {
              success: false,
              fieldErrors: { status: "作成時は下書きまたは送付済みのみ選択できます" },
              timestamp: Date.now(),
            };
          }
          const req: CreateEstimateRequest = {
            title: current.title,
            status: current.status,
            owner_id: current.ownerId ? Number(current.ownerId) : null,
            pet_id: current.petId ? Number(current.petId) : null,
            medical_record_id: current.medicalRecordId ? Number(current.medicalRecordId) : null,
            subtotal: current.subtotal,
            tax_total: current.taxTotal,
            total_amount: current.totalAmount,
            insurance_amount: current.insuranceAmount,
            discount_amount: current.discountAmount,
            valid_until: current.validUntil || null,
            comment: current.comment,
            notes: current.notes,
          };
          // FE-RC-005: 成功 toast は useCreateEstimate の onSuccess に一元化済み（ここでは通知しない）。
          await createEstimate(req);
          // Navigation is handled in the component via useEffect
        }
        return { success: true, timestamp: Date.now() };
      } catch {
        // FE-RC-005: useCreateEstimate/useUpdateEstimate の onError が既に handleApiError で通知済み。
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const handleChange = useCallback(
    <K extends keyof EstimateFormState>(field: K, value: EstimateFormState[K]) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleCancel = useCallback(() => {
    if (isEdit && estimate) {
      navigate(paths.estimates.detail.getHref(estimate.id));
    } else {
      navigate(paths.estimates.getHref());
    }
  }, [isEdit, estimate, navigate]);

  return { form, handleChange, formAction, formState, handleCancel, isPending, isEdit };
}
