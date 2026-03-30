import { useState, useEffect, useRef, useTransition, useCallback, useActionState } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateTreatmentPlan } from "../api/treatment-plans";
import type { UpdateMedicalRecordRequest } from "../api/types";
import type { TreatmentItem } from "../components/TreatmentTable";

const DEFAULT_CHIEF_COMPLAINT = "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
const DEFAULT_TREATMENT_POLICY = "# 治療方針";
const DEFAULT_PLAN = "# 治療方針";
const DEFAULT_ASSESSMENT = "# 診断詳細";

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;

  const [activeTab, setActiveTab] = useState("問診");
  const [visitType, setVisitType] = useState("再診");
  const [isCreating, setIsCreating] = useState(false);
  const hasAutoCreatedRef = useRef(false);
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [chiefComplaintCategoryId, setChiefComplaintCategoryId] = useState<number | null>(null);
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
        if (draft.chiefComplaint) setChiefComplaint(draft.chiefComplaint);
        if (draft.treatmentPolicy) setTreatmentPolicy(draft.treatmentPolicy);
        if (draft.plan) setPlan(draft.plan);
        if (draft.assessment) setAssessment(draft.assessment);
        if (draft.visitType) setVisitType(draft.visitType);
        if (draft.diagnosis1CategoryId) setDiagnosis1CategoryId(draft.diagnosis1CategoryId);
        if (draft.diagnosis1NameId) setDiagnosis1NameId(draft.diagnosis1NameId);
        if (draft.diagnosis2CategoryId) setDiagnosis2CategoryId(draft.diagnosis2CategoryId);
        if (draft.diagnosis2NameId) setDiagnosis2NameId(draft.diagnosis2NameId);
        toast.info("未保存の下書きを復元しました", { duration: 2000 });
      } catch { /* ignore */ }
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

  // 既存カルテデータをフォームに反映
  useEffect(() => {
    if (!existingRecord) return;
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
  }, [existingRecord]);

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { data: selectedPet, isLoading: isPetLoading } = useGetPet(resolvedPetId);

  // Ownerデータを取得（飼主割引率用）
  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { owner } = useGetOwner(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;

  const createMutation = useCreateMedicalRecord();
  const updateMutation = useUpdateMedicalRecord();
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateTreatmentPlan(recordId ?? "");

  // useTransition: save の pending 管理 (rerender-transitions)
  const [, startSaveTransition] = useTransition();

  interface FormState {
    success: boolean;
    timestamp: number;
  }

  /**
   * React 19 useActionState を使用したタブ別保存アクション
   * 自動作成でカルテは必ず存在するため isNewRecord 分岐なし
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      if (!recordId) return { success: false, timestamp: Date.now() };

      try {
        switch (activeTab) {
          case "問診":
            await updateInquiryMutation.mutateAsync({
              chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
              chief_complaint_category_id: chiefComplaintCategoryId,
              notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
            });
            break;

          case "診察/治療プラン": {
            if (diagnosis1CategoryId && !diagnosis1NameId) {
              toast.error("診断名を選択してください");
              return { success: false, timestamp: Date.now() };
            }
            await updateTreatmentPlanMutation.mutateAsync({
              treatment_policy: plan !== DEFAULT_PLAN ? plan : undefined,
              diagnosis_details: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
              diagnosis_category_id: diagnosis1CategoryId ?? undefined,
              diagnosis_name_id: diagnosis1NameId ?? undefined,
              diagnosis_2_category_id: diagnosis2CategoryId ?? undefined,
              diagnosis_2_name_id: diagnosis2NameId ?? undefined,
            });
            break;
          }

          // 治療・予防接種・定期健診・検査・画像タブは
          // インラインCRUDが行単位で保存済み → formAction では何もしない
          default:
            break;
        }

        localStorage.removeItem(DRAFT_KEY);
        toast.success("保存しました");
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
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
  const handleChangeDoctor = useCallback(
    (newDoctorId: string, newDoctorName: string) => {
      if (!recordId) return;
      startSaveTransition(async () => {
        try {
          await updateMutation.mutateAsync({
            id: recordId,
            req: { doctor_id: Number(newDoctorId) } as UpdateMedicalRecordRequest,
          });
          toast.success(`担当医を ${newDoctorName} に変更しました`);
        } catch (error) {
          handleApiError(error, "担当医変更");
        }
      });
    },
    [recordId, updateMutation, startSaveTransition],
  );

  // 飼主変更ハンドラ
  const handleChangeOwner = useCallback(
    (newOwner: { id: string; name: string }) => {
      if (!recordId) return;
      startSaveTransition(async () => {
        try {
          await updateMutation.mutateAsync({
            id: recordId,
            req: { owner_id: Number(newOwner.id) } as UpdateMedicalRecordRequest,
          });
          toast.success(`飼主を ${newOwner.name} に変更しました`);
        } catch (error) {
          handleApiError(error, "飼主変更");
        }
      });
    },
    [recordId, updateMutation, startSaveTransition],
  );

  // 新規作成時: ページ表示と同時にカルテを自動作成して編集URLに置換ナビゲーション
  useEffect(() => {
    if (!isNewRecord || !selectedPet || hasAutoCreatedRef.current) return;
    hasAutoCreatedRef.current = true;

    const autoCreate = async () => {
      setIsCreating(true);
      try {
        const today = new Date().toISOString().split("T")[0];
        const record = await createMutation.mutateAsync({
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: visitType,
          status: "draft",
        });
        // 編集URLに置換ナビゲーション（ブラウザ履歴を汚さない）
        navigate(paths.medicalRecords.detail.getHref(record.id), { replace: true });
      } catch (error) {
        handleApiError(error, "カルテ作成");
        hasAutoCreatedRef.current = false; // エラー時はリトライ可能に
      } finally {
        setIsCreating(false);
      }
    };

    autoCreate();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isNewRecord, selectedPet?.id]);

  // petIdがなく新規作成の場合はselect-petへリダイレクト（呼び出し元で判定）
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
    isSaving: isPending,
    isCreating,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems,
    // 問診タブ
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintCategoryId,
    setChiefComplaintCategoryId,
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
    // 担当医変更
    handleChangeDoctor,
    // 飼主変更
    handleChangeOwner,
  };
}
