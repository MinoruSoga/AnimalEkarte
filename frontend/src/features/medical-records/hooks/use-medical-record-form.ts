import { useState, useEffect, useTransition, useCallback } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { usePetInfo } from "@/hooks/use-pet";
import { useOwnerInfo } from "@/hooks/use-owner";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateTreatmentPlan } from "../api/treatment-plans";
import type { CreateMedicalRecordRequest, UpdateMedicalRecordRequest } from "../api/types";
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

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord } = useGetMedicalRecord(recordId ?? "");

  // 既存カルテデータをフォームに反映
  useEffect(() => {
    if (!existingRecord) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン。React 18 が自動バッチするため実害なし
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
  }, [existingRecord]);

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { pet: selectedPet, isLoading: isPetLoading } = usePetInfo(resolvedPetId);

  // Ownerデータを取得（飼主割引率用）
  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { owner } = useOwnerInfo(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;

  const createMutation = useCreateMedicalRecord();
  const updateMutation = useUpdateMedicalRecord();
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateTreatmentPlan(recordId ?? "");

  // useTransition: save の pending 管理 (rerender-transitions)
  const [isSaveTransitionPending, startSaveTransition] = useTransition();

  const handleBack = () => {
    if (location.state?.from) {
      navigate(location.state.from);
      return;
    }

    if (!recordId) {
      navigate(paths.medicalRecords.selectPet.getHref());
    } else {
      navigate(paths.medicalRecords.getHref());
    }
  };

  const handleSave = () => {
    if (!selectedPet) return;

    // ── バリデーション ──
    const isChiefComplaintEmpty = !chiefComplaint || chiefComplaint.trim() === "" || chiefComplaint === DEFAULT_CHIEF_COMPLAINT;
    const isPlanEmpty = !plan || plan.trim() === "" || plan === DEFAULT_PLAN;

    if (isChiefComplaintEmpty) {
      toast.error("問診内容（主訴）を入力してください");
      return;
    }

    if (isPlanEmpty) {
      toast.error("診察/治療プランを入力してください");
      return;
    }

    if (diagnosis1CategoryId && !diagnosis1NameId) {
      toast.error("診断名を選択してください");
      return;
    }

    startSaveTransition(async () => {
      const today = new Date().toISOString().split("T")[0];

      if (isNewRecord) {
        const req: CreateMedicalRecordRequest = {
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: "再診",
          status: "draft",
          chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
          chief_complaint_category_id: chiefComplaintCategoryId,
          plan: plan !== DEFAULT_PLAN ? plan : undefined,
          assessment: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
          notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
          diagnosis_1_category_id: diagnosis1CategoryId,
          diagnosis_1_name_id: diagnosis1NameId,
          diagnosis_2_category_id: diagnosis2CategoryId,
          diagnosis_2_name_id: diagnosis2NameId,
        };

        try {
          await createMutation.mutateAsync(req);
          toast.success("カルテを作成しました");
          navigate(location.state?.from ?? paths.medicalRecords.getHref());
        } catch (error) {
          handleApiError(error, "作成");
        }
      } else if (recordId) {
        const mainReq: UpdateMedicalRecordRequest = {
          status: "draft",
        };

        const inquiriesReq = {
          chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
          chief_complaint_category_id: chiefComplaintCategoryId,
          notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
        };

        const treatmentPlanReq = {
          treatment_policy: plan !== DEFAULT_PLAN ? plan : undefined,
          diagnosis_details: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
          diagnosis_category_id: diagnosis1CategoryId ?? undefined,
          diagnosis_name_id: diagnosis1NameId ?? undefined,
          diagnosis_2_category_id: diagnosis2CategoryId ?? undefined,
          diagnosis_2_name_id: diagnosis2NameId ?? undefined,
        };

        try {
          // 並列で複数の API を呼び出し
          await Promise.all([
            updateMutation.mutateAsync({ id: recordId, req: mainReq }),
            updateInquiryMutation.mutateAsync(inquiriesReq),
            updateTreatmentPlanMutation.mutateAsync(treatmentPlanReq),
          ]);
          toast.success("カルテを更新しました");
          navigate(location.state?.from ?? paths.medicalRecords.getHref());
        } catch (error) {
          handleApiError(error, "更新");
        }
      }
    });
  };

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

  // petIdがなく新規作成の場合はselect-petへリダイレクト（呼び出し元で判定）
  const shouldRedirectToSelectPet = isNewRecord && !petId;

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    selectedPet: selectedPet ?? null,
    isPetLoading,
    shouldRedirectToSelectPet,
    handleBack,
    handleSave,
    isSaving: isSaveTransitionPending,
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
    // 飼主変更
    handleChangeOwner,
  };
}
