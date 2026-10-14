import {
  useState,
  useCallback,
  useActionState,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
} from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { INITIAL_ACTION_STATE, type ActionState } from "@/types/form";
import { calcEstimateTotals } from "@/lib/estimate-totals";
import type { Estimate, EstimateLineItem, EstimateStatus } from "../types";
import { useCreateEstimate } from "../api/create-estimate";
import { useUpdateEstimate } from "../api/update-estimate";
import type {
  CreateEstimateRequest,
  EstimateItemRequest,
  UpdateEstimateRequest,
} from "../api/types";
import { CREATE_ALLOWED_STATUSES } from "../constants/estimate-status-options";
import {
  ESTIMATE_LOCKED_EDIT_MESSAGE,
  isEstimateLockedStatus,
} from "../lib/is-estimate-locked-status";

export type EstimateItemDraft = Pick<
  EstimateLineItem,
  | "id"
  | "name"
  | "category"
  | "unitPrice"
  | "quantity"
  | "taxType"
  | "taxRate"
  | "discountRate"
  | "discountAmount"
  | "isInsuranceApplicable"
  | "sortOrder"
> & {
  localId: string;
};

let localIdSeq = 0;
const nextLocalId = () => `local-item-${++localIdSeq}`;

function itemToDraft(item: EstimateLineItem): EstimateItemDraft {
  return {
    id: item.id,
    localId: item.id ? `saved-${item.id}` : nextLocalId(),
    name: item.name,
    category: item.category,
    unitPrice: item.unitPrice,
    quantity: item.quantity,
    taxType: item.taxType,
    taxRate: item.taxRate,
    discountRate: item.discountRate,
    discountAmount: item.discountAmount,
    isInsuranceApplicable: item.isInsuranceApplicable,
    sortOrder: item.sortOrder,
  };
}

function deriveTotals(items: EstimateItemDraft[]) {
  const t = calcEstimateTotals(items);
  return { subtotal: t.subtotal, taxTotal: t.tax, totalAmount: t.total };
}

function serializeItems(items: EstimateItemDraft[]): string {
  return JSON.stringify(items.map((item, index) => draftToRequestItem(item, index)));
}

function draftToRequestItem(item: EstimateItemDraft, index: number): EstimateItemRequest {
  const persistedId = Number(item.id);
  const request: EstimateItemRequest = {
    name: item.name.trim(),
    category: item.category,
    unit_price: item.unitPrice,
    quantity: item.quantity,
    tax_type: item.taxType,
    tax_rate: item.taxRate,
    discount_rate: item.discountRate,
    discount_amount: item.discountAmount,
    is_insurance_applicable: item.isInsuranceApplicable,
    sort_order: index,
  };
  // PR665: 既存行は安定 ID を一致ヒントとして送り、改名・並べ替えでも
  // サーバ側が同一行として割引権限・既存税設定を判定できるようにする。
  if (item.id !== "" && Number.isSafeInteger(persistedId) && persistedId > 0) {
    request.existing_item_id = persistedId;
  }
  return request;
}

function validateItems(items: EstimateItemDraft[]): string | null {
  for (let i = 0; i < items.length; i++) {
    const item = items[i]!;
    // PR665: 置換送信時は未変更行（旧データの不正行を含む）も全行検証する。
    // メタデータのみの保存（items 未送信）は検証対象外。
    const label = `${i + 1}行目`;
    const name = item.name.trim();
    if (!name) return `${label}: 品目名を入力してください`;
    if (name.length > 255) return `${label}: 品目名は255文字以内で入力してください`;
    if (!Number.isSafeInteger(item.unitPrice) || item.unitPrice < 0) {
      return `${label}: 単価は0以上の整数で入力してください`;
    }
    if (
      !Number.isFinite(item.quantity) ||
      item.quantity <= 0 ||
      item.quantity > 999999999.9 ||
      Math.abs(item.quantity * 10 - Math.round(item.quantity * 10)) > 1e-9
    ) {
      return `${label}: 数量は0より大きい0.1刻みの値（999999999.9以下）で入力してください`;
    }
    if (item.taxType !== "included" && item.taxType !== "excluded" && item.taxType !== "exempt") {
      return `${label}: 課税区分が不正です`;
    }
    if (
      !Number.isFinite(item.taxRate) ||
      item.taxRate < 0 ||
      item.taxRate > 1 ||
      Math.abs(item.taxRate * 100 - Math.round(item.taxRate * 100)) > 1e-9
    ) {
      return `${label}: 税率は0〜1の小数第2位までの値で入力してください`;
    }
    if (!Number.isFinite(item.discountRate) || item.discountRate < 0 || item.discountRate > 100) {
      return `${label}: 割引率は0〜100で入力してください`;
    }
    if (!Number.isSafeInteger(item.discountAmount) || item.discountAmount < 0) {
      return `${label}: 値引額は0以上の整数で入力してください`;
    }
    const base = item.unitPrice * item.quantity;
    if (!Number.isFinite(base) || base > Number.MAX_SAFE_INTEGER) {
      return `${label}: 単価×数量が上限を超えています`;
    }
    if (item.discountAmount > base) {
      return `${label}: 値引額は単価×数量以下で入力してください`;
    }
  }
  const totals = calcEstimateTotals(items);
  if (!Number.isSafeInteger(totals.subtotal) || totals.subtotal < 0) {
    return "小計が安全な金額範囲を超えています";
  }
  if (!Number.isSafeInteger(totals.tax) || totals.tax < 0) {
    return "税額合計が安全な金額範囲を超えています";
  }
  if (!Number.isSafeInteger(totals.total) || totals.total < 0) {
    return "合計が安全な金額範囲を超えています";
  }
  return null;
}

