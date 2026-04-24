import { useState, useEffect, useRef, useTransition, useCallback, useActionState } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateClinicalPlan } from "../api/clinical-plan";
import type { UpdateMedicalRecordRequest } from "../api/types";
import type { TreatmentItem } from "../components/TreatmentTable";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

const DEFAULT_CHIEF_COMPLAINT = "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
const DEFAULT_TREATMENT_POLICY = "# 治療方針";
const DEFAULT_PLAN = "# 治療方針";
const DEFAULT_ASSESSMENT = "# 診断詳細";

// rendering-hoist-jsx: アクセシビリティ用定数をモジュールレベルに巻き上げ（毎レンダー再生成を回避）
const MEDICAL_RECORD_PRIORITY_FIELDS = ["treatment_policy", "diagnosis1_category_id"] as const;

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;
  const { canEdit } = usePermission("medical-records");

  const [activeTab, setActiveTab] = useState("問診");
  const [visitType, setVisitType] = useState("再診");
  const [isCreating, startCreateTransition] = useTransition();
  const [manualErrors, setManualErrors] = useState<Record<string, string>>({});

  // --- Focus Management (Accessibility) ---
  // Tab switching: previous-value pattern (no side effects during render)
  const [prevManualErrors, setPrevManualErrors] = useState(manualErrors);
  if (prevManualErrors !== manualErrors) {
    setPrevManualErrors(manualErrors);
    const errorFields = Object.keys(manualErrors);
    if (errorFields.length > 0) {
      const firstError = MEDICAL_RECORD_PRIORITY_FIELDS.find(f => errorFields.includes(f)) || errorFields[0];
      if (["treatment_policy"].includes(firstError)) {
        setActiveTab("問診");
      } else if (["diagnosis1_category_id", "diagnosis1_name_id"].includes(firstError)) {
        setActiveTab("診察/治療プラン");
      }
    }
  }

  // DOM focus side effect (no setState)
  useEffect(() => {
    const errorFields = Object.keys(manualErrors);
    if (errorFields.length === 0) return;
    const firstError = MEDICAL_RECORD_PRIORITY_FIELDS.find(f => errorFields.includes(f)) || errorFields[0];
    setTimeout(() => {
      const element = document.getElementById(firstError);
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 50);
  }, [manualErrors]);

  const hasAutoCreatedRef = useRef(false);
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [chiefComplaintTypeId, setChiefComplaintTypeId] = useState<number | null>(null);
  const [treatmentPolicy, setTreatmentPolicy] = useState(DEFAULT_TREATMENT_POLICY);

  // 診察/治療プランタブの状態（SOAPS）
  const [plan, setPlan] = useState(DEFAULT_PLAN);
  const [assessment, setAssessment] = useState(DEFAULT_ASSESSMENT);

  // 診断マスタの状態
  const [diagnosis1CategoryId, setDiagnosis1CategoryId] = useState<number | null>(null);
  const [diagnosis1NameId, setDiagnosis1NameId] = useState<number | null>(null);
  const [diagnosis2CategoryId, setDiagnosis2CategoryId] = useState<number | null>(null);
  const [diagnosis2NameId, setDiagnosis2NameId] = useState<number | null>(null);

  // --- Draft Persistence (Local Storage) ---
  const DRAFT_KEY = `medical-record-draft-${recordId}`;

  // Load draft on mount
  useEffect(() => {
    if (!recordId) return;
    const saved = localStorage.getItem(DRAFT_KEY);
    if (saved) {
      try {
        const draft = JSON.parse(saved);
        /* eslint-disable react-hooks/set-state-in-effect */
        if (draft.chiefComplaint) setChiefComplaint(draft.chiefComplaint);
        if (draft.treatmentPolicy) setTreatmentPolicy(draft.treatmentPolicy);
        if (draft.plan) setPlan(draft.plan);
        if (draft.assessment) setAssessment(draft.assessment);
        if (draft.visitType) setVisitType(draft.visitType);
        if (draft.diagnosis1CategoryId) setDiagnosis1CategoryId(draft.diagnosis1CategoryId);
        if (draft.diagnosis1NameId) setDiagnosis1NameId(draft.diagnosis1NameId);
        if (draft.diagnosis2CategoryId) setDiagnosis2CategoryId(draft.diagnosis2CategoryId);
        if (draft.diagnosis2NameId) setDiagnosis2NameId(draft.diagnosis2NameId);
        /* eslint-enable react-hooks/set-state-in-effect */
        toast.info("未保存の下書きを復元しました", { duration: 2000 });
      } catch {
        // localStorage の下書きが破損している場合は静かにスキップ（復元失敗は非致命的）
        localStorage.removeItem(DRAFT_KEY);
      }
    }
  }, [recordId, DRAFT_KEY]);

  // Save draft on changes
  useEffect(() => {
    if (!recordId) return;
    const draft = { 
      chiefComplaint, treatmentPolicy, plan, assessment, 
      visitType,
      diagnosis1CategoryId, diagnosis1NameId,
      diagnosis2CategoryId, diagnosis2NameId 
    };
    localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
  }, [
    recordId, DRAFT_KEY, chiefComplaint, treatmentPolicy, plan, assessment, 
    visitType, diagnosis1CategoryId, diagnosis1NameId, diagnosis2CategoryId, diagnosis2NameId
  ]);

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord } = useGetMedicalRecord(recordId ?? "");

  // 既存カルテデータをフォームに反映 — previous-value pattern
  const [prevExistingRecord, setPrevExistingRecord] = useState(existingRecord);
  if (prevExistingRecord !== existingRecord && existingRecord) {
    setPrevExistingRecord(existingRecord);
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
  }

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { data: selectedPet, isLoading: isPetLoading } = useGetPet(resolvedPetId);

  // Ownerデータを取得（飼主割引率用）
  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { data: owner } = useGetOwner(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;

  const queryClient = useQueryClient();
  const createMutation = useCreateMedicalRecord();
  const updateMutation = useUpdateMedicalRecord();
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateClinicalPlan(recordId ?? "");

  // useTransition: save の pending 管理 (rerender-transitions)
  const [isSavingTransition, startSaveTransition] = useTransition();

  // activeTab を保存時に正確に参照するための ref
  const activeTabRef = useRef(activeTab);
  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  /**
   * React 19 useActionState を使用したタブ別保存アクション
   */
  const [formState, formAction, isSaving] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (!recordId) return { success: false, timestamp: Date.now() };

      try {
        setManualErrors({});
        const currentTab = activeTabRef.current;

        switch (currentTab) {
          case "問診":
            await updateInquiryMutation.mutateAsync({
              chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
              chief_complaint_type_id: chiefComplaintTypeId,
              notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
            });
            break;

          case "診察/治療プラン": {
            if (!canEdit) break;
            if (diagnosis1CategoryId && !diagnosis1NameId) {
              const diagError = { diagnosis1_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-102: DEFAULT値でも常に送信する（undefined を送ると BE が 400 を返す）
            const treatmentPlanPayload = {
              treatment_policy: plan,
              diagnosis_details: assessment,
              diagnosis_type_id: diagnosis1CategoryId ?? undefined,
              diagnosis_name_id: diagnosis1NameId ?? undefined,
              diagnosis_2_type_id: diagnosis2CategoryId ?? undefined,
              diagnosis_2_name_id: diagnosis2NameId ?? undefined,
            };
            await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);
            break;
          }

          default:
            break;
        }

        localStorage.removeItem(DRAFT_KEY);
        toast.success("保存しました");
        queryClient.invalidateQueries({ queryKey: ["reception"] });
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  const handleBack = useCallback(() => {
    if (location.state?.from) {
      navigate(location.state.from);
      return;
    }

    if (!recordId) {
      navigate(paths.medicalRecords.selectPet.getHref());
    } else {
      navigate(paths.medicalRecords.getHref());
    }
  }, [location.state, navigate, recordId]);

  // 担当医変更ハンドラ
  const handleChangeDoctor = (newDoctorId: string, newDoctorName: string) => {
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            doctor_id: Number(newDoctorId),
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`担当医を ${newDoctorName} に変更しました`);
      } catch (error) {
        handleApiError(error, "担当医変更");
      }
    });
  };

  // BUG-373: 飼主変更 — 確認モーダル経由で実行
  const [pendingOwnerChange, setPendingOwnerChange] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const requestOwnerChange = useCallback(
    (newOwner: { id: string; name: string }) => {
      setPendingOwnerChange(newOwner);
    },
    [],
  );

  const confirmOwnerChange = () => {
    if (!pendingOwnerChange || !recordId) return;
    const newOwner = pendingOwnerChange;
    setPendingOwnerChange(null);
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            owner_id: Number(newOwner.id),
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`飼主を ${newOwner.name} に変更しました`);
      } catch (error) {
        handleApiError(error, "飼主変更");
      }
    });
  };

  const cancelOwnerChange = useCallback(() => setPendingOwnerChange(null), []);

  // 新規作成時: ページ表示と同時にカルテを自動作成
  useEffect(() => {
    if (!isNewRecord || !selectedPet || hasAutoCreatedRef.current) return;
    hasAutoCreatedRef.current = true;

    startCreateTransition(async () => {
      try {
        const today = new Date().toISOString().split("T")[0];
        const record = await createMutation.mutateAsync({
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: visitType,
          status: "draft",
        });
        navigate(paths.medicalRecords.detail.getHref(record.id), { replace: true });
      } catch (error) {
        handleApiError(error, "カルテ作成");
        hasAutoCreatedRef.current = false;
      }
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional: run only when isNewRecord or petId changes; createMutation/navigate/visitType are stable references
  }, [isNewRecord, selectedPet?.id]);

  const shouldRedirectToSelectPet = isNewRecord && !petId;

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    visitType,
    setVisitType,
    selectedPet: selectedPet ?? null,
    isPetLoading,
    shouldRedirectToSelectPet,
    handleBack,
    formAction,
    formState,
    isSaving: isSaving || isSavingTransition,
    isCreating,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems,
    // 問診タブ
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    // 診察/治療プランタブ（SOAPS）
    plan,
    setPlan,
    assessment,
    setAssessment,
    // 診断マスタ
    diagnosis1CategoryId,
    setDiagnosis1CategoryId,
    diagnosis1NameId,
    setDiagnosis1NameId,
    diagnosis2CategoryId,
    setDiagnosis2CategoryId,
    diagnosis2NameId,
    setDiagnosis2NameId,
    // 飼主割引率
    ownerDiscountRate,
    // 医療記録
    visitCount: existingRecord?.visitCount,
    // 担当医変更
    handleChangeDoctor,
    // 飼主変更
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
    fieldErrors: manualErrors,
  };
}