interface EstimateFormState {
  title: string;
  status: EstimateStatus;
  ownerId: string;
  petId: string;
  medicalRecordId: string;
  insuranceAmount: number;
  discountAmount: number;
  validUntil: string;
  comment: string;
  notes: string;
  items: EstimateItemDraft[];
}

function buildInitialState(estimate?: Estimate): EstimateFormState {
  return {
    title: estimate?.title ?? "",
    status: estimate?.status ?? "draft",
    ownerId: estimate?.ownerId ?? "",
    petId: estimate?.petId ?? "",
    medicalRecordId: estimate?.medicalRecordId ?? "",
    insuranceAmount: estimate?.insuranceAmount ?? 0,
    discountAmount: estimate?.discountAmount ?? 0,
    validUntil: estimate?.validUntil ?? "",
    comment: estimate?.comment ?? "",
    notes: estimate?.notes ?? "",
    items: (estimate?.items ?? []).map(itemToDraft),
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
  standardTaxRate?: number;
}

export function useEstimateForm(args: UseEstimateFormArgs = { mode: "create" }) {
  const navigate = useNavigate();
  // BUG-019: mode is route-param driven, not `!!estimate`
  const isEdit = args.mode === "edit";
  const estimate = args.estimate;

  const standardTaxRate = args.standardTaxRate ?? 0.1;

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

  const [baseline, setBaseline] = useState(() => ({
    itemsKey: serializeItems(form.items),
  }));

  // Sync with estimate data if it loads later — previous-value pattern
  const [prevEstimateId, setPrevEstimateId] = useState(estimate?.id);
  if (prevEstimateId !== estimate?.id) {
    setPrevEstimateId(estimate?.id);
    if (estimate) {
      const hydratedForm = buildInitialState(estimate);
      setBaseline({
        itemsKey: serializeItems(hydratedForm.items),
      });
      setForm(hydratedForm);
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

      const itemsChanged = !isEdit || serializeItems(current.items) !== baseline.itemsKey;
      if (itemsChanged) {
        const itemsError = validateItems(current.items);
        if (itemsError) {
          return {
            success: false,
            fieldErrors: { items: itemsError },
            timestamp: Date.now(),
          };
        }
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
            insurance_amount: current.insuranceAmount,
            discount_amount: current.discountAmount,
            valid_until: current.validUntil || null,
            comment: current.comment,
            notes: current.notes,
          };
          if (itemsChanged) {
            const derived = deriveTotals(current.items);
            req.subtotal = derived.subtotal;
            req.tax_total = derived.taxTotal;
            req.total_amount = derived.totalAmount;
            req.items = current.items.map(draftToRequestItem);
          }
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
          const derived = deriveTotals(current.items);
          const req: CreateEstimateRequest = {
            title: current.title,
            status: current.status,
            owner_id: current.ownerId ? Number(current.ownerId) : null,
            pet_id: current.petId ? Number(current.petId) : null,
            medical_record_id: current.medicalRecordId ? Number(current.medicalRecordId) : null,
            subtotal: derived.subtotal,
            tax_total: derived.taxTotal,
            total_amount: derived.totalAmount,
            insurance_amount: current.insuranceAmount,
            discount_amount: current.discountAmount,
            valid_until: current.validUntil || null,
            comment: current.comment,
            notes: current.notes,
            items: current.items.map(draftToRequestItem),
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

  const createEmptyItem = useCallback((): EstimateItemDraft => {
    return {
      id: "",
      localId: nextLocalId(),
      name: "",
      category: "other",
      unitPrice: 0,
      quantity: 1,
      taxType: "included",
      taxRate: standardTaxRate,
      discountRate: 0,
      discountAmount: 0,
      isInsuranceApplicable: false,
      sortOrder: 0,
    };
  }, [standardTaxRate]);

  const totals = useMemo(() => {
    if (
      isEdit &&
      estimate &&
      form.items.length === 0 &&
      serializeItems(form.items) === baseline.itemsKey
    ) {
      return {
        subtotal: estimate.subtotal,
        taxTotal: estimate.taxTotal,
        totalAmount: estimate.totalAmount,
      };
    }
    return deriveTotals(form.items);
  }, [isEdit, estimate, form.items, baseline.itemsKey]);

  return {
    form,
    handleChange,
    formAction,
    formState,
    handleCancel,
    isPending,
    isEdit,
    totals,
    createEmptyItem,
  };
}
